package ui

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
	"github.com/yuwen002/meteor-ch-tui/internal/configstore"
)

func (m *Model) openAddModal() {
	m.mgrAddModal = true
	m.mgrEditOrigName = ""
	m.mgrAddName = textinput.New()
	m.mgrAddName.Placeholder = "连接名称（如 db-prod-01）"
	m.mgrAddName.CharLimit = 64
	m.mgrAddName.Width = 40

	m.mgrAddGroup = textinput.New()
	m.mgrAddGroup.Placeholder = "分组（如 生产环境）"
	m.mgrAddGroup.CharLimit = 64
	m.mgrAddGroup.Width = 40

	m.mgrAddHost = textinput.New()
	m.mgrAddHost.Placeholder = "127.0.0.1"
	m.mgrAddHost.CharLimit = 128
	m.mgrAddHost.Width = 40

	m.mgrAddPort = textinput.New()
	m.mgrAddPort.Placeholder = "9000"
	m.mgrAddPort.CharLimit = 5
	m.mgrAddPort.Width = 20

	m.mgrAddUser = textinput.New()
	m.mgrAddUser.Placeholder = "default"
	m.mgrAddUser.CharLimit = 64
	m.mgrAddUser.Width = 40

	m.mgrAddPass = textinput.New()
	m.mgrAddPass.EchoMode = textinput.EchoPassword
	m.mgrAddPass.EchoCharacter = '•'
	m.mgrAddPass.CharLimit = 128
	m.mgrAddPass.Width = 40

	m.mgrAddDB = textinput.New()
	m.mgrAddDB.Placeholder = "留空可浏览全部库"
	m.mgrAddDB.CharLimit = 64
	m.mgrAddDB.Width = 40

	m.mgrAddProtocol = textinput.New()
	m.mgrAddProtocol.SetValue(ch.ProtocolNative)
	m.mgrAddProtocol.CharLimit = 16
	m.mgrAddProtocol.Width = 20

	m.mgrTestSeq++
	m.mgrTesting = false
	m.mgrTestOK = false
	m.mgrTestResult = ""
	m.mgrAddName.Focus()
	m.mgrAddFieldIdx = 0
}

func (m *Model) openEditModal() {
	conn := m.currentMgrConn()
	if conn == nil {
		return
	}
	m.mgrEditModal = true
	m.mgrEditOrigName = conn.Name
	m.mgrAddModal = true

	m.mgrAddName = textinput.New()
	m.mgrAddName.SetValue(conn.Name)
	m.mgrAddName.CharLimit = 64
	m.mgrAddName.Width = 40

	m.mgrAddGroup = textinput.New()
	m.mgrAddGroup.SetValue(conn.Group)
	m.mgrAddGroup.CharLimit = 64
	m.mgrAddGroup.Width = 40

	m.mgrAddHost = textinput.New()
	m.mgrAddHost.SetValue(conn.Host)
	m.mgrAddHost.CharLimit = 128
	m.mgrAddHost.Width = 40

	m.mgrAddPort = textinput.New()
	m.mgrAddPort.SetValue(strconv.Itoa(conn.Port))
	m.mgrAddPort.CharLimit = 5
	m.mgrAddPort.Width = 20

	m.mgrAddUser = textinput.New()
	m.mgrAddUser.SetValue(conn.Username)
	m.mgrAddUser.CharLimit = 64
	m.mgrAddUser.Width = 40

	m.mgrAddPass = textinput.New()
	m.mgrAddPass.EchoMode = textinput.EchoPassword
	m.mgrAddPass.EchoCharacter = '•'
	m.mgrAddPass.SetValue(conn.Password)
	m.mgrAddPass.CharLimit = 128
	m.mgrAddPass.Width = 40

	m.mgrAddDB = textinput.New()
	m.mgrAddDB.SetValue(conn.Database)
	m.mgrAddDB.CharLimit = 64
	m.mgrAddDB.Width = 40

	m.mgrAddProtocol = textinput.New()
	m.mgrAddProtocol.SetValue(ch.NormalizeProtocol(conn.Protocol))
	m.mgrAddProtocol.CharLimit = 16
	m.mgrAddProtocol.Width = 20

	m.mgrTestSeq++
	m.mgrTesting = false
	m.mgrTestOK = false
	m.mgrTestResult = ""
	m.mgrAddName.Focus()
	m.mgrAddFieldIdx = 0
}

