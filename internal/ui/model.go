package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
	"github.com/yuwen002/meteor-ch-tui/internal/configstore"
)

type focusArea int

const (
	focusSidebar focusArea = iota
	focusEditor
	focusTable
)

type paneMode int

const (
	modeResult paneMode = iota
	modeSlow
)

type sidebarTab int

const (
	tabDatabases sidebarTab = iota
	tabViews
	tabNodes
)

type sidebarItemKind int

const (
	itemDatabase sidebarItemKind = iota
	itemTable
	itemView
	itemNode
)

// sidebarItem 把侧边栏渲染出的可选项关联到底层数据与行号。
type sidebarItem struct {
	kind sidebarItemKind
	idx  int
	row  int
}

// mgrLine 记录连接管理器列表每一行对应的分组/连接下标。
// group < 0 表示标题行；conn < 0 表示分组标题行。
type mgrLine struct {
	group int
	conn  int
}

type confirmKind int

const (
	confirmNone confirmKind = iota
	confirmDeleteConn
	confirmDeleteGroup
)

// pendingConfirm 保存待确认的删除动作，确认后才真正执行。
type pendingConfirm struct {
	kind confirmKind
	name string
	info string
}

type contextAction int

const (
	contextPreview contextAction = iota
	contextCopyName
	contextRefresh
	contextResult
	contextManager
)

type contextMenuItem struct {
	label   string
	action  contextAction
	enabled bool
}

type fileBrowserItem struct {
	name    string
	isDir   bool
	modTime time.Time
}

type cfgField int

const (
	cfgFieldHost cfgField = iota
	cfgFieldPort
	cfgFieldUser
	cfgFieldPass
	cfgFieldDB
	cfgFieldProtocol
	cfgFieldCount
)

type appMode int

const (
	appManager appMode = iota
	appMain
)

const (
	sidebarOuterWidth = 32
	editorPaneHeight  = 7
)

type connectedMsg struct {
	client             *ch.Client
	nodes              []ch.Node
	node               *ch.Node
	databases          []string
	tables             []string
	views              []string
	databasesTruncated bool
	objectsTruncated   bool
	database           string
	catalogErr         error
	err                error
}

// testConnMsg 是「测试连接」的异步结果，seq 用于丢弃过期结果。
type testConnMsg struct {
	seq     int
	version string
	elapsed time.Duration
	err     error
}

type queryDoneMsg struct {
	res *ch.Result
	sql string
	err error
}

type slowDoneMsg struct {
	res *ch.Result
	err error
}

