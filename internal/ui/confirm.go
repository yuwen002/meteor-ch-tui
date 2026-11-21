package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuwen002/meteor-ch-tui/internal/configstore"
)

func (m *Model) requestDeleteConn() {
	if m.markedCount() > 0 {
		m.requestBatchDeleteConns()
		return
	}
	conn := m.currentMgrConn()
	if conn == nil {
		return
	}
	m.confirm = pendingConfirm{kind: confirmDeleteConn, name: conn.Name}
}

func (m *Model) requestDeleteGroup() {
	groups := m.cs.Groups()
	if m.mgrGroupSel < 0 || m.mgrGroupSel >= len(groups) {
		return
	}
	name := groups[m.mgrGroupSel]
	if name == configstore.DefaultGroupName {
		m.status = "默认分组不能删除"
		m.statusOK = false
		return
	}
	m.confirm = pendingConfirm{
		kind: confirmDeleteGroup,
		name: name,
		info: fmt.Sprintf("该组下的 %d 个连接将移回默认分组", len(m.cs.ByGroup(name))),
	}
}

func (m *Model) requestExportPasswords() {
	m.confirm = pendingConfirm{
		kind: confirmExportPasswords,
		info: "明文密码会写入导出文件，请妥善保管导出结果",
	}
}

func (m *Model) cancelConfirm() {
	m.confirm = pendingConfirm{}
}

func (m *Model) acceptConfirm() {
	c := m.confirm
	m.confirm = pendingConfirm{}
	switch c.kind {
	case confirmDeleteConn:
		m.cs.Remove(c.name)
		if err := m.cs.Save(); err != nil {
			m.status = "删除失败: " + err.Error()
			m.statusOK = false
			return
		}
		if m.cs.LastUsed == c.name {
			m.cs.SetLastUsed("")
			_ = m.cs.Save()
		}
		m.status = "已删除连接: " + c.name
		m.statusOK = true
		clampMgrSelection(m.cs, &m.mgrGroupIdx, &m.mgrConnIdx)
	case confirmDeleteGroup:
		m.performDeleteGroup(c.name)
	case confirmBatchDeleteConns:
		next := m.cs.Clone()
		removed := next.RemoveConnections(c.names)
		if err := next.Save(); err != nil {
			m.status = "删除失败: " + err.Error()
			m.statusOK = false
			return
		}
		*m.cs = *next
		m.clearConnMarks()
		clampMgrSelection(m.cs, &m.mgrGroupIdx, &m.mgrConnIdx)
		m.status = fmt.Sprintf("已删除 %d 个连接", removed)
		m.statusOK = true
	case confirmExportPasswords:
		m.mgrExportPasswords = true
		m.refreshMgrExportPreview()
	}
}

func (m *Model) handleConfirmKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter", "y", "Y":
		m.acceptConfirm()
	case "esc", "n", "N":
		m.cancelConfirm()
	}
	return m, nil
}

func (m *Model) handleConfirmMouse(msg tea.MouseMsg, isClick bool) (tea.Model, tea.Cmd) {
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonRight {
		m.cancelConfirm()
		return m, nil
	}
	if !isClick {
		return m, nil
	}
	r := m.confirmRect
	if r.w == 0 || r.h == 0 {
		return m, nil
	}
	if msg.X < r.left || msg.X >= r.left+r.w || msg.Y < r.top || msg.Y >= r.top+r.h {
		return m, nil
	}
	relX, relY := msg.X-r.left, msg.Y-r.top
	if role, ok := hitFooter(relX, relY, r, footerRoleOK, footerRoleCancel); ok {
		switch role {
		case footerRoleOK:
			m.acceptConfirm()
		case footerRoleCancel:
			m.cancelConfirm()
		}
	}
	return m, nil
}