func (m *Model) currentMgrConn() *configstore.SavedConnection {
	groups := m.cs.Groups()
	if m.mgrGroupIdx >= len(groups) {
		return nil
	}
	conns := m.cs.ByGroup(groups[m.mgrGroupIdx])
	if m.mgrConnIdx >= len(conns) {
		return nil
	}
	return &conns[m.mgrConnIdx]
}

func (m *Model) handleMgrKey(msg tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	if m.mgrImportModal {
		return m.handleMgrImportKey(msg, key)
	}
	if m.mgrAddModal {
		return m.handleMgrAddKey(msg, key)
	}
	if m.mgrGroupModal {
		return m.handleMgrGroupKey(msg, key)
	}

	switch key {
	case "g":
		m.openGroupModal()
		return m, nil
	case "up", "k":
		moveMgrSelection(m.cs, &m.mgrGroupIdx, &m.mgrConnIdx, -1)
		return m, nil
	case "down", "j":
		moveMgrSelection(m.cs, &m.mgrGroupIdx, &m.mgrConnIdx, 1)
		return m, nil
	case "enter":
		conn := m.currentMgrConn()
		if conn == nil {
			m.status = "请先选择一个连接"
			m.statusOK = false
			return m, nil
		}
		m.cfg = conn.ToCfg()
		m.cs.SetLastUsed(conn.Name)
		_ = m.cs.Save()
		m.initMainMode()
		m.appMode = appMain
		m.loading = true
		m.status = "正在连接 " + conn.Name + " (" + m.cfg.Addr() + ")..."
		m.statusOK = true
		return m, m.connectCmd()
	case "+":
		m.mgrEditModal = false
		m.openAddModal()
		return m, nil
	case "e":
		m.openEditModal()
		return m, nil
	case " ":
		m.toggleConnMark()
		return m, nil
	case "ctrl+a":
		m.markAllConns()
		return m, nil
	case "m":
		m.openBatchMoveModal()
		return m, nil
	case "delete", "backspace":
		m.requestDeleteConn()
		return m, nil
	case "ctrl+e":
		m.openMgrExportModal()
		return m, nil
	// Ctrl+I 与 Tab 同为控制码 9，Bubble Tea 只会解析成 "tab"，故两者一并匹配。
	case "tab", "ctrl+i":
		m.openMgrImportModal()
		return m, nil
	case "?", "f1":
		m.openHelp()
		return m, nil
	}
	return m, nil
}

func moveMgrSelection(cs *configstore.ConfigFile, gIdx, cIdx *int, delta int) {
	groups := cs.Groups()
	if len(groups) == 0 {
		return
	}
	conns := cs.ByGroup(groups[*gIdx])
	newG, newC := *gIdx, *cIdx
	if delta > 0 {
		newC++
		for newG < len(groups) && newC >= len(cs.ByGroup(groups[newG])) {
			newC = 0
			newG++
			if newG < len(groups) {
				conns = cs.ByGroup(groups[newG])
			}
		}
	} else {
		newC--
		for newG >= 0 && newC < 0 {
			newG--
			if newG >= 0 {
				conns = cs.ByGroup(groups[newG])
				newC = len(conns) - 1
			}
		}
	}
	if newG < 0 || newG >= len(groups) {
		return
	}
	*gIdx = newG
	*cIdx = newC
}

func clampMgrSelection(cs *configstore.ConfigFile, gIdx, cIdx *int) {
	groups := cs.Groups()
	if len(groups) == 0 {
		*gIdx = 0
		*cIdx = 0
		return
	}
	if *gIdx >= len(groups) {
		*gIdx = len(groups) - 1
	}
	if *gIdx < 0 {
		*gIdx = 0
	}
	conns := cs.ByGroup(groups[*gIdx])
	if len(conns) == 0 {
		*cIdx = 0
		return
	}
	if *cIdx >= len(conns) {
		*cIdx = len(conns) - 1
	}
	if *cIdx < 0 {
		*cIdx = 0
	}
}