// Model 是 bubbletea 应用的根模型，持有全部 UI 状态。
type Model struct {
	appMode appMode

	cfg ch.Config
	cs  *configstore.ConfigFile

	client  *ch.Client
	nodes   []ch.Node
	nodeIdx int
	curNode *ch.Node

	// 数据库浏览：账号可见的全部库、当前库的表/视图与当前会话库。
	databases          []string
	tables             []string
	views              []string
	databasesTruncated bool
	objectsTruncated   bool
	currentDB          string

	sidebarTab     sidebarTab
	sidebarIdx     int
	sidebarItems   []sidebarItem
	sidebarLineTop int

	sidebarFirstLineY int
	sidebarTabRowY    int
	sidebarTabDBX     int
	sidebarTabDBW     int
	sidebarTabViewX   int
	sidebarTabViewW   int
	sidebarTabNodeX   int
	sidebarTabNodeW   int

	focus focusArea
	mode  paneMode

	editor textarea.Model
	table  tableState
	slow   tableState

	queryHistory      []string
	historyModal      bool
	historySel        int
	historyRect       modalRect
	resultExportModal bool
	resultExportSel   int
	resultExportRect  modalRect
	exportPathInput   textinput.Model
	exportPathFocused bool
	lastExportDir     string
	dirBrowserOpen    bool
	dirBrowserPath    string
	dirBrowserEntries []string
	dirBrowserSel     int
	dirBrowserRect    modalRect
	fileBrowserOpen   bool
	fileBrowserPath   string
	fileBrowserItems  []fileBrowserItem
	fileBrowserSel    int
	fileBrowserRect   modalRect
	lastImportDir     string

	fileBrowserFilter        textinput.Model
	fileBrowserFilterFocused bool
	fileBrowserSortMode      int

	sheetPickerOpen   bool
	sheetPickerPath   string
	sheetPickerSheets []string
	sheetPickerSel    int
	sheetPickerRect   modalRect

	width  int
	height int

	status   string
	statusOK bool
	err      error

	loading   bool
	rowModal  bool
	rowScroll int

	slowErr error

	isTTY        bool
	quickConnect bool

	cfgModal  bool
	cfgFields [cfgFieldCount]textinput.Model
	cfgActive cfgField

	mgrGroupIdx int
	mgrConnIdx  int

	mgrAddModal    bool
	mgrAddName     textinput.Model
	mgrAddGroup    textinput.Model
	mgrAddHost     textinput.Model
	mgrAddPort     textinput.Model
	mgrAddProtocol textinput.Model
	mgrAddUser     textinput.Model
	mgrAddPass     textinput.Model
	mgrAddDB       textinput.Model
	mgrAddFieldIdx int

	mgrEditModal    bool
	mgrEditOrigName string

	// 「测试连接」状态；seq 用于忽略已取消或已重开的弹窗的测试结果。
	mgrTestSeq    int
	mgrTesting    bool
	mgrTestOK     bool
	mgrTestResult string

	// 分组管理弹窗。
	mgrGroupModal bool
	mgrGroupInput textinput.Model
	mgrGroupSel   int
	mgrGroupRect  modalRect

	// 删除确认弹窗，叠加在其它弹窗之上。
	confirm     pendingConfirm
	confirmRect modalRect

	// 快捷键帮助浮层。
	helpOpen   bool
	helpScroll int
	helpRect   modalRect

	// 右键上下文菜单。
	ctxMenuOpen  bool
	ctxMenuX     int
	ctxMenuY     int
	ctxMenuSel   int
	ctxMenuItems []contextMenuItem
	ctxMenuRect  modalRect

	mgrImportModal     bool
	mgrImportText      textarea.Model
	mgrImportOverwrite bool
	mgrExportMode      bool
	mgrExportPasswords bool
	mgrExportFormat    configExportFormat
	mgrImportRect      modalRect

	// 管理器列表的渲染热区：行映射、视口起始行、内容首行屏幕 Y、工具栏按钮。
	mgrLines          []mgrLine
	mgrListTop        int
	mgrListFirstLineY int
	mgrToolRow        int
	mgrToolNewX       int
	mgrToolNewW       int
	mgrToolEditX      int
	mgrToolEditW      int
	mgrToolDelX       int
	mgrToolDelW       int
	mgrToolGroupX     int
	mgrToolGroupW     int
	mgrToolImportX    int
	mgrToolImportW    int
	mgrToolExportX    int
	mgrToolExportW    int

	lastClickX    int
	lastClickY    int
	lastClickTime int64

	addModalRect modalRect
	cfgModalRect modalRect
	backBtnX     int
	backBtnW     int
}

// NewModel 构造根模型。
func NewModel(cs *configstore.ConfigFile, isTTY bool) *Model {
	m := &Model{
		cs:    cs,
		isTTY: isTTY,
	}
	if state, err := configstore.LoadUIState(); err == nil {
		m.lastExportDir = state.LastExportDir
		m.lastImportDir = state.LastImportDir
	}

	if len(cs.Connections) == 0 {
		m.appMode = appManager
		m.openAddModal()
		m.status = "尚无连接，按 Tab 在字段间切换，Enter 保存"
		m.statusOK = true
		return m
	}

	m.appMode = appManager
	m.status = "选择一个连接按 Enter 连接 · + 新建 · e 编辑 · Del 删除 · Ctrl+E 导出 · Ctrl+I 导入"
	m.statusOK = true

	if cs.LastUsed != "" {
		for i, conn := range cs.AllFlat() {
			if conn.Name == cs.LastUsed {
				flat := cs.AllFlat()
				for gIdx, g := range cs.Groups() {
					for cIdx, c := range cs.ByGroup(g) {
						if c.Name == conn.Name {
							m.mgrGroupIdx = gIdx
							m.mgrConnIdx = cIdx
							_ = i
							break
						}
					}
				}
				_ = flat
				break
			}
		}
	}

	return m
}

