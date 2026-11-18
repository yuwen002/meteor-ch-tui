package ui

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) cycleFocus() tea.Cmd {
	switch m.focus {
	case focusSidebar:
		m.focus = focusEditor
		return m.editor.Focus()
	case focusEditor:
		m.focus = focusTable
		m.editor.Blur()
	default:
		m.focus = focusSidebar
		m.editor.Blur()
	}
	return nil
}

func (m *Model) cycleFocusBack() tea.Cmd {
	switch m.focus {
	case focusSidebar:
		m.focus = focusTable
	case focusEditor:
		m.focus = focusSidebar
		m.editor.Blur()
	default:
		m.focus = focusEditor
		return m.editor.Focus()
	}
	return nil
}

func (m *Model) backToManager() {
	if m.client != nil {
		m.client.Close()
	}
	m.appMode = appManager
	m.cfgModal = false
	m.rowModal = false
	m.ctxMenuOpen = false
	m.table = tableState{}
	m.slow = tableState{}
	m.client = nil
	m.nodes = nil
	m.curNode = nil
	m.nodeIdx = 0
	m.databases = nil
	m.tables = nil
	m.views = nil
	m.databasesTruncated = false
	m.objectsTruncated = false
	m.currentDB = ""
	m.sidebarTab = tabDatabases
	m.sidebarIdx = 0
	m.sidebarItems = nil
	m.status = "已断开，返回连接管理器"
	m.statusOK = true
}
