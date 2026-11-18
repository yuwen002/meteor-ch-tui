package ui

import (
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) openContextMenu(x, y int) {
	items := make([]contextMenuItem, 0, 5)
	previewLabel := "预览选中项"
	previewOK := false
	switch m.sidebarTab {
	case tabDatabases:
		idx := m.sidebarIdx - len(m.databases)
		previewOK = idx >= 0 && idx < len(m.tables)
	case tabViews:
		previewOK = m.sidebarIdx >= 0 && m.sidebarIdx < len(m.views)
	case tabNodes:
		previewLabel = "切换选中节点"
		previewOK = m.sidebarIdx >= 0 && m.sidebarIdx < len(m.nodes)
	}
	if previewOK && m.client != nil {
		items = append(items, contextMenuItem{label: previewLabel, action: contextPreview, enabled: true})
	}
	if name := m.selectedSidebarName(); name != "" {
		items = append(items, contextMenuItem{label: "复制名称: " + name, action: contextCopyName, enabled: true})
	}
	items = append(items,
		contextMenuItem{label: "刷新元数据", action: contextRefresh, enabled: true},
		contextMenuItem{label: "返回结果面板", action: contextResult, enabled: true},
		contextMenuItem{label: "返回连接管理器", action: contextManager, enabled: true},
	)
	if len(items) == 0 {
		return
	}
	m.ctxMenuItems = items
	m.ctxMenuX = x
	m.ctxMenuY = y
	m.ctxMenuSel = 0
	m.ctxMenuOpen = true
}

func (m *Model) closeContextMenu() {
	m.ctxMenuOpen = false
	m.ctxMenuItems = nil
	m.ctxMenuRect = modalRect{}
}

func (m *Model) moveContextMenu(delta int) {
	if len(m.ctxMenuItems) == 0 {
		return
	}
	m.ctxMenuSel = clampInt(m.ctxMenuSel+delta, 0, len(m.ctxMenuItems)-1)
}

func (m *Model) executeContextMenu() tea.Cmd {
	if m.ctxMenuSel < 0 || m.ctxMenuSel >= len(m.ctxMenuItems) {
		m.closeContextMenu()
		return nil
	}
	action := m.ctxMenuItems[m.ctxMenuSel].action
	m.closeContextMenu()
	switch action {
	case contextPreview:
		_, cmd := m.activateSidebarItem()
		return cmd
	case contextCopyName:
		name := m.selectedSidebarName()
		if err := clipboard.WriteAll(name); err != nil {
			m.status = "复制失败: " + err.Error()
			m.statusOK = false
			return nil
		}
		m.status = "已复制: " + name
		m.statusOK = true
	case contextRefresh:
		m.loading = true
		m.status = "正在刷新节点、库与视图列表..."
		m.statusOK = true
		return m.connectCmd()
	case contextResult:
		m.mode = modeResult
		m.focus = focusTable
		m.layout()
		m.status = "已返回结果面板"
		m.statusOK = true
	case contextManager:
		m.backToManager()
	}
	return nil
}

func (m *Model) handleContextMenuKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.closeContextMenu()
	case "up", "k":
		m.moveContextMenu(-1)
	case "down", "j":
		m.moveContextMenu(1)
	case "enter":
		return m, m.executeContextMenu()
	}
	return m, nil
}

func (m *Model) handleContextMenuMouse(msg tea.MouseMsg, isClick, isWheelUp, isWheelDown bool) (tea.Model, tea.Cmd) {
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonRight {
		m.closeContextMenu()
		return m, nil
	}
	if isWheelUp || isWheelDown {
		if isWheelUp {
			m.moveContextMenu(-1)
		} else {
			m.moveContextMenu(1)
		}
		return m, nil
	}
	if !isClick {
		return m, nil
	}
	r := m.ctxMenuRect
	if r.w == 0 || r.h == 0 || msg.X < r.left || msg.X >= r.left+r.w || msg.Y < r.top || msg.Y >= r.top+r.h {
		m.closeContextMenu()
		return m, nil
	}
	idx := (msg.Y - r.top) - r.listStart
	if idx >= 0 && idx < r.listCount {
		m.ctxMenuSel = idx
		return m, m.executeContextMenu()
	}
	return m, nil
}
