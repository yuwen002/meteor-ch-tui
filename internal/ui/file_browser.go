package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/yuwen002/meteor-ch-tui/internal/configstore"
)

func isImportFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".json" || ext == ".csv" || ext == ".tsv" || ext == ".xlsx"
}

func (m *Model) openFileBrowser() {
	start := m.lastImportDir
	if start == "" {
		if cwd, err := os.Getwd(); err == nil {
			start = cwd
		}
	}
	if expanded, err := expandPathInput(start); err == nil && expanded != "" {
		start = expanded
	}
	if info, err := os.Stat(start); err != nil || !info.IsDir() {
		if cwd, err := os.Getwd(); err == nil {
			start = cwd
		}
	}
	m.fileBrowserPath = filepath.Clean(start)
	m.fileBrowserSel = 0
	m.fileBrowserOpen = true
	ti := textinput.New()
	ti.Placeholder = "搜索文件或目录..."
	ti.CharLimit = 0
	ti.Width = 32
	m.fileBrowserFilter = ti
	m.fileBrowserFilterFocused = false
	m.loadFileBrowser()
}

func (m *Model) closeFileBrowser() {
	m.fileBrowserOpen = false
	m.fileBrowserRect = modalRect{}
}

func (m *Model) focusFileBrowserFilter() tea.Cmd {
	m.fileBrowserFilterFocused = true
	return m.fileBrowserFilter.Focus()
}

func (m *Model) blurFileBrowserFilter() {
	m.fileBrowserFilterFocused = false
	m.fileBrowserFilter.Blur()
}

func (m *Model) toggleFileBrowserSort() {
	m.setFileBrowserSortMode((m.fileBrowserSortMode + 1) % 2)
	m.fileBrowserSel = 0
	m.loadFileBrowser()
}

func (m *Model) loadFileBrowser() {
	dirItems := []fileBrowserItem{}
	fileItems := []fileBrowserItem{}
	if dirs, err := os.ReadDir(m.fileBrowserPath); err == nil {
		for _, d := range dirs {
			var modTime time.Time
			if info, err := d.Info(); err == nil {
				modTime = info.ModTime()
			}
			if d.IsDir() {
				dirItems = append(dirItems, fileBrowserItem{name: d.Name(), isDir: true, modTime: modTime})
			} else if isImportFile(d.Name()) {
				fileItems = append(fileItems, fileBrowserItem{name: d.Name(), modTime: modTime})
			}
		}
	}

	filter := strings.ToLower(strings.TrimSpace(m.fileBrowserFilter.Value()))
	match := func(name string) bool {
		return filter == "" || strings.Contains(strings.ToLower(name), filter)
	}
	byName := func(items []fileBrowserItem) {
		sort.Slice(items, func(i, j int) bool {
			return strings.ToLower(items[i].name) < strings.ToLower(items[j].name)
		})
	}
	byTime := func(items []fileBrowserItem) {
		sort.Slice(items, func(i, j int) bool {
			if items[i].modTime.Equal(items[j].modTime) {
				return strings.ToLower(items[i].name) < strings.ToLower(items[j].name)
			}
			return items[i].modTime.After(items[j].modTime)
		})
	}
	sortItems := byName
	if m.fileBrowserSortMode == 1 {
		sortItems = byTime
	}

	filteredDirs := make([]fileBrowserItem, 0, len(dirItems))
	for _, item := range dirItems {
		if match(item.name) {
			filteredDirs = append(filteredDirs, item)
		}
	}
	filteredFiles := make([]fileBrowserItem, 0, len(fileItems))
	for _, item := range fileItems {
		if match(item.name) {
			filteredFiles = append(filteredFiles, item)
		}
	}
	sortItems(filteredDirs)
	sortItems(filteredFiles)

	items := append([]fileBrowserItem{{name: "..", isDir: true}}, filteredDirs...)
	items = append(items, filteredFiles...)
	m.fileBrowserItems = items
	m.fileBrowserSel = clampInt(m.fileBrowserSel, 0, len(items)-1)
}

func (m *Model) moveFileBrowser(delta int) {
	if len(m.fileBrowserItems) == 0 {
		return
	}
	m.fileBrowserSel = clampInt(m.fileBrowserSel+delta, 0, len(m.fileBrowserItems)-1)
}

