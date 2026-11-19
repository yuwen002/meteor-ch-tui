package ui

import (
	"fmt"
	"strconv"
	"time"

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
	uiState                  configstore.UIState

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

	columnModalOpen bool
	columnModalSel  int
	columnModalRect modalRect

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
		m.uiState = state
		m.lastExportDir = state.LastExportDir
		m.lastImportDir = state.LastImportDir
		m.fileBrowserSortMode = state.FileBrowserSortMode
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
		m.applyResultLayout(&m.table)
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
		m.applyResultLayout(&m.slow)
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
	if m.columnModalOpen {
		return m.handleColumnModalKey(key)
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

// clearModalError 清除弹窗校验留下的错误提示，让状态栏回到默认文案。
// connTestConfig 只取连接相关字段做测试，不要求填写连接名称。

// startConnTest 用当前表单内容发起一次异步连接测试，不关闭弹窗。

// defaultPortFor 返回协议对应的默认端口。

// applyProtocolPort 只在端口为空或仍是某个协议默认端口时联动改写。
// 用户自定义端口（如 28123）原样保留，由 protocolPortNote 给出提示。

// protocolPortNote 在端口与该协议默认端口不一致时提示，
// 避免从 HTTP(28123) 切回 Native 后忘记改端口，复现握手错误。

// ============ main mode handlers ============

// sidebarItemCount 返回当前标签中的可选项数量。

// syncSidebarSelection 让选中项对齐当前库 / 当前节点。

// activateSidebarItem 执行当前选中项：切换库、预览表/视图或切换节点。

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
	if m.columnModalOpen {
		return m.handleColumnModalMouse(msg, isClick, isWheelUp, isWheelDown)
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

// selectMgrLineAt 把屏幕 Y 换算成列表行，再选中对应连接。

// handleMgrGroupMouse 处理分组弹窗的点击与滚轮。
// 热区复用：testX/testW = 重命名，okX/okW = 新建，cancelX/cancelW = 删除，extraX/extraW = 取消。

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