var mgrAddFieldsInfo = []struct {
	name string
	ti   func(m *Model) *textinput.Model
}{
	{"名称", func(m *Model) *textinput.Model { return &m.mgrAddName }},
	{"分组", func(m *Model) *textinput.Model { return &m.mgrAddGroup }},
	{"Host", func(m *Model) *textinput.Model { return &m.mgrAddHost }},
	{"Port", func(m *Model) *textinput.Model { return &m.mgrAddPort }},
	{"协议", func(m *Model) *textinput.Model { return &m.mgrAddProtocol }},
	{"Username", func(m *Model) *textinput.Model { return &m.mgrAddUser }},
	{"Password", func(m *Model) *textinput.Model { return &m.mgrAddPass }},
	{"Database", func(m *Model) *textinput.Model { return &m.mgrAddDB }},
}

// mgrAddGroupValues 返回新建/编辑弹窗中可快捷填入的分组值（默认组用空字符串表示）。
func (m *Model) mgrAddGroupValues() []string {
	groups := m.cs.Groups()
	vals := make([]string, 0, len(groups))
	for _, g := range groups {
		if g == configstore.DefaultGroupName {
			vals = append(vals, "")
		} else {
			vals = append(vals, g)
		}
	}
	return vals
}

// cycleMgrAddGroup 在已有分组中按方向循环，填入分组输入框。
func (m *Model) cycleMgrAddGroup(next bool) {
	vals := m.mgrAddGroupValues()
	if len(vals) == 0 {
		return
	}
	cur := strings.TrimSpace(m.mgrAddGroup.Value())
	idx := -1
	for i, v := range vals {
		if v == cur {
			idx = i
			break
		}
	}
	if next {
		idx++
	} else {
		idx--
	}
	idx = (idx + len(vals)) % len(vals)
	m.mgrAddGroup.SetValue(vals[idx])
	m.mgrAddGroup.CursorEnd()
}

func (m *Model) handleMgrAddKey(msg tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.closeMgrAddModal()
		return m, nil
	case "enter":
		m.saveMgrAddConn()
		return m, nil
	case "ctrl+t":
		return m.startConnTest()
	case "tab":
		m.mgrAddFieldIdx = (m.mgrAddFieldIdx + 1) % len(mgrAddFieldsInfo)
		m.focusMgrAddField()
		return m, nil
	case "shift+tab":
		m.mgrAddFieldIdx = (m.mgrAddFieldIdx - 1 + len(mgrAddFieldsInfo)) % len(mgrAddFieldsInfo)
		m.focusMgrAddField()
		return m, nil
	case "left", "right", " ":
		if mgrAddFieldsInfo[m.mgrAddFieldIdx].name == "协议" {
			m.toggleMgrAddProtocol()
			return m, nil
		}
		if mgrAddFieldsInfo[m.mgrAddFieldIdx].name == "分组" && (key == "left" || key == "right") {
			m.cycleMgrAddGroup(key == "right")
			return m, nil
		}
	}
	if mgrAddFieldsInfo[m.mgrAddFieldIdx].name == "协议" {
		return m, nil
	}
	ti := mgrAddFieldsInfo[m.mgrAddFieldIdx].ti(m)
	var cmd tea.Cmd
	*ti, cmd = ti.Update(msg)
	return m, cmd
}

func (m *Model) focusMgrAddField() {
	for i, info := range mgrAddFieldsInfo {
		if i == m.mgrAddFieldIdx {
			info.ti(m).Focus()
		} else {
			info.ti(m).Blur()
		}
	}
}

func (m *Model) buildSavedConn() (configstore.SavedConnection, error) {
	name := strings.TrimSpace(m.mgrAddName.Value())
	group := strings.TrimSpace(m.mgrAddGroup.Value())
	host := strings.TrimSpace(m.mgrAddHost.Value())
	portStr := strings.TrimSpace(m.mgrAddPort.Value())
	user := strings.TrimSpace(m.mgrAddUser.Value())
	pass := m.mgrAddPass.Value()
	db := strings.TrimSpace(m.mgrAddDB.Value())

	if name == "" {
		return configstore.SavedConnection{}, errCfg("名称不能为空")
	}
	if host == "" {
		return configstore.SavedConnection{}, errCfg("Host 不能为空")
	}
	port := 9000
	if portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil || p <= 0 || p > 65535 {
			return configstore.SavedConnection{}, errCfg("Port 必须是 1-65535")
		}
		port = p
	}
	if user == "" {
		user = "default"
	}
	return configstore.SavedConnection{
		Name:     name,
		Group:    group,
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
		Database: db,
		Protocol: ch.NormalizeProtocol(m.mgrAddProtocol.Value()),
	}, nil
}

