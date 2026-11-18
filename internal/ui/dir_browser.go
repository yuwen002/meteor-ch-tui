package ui

import (
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m *Model) openDirBrowser() {
	start := strings.TrimSpace(m.exportPathInput.Value())
	if start == "" {
		start = m.lastExportDir
	}
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
	m.dirBrowserPath = filepath.Clean(start)
	m.dirBrowserSel = 0
	m.dirBrowserOpen = true
	m.loadDirBrowser()
}

func (m *Model) closeDirBrowser() {
	m.dirBrowserOpen = false
	m.dirBrowserRect = modalRect{}
}

func (m *Model) loadDirBrowser() {
	entries := []string{".."}
	if dirs, err := os.ReadDir(m.dirBrowserPath); err == nil {
		for _, d := range dirs {
			if d.IsDir() {
				entries = append(entries, d.Name())
			}
		}
	}
	m.dirBrowserEntries = entries
	m.dirBrowserSel = clampInt(m.dirBrowserSel, 0, len(entries)-1)
}

func (m *Model) moveDirBrowser(delta int) {
	if len(m.dirBrowserEntries) == 0 {
		return
	}
	m.dirBrowserSel = clampInt(m.dirBrowserSel+delta, 0, len(m.dirBrowserEntries)-1)
}

func (m *Model) enterDirBrowser() {
	if m.dirBrowserSel < 0 || m.dirBrowserSel >= len(m.dirBrowserEntries) {
		return
	}
	entry := m.dirBrowserEntries[m.dirBrowserSel]
	if entry == ".." {
		m.parentDirBrowser()
		return
	}
	m.dirBrowserPath = filepath.Join(m.dirBrowserPath, entry)
	m.dirBrowserSel = 0
	m.loadDirBrowser()
}

func (m *Model) parentDirBrowser() {
	parent := filepath.Dir(m.dirBrowserPath)
	if parent == m.dirBrowserPath {
		return
	}
	m.dirBrowserPath = parent
	m.dirBrowserSel = 0
	m.loadDirBrowser()
}

func (m *Model) selectDirBrowser() (tea.Model, tea.Cmd) {
	path := m.dirBrowserPath
	m.exportPathInput.SetValue(path)
	m.setLastExportDir(path)
	m.closeDirBrowser()
	m.blurExportPath()
	m.status = "已选择保存目录: " + path
	m.statusOK = true
	return m, nil
}

func (m *Model) handleDirBrowserKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.closeDirBrowser()
	case "up", "k":
		m.moveDirBrowser(-1)
	case "down", "j":
		m.moveDirBrowser(1)
	case "enter", "right", "l":
		m.enterDirBrowser()
	case "left", "h", "backspace":
		m.parentDirBrowser()
	case "ctrl+s":
		return m.selectDirBrowser()
	}
	return m, nil
}

func (m *Model) handleDirBrowserMouse(msg tea.MouseMsg, isClick, isWheelUp, isWheelDown bool) (tea.Model, tea.Cmd) {
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonRight {
		m.closeDirBrowser()
		return m, nil
	}
	if isWheelUp || isWheelDown {
		if isWheelUp {
			m.moveDirBrowser(-1)
		} else {
			m.moveDirBrowser(1)
		}
		return m, nil
	}
	if !isClick {
		return m, nil
	}
	r := m.dirBrowserRect
	if r.w == 0 || r.h == 0 || msg.X < r.left || msg.X >= r.left+r.w || msg.Y < r.top || msg.Y >= r.top+r.h {
		return m, nil
	}
	relX, relY := msg.X-r.left, msg.Y-r.top
	if role, ok := hitFooter(relX, relY, r, footerRoleOK, footerRoleExtra, footerRoleCancel); ok {
		switch role {
		case footerRoleOK:
			return m.selectDirBrowser()
		case footerRoleExtra:
			m.parentDirBrowser()
		case footerRoleCancel:
			m.closeDirBrowser()
		}
		return m, nil
	}
	idx := relY - r.listStart
	if idx >= 0 && idx < r.listCount {
		m.dirBrowserSel = idx
		m.enterDirBrowser()
	}
	return m, nil
}

func (m *Model) renderDirBrowser(bg string) string {
	innerW := 76
	if innerW > m.width-8 {
		innerW = maxInt(36, m.width-8)
	}
	boxW := innerW + 2

	title := styleTitle.Render(" 选择保存目录 ")
	hint := styleMuted.Render(" ↑↓ · Enter 进入 · Ctrl+S 选择 · Esc 取消 ")
	lines := make([]string, 0, len(m.dirBrowserEntries)+6)
	lines = append(lines, padCell(title, maxInt(0, innerW-lipgloss.Width(hint)))+hint)
	lines = append(lines, "")
	pathText := ansi.Truncate(m.dirBrowserPath, innerW-2, "~")
	lines = append(lines, styleOkText.Render(padCell(" "+pathText, innerW)))
	lines = append(lines, "")

	listStart := len(lines) + 1
	if len(m.dirBrowserEntries) == 0 {
		lines = append(lines, styleMuted.Render(padCell(" (无子目录)", innerW)))
	}
	for i, entry := range m.dirBrowserEntries {
		icon := "📁 "
		if entry == ".." {
			icon = ".. "
		}
		text := padCell("   "+icon+entry, innerW)
		if i == m.dirBrowserSel {
			lines = append(lines, styleRowSelected.Render(text))
		} else {
			lines = append(lines, text)
		}
	}
	listCount := len(m.dirBrowserEntries)
	lines = append(lines, "")

	dirBtns := []footerButton{
		{label: styleBtnOK.Render("[Ctrl+S] 选择当前"), role: footerRoleOK},
		{label: styleBtnMuted.Render("[Backspace] 上级"), role: footerRoleExtra},
		{label: styleBtnMuted.Render("[Esc] 取消"), role: footerRoleCancel},
	}
	footer, footerPositions := footerLayout(innerW, footerAlignFirstLeftRestRight, dirBtns)
	footerContentRow := len(lines)
	lines = append(lines, footer)

	content := strings.Join(lines, "\n")
	box := styleModal.Width(innerW).Height(len(lines)).Render(content)
	out, rect := overlayRect(bg, box, m.width)
	rect.listStart = listStart
	rect.listCount = listCount
	applyFooter(&rect, footerContentRow, dirBtns, footerPositions)
	rect.w = boxW
	m.dirBrowserRect = rect
	return out
}
