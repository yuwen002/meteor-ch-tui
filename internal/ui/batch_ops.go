package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) ensureConnMarks() {
	if m.mgrMarked == nil {
		m.mgrMarked = map[string]bool{}
	}
}

func (m *Model) toggleConnMark() {
	conn := m.currentMgrConn()
	if conn == nil {
		return
	}
	m.ensureConnMarks()
	if m.mgrMarked[conn.Name] {
		delete(m.mgrMarked, conn.Name)
	} else {
		m.mgrMarked[conn.Name] = true
	}
	m.status = fmt.Sprintf("已选 %d 个连接", m.markedCount())
	m.statusOK = true
}

func (m *Model) markAllConns() {
	m.ensureConnMarks()
	for _, conn := range m.cs.Connections {
		m.mgrMarked[conn.Name] = true
	}
	m.status = fmt.Sprintf("已选 %d 个连接", m.markedCount())
	m.statusOK = true
}

func (m *Model) clearConnMarks() {
	m.mgrMarked = nil
}

func (m *Model) markedNames() []string {
	var out []string
	for _, conn := range m.cs.Connections {
		if m.mgrMarked[conn.Name] {
			out = append(out, conn.Name)
		}
	}
	return out
}

func (m *Model) markedCount() int {
	return len(m.markedNames())
}

func (m *Model) requestBatchDeleteConns() {
	names := m.markedNames()
	if len(names) == 0 {
		return
	}
	m.confirm = pendingConfirm{
		kind:  confirmBatchDeleteConns,
		names: names,
		info:  fmt.Sprintf("共 %d 个连接，此操作不可撤销", len(names)),
	}
}

func (m *Model) openBatchMoveModal() {
	names := m.markedNames()
	if len(names) == 0 {
		m.status = "请先标记要移动的连接（空格）"
		m.statusOK = false
		return
	}
	groups := m.cs.Groups()
	m.mgrBatchMoveSel = clampInt(m.mgrGroupIdx, 0, maxInt(0, len(groups)-1))
	m.mgrBatchMoveOpen = true
}

func (m *Model) closeBatchMoveModal() {
	m.mgrBatchMoveOpen = false
	m.mgrBatchMoveRect = modalRect{}
}

func (m *Model) moveBatchSelection(delta int) {
	groups := m.cs.Groups()
	if len(groups) == 0 {
		return
	}
	m.mgrBatchMoveSel = clampInt(m.mgrBatchMoveSel+delta, 0, len(groups)-1)
}

func (m *Model) batchMoveTarget() string {
	groups := m.cs.Groups()
	if m.mgrBatchMoveSel < 0 || m.mgrBatchMoveSel >= len(groups) {
		return ""
	}
	return groups[m.mgrBatchMoveSel]
}

func (m *Model) confirmBatchMove() {
	target := m.batchMoveTarget()
	if target == "" {
		return
	}
	names := m.markedNames()
	next := m.cs.Clone()
	moved := next.MoveConnectionsToGroup(names, target)
	if moved == 0 {
		m.status = "选中的连接已在目标分组"
		m.statusOK = true
		m.closeBatchMoveModal()
		return
	}
	if err := next.Save(); err != nil {
		m.status = "移动失败: " + err.Error()
		m.statusOK = false
		return
	}
	*m.cs = *next
	m.clearConnMarks()
	m.closeBatchMoveModal()
	m.status = fmt.Sprintf("已将 %d 个连接移动到 %s", moved, target)
	m.statusOK = true
}

func (m *Model) handleBatchMoveKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.closeBatchMoveModal()
	case "up", "k":
		m.moveBatchSelection(-1)
	case "down", "j":
		m.moveBatchSelection(1)
	case "enter":
		m.confirmBatchMove()
	}
	return m, nil
}

func (m *Model) handleBatchMoveMouse(msg tea.MouseMsg, isClick, isWheelUp, isWheelDown bool) (tea.Model, tea.Cmd) {
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonRight {
		m.closeBatchMoveModal()
		return m, nil
	}
	if isWheelUp || isWheelDown {
		if isWheelUp {
			m.moveBatchSelection(-1)
		} else {
			m.moveBatchSelection(1)
		}
		return m, nil
	}
	if !isClick {
		return m, nil
	}
	r := m.mgrBatchMoveRect
	if r.w == 0 || r.h == 0 || msg.X < r.left || msg.X >= r.left+r.w || msg.Y < r.top || msg.Y >= r.top+r.h {
		return m, nil
	}
	relX, relY := msg.X-r.left, msg.Y-r.top
	if role, ok := hitFooter(relX, relY, r, footerRoleOK, footerRoleCancel); ok {
		switch role {
		case footerRoleOK:
			m.confirmBatchMove()
		case footerRoleCancel:
			m.closeBatchMoveModal()
		}
		return m, nil
	}
	idx := relY - r.listStart
	if idx >= 0 && idx < r.listCount {
		m.mgrBatchMoveSel = idx
	}
	return m, nil
}

func (m *Model) renderBatchMoveModal(bg string) string {
	innerW := minInt(70, m.width-8)
	if innerW < 36 {
		innerW = maxInt(30, m.width-4)
	}
	groups := m.cs.Groups()
	title := styleTitle.Render(" 移动连接到分组 ")
	hint := styleMuted.Render(fmt.Sprintf(" 已选 %d 个 · ↑↓ 选择 · Enter 移动", m.markedCount()))
	lines := []string{padCell(title, maxInt(0, innerW-lipgloss.Width(hint))) + hint, ""}

	listStart := len(lines) + 1
	for i, group := range groups {
		text := padCell(fmt.Sprintf("   %s (%d)", group, len(m.cs.ByGroup(group))), innerW)
		if i == m.mgrBatchMoveSel {
			lines = append(lines, styleRowSelected.Render(text))
		} else {
			lines = append(lines, text)
		}
	}
	lines = append(lines, "")

	btns := []footerButton{
		{label: styleBtnOK.Render("[Enter] 移动"), role: footerRoleOK},
		{label: styleBtnMuted.Render("[Esc] 取消"), role: footerRoleCancel},
	}
	footer, positions := footerLayout(innerW, footerAlignFirstLeftRestRight, btns)
	footerContentRow := len(lines)
	lines = append(lines, footer)

	box := styleModal.Width(innerW).Height(len(lines)).Render(strings.Join(lines, "\n"))
	out, rect := overlayRect(bg, box, m.width)
	rect.listStart = listStart
	rect.listCount = len(groups)
	applyFooter(&rect, footerContentRow, btns, positions)
	m.mgrBatchMoveRect = rect
	return out
}