func (m *Model) connTestConfig() (ch.Config, error) {
	host := strings.TrimSpace(m.mgrAddHost.Value())
	portStr := strings.TrimSpace(m.mgrAddPort.Value())
	user := strings.TrimSpace(m.mgrAddUser.Value())
	if host == "" {
		return ch.Config{}, errCfg("Host 不能为空")
	}
	port := 9000
	if portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil || p <= 0 || p > 65535 {
			return ch.Config{}, errCfg("Port 必须是 1-65535")
		}
		port = p
	}
	if user == "" {
		user = "default"
	}
	return ch.Config{
		Host: host, Port: port, Username: user,
		Password: m.mgrAddPass.Value(),
		Database: strings.TrimSpace(m.mgrAddDB.Value()),
		Protocol: ch.NormalizeProtocol(m.mgrAddProtocol.Value()),
	}, nil
}

func (m *Model) testConnCmd(seq int, cfg ch.Config) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		start := time.Now()
		version, err := ch.Ping(ctx, cfg)
		return testConnMsg{seq: seq, version: version, elapsed: time.Since(start), err: err}
	}
}

func (m *Model) startConnTest() (tea.Model, tea.Cmd) {
	if m.mgrTesting {
		return m, nil
	}
	cfg, err := m.connTestConfig()
	if err != nil {
		m.mgrTestOK = false
		m.mgrTestResult = err.Error()
		return m, nil
	}
	m.mgrTestSeq++
	m.mgrTesting = true
	m.mgrTestResult = ""
	return m, m.testConnCmd(m.mgrTestSeq, cfg)
}

func (m *Model) clearModalError() {
	if !m.statusOK {
		m.status = ""
		m.statusOK = true
	}
}

func (m *Model) closeMgrAddModal() {
	m.mgrAddModal = false
	m.mgrEditModal = false
	m.mgrEditOrigName = ""
	m.clearModalError()
	m.mgrTestSeq++
	m.mgrTesting = false
}

func (m *Model) saveMgrAddConn() {
	conn, err := m.buildSavedConn()
	if err != nil {
		m.status = "保存失败: " + err.Error()
		m.statusOK = false
		return
	}
	if m.mgrEditModal {
		orig := m.mgrEditOrigName
		if orig == "" {
			orig = conn.Name
		}
		if other := m.cs.Find(conn.Name); other != nil && conn.Name != orig {
			m.status = "名称已存在，请换一个"
			m.statusOK = false
			return
		}
		if !m.cs.Replace(orig, conn) {
			m.cs.Add(conn)
		}
		if m.cs.LastUsed == orig {
			m.cs.SetLastUsed(conn.Name)
		}
	} else {
		if m.cs.Find(conn.Name) != nil {
			m.status = "名称已存在，请换一个"
			m.statusOK = false
			return
		}
		m.cs.Add(conn)
	}
	if err := m.cs.Save(); err != nil {
		m.status = "保存失败: " + err.Error()
		m.statusOK = false
		return
	}
	m.closeMgrAddModal()
	m.status = "已保存连接: " + conn.Name
	m.statusOK = true
	for i, g := range m.cs.Groups() {
		for j, c := range m.cs.ByGroup(g) {
			if c.Name == conn.Name {
				m.mgrGroupIdx = i
				m.mgrConnIdx = j
				return
			}
		}
	}
}

func (m *Model) openMgrImportModal() {
	m.mgrImportModal = true
	m.mgrExportMode = false
	ta := textarea.New()
	ta.Placeholder = "粘贴 JSON/CSV/TSV，或用 Ctrl+B 选择文件..."
	ta.SetHeight(12)
	ta.CharLimit = 0
	m.mgrImportText = ta
	m.mgrImportOverwrite = false
	m.mgrImportText.Focus()
}

