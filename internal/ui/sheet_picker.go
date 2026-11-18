package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m *Model) openSheetPicker(path string, sheets []string) {
	m.sheetPickerPath = path
	m.sheetPickerSheets = append([]string(nil), sheets...)
	m.sheetPickerSel = 0
	m.sheetPickerOpen = true
}

func (m *Model) closeSheetPicker() {
	m.sheetPickerOpen = false
	m.sheetPickerPath = ""
	m.sheetPickerSheets = nil
	m.sheetPickerSel = 0
	m.sheetPickerRect = modalRect{}
}

func (m *Model) moveSheetPicker(delta int) {
	if len(m.sheetPickerSheets) == 0 {
		return
	}
	m.sheetPickerSel = clampInt(m.sheetPickerSel+delta, 0, len(m.sheetPickerSheets)-1)
}

func (m *Model) loadSelectedSheet() (tea.Model, tea.Cmd) {
	if m.sheetPickerSel < 0 || m.sheetPickerSel >= len(m.sheetPickerSheets) {
		return m, nil
	}
	sheet := m.sheetPickerSheets[m.sheetPickerSel]
	if err := m.setImportFromXLSX(m.sheetPickerPath, sheet); err != nil {
		m.status = "导入工作表失败: " + err.Error()
		m.statusOK = false
		return m, nil
	}
	m.closeSheetPicker()
	m.status = "已载入工作表: " + sheet
	m.statusOK = true
	return m, nil
}

func (m *Model) handleSheetPickerKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.closeSheetPicker()
	case "up", "k":
		m.moveSheetPicker(-1)
	case "down", "j":
		m.moveSheetPicker(1)
	case "enter", "right", "l":
		return m.loadSelectedSheet()
	case "left", "h", "backspace":
		m.closeSheetPicker()
	}
	return m, nil
}

func (m *Model) handleSheetPickerMouse(msg tea.MouseMsg, isClick, isWheelUp, isWheelDown bool) (tea.Model, tea.Cmd) {
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonRight {
		m.closeSheetPicker()
		return m, nil
	}
	if isWheelUp || isWheelDown {
		if isWheelUp {
			m.moveSheetPicker(-1)
		} else {
			m.moveSheetPicker(1)
		}
		return m, nil
	}
	if !isClick {
		return m, nil
	}
	r := m.sheetPickerRect
	if r.w == 0 || r.h == 0 || msg.X < r.left || msg.X >= r.left+r.w || msg.Y < r.top || msg.Y >= r.top+r.h {
		return m, nil
	}
	relX, relY := msg.X-r.left, msg.Y-r.top
	if role, ok := hitFooter(relX, relY, r, footerRoleOK, footerRoleCancel); ok {
		switch role {
		case footerRoleOK:
			return m.loadSelectedSheet()
		case footerRoleCancel:
			m.closeSheetPicker()
		}
		return m, nil
	}
	idx := relY - r.listStart
	if idx >= 0 && idx < r.listCount {
		m.sheetPickerSel = idx
		return m.loadSelectedSheet()
	}
	return m, nil
}

func (m *Model) renderSheetPicker(bg string) string {
	innerW := minInt(72, m.width-8)
	if innerW < 36 {
		innerW = maxInt(28, m.width-4)
	}

	title := styleTitle.Render(" 选择 Excel 工作表 ")
	hint := styleMuted.Render(" ↑↓ · Enter 导入 · Esc 取消 ")
	lines := make([]string, 0, len(m.sheetPickerSheets)+6)
	lines = append(lines, padCell(title, maxInt(0, innerW-lipgloss.Width(hint)))+hint)
	lines = append(lines, "")
	pathText := ansi.Truncate(m.sheetPickerPath, innerW-2, "~")
	lines = append(lines, styleOkText.Render(padCell(" "+pathText, innerW)))
	lines = append(lines, "")

	listStart := len(lines) + 1
	for i, sheet := range m.sheetPickerSheets {
		text := padCell("   "+sheet, innerW)
		if i == m.sheetPickerSel {
			lines = append(lines, styleRowSelected.Render(text))
		} else {
			lines = append(lines, text)
		}
	}
	lines = append(lines, "")

	btns := []footerButton{
		{label: styleBtnOK.Render("[Enter] 导入"), role: footerRoleOK},
		{label: styleBtnMuted.Render("[Esc] 取消"), role: footerRoleCancel},
	}
	footer, positions := footerLayout(innerW, footerAlignFirstLeftRestRight, btns)
	footerContentRow := len(lines)
	lines = append(lines, footer)

	content := strings.Join(lines, "\n")
	box := styleModal.Width(innerW).Height(len(lines)).Render(content)
	out, rect := overlayRect(bg, box, m.width)
	rect.listStart = listStart
	rect.listCount = len(m.sheetPickerSheets)
	applyFooter(&rect, footerContentRow, btns, positions)
	m.sheetPickerRect = rect
	return out
}
