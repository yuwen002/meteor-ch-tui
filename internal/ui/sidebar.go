package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
)

func (m *Model) handleSidebarKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		m.moveSidebar(-1)
		return m, nil
	case "down", "j":
		m.moveSidebar(1)
		return m, nil
	case "pgup":
		m.moveSidebar(-10)
		return m, nil
	case "pgdown":
		m.moveSidebar(10)
		return m, nil
	case "home":
		m.sidebarIdx = 0
		return m, nil
	case "end":
		m.sidebarIdx = maxInt(0, m.sidebarItemCount()-1)
		return m, nil
	case "1", "d":
		m.setSidebarTab(tabDatabases)
		return m, nil
	case "2", "v":
		m.setSidebarTab(tabViews)
		return m, nil
	case "3", "n":
		m.setSidebarTab(tabNodes)
		return m, nil
	case "enter":
		return m.activateSidebarItem()
	case "r":
		m.loading = true
		m.status = "正在刷新节点、库与视图列表..."
		m.statusOK = true
		return m, m.connectCmd()
	}
	return m, nil
}

func (m *Model) sidebarItemCount() int {
	switch m.sidebarTab {
	case tabDatabases:
		return len(m.databases) + len(m.tables)
	case tabViews:
		return len(m.views)
	default:
		return len(m.nodes)
	}
}

func (m *Model) moveSidebar(delta int) {
	count := m.sidebarItemCount()
	if count == 0 {
		return
	}
	m.sidebarIdx = clampInt(m.sidebarIdx+delta, 0, count-1)
}

func (m *Model) setSidebarTab(t sidebarTab) {
	if m.sidebarTab == t {
		return
	}
	m.sidebarTab = t
	m.sidebarItems = nil
	m.syncSidebarSelection()
}

func (m *Model) syncSidebarSelection() {
	switch m.sidebarTab {
	case tabDatabases:
		for i, db := range m.databases {
			if db == m.currentDB {
				m.sidebarIdx = i
				return
			}
		}
		m.sidebarIdx = 0
	case tabViews:
		m.sidebarIdx = 0
	default:
		m.sidebarIdx = clampInt(m.nodeIdx, 0, maxInt(0, len(m.nodes)-1))
	}
}

func (m *Model) activateSidebarItem() (tea.Model, tea.Cmd) {
	count := m.sidebarItemCount()
	if count == 0 || m.sidebarIdx < 0 || m.sidebarIdx >= count {
		return m, nil
	}
	if m.sidebarTab == tabNodes {
		if m.sidebarIdx >= len(m.nodes) {
			return m, nil
		}
		n := m.nodes[m.sidebarIdx]
		m.loading = true
		m.status = "正在切换节点 " + n.Label() + "..."
		m.statusOK = true
		return m, m.switchNodeCmd(n)
	}
	if m.sidebarTab == tabViews {
		if m.sidebarIdx >= len(m.views) {
			return m, nil
		}
		return m.previewView(m.views[m.sidebarIdx])
	}
	if m.sidebarIdx < len(m.databases) {
		db := m.databases[m.sidebarIdx]
		if db == m.currentDB {
			m.status = "当前已在库 " + db + " · 下方可选表预览 · 按 2 查看视图"
			m.statusOK = true
			return m, nil
		}
		m.loading = true
		m.status = "正在切换到库 " + db + "..."
		m.statusOK = true
		return m, m.switchDatabaseCmd(db)
	}
	idx := m.sidebarIdx - len(m.databases)
	if idx < 0 || idx >= len(m.tables) {
		return m, nil
	}
	return m.previewTable(m.tables[idx])
}

func (m *Model) previewTable(name string) (tea.Model, tea.Cmd) {
	return m.previewObject(name, "表")
}

func (m *Model) previewView(name string) (tea.Model, tea.Cmd) {
	return m.previewObject(name, "视图")
}

func (m *Model) previewObject(name, kind string) (tea.Model, tea.Cmd) {
	if m.client == nil {
		m.status = "尚未连接 ClickHouse"
		m.statusOK = false
		return m, nil
	}
	sql := fmt.Sprintf("SELECT * FROM %s.%s LIMIT 100", quoteIdent(m.currentDB), quoteIdent(name))
	m.editor.SetValue(sql)
	m.loading = true
	m.status = "正在预览" + kind + " " + m.currentDB + "." + name + "..."
	m.statusOK = true
	return m, m.runQueryCmd(sql)
}

func (m *Model) moveNode(delta int) {
	if len(m.nodes) == 0 {
		return
	}
	m.nodeIdx = clampInt(m.nodeIdx+delta, 0, len(m.nodes)-1)
}

func (m *Model) indexOfNode(n *ch.Node) int {
	if n == nil {
		return 0
	}
	for i := range m.nodes {
		if m.nodes[i].Cluster == n.Cluster && m.nodes[i].Shard == n.Shard &&
			m.nodes[i].Replica == n.Replica && m.nodes[i].HostName == n.HostName {
			return i
		}
	}
	return 0
}

func (m *Model) selectedSidebarName() string {
	if m.appMode != appMain || m.sidebarIdx < 0 {
		return ""
	}
	switch m.sidebarTab {
	case tabDatabases:
		if m.sidebarIdx < len(m.databases) {
			return m.databases[m.sidebarIdx]
		}
		idx := m.sidebarIdx - len(m.databases)
		if idx >= 0 && idx < len(m.tables) {
			return m.tables[idx]
		}
	case tabViews:
		if m.sidebarIdx < len(m.views) {
			return m.views[m.sidebarIdx]
		}
	case tabNodes:
		if m.sidebarIdx < len(m.nodes) {
			return m.nodes[m.sidebarIdx].Label()
		}
	}
	return ""
}

func (m *Model) handleSidebarClick(x, y int) (tea.Model, tea.Cmd) {
	m.focus = focusSidebar
	if y == m.sidebarTabRowY {
		if x >= m.sidebarTabDBX && x < m.sidebarTabDBX+m.sidebarTabDBW {
			m.setSidebarTab(tabDatabases)
			return m, nil
		}
		if x >= m.sidebarTabViewX && x < m.sidebarTabViewX+m.sidebarTabViewW {
			m.setSidebarTab(tabViews)
			return m, nil
		}
		if x >= m.sidebarTabNodeX && x < m.sidebarTabNodeX+m.sidebarTabNodeW {
			m.setSidebarTab(tabNodes)
			return m, nil
		}
		return m, nil
	}
	if len(m.sidebarItems) == 0 {
		return m, nil
	}
	lineIdx := m.sidebarLineTop + (y - m.sidebarFirstLineY)
	for i, it := range m.sidebarItems {
		if it.row != lineIdx {
			continue
		}
		switch m.sidebarTab {
		case tabDatabases:
			if it.kind != itemDatabase && it.kind != itemTable {
				return m, nil
			}
		case tabViews:
			if it.kind != itemView {
				return m, nil
			}
		case tabNodes:
			if it.kind != itemNode {
				return m, nil
			}
		}
		m.sidebarIdx = i
		if it.kind == itemNode {
			// 切换节点需要重新建连，单击仅选中，Enter 确认。
			if it.idx >= 0 && it.idx < len(m.nodes) {
				m.nodeIdx = it.idx
				m.status = "已选中节点 " + m.nodes[it.idx].Label() + " · Enter 切换"
				m.statusOK = true
			}
			return m, nil
		}
		return m.activateSidebarItem()
	}
	return m, nil
}