func (m *Model) closeMgrImportModal() {
	m.mgrImportModal = false
	m.mgrExportMode = false
	m.mgrImportRect = modalRect{}
	m.closeSheetPicker()
	m.blurExportPath()
}

func (m *Model) commitMgrImport() (tea.Model, tea.Cmd) {
	if m.mgrExportMode {
		return m.copyMgrExport()
	}
	imported, err := configstore.ImportAuto(m.mgrImportText.Value())
	if err != nil {
		m.status = "导入失败: " + err.Error()
		m.statusOK = false
		return m, nil
	}
	merged := m.cs.Clone()
	added, updated := merged.Merge(imported, m.mgrImportOverwrite)
	if err := merged.Save(); err != nil {
		m.status = "保存失败: " + err.Error()
		m.statusOK = false
		return m, nil
	}
	*m.cs = *merged
	m.closeMgrImportModal()
	m.status = "已导入 " + strconv.Itoa(added) + " 个连接"
	if updated > 0 {
		m.status += "，覆盖 " + strconv.Itoa(updated) + " 个同名连接"
	}
	m.statusOK = true
	return m, nil
}

func (m *Model) toggleMgrImportOption() {
	if m.mgrExportMode {
		if m.mgrExportPasswords {
			m.mgrExportPasswords = false
			m.refreshMgrExportPreview()
			return
		}
		m.requestExportPasswords()
		return
	}
	m.mgrImportOverwrite = !m.mgrImportOverwrite
}

func (m *Model) handleMgrImportKey(msg tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	if m.mgrExportMode && m.exportPathFocused {
		switch key {
		case "esc", "enter", "ctrl+p":
			m.blurExportPath()
			return m, nil
		case "ctrl+b":
			m.blurExportPath()
			m.openDirBrowser()
			return m, nil
		}
		var cmd tea.Cmd
		m.exportPathInput, cmd = m.exportPathInput.Update(msg)
		return m, cmd
	}
	switch key {
	case "esc":
		m.closeMgrImportModal()
		return m, nil
	case "tab":
		m.toggleMgrImportOption()
		return m, nil
	case "ctrl+o":
		return m.commitMgrImport()
	case "ctrl+p":
		if m.mgrExportMode {
			return m, m.focusExportPath()
		}
	case "ctrl+b":
		if m.mgrExportMode {
			m.openDirBrowser()
		} else {
			m.openFileBrowser()
		}
		return m, nil
	case "left":
		if m.mgrExportMode {
			m.cycleMgrExportFormat(-1)
			return m, nil
		}
	case "right":
		if m.mgrExportMode {
			m.cycleMgrExportFormat(1)
			return m, nil
		}
	case "enter":
		if m.mgrExportMode {
			return m.saveMgrExport()
		}
	}
	var cmd tea.Cmd
	m.mgrImportText, cmd = m.mgrImportText.Update(msg)
	return m, cmd
}

func (m *Model) handleMgrImportMouse(msg tea.MouseMsg, isClick bool) (tea.Model, tea.Cmd) {
	if !isClick {
		return m, nil
	}
	r := m.mgrImportRect
	if r.w == 0 || r.h == 0 {
		return m, nil
	}
	if msg.X < r.left || msg.X >= r.left+r.w || msg.Y < r.top || msg.Y >= r.top+r.h {
		return m, nil
	}
	relX, relY := msg.X-r.left, msg.Y-r.top
	if m.mgrExportMode && r.browseW > 0 && relY == r.pathRow &&
		relX >= r.browseX && relX < r.browseX+r.browseW {
		m.blurExportPath()
		m.openDirBrowser()
		return m, nil
	}
	if m.mgrExportMode && r.pathRow > 0 && relY == r.pathRow &&
		r.pathW > 0 && relX >= r.pathX && relX < r.pathX+r.pathW {
		return m, m.focusExportPath()
	}
	if m.mgrExportMode && r.listCount > 0 && relY == r.listStart {
		m.cycleMgrExportFormat(1)
		return m, nil
	}
	if r.protocolRow > 0 && relY == r.protocolRow &&
		r.testW > 0 && relX >= r.testX && relX < r.testX+r.testW {
		m.toggleMgrImportOption()
		return m, nil
	}
	if role, ok := hitFooter(relX, relY, r, footerRoleOK, footerRoleExtra, footerRoleCancel); ok {
		switch role {
		case footerRoleOK:
			if m.mgrExportMode {
				return m.saveMgrExport()
			}
			return m.commitMgrImport()
		case footerRoleExtra:
			if m.mgrExportMode {
				return m.copyMgrExport()
			}
			m.openFileBrowser()
		case footerRoleCancel:
			m.closeMgrImportModal()
		}
		return m, nil
	}
	if relY >= r.fieldStart && relY < r.fieldStart+r.fieldCount {
		return m, m.mgrImportText.Focus()
	}
	return m, nil
}