func (m *Model) enterFileBrowser() {
	if m.fileBrowserSel < 0 || m.fileBrowserSel >= len(m.fileBrowserItems) {
		return
	}
	item := m.fileBrowserItems[m.fileBrowserSel]
	if item.isDir {
		if item.name == ".." {
			parent := filepath.Dir(m.fileBrowserPath)
			if parent != m.fileBrowserPath {
				m.fileBrowserPath = parent
			}
		} else {
			m.fileBrowserPath = filepath.Join(m.fileBrowserPath, item.name)
		}
		m.fileBrowserSel = 0
		m.loadFileBrowser()
		return
	}
	m.selectFileBrowser()
}

func (m *Model) parentFileBrowser() {
	parent := filepath.Dir(m.fileBrowserPath)
	if parent == m.fileBrowserPath {
		return
	}
	m.fileBrowserPath = parent
	m.fileBrowserSel = 0
	m.loadFileBrowser()
}

func (m *Model) setImportFromXLSX(path, sheet string) error {
	cfg, err := configstore.ImportXLSXSheet(path, sheet)
	if err != nil {
		return err
	}
	data, err := cfg.ExportJSON(true)
	if err != nil {
		return err
	}
	m.mgrImportText.SetValue(data)
	return nil
}

func (m *Model) selectFileBrowser() (tea.Model, tea.Cmd) {
	if m.fileBrowserSel < 0 || m.fileBrowserSel >= len(m.fileBrowserItems) {
		return m, nil
	}
	item := m.fileBrowserItems[m.fileBrowserSel]
	if item.isDir {
		m.enterFileBrowser()
		return m, nil
	}
	path := filepath.Join(m.fileBrowserPath, item.name)
	m.setLastImportDir(m.fileBrowserPath)
	m.closeFileBrowser()

	if strings.EqualFold(filepath.Ext(path), ".xlsx") {
		sheets, err := configstore.XLSXSheets(path)
		if err != nil {
			m.status = "读取文件失败: " + err.Error()
			m.statusOK = false
			return m, nil
		}
		if len(sheets) > 1 {
			m.openSheetPicker(path, sheets)
			m.status = "请选择要导入的工作表"
			m.statusOK = true
			return m, nil
		}
		if err := m.setImportFromXLSX(path, sheets[0]); err != nil {
			m.status = "读取文件失败: " + err.Error()
			m.statusOK = false
			return m, nil
		}
		m.status = "已载入文件: " + path
		m.statusOK = true
		return m, nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		m.status = "读取文件失败: " + err.Error()
		m.statusOK = false
		return m, nil
	}
	m.mgrImportText.SetValue(string(b))
	m.status = "已载入文件: " + path
	m.statusOK = true
	return m, nil
}

func (m *Model) handleFileBrowserKey(msg tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	if m.fileBrowserFilterFocused {
		switch key {
		case "esc", "enter":
			m.blurFileBrowserFilter()
			return m, nil
		case "ctrl+f":
			m.blurFileBrowserFilter()
			return m, nil
		}
		var cmd tea.Cmd
		m.fileBrowserFilter, cmd = m.fileBrowserFilter.Update(msg)
		m.fileBrowserSel = 0
		m.loadFileBrowser()
		return m, cmd
	}
	switch key {
	case "esc":
		m.closeFileBrowser()
	case "/", "ctrl+f":
		return m, m.focusFileBrowserFilter()
	case "s":
		m.toggleFileBrowserSort()
	case "up", "k":
		m.moveFileBrowser(-1)
	case "down", "j":
		m.moveFileBrowser(1)
	case "enter", "right", "l":
		return m.selectFileBrowser()
	case "left", "h", "backspace":
		m.parentFileBrowser()
	}
	return m, nil
}

