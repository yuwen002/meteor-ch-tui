package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func quoteIdent(s string) string {
	return "`" + strings.ReplaceAll(s, "`", "``") + "`"
}

func (m *Model) handleEditorKey(msg tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	switch key {
	case "alt+enter", "ctrl+s", "f5":
		sql := strings.TrimSpace(m.editor.Value())
		if sql == "" {
			m.status = "SQL 为空"
			m.statusOK = false
			return m, nil
		}
		if m.client == nil {
			m.status = "尚未连接 ClickHouse"
			m.statusOK = false
			return m, nil
		}
		m.loading = true
		m.status = "执行中..."
		m.statusOK = true
		return m, m.runQueryCmd(sql)
	case "pgup", "pgdown", "home", "end":
		return m, nil
	}
	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(msg)
	return m, cmd
}

func (m *Model) handleTableKey(key string) (tea.Model, tea.Cmd) {
	t := m.activeTable()
	switch key {
	case "up", "k":
		t.moveCursor(-1, m.tableRowsVisible())
		return m, nil
	case "down", "j":
		t.moveCursor(1, m.tableRowsVisible())
		return m, nil
	case "pgup":
		t.moveCursor(-m.tableRowsVisible(), m.tableRowsVisible())
		return m, nil
	case "pgdown":
		t.moveCursor(m.tableRowsVisible(), m.tableRowsVisible())
		return m, nil
	case "left", "h":
		t.scrollCols(-1, m.tableContentWidth())
		return m, nil
	case "right", "l":
		t.scrollCols(1, m.tableContentWidth())
		return m, nil
	case "home":
		t.startCol = 0
		t.topRow = 0
		t.cursor = 0
		return m, nil
	case "c":
		m.openColumnModal()
		return m, nil
	case "v":
		if m.mode != modeResult || t.res == nil || len(t.res.Rows) == 0 {
			m.status = "当前没有可查看的行"
			m.statusOK = false
			return m, nil
		}
		m.rowModal = true
		m.rowScroll = 0
		return m, nil
	case "e", "i":
		m.focus = focusEditor
		return m, m.editor.Focus()
	}
	return m, nil
}

func (m *Model) handleRowModalKey(key string) (tea.Model, tea.Cmd) {
	total := m.rowModalLines()
	page := m.modalPageSize()
	switch key {
	case "esc", "v", "q":
		m.rowModal = false
		return m, nil
	case "up", "k":
		if m.rowScroll > 0 {
			m.rowScroll--
		}
		return m, nil
	case "down", "j":
		if m.rowScroll < total-page {
			m.rowScroll++
		}
		return m, nil
	case "pgup":
		m.rowScroll -= page
		if m.rowScroll < 0 {
			m.rowScroll = 0
		}
		return m, nil
	case "pgdown":
		m.rowScroll += page
		if m.rowScroll > total-page {
			m.rowScroll = total - page
		}
		if m.rowScroll < 0 {
			m.rowScroll = 0
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) activeTable() *tableState {
	if m.mode == modeSlow {
		return &m.slow
	}
	return &m.table
}