func (m *Model) handleMgrMouse(msg tea.MouseMsg, isClick, isWheelUp, isWheelDown bool) (tea.Model, tea.Cmd) {
	if m.mgrGroupModal {
		return m.handleMgrGroupMouse(msg, isClick, isWheelUp, isWheelDown)
	}
	if m.mgrAddModal {
		return m.handleMgrAddModalMouse(msg, isClick)
	}
	if m.mgrImportModal {
		return m.handleMgrImportMouse(msg, isClick)
	}

	if isWheelUp {
		moveMgrSelection(m.cs, &m.mgrGroupIdx, &m.mgrConnIdx, -1)
		return m, nil
	}
	if isWheelDown {
		moveMgrSelection(m.cs, &m.mgrGroupIdx, &m.mgrConnIdx, 1)
		return m, nil
	}
	if !isClick {
		return m, nil
	}
	if m.handleMgrToolbarClick(msg.X, msg.Y) {
		return m, nil
	}

	// 单击与双击都先按坐标命中，避免双击连到上一次选中的连接。
	m.selectMgrLineAt(msg.Y)
	if !m.isDoubleClick(msg.X, msg.Y) {
		return m, nil
	}
	conn := m.currentMgrConn()
	if conn == nil {
		return m, nil
	}
	m.cfg = conn.ToCfg()
	m.cs.SetLastUsed(conn.Name)
	_ = m.cs.Save()
	m.initMainMode()
	m.appMode = appMain
	m.loading = true
	m.status = "正在连接 " + conn.Name + " (" + m.cfg.Addr() + ")..."
	m.statusOK = true
	return m, m.connectCmd()
}

func (m *Model) handleMgrToolbarClick(x, y int) bool {
	if m.mgrToolRow == 0 || y != m.mgrToolRow {
		return false
	}
	switch {
	case m.mgrToolNewW > 0 && x >= m.mgrToolNewX && x < m.mgrToolNewX+m.mgrToolNewW:
		m.openAddModal()
		return true
	case m.mgrToolEditW > 0 && x >= m.mgrToolEditX && x < m.mgrToolEditX+m.mgrToolEditW:
		if m.currentMgrConn() == nil {
			m.status = "请先选择一个连接"
			m.statusOK = false
			return true
		}
		m.openEditModal()
		return true
	case m.mgrToolDelW > 0 && x >= m.mgrToolDelX && x < m.mgrToolDelX+m.mgrToolDelW:
		if m.currentMgrConn() == nil {
			m.status = "请先选择一个连接"
			m.statusOK = false
			return true
		}
		m.requestDeleteConn()
		return true
	case m.mgrToolGroupW > 0 && x >= m.mgrToolGroupX && x < m.mgrToolGroupX+m.mgrToolGroupW:
		m.openGroupModal()
		return true
	case m.mgrToolImportW > 0 && x >= m.mgrToolImportX && x < m.mgrToolImportX+m.mgrToolImportW:
		m.openMgrImportModal()
		return true
	case m.mgrToolExportW > 0 && x >= m.mgrToolExportX && x < m.mgrToolExportX+m.mgrToolExportW:
		m.openMgrExportModal()
		return true
	}
	return false
}