func (m *Model) persistDirs() {
	_ = configstore.SaveUIState(configstore.UIState{
		LastExportDir: m.lastExportDir,
		LastImportDir: m.lastImportDir,
	})
}

func (m *Model) setLastExportDir(dir string) {
	if m.lastExportDir == dir {
		return
	}
	m.lastExportDir = dir
	m.persistDirs()
}

func (m *Model) setLastImportDir(dir string) {
	if m.lastImportDir == dir {
		return
	}
	m.lastImportDir = dir
	m.persistDirs()
}

func NewModelQuickConnect(cfg ch.Config, cs *configstore.ConfigFile, isTTY bool) *Model {
	m := &Model{
		cfg:          cfg,
		cs:           cs,
		isTTY:        isTTY,
		appMode:      appMain,
		quickConnect: true,
	}
	m.initMainMode()
	m.loading = true
	m.status = "正在连接 " + cfg.Addr() + "..."
	m.statusOK = true
	return m
}

func (m *Model) initMainMode() {
	ta := textarea.New()
	ta.Placeholder = "SELECT ... FROM ... (Alt+Enter 执行)"
	ta.Prompt = "SQL> "
	ta.CharLimit = 0
	ta.ShowLineNumbers = false
	ta.MaxHeight = 4
	ta.SetHeight(4)
	ta.Focus()
	m.editor = ta
	m.focus = focusEditor
	m.initCfgFields()
	m.databases = nil
	m.tables = nil
	m.views = nil
	m.databasesTruncated = false
	m.objectsTruncated = false
	m.currentDB = ""
	m.sidebarTab = tabDatabases
	m.sidebarIdx = 0
	m.sidebarItems = nil
	m.nodeIdx = 0
}

func (m *Model) initCfgFields() {
	labels := []string{"Host", "Port", "Username", "Password", "Database", "协议"}
	values := []string{m.cfg.Host, strconv.Itoa(m.cfg.Port), m.cfg.Username, m.cfg.Password, m.cfg.Database, ch.NormalizeProtocol(m.cfg.Protocol)}
	for i := 0; i < int(cfgFieldCount); i++ {
		ti := textinput.New()
		ti.CharLimit = 128
		ti.Width = 40
		ti.Placeholder = labels[i]
		ti.SetValue(values[i])
		if i == int(cfgFieldPass) {
			ti.EchoMode = textinput.EchoPassword
			ti.EchoCharacter = '•'
		}
		m.cfgFields[i] = ti
	}
	m.cfgFields[cfgFieldDB].Placeholder = "留空可浏览全部库"
	m.cfgFields[cfgFieldHost].Focus()
	m.cfgActive = cfgFieldHost
}

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

// Init 启动应用。
func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{tea.SetWindowTitle("meteor-ch-tui")}
	if !m.isTTY {
		cmds = append(cmds, func() tea.Msg {
			return tea.WindowSizeMsg{Width: 120, Height: 40}
		})
	}
	if m.quickConnect {
		m.quickConnect = false
		cmds = append(cmds, m.connectCmd())
	}
	return tea.Batch(cmds...)
}

