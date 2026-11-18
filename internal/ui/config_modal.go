package ui

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
)

func (m *Model) closeCfgModal() {
	m.cfgModal = false
	m.cfgFields[m.cfgActive].Blur()
	m.clearModalError()
}

func (m *Model) commitCfgModal() tea.Cmd {
	newCfg, err := m.buildCfg()
	if err != nil {
		m.status = "配置无效: " + err.Error()
		m.statusOK = false
		return nil
	}
	m.cfg = newCfg
	m.closeCfgModal()
	m.loading = true
	m.status = "正在连接 " + m.cfg.Addr() + "..."
	m.statusOK = true
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
	m.table = tableState{}
	m.slow = tableState{}
	m.mode = modeResult
	return m.connectCmd()
}

func (m *Model) handleCfgModalKey(msg tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.closeCfgModal()
		return m, nil
	case "enter":
		return m, m.commitCfgModal()
	case "tab":
		m.cfgFields[m.cfgActive].Blur()
		m.cfgActive = (m.cfgActive + 1) % cfgFieldCount
		m.cfgFields[m.cfgActive].Focus()
		return m, nil
	case "shift+tab":
		m.cfgFields[m.cfgActive].Blur()
		m.cfgActive = (m.cfgActive - 1 + cfgFieldCount) % cfgFieldCount
		m.cfgFields[m.cfgActive].Focus()
		return m, nil
	case "left", "right", " ":
		if m.cfgActive == cfgFieldProtocol {
			m.toggleCfgProtocol()
			return m, nil
		}
	}

	if m.cfgActive == cfgFieldProtocol {
		return m, nil
	}
	var cmd tea.Cmd
	m.cfgFields[m.cfgActive], cmd = m.cfgFields[m.cfgActive].Update(msg)
	return m, cmd
}

func (m *Model) buildCfg() (ch.Config, error) {
	cfg := ch.Config{
		Host:     strings.TrimSpace(m.cfgFields[cfgFieldHost].Value()),
		Username: strings.TrimSpace(m.cfgFields[cfgFieldUser].Value()),
		Password: m.cfgFields[cfgFieldPass].Value(),
		Database: strings.TrimSpace(m.cfgFields[cfgFieldDB].Value()),
		Protocol: ch.NormalizeProtocol(m.cfgFields[cfgFieldProtocol].Value()),
	}
	if cfg.Host == "" {
		return cfg, errCfg("Host 不能为空")
	}
	port, err := strconv.Atoi(strings.TrimSpace(m.cfgFields[cfgFieldPort].Value()))
	if err != nil || port <= 0 || port > 65535 {
		return cfg, errCfg("Port 必须是 1-65535 的数字")
	}
	cfg.Port = port
	return cfg, nil
}

type cfgErr string

func (e cfgErr) Error() string { return string(e) }

func errCfg(s string) error { return cfgErr(s) }

func (m *Model) handleCfgModalMouse(msg tea.MouseMsg, isClick bool) (tea.Model, tea.Cmd) {
	if !isClick {
		return m, nil
	}

	r := m.cfgModalRect
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
			return m, m.commitCfgModal()
		case footerRoleCancel:
			m.closeCfgModal()
		}
		return m, nil
	}

	fieldIdx := relY - r.fieldStart
	if fieldIdx < 0 || fieldIdx >= r.fieldCount {
		return m, nil
	}
	m.cfgFields[m.cfgActive].Blur()
	m.cfgActive = cfgField(fieldIdx)
	m.cfgFields[m.cfgActive].Focus()
	if relY == r.protocolRow {
		m.toggleCfgProtocol()
	}
	return m, nil
}