func (m *Model) selectMgrLineAt(screenY int) {
	if len(m.mgrLines) == 0 || m.mgrListFirstLineY == 0 {
		return
	}
	row := m.mgrListTop + (screenY - m.mgrListFirstLineY)
	if row < 0 || row >= len(m.mgrLines) {
		return
	}
	line := m.mgrLines[row]
	if line.group < 0 {
		return
	}
	m.mgrGroupIdx = line.group
	if line.conn >= 0 {
		m.mgrConnIdx = line.conn
	} else {
		m.mgrConnIdx = 0
	}
}

func (m *Model) handleMgrAddModalMouse(msg tea.MouseMsg, isClick bool) (tea.Model, tea.Cmd) {
	if !isClick {
		return m, nil
	}

	r := m.addModalRect
	if r.w == 0 || r.h == 0 {
		return m, nil
	}
	if msg.X < r.left || msg.X >= r.left+r.w || msg.Y < r.top || msg.Y >= r.top+r.h {
		return m, nil
	}

	relX, relY := msg.X-r.left, msg.Y-r.top
	if role, ok := hitFooter(relX, relY, r, footerRoleTest, footerRoleOK, footerRoleCancel); ok {
		switch role {
		case footerRoleTest:
			return m.startConnTest()
		case footerRoleOK:
			m.saveMgrAddConn()
		case footerRoleCancel:
			m.closeMgrAddModal()
		}
		return m, nil
	}

	// 已有分组快捷项：点击后填入分组输入框。
	if relY == r.groupRow {
		for i, x := range r.groupXs {
			if i < len(r.groupWs) && i < len(r.groupVals) &&
				relX >= x && relX < x+r.groupWs[i] {
				m.mgrAddGroup.SetValue(r.groupVals[i])
				m.mgrAddGroup.CursorEnd()
				return m, nil
			}
		}
	}

	fieldIdx := relY - r.fieldStart
	if fieldIdx < 0 || fieldIdx >= r.fieldCount {
		return m, nil
	}
	m.mgrAddFieldIdx = fieldIdx
	m.focusMgrAddField()
	if relY == r.protocolRow {
		m.toggleMgrAddProtocol()
	}
	return m, nil
}

func (m *Model) openGroupModal() {
	m.mgrGroupModal = true
	groups := m.cs.Groups()
	m.mgrGroupSel = clampInt(m.mgrGroupIdx, 0, maxInt(0, len(groups)-1))
	m.mgrGroupInput = textinput.New()
	m.mgrGroupInput.Placeholder = "输入分组名称"
	m.mgrGroupInput.CharLimit = 64
	m.mgrGroupInput.Width = 28
	m.mgrGroupInput.Focus()
}

func (m *Model) closeGroupModal() {
	m.mgrGroupModal = false
	m.mgrGroupInput.Blur()
}

func (m *Model) saveGroups() error {
	if err := m.cs.Save(); err != nil {
		m.status = "分组保存失败: " + err.Error()
		m.statusOK = false
		return err
	}
	return nil
}

func (m *Model) addGroupFromInput() {
	name := strings.TrimSpace(m.mgrGroupInput.Value())
	if name == "" {
		m.status = "请输入分组名称"
		m.statusOK = false
		return
	}
	if !m.cs.AddGroup(name) {
		m.status = "分组已存在或名称非法"
		m.statusOK = false
		return
	}
	if err := m.saveGroups(); err != nil {
		return
	}
	m.mgrGroupSel = len(m.cs.Groups()) - 1
	m.mgrGroupInput.SetValue("")
	m.status = "已新建分组: " + name
	m.statusOK = true
}

func (m *Model) renameGroupFromInput() {
	name := strings.TrimSpace(m.mgrGroupInput.Value())
	groups := m.cs.Groups()
	if name == "" {
		m.status = "请输入新分组名称"
		m.statusOK = false
		return
	}
	if m.mgrGroupSel < 0 || m.mgrGroupSel >= len(groups) {
		m.status = "请先选择一个分组"
		m.statusOK = false
		return
	}
	old := groups[m.mgrGroupSel]
	if !m.cs.RenameGroup(old, name) {
		m.status = "重命名失败（默认组不可改名，或名称已存在）"
		m.statusOK = false
		return
	}
	if err := m.saveGroups(); err != nil {
		return
	}
	m.mgrGroupInput.SetValue("")
	m.status = "已将分组 " + old + " 重命名为 " + name
	m.statusOK = true
}