// Update 处理所有消息。
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.appMode == appMain {
			m.layout()
		}
		return m, nil

	case connectedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.status = "连接失败: " + msg.err.Error()
			m.statusOK = false
			m.syncSidebarSelection()
			return m, nil
		}
		if m.client != nil {
			m.client.Close()
		}
		prevDB := m.currentDB
		m.client = msg.client
		m.nodes = msg.nodes
		m.databases = msg.databases
		m.tables = msg.tables
		m.views = msg.views
		m.databasesTruncated = msg.databasesTruncated
		m.objectsTruncated = msg.objectsTruncated
		m.currentDB = msg.database
		if m.currentDB == "" {
			m.currentDB = m.cfg.EffectiveDatabase()
		}
		m.err = nil
		m.curNode = msg.node
		m.nodeIdx = m.indexOfNode(msg.node)
		m.syncSidebarSelection()
		m.status = fmtConnected(m.client.Addr(), len(m.nodes), len(m.databases))
		if prevDB != "" && prevDB != m.currentDB {
			m.status = fmtDBSwitched(m.currentDB, len(m.tables), len(m.views)) + " · " + m.status
		}
		if msg.catalogErr != nil {
			m.status += "（" + msg.catalogErr.Error() + "）"
		}
		m.statusOK = true
		m.layout()
		return m, nil

	case testConnMsg:
		if msg.seq != m.mgrTestSeq {
			return m, nil
		}
		m.mgrTesting = false
		if msg.err != nil {
			m.mgrTestOK = false
			m.mgrTestResult = "测试失败: " + msg.err.Error()
			return m, nil
		}
		m.mgrTestOK = true
		m.mgrTestResult = fmt.Sprintf("测试成功 · 服务端 %s · 耗时 %s", msg.version, msg.elapsed.Round(time.Millisecond))
		return m, nil

	case queryDoneMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.status = "查询失败: " + msg.err.Error()
			m.statusOK = false
			return m, nil
		}
		m.mode = modeResult
		m.table.setResult(msg.res)
		m.status = fmtQueryDone(msg.res)
		m.statusOK = true
		m.focus = focusTable
		m.layout()
		return m, nil

	case slowDoneMsg:
		m.loading = false
		if msg.err != nil {
			m.slowErr = msg.err
			m.status = "读取活跃查询失败: " + msg.err.Error()
			m.statusOK = false
			return m, nil
		}
		m.slowErr = nil
		m.mode = modeSlow
		m.slow.setResult(msg.res)
		m.status = fmtSlowDone(len(msg.res.Rows))
		m.statusOK = true
		m.focus = focusTable
		m.layout()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" {
		if m.appMode == appMain && m.client != nil {
			m.client.Close()
		}
		return m, tea.Quit
	}

	if m.helpOpen {
		return m.handleHelpKey(key)
	}
	if m.confirm.kind != confirmNone {
		return m.handleConfirmKey(key)
	}
	if m.fileBrowserOpen {
		return m.handleFileBrowserKey(msg, key)
	}
	if m.sheetPickerOpen {
		return m.handleSheetPickerKey(key)
	}
	if m.dirBrowserOpen {
		return m.handleDirBrowserKey(key)
	}
	if m.ctxMenuOpen {
		return m.handleContextMenuKey(key)
	}
	if m.historyModal {
		return m.handleHistoryKey(key)
	}
	if m.resultExportModal {
		return m.handleResultExportKey(msg, key)
	}

	if m.appMode == appManager {
		return m.handleMgrKey(msg, key)
	}

	return m.handleMainKey(msg, key)
}

