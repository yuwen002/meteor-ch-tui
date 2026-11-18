package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) addQueryHistory(sql string) {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return
	}
	if n := len(m.queryHistory); n > 0 && m.queryHistory[n-1] == sql {
		return
	}
	m.queryHistory = append(m.queryHistory, sql)
	if len(m.queryHistory) > 50 {
		m.queryHistory = m.queryHistory[len(m.queryHistory)-50:]
	}
}

func (m *Model) openHistoryModal() {
	if len(m.queryHistory) == 0 {
		m.status = "暂无查询历史"
		m.statusOK = true
		return
	}
	m.historySel = len(m.queryHistory) - 1
	m.historyModal = true
}

func (m *Model) closeHistoryModal() {
	m.historyModal = false
	m.historyRect = modalRect{}
}

func (m *Model) moveHistory(delta int) {
	if len(m.queryHistory) == 0 {
		return
	}
	m.historySel = clampInt(m.historySel+delta, 0, len(m.queryHistory)-1)
}

func (m *Model) useHistory() {
	if m.historySel < 0 || m.historySel >= len(m.queryHistory) {
		m.closeHistoryModal()
		return
	}
	sql := m.queryHistory[m.historySel]
	m.closeHistoryModal()
	m.editor.SetValue(sql)
	m.focus = focusEditor
	m.editor.Focus()
	m.status = "已载入历史 SQL"
	m.statusOK = true
}

func (m *Model) handleHistoryKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.closeHistoryModal()
	case "up", "k":
		m.moveHistory(-1)
	case "down", "j":
		m.moveHistory(1)
	case "enter":
		m.useHistory()
	}
	return m, nil
}

func (m *Model) handleHistoryMouse(msg tea.MouseMsg, isClick, isWheelUp, isWheelDown bool) (tea.Model, tea.Cmd) {
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonRight {
		m.closeHistoryModal()
		return m, nil
	}
	if isWheelUp || isWheelDown {
		if isWheelUp {
			m.moveHistory(-1)
		} else {
			m.moveHistory(1)
		}
		return m, nil
	}
	if !isClick {
		return m, nil
	}
	r := m.historyRect
	if r.w == 0 || r.h == 0 || msg.X < r.left || msg.X >= r.left+r.w || msg.Y < r.top || msg.Y >= r.top+r.h {
		return m, nil
	}
	relX, relY := msg.X-r.left, msg.Y-r.top
	if role, ok := hitFooter(relX, relY, r, footerRoleOK, footerRoleCancel); ok {
		switch role {
		case footerRoleOK:
			m.useHistory()
		case footerRoleCancel:
			m.closeHistoryModal()
		}
		return m, nil
	}
	idx := relY - r.listStart
	if idx >= 0 && idx < r.listCount {
		m.historySel = idx
	}
	return m, nil
}