func (m *Model) handleFileBrowserMouse(msg tea.MouseMsg, isClick, isWheelUp, isWheelDown bool) (tea.Model, tea.Cmd) {
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonRight {
		m.closeFileBrowser()
		return m, nil
	}
	if isWheelUp || isWheelDown {
		if isWheelUp {
			m.moveFileBrowser(-1)
		} else {
			m.moveFileBrowser(1)
		}
		return m, nil
	}
	if !isClick {
		return m, nil
	}
	r := m.fileBrowserRect
	if r.w == 0 || r.h == 0 || msg.X < r.left || msg.X >= r.left+r.w || msg.Y < r.top || msg.Y >= r.top+r.h {
		return m, nil
	}
	relX, relY := msg.X-r.left, msg.Y-r.top
	if r.filterRow > 0 && relY == r.filterRow {
		if r.filterW > 0 && relX >= r.filterX && relX < r.filterX+r.filterW {
			return m, m.focusFileBrowserFilter()
		}
		if r.sortW > 0 && relX >= r.sortX && relX < r.sortX+r.sortW {
			m.toggleFileBrowserSort()
			return m, nil
		}
		return m, nil
	}
	if role, ok := hitFooter(relX, relY, r, footerRoleOK, footerRoleExtra, footerRoleCancel); ok {
		switch role {
		case footerRoleOK:
			return m.selectFileBrowser()
		case footerRoleExtra:
			m.parentFileBrowser()
		case footerRoleCancel:
			m.closeFileBrowser()
		}
		return m, nil
	}
	idx := relY - r.listStart
	if idx >= 0 && idx < r.listCount {
		m.fileBrowserSel = idx
		return m.selectFileBrowser()
	}
	return m, nil
}

func (m *Model) renderFileBrowser(bg string) string {
	innerW := 76
	if innerW > m.width-8 {
		innerW = maxInt(36, m.width-8)
	}

	title := styleTitle.Render(" 选择导入文件 ")
	hint := styleMuted.Render(" ↑↓ · / 搜索 · s 排序 · Enter 选择 · Esc 取消 ")
	lines := make([]string, 0, len(m.fileBrowserItems)+8)
	lines = append(lines, padCell(title, maxInt(0, innerW-lipgloss.Width(hint)))+hint)
	lines = append(lines, "")
	pathText := ansi.Truncate(m.fileBrowserPath, innerW-2, "~")
	lines = append(lines, styleOkText.Render(padCell(" "+pathText, innerW)))
	lines = append(lines, "")

	filterMark := "  "
	if m.fileBrowserFilterFocused {
		filterMark = "> "
	}
	filterLabel := filterMark + "搜索: "
	sortLabel := "[S] 名称"
	if m.fileBrowserSortMode == 1 {
		sortLabel = "[S] 时间"
	}
	inputW := maxInt(10, innerW-lipgloss.Width(filterLabel)-lipgloss.Width(sortLabel)-3)
	m.fileBrowserFilter.Width = inputW
	filterView := m.fileBrowserFilter.View()
	filterLine := filterLabel + filterView + " " + sortLabel
	filterContentRow := len(lines)
	lines = append(lines, padCell(filterLine, innerW))
	lines = append(lines, "")
	filterRow := filterContentRow + 1
	filterX := lipgloss.Width(filterLabel)
	filterW := lipgloss.Width(filterView)
	sortX := filterX + filterW + 1
	sortW := lipgloss.Width(sortLabel)

	listStart := len(lines) + 1
	if len(m.fileBrowserItems) == 0 {
		lines = append(lines, styleMuted.Render(padCell(" (无可用文件)", innerW)))
	}
	for i, item := range m.fileBrowserItems {
		icon := "📄 "
		if item.isDir {
			icon = "📁 "
			if item.name == ".." {
				icon = ".. "
			}
		}
		text := padCell("   "+icon+item.name, innerW)
		if i == m.fileBrowserSel {
			lines = append(lines, styleRowSelected.Render(text))
		} else {
			lines = append(lines, text)
		}
	}
	listCount := len(m.fileBrowserItems)
	lines = append(lines, "")

	fileBtns := []footerButton{
		{label: styleBtnOK.Render("[Enter] 选择"), role: footerRoleOK},
		{label: styleBtnMuted.Render("[Backspace] 上级"), role: footerRoleExtra},
		{label: styleBtnMuted.Render("[Esc] 取消"), role: footerRoleCancel},
	}
	footer, footerPositions := footerLayout(innerW, footerAlignFirstLeftRestRight, fileBtns)
	footerContentRow := len(lines)
	lines = append(lines, footer)

	content := strings.Join(lines, "\n")
	box := styleModal.Width(innerW).Height(len(lines)).Render(content)
	out, rect := overlayRect(bg, box, m.width)
	rect.listStart = listStart
	rect.listCount = listCount
	rect.filterRow = filterRow
	rect.filterX = filterX
	rect.filterW = filterW
	rect.sortX = sortX
	rect.sortW = sortW
	applyFooter(&rect, footerContentRow, fileBtns, footerPositions)
	m.fileBrowserRect = rect
	return out
}
