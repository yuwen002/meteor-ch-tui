package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// openGroupPicker 打开新建/编辑连接弹窗中的分组选择弹层。
func (m *Model) openGroupPicker() {
	groups := m.cs.Groups()
	m.mgrGroupPickerSel = 0
	current := strings.TrimSpace(m.mgrAddGroup.Value())
	for i, g := range groups {
		if g == current {
			m.mgrGroupPickerSel = i
			break
		}
	}
	m.mgrGroupPickerOpen = true
}

func (m *Model) closeGroupPicker() {
	m.mgrGroupPickerOpen = false
}

// groupPickerValues 返回弹层中的分组值（默认组用空字符串表示）。
func (m *Model) groupPickerValues() []string {
	return m.mgrAddGroupValues()
}

func (m *Model) confirmGroupPicker() {
	vals := m.groupPickerValues()
	if m.mgrGroupPickerSel < 0 || m.mgrGroupPickerSel >= len(vals) {
		m.closeGroupPicker()
		return
	}
	m.mgrAddGroup.SetValue(vals[m.mgrGroupPickerSel])
	m.mgrAddGroup.CursorEnd()
	m.closeGroupPicker()
}

func (m *Model) handleGroupPickerKey(msg tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	vals := m.groupPickerValues()
	switch key {
	case "esc":
		m.closeGroupPicker()
		return m, nil
	case "up":
		if len(vals) > 0 {
			m.mgrGroupPickerSel = (m.mgrGroupPickerSel - 1 + len(vals)) % len(vals)
		}
		return m, nil
	case "down":
		if len(vals) > 0 {
			m.mgrGroupPickerSel = (m.mgrGroupPickerSel + 1) % len(vals)
		}
		return m, nil
	case "enter":
		m.confirmGroupPicker()
		return m, nil
	}
	return m, nil
}

func (m *Model) handleGroupPickerMouse(msg tea.MouseMsg, isClick, isWheelUp, isWheelDown bool) (tea.Model, tea.Cmd) {
	vals := m.groupPickerValues()
	if isWheelUp && len(vals) > 0 {
		m.mgrGroupPickerSel = (m.mgrGroupPickerSel - 1 + len(vals)) % len(vals)
		return m, nil
	}
	if isWheelDown && len(vals) > 0 {
		m.mgrGroupPickerSel = (m.mgrGroupPickerSel + 1) % len(vals)
		return m, nil
	}
	if !isClick {
		return m, nil
	}
	r := m.mgrGroupPickerRect
	if r.w == 0 || r.h == 0 || msg.X < r.left || msg.X >= r.left+r.w || msg.Y < r.top || msg.Y >= r.top+r.h {
		return m, nil
	}
	idx := msg.Y - r.top - r.listStart
	if idx >= 0 && idx < len(vals) {
		m.mgrGroupPickerSel = idx
		m.confirmGroupPicker()
	}
	return m, nil
}

func (m *Model) renderGroupPicker(bg string) string {
	groups := m.cs.Groups()
	innerW := minInt(48, m.width-8)
	if innerW < 24 {
		innerW = maxInt(20, m.width-4)
	}
	title := styleTitle.Render(" 选择分组 ")
	hint := styleMuted.Render(" ↑↓ 选择 · Enter 确认 · Esc 取消 ")
	lines := []string{padCell(title, maxInt(0, innerW-lipgloss.Width(hint))) + hint, ""}

	listStart := len(lines) + 1
	for i, g := range groups {
		text := padCell(fmt.Sprintf("   %s (%d)", g, len(m.cs.ByGroup(g))), innerW)
		if i == m.mgrGroupPickerSel {
			lines = append(lines, styleRowSelected.Render(text))
		} else {
			lines = append(lines, text)
		}
	}
	if len(groups) == 0 {
		lines = append(lines, styleMuted.Render(padCell(" (暂无分组，可手动输入新分组)", innerW)))
	}

	box := styleModal.Width(innerW).Height(len(lines)).Render(strings.Join(lines, "\n"))
	out, rect := overlayRect(bg, box, m.width)
	rect.listStart = listStart
	m.mgrGroupPickerRect = rect
	return out
}
