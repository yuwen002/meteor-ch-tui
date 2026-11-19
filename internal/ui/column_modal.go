package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) openColumnModal() {
	t := m.activeTable()
	if t.res == nil || len(t.res.Columns) == 0 {
		m.status = "当前没有可设置的列"
		m.statusOK = false
		return
	}
	m.columnModalOpen = true
	m.columnModalSel = 0
}

func (m *Model) closeColumnModal() {
	m.columnModalOpen = false
	m.columnModalRect = modalRect{}
}

func (m *Model) moveColumnModal(delta int) {
	t := m.activeTable()
	if t.res == nil {
		return
	}
	m.columnModalSel = clampInt(m.columnModalSel+delta, 0, len(t.res.Columns)-1)
}

func (m *Model) selectedColumnName() string {
	t := m.activeTable()
	if t.res == nil || m.columnModalSel < 0 || m.columnModalSel >= len(t.res.Columns) {
		return ""
	}
	return t.res.Columns[m.columnModalSel]
}

func (m *Model) toggleSelectedColumn() {
	name := m.selectedColumnName()
	if name == "" {
		return
	}
	t := m.activeTable()
	t.toggleColumn(name)
	m.persistResultLayout(t)
	if t.hidden[name] {
		m.status = "已隐藏列: " + name
	} else {
		m.status = "已显示列: " + name
	}
	m.statusOK = true
}

func (m *Model) adjustSelectedColumnWidth(delta int) {
	name := m.selectedColumnName()
	if name == "" {
		return
	}
	t := m.activeTable()
	t.adjustColumnWidth(name, delta)
	m.persistResultLayout(t)
	m.status = fmt.Sprintf("列宽 %s: %d", name, t.widths[m.columnModalSel])
	m.statusOK = true
}

func (m *Model) resetSelectedColumnWidth() {
	name := m.selectedColumnName()
	if name == "" {
		return
	}
	t := m.activeTable()
	t.resetColumnWidth(name)
	m.persistResultLayout(t)
	m.status = "已重置列宽: " + name
	m.statusOK = true
}

func (m *Model) showAllResultColumns() {
	t := m.activeTable()
	t.showAllColumns()
	m.persistResultLayout(t)
	m.status = "已显示全部列"
	m.statusOK = true
}

func (m *Model) handleColumnModalKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.closeColumnModal()
	case "up", "k":
		m.moveColumnModal(-1)
	case "down", "j":
		m.moveColumnModal(1)
	case "enter", " ", "space":
		m.toggleSelectedColumn()
	case "left", "[":
		m.adjustSelectedColumnWidth(-2)
	case "right", "]":
		m.adjustSelectedColumnWidth(2)
	case "r":
		m.resetSelectedColumnWidth()
	case "a":
		m.showAllResultColumns()
	}
	return m, nil
}

func (m *Model) handleColumnModalMouse(msg tea.MouseMsg, isClick, isWheelUp, isWheelDown bool) (tea.Model, tea.Cmd) {
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonRight {
		m.closeColumnModal()
		return m, nil
	}
	if isWheelUp || isWheelDown {
		if isWheelUp {
			m.moveColumnModal(-1)
		} else {
			m.moveColumnModal(1)
		}
		return m, nil
	}
	if !isClick {
		return m, nil
	}
	r := m.columnModalRect
	if r.w == 0 || r.h == 0 || msg.X < r.left || msg.X >= r.left+r.w || msg.Y < r.top || msg.Y >= r.top+r.h {
		return m, nil
	}
	relX, relY := msg.X-r.left, msg.Y-r.top
	if role, ok := hitFooter(relX, relY, r, footerRoleOK, footerRoleExtra, footerRoleCancel); ok {
		switch role {
		case footerRoleOK:
			m.showAllResultColumns()
		case footerRoleExtra:
			m.resetSelectedColumnWidth()
		case footerRoleCancel:
			m.closeColumnModal()
		}
		return m, nil
	}
	idx := relY - r.listStart
	if idx >= 0 && idx < r.listCount {
		m.columnModalSel = idx
		m.toggleSelectedColumn()
	}
	return m, nil
}

func (m *Model) renderColumnModal(bg string) string {
	innerW := minInt(78, m.width-8)
	if innerW < 40 {
		innerW = maxInt(32, m.width-4)
	}
	t := m.activeTable()

	title := styleTitle.Render(" 列设置 ")
	hint := styleMuted.Render(" ↑↓ · Space 显示/隐藏 · [ ] 宽度 · Esc 关闭 ")
	lines := []string{padCell(title, maxInt(0, innerW-lipgloss.Width(hint))) + hint, ""}

	listStart := len(lines) + 1
	if t.res != nil {
		for i, name := range t.res.Columns {
			mark := "[ ]"
			if !t.hidden[name] {
				mark = "[x]"
			}
			width := 0
			if i < len(t.widths) {
				width = t.widths[i]
			}
			text := padCell(fmt.Sprintf("  %s  %-18s  宽度 %d", mark, name, width), innerW)
			if i == m.columnModalSel {
				lines = append(lines, styleRowSelected.Render(text))
			} else {
				lines = append(lines, text)
			}
		}
	}
	lines = append(lines, "")

	btns := []footerButton{
		{label: styleBtnOK.Render("[A] 全部显示"), role: footerRoleOK},
		{label: styleBtnMuted.Render("[R] 重置列宽"), role: footerRoleExtra},
		{label: styleBtnMuted.Render("[Esc] 关闭"), role: footerRoleCancel},
	}
	footer, positions := footerLayout(innerW, footerAlignFirstLeftRestRight, btns)
	footerContentRow := len(lines)
	lines = append(lines, footer)

	content := strings.Join(lines, "\n")
	box := styleModal.Width(innerW).Height(len(lines)).Render(content)
	out, rect := overlayRect(bg, box, m.width)
	rect.listStart = listStart
	rect.listCount = 0
	if t.res != nil {
		rect.listCount = len(t.res.Columns)
	}
	applyFooter(&rect, footerContentRow, btns, positions)
	m.columnModalRect = rect
	return out
}