func (m *Model) handleMainKey(msg tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	if m.cfgModal {
		return m.handleCfgModalKey(msg, key)
	}

	if m.rowModal {
		return m.handleRowModalKey(key)
	}

	switch key {
	case "ctrl+n":
		m.cfgModal = true
		m.cfgFields[m.cfgActive].Focus()
		return m, nil
	case "tab":
		return m, m.cycleFocus()
	case "shift+tab":
		return m, m.cycleFocusBack()
	case "ctrl+r":
		if m.client == nil {
			m.status = "尚未连接"
			m.statusOK = false
			return m, nil
		}
		m.loading = true
		m.status = "正在读取活跃查询..."
		return m, m.slowQueryCmd()
	case "ctrl+y":
		return m.copyResultTSV()
	case "ctrl+e":
		m.openResultExportModal()
		return m, nil
	case "ctrl+h":
		m.openHistoryModal()
		return m, nil
	case "esc":
		if m.mode == modeSlow {
			m.mode = modeResult
			m.focus = focusTable
			m.layout()
			m.status = "已返回结果面板"
			m.statusOK = true
			return m, nil
		}
		m.editor.Blur()
		m.focus = focusTable
		return m, nil
	case "ctrl+q":
		m.backToManager()
		return m, nil
	case "?", "f1":
		m.openHelp()
		return m, nil
	}

	switch m.focus {
	case focusSidebar:
		return m.handleSidebarKey(key)
	case focusEditor:
		return m.handleEditorKey(msg, key)
	default:
		return m.handleTableKey(key)
	}
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

// clearModalError 清除弹窗校验留下的错误提示，让状态栏回到默认文案。
// connTestConfig 只取连接相关字段做测试，不要求填写连接名称。
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

// startConnTest 用当前表单内容发起一次异步连接测试，不关闭弹窗。
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

func nextProtocol(cur string) string {
	switch ch.NormalizeProtocol(cur) {
	case ch.ProtocolNative:
		return ch.ProtocolHTTP
	case ch.ProtocolHTTP:
		return ch.ProtocolHTTPS
	default:
		return ch.ProtocolNative
	}
}

// defaultPortFor 返回协议对应的默认端口。
func defaultPortFor(protocol string) string {
	switch ch.NormalizeProtocol(protocol) {
	case ch.ProtocolHTTP:
		return "8123"
	case ch.ProtocolHTTPS:
		return "8443"
	default:
		return "9000"
	}
}

// applyProtocolPort 只在端口为空或仍是某个协议默认端口时联动改写。
// 用户自定义端口（如 28123）原样保留，由 protocolPortNote 给出提示。
func applyProtocolPort(ti *textinput.Model, protocol string) {
	port := strings.TrimSpace(ti.Value())
	switch port {
	case "", "9000", "8123", "8443":
		ti.SetValue(defaultPortFor(protocol))
	}
}

// protocolPortNote 在端口与该协议默认端口不一致时提示，
// 避免从 HTTP(28123) 切回 Native 后忘记改端口，复现握手错误。
func protocolPortNote(protocol, portStr string) string {
	port := strings.TrimSpace(portStr)
	def := defaultPortFor(protocol)
	if port == "" || port == def {
		return ""
	}
	return fmt.Sprintf("  ! 端口 %s（默认 %s）", port, def)
}

func (m *Model) toggleMgrAddProtocol() {
	next := nextProtocol(m.mgrAddProtocol.Value())
	m.mgrAddProtocol.SetValue(next)
	applyProtocolPort(&m.mgrAddPort, next)
}

func (m *Model) toggleCfgProtocol() {
	ti := &m.cfgFields[cfgFieldProtocol]
	next := nextProtocol(ti.Value())
	ti.SetValue(next)
	applyProtocolPort(&m.cfgFields[cfgFieldPort], next)
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
		m.mgrExportPasswords = !m.mgrExportPasswords
		m.refreshMgrExportPreview()
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

// ============ main mode handlers ============

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

// sidebarItemCount 返回当前标签中的可选项数量。
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

// syncSidebarSelection 让选中项对齐当前库 / 当前节点。
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

// activateSidebarItem 执行当前选中项：切换库、预览表/视图或切换节点。
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

func quoteIdent(s string) string {
	return "`" + strings.ReplaceAll(s, "`", "``") + "`"
}

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

func (m *Model) activeTable() *tableState {
	if m.mode == modeSlow {
		return &m.slow
	}
	return &m.table
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

// handleRightClick 在管理器中逐层关弹窗；在主界面打开上下文菜单。
func (m *Model) handleRightClick(x, y int) (tea.Model, tea.Cmd) {
	if m.appMode == appManager {
		if m.mgrGroupModal {
			m.closeGroupModal()
		} else if m.mgrAddModal {
			m.closeMgrAddModal()
		} else if m.mgrImportModal {
			m.closeMgrImportModal()
		}
		return m, nil
	}
	if m.cfgModal {
		m.closeCfgModal()
		return m, nil
	}
	if m.rowModal {
		m.rowModal = false
		return m, nil
	}
	m.openContextMenu(x, y)
	return m, nil
}

func (m *Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	isClick := msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft
	isRight := msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonRight
	isWheelUp := msg.Button == tea.MouseButtonWheelUp
	isWheelDown := msg.Button == tea.MouseButtonWheelDown

	if m.helpOpen {
		return m.handleHelpMouse(msg, isClick, isWheelUp, isWheelDown)
	}
	if m.confirm.kind != confirmNone {
		return m.handleConfirmMouse(msg, isClick)
	}
	if m.fileBrowserOpen {
		return m.handleFileBrowserMouse(msg, isClick, isWheelUp, isWheelDown)
	}
	if m.sheetPickerOpen {
		return m.handleSheetPickerMouse(msg, isClick, isWheelUp, isWheelDown)
	}
	if m.dirBrowserOpen {
		return m.handleDirBrowserMouse(msg, isClick, isWheelUp, isWheelDown)
	}
	if m.historyModal {
		return m.handleHistoryMouse(msg, isClick, isWheelUp, isWheelDown)
	}
	if m.resultExportModal {
		return m.handleResultExportMouse(msg, isClick, isWheelUp, isWheelDown)
	}
	if m.ctxMenuOpen {
		return m.handleContextMenuMouse(msg, isClick, isWheelUp, isWheelDown)
	}
	if isRight {
		return m.handleRightClick(msg.X, msg.Y)
	}
	if !isClick && !isWheelUp && !isWheelDown {
		return m, nil
	}

	if m.appMode == appManager {
		return m.handleMgrMouse(msg, isClick, isWheelUp, isWheelDown)
	}
	return m.handleMainMouse(msg, isClick, isWheelUp, isWheelDown)
}

func (m *Model) isDoubleClick(x, y int) bool {
	now := time.Now().UnixMilli()
	if m.lastClickX == x && m.lastClickY == y && now-m.lastClickTime < 700 {
		m.lastClickX, m.lastClickY, m.lastClickTime = -1, -1, 0
		return true
	}
	m.lastClickX, m.lastClickY, m.lastClickTime = x, y, now
	return false
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

// selectMgrLineAt 把屏幕 Y 换算成列表行，再选中对应连接。
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

func (m *Model) requestDeleteConn() {
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

// handleMgrGroupMouse 处理分组弹窗的点击与滚轮。
// 热区复用：testX/testW = 重命名，okX/okW = 新建，cancelX/cancelW = 删除，extraX/extraW = 取消。
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

func (m *Model) handleMainMouse(msg tea.MouseMsg, isClick, isWheelUp, isWheelDown bool) (tea.Model, tea.Cmd) {
	if m.cfgModal {
		return m.handleCfgModalMouse(msg, isClick)
	}
	if m.rowModal {
		if isWheelUp || isWheelDown {
			delta := 3
			if isWheelUp {
				delta = -3
			}
			total := m.rowModalLines()
			page := m.modalPageSize()
			m.rowScroll = clampInt(m.rowScroll+delta, 0, maxInt(0, total-page))
		}
		return m, nil
	}
	if isClick && msg.Y == 0 && msg.X >= m.backBtnX && msg.X < m.backBtnX+m.backBtnW {
		m.backToManager()
		return m, nil
	}
	if isWheelUp || isWheelDown {
		delta := 3
		if isWheelUp {
			delta = -3
		}
		if msg.X <= m.sidebarWidth() {
			m.focus = focusSidebar
			m.moveSidebar(delta)
			return m, nil
		}
		m.activeTable().moveCursor(delta, m.tableRowsVisible())
		return m, nil
	}
	if !isClick {
		return m, nil
	}

	sidebarW := m.sidebarWidth()
	if msg.X <= sidebarW {
		return m.handleSidebarClick(msg.X, msg.Y)
	}

	editorY1 := 3
	editorY2 := editorPaneHeight + 1
	if msg.Y >= editorY1 && msg.Y < editorY2 {
		m.focus = focusEditor
		return m, m.editor.Focus()
	}

	m.focus = focusTable
	return m, nil
}

// handleSidebarClick 处理侧边栏 tab 切换与条目点击，行号由渲染时记录的热区换算。
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