func (m *Model) performDeleteGroup(name string) {
	if !m.cs.RemoveGroup(name) {
		m.status = "默认分组不能删除"
		m.statusOK = false
		return
	}
	if err := m.saveGroups(); err != nil {
		return
	}
	groups := m.cs.Groups()
	m.mgrGroupSel = clampInt(m.mgrGroupSel, 0, maxInt(0, len(groups)-1))
	m.mgrGroupIdx = clampInt(m.mgrGroupIdx, 0, maxInt(0, len(groups)-1))
	m.mgrConnIdx = 0
	m.status = "已删除分组 " + name + "，其连接已移回默认分组"
	m.statusOK = true
}

func (m *Model) reorderSelectedGroup(delta int) {
	groups := m.cs.Groups()
	if m.mgrGroupSel < 0 || m.mgrGroupSel >= len(groups) {
		return
	}
	name := groups[m.mgrGroupSel]
	if !m.cs.MoveGroup(name, delta) {
		m.status = "已到边界，无法移动"
		m.statusOK = false
		return
	}
	if err := m.saveGroups(); err != nil {
		return
	}
	oldSel := m.mgrGroupSel
	m.mgrGroupSel = clampInt(oldSel+delta, 0, maxInt(0, len(m.cs.Groups())-1))
	if m.mgrGroupIdx == oldSel {
		m.mgrGroupIdx = m.mgrGroupSel
	}
	m.status = "已调整分组顺序: " + name
	m.statusOK = true
}

func (m *Model) handleMgrGroupKey(msg tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	groups := m.cs.Groups()
	switch key {
	case "esc":
		m.closeGroupModal()
		return m, nil
	case "up":
		if len(groups) > 0 {
			m.mgrGroupSel = (m.mgrGroupSel - 1 + len(groups)) % len(groups)
		}
		return m, nil
	case "down":
		if len(groups) > 0 {
			m.mgrGroupSel = (m.mgrGroupSel + 1) % len(groups)
		}
		return m, nil
	case "ctrl+k":
		m.reorderSelectedGroup(-1)
		return m, nil
	case "ctrl+j":
		m.reorderSelectedGroup(1)
		return m, nil
	case "enter":
		m.addGroupFromInput()
		return m, nil
	case "ctrl+r":
		m.renameGroupFromInput()
		return m, nil
	case "ctrl+d":
		m.requestDeleteGroup()
		return m, nil
	}
	var cmd tea.Cmd
	m.mgrGroupInput, cmd = m.mgrGroupInput.Update(msg)
	return m, cmd
}

func (m *Model) handleMgrGroupMouse(msg tea.MouseMsg, isClick, isWheelUp, isWheelDown bool) (tea.Model, tea.Cmd) {
	groups := m.cs.Groups()
	if isWheelUp && len(groups) > 0 {
		m.mgrGroupSel = (m.mgrGroupSel - 1 + len(groups)) % len(groups)
		return m, nil
	}
	if isWheelDown && len(groups) > 0 {
		m.mgrGroupSel = (m.mgrGroupSel + 1) % len(groups)
		return m, nil
	}
	if !isClick {
		return m, nil
	}
	r := m.mgrGroupRect
	if r.w == 0 || r.h == 0 {
		return m, nil
	}
	if msg.X < r.left || msg.X >= r.left+r.w || msg.Y < r.top || msg.Y >= r.top+r.h {
		return m, nil
	}
	relX, relY := msg.X-r.left, msg.Y-r.top
	if role, ok := hitFooter(relX, relY, r, footerRoleTest, footerRoleOK, footerRoleCancel, footerRoleExtra); ok {
		switch role {
		case footerRoleTest:
			m.renameGroupFromInput()
		case footerRoleOK:
			m.addGroupFromInput()
		case footerRoleCancel:
			m.requestDeleteGroup()
		case footerRoleExtra:
			m.closeGroupModal()
		}
		return m, nil
	}
	idx := relY - r.listStart
	if idx >= 0 && idx < r.listCount {
		m.mgrGroupSel = idx
		return m, nil
	}
	if relY == r.fieldStart {
		return m, m.mgrGroupInput.Focus()
	}
	return m, nil
}
