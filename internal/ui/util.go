package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/yuwen002/meteor-ch-tui/internal/configstore"
)

// padCell 基于 ansi.Truncate/StringWidth 实现 ANSI 安全的单元格截断与填充。
func padCell(s string, width int) string {
	s = ansi.Truncate(s, width, "~")
	w := ansi.StringWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// modalRect 记录弹窗叠加到屏幕后的实际位置、字段行与可点击区域，供鼠标命中检测使用。
type modalRect struct {
	top, left, w, h int

	fieldStart       int
	fieldCount       int
	listStart        int
	listCount        int
	protocolRow      int
	footerRow        int
	testX, testW     int
	okX, okW         int
	cancelX          int
	cancelW          int
	extraX, extraW   int
	pathRow          int
	pathX, pathW     int
	browseX, browseW int

	filterRow        int
	filterX, filterW int
	sortX, sortW     int
}

// overlayRect 将 box 叠加到 bg 的中央区域，返回合成结果与 box 的实际矩形。
func overlayRect(bg, box string, totalW int) (string, modalRect) {
	bgLines := strings.Split(bg, "\n")
	boxLines := strings.Split(box, "\n")
	boxW := ansi.StringWidth(boxLines[0])
	top := (len(bgLines) - len(boxLines)) / 2
	if top < 0 {
		top = 0
	}
	left := (totalW - boxW) / 2
	if left < 0 {
		left = 0
	}
	for i, bl := range boxLines {
		row := top + i
		if row >= len(bgLines) {
			break
		}
		line := bgLines[row]
		prefix := ansi.Truncate(line, left, "")
		suffix := ansi.TruncateLeft(line, left+boxW, "")
		bgLines[row] = prefix + bl + suffix
	}
	return strings.Join(bgLines, "\n"), modalRect{top: top, left: left, w: boxW, h: len(boxLines)}
}

// overlayAt 将 box 叠加到 bg 的中央区域，返回合成后的字符串。
func overlayAt(bg, box string, totalW int) string {
	out, _ := overlayRect(bg, box, totalW)
	return out
}

// overlayRectAt 将 box 叠加到 bg 的指定位置，并自动避开屏幕边缘。
func overlayRectAt(bg, box string, totalW, left, top int) (string, modalRect) {
	bgLines := strings.Split(bg, "\n")
	boxLines := strings.Split(box, "\n")
	boxW := ansi.StringWidth(boxLines[0])
	if left < 0 {
		left = 0
	}
	if left+boxW > totalW {
		left = maxInt(0, totalW-boxW)
	}
	if top < 0 {
		top = 0
	}
	if top+len(boxLines) > len(bgLines) {
		top = maxInt(0, len(bgLines)-len(boxLines))
	}
	for i, bl := range boxLines {
		row := top + i
		if row >= len(bgLines) {
			break
		}
		line := bgLines[row]
		prefix := ansi.Truncate(line, left, "")
		suffix := ansi.TruncateLeft(line, left+boxW, "")
		bgLines[row] = prefix + bl + suffix
	}
	return strings.Join(bgLines, "\n"), modalRect{top: top, left: left, w: boxW, h: len(boxLines)}
}

// clipLines 仅保留前 n 行。
func clipLines(s string, n int) string {
	if n <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}

// padToHeight 将内容填充/裁剪到恰好 h 行宽 w。
func padToHeight(s string, h, w int) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = padCell(lines[i], w)
	}
	for len(lines) < h {
		lines = append(lines, strings.Repeat(" ", w))
	}
	if len(lines) > h {
		lines = lines[:h]
	}
	return strings.Join(lines, "\n")
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clampInt(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

// runeWidth 返回字符串的显示宽度（含全角/emoji）。
func runeWidth(s string) int {
	return lipgloss.Width(s)
}

func (m *Model) persistUIState() {
	_ = configstore.SaveUIState(m.uiState)
}

func (m *Model) persistDirs() {
	m.uiState.LastExportDir = m.lastExportDir
	m.uiState.LastImportDir = m.lastImportDir
	m.persistUIState()
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

func (m *Model) isDoubleClick(x, y int) bool {
	now := time.Now().UnixMilli()
	if m.lastClickX == x && m.lastClickY == y && now-m.lastClickTime < 700 {
		m.lastClickX, m.lastClickY, m.lastClickTime = -1, -1, 0
		return true
	}
	m.lastClickX, m.lastClickY, m.lastClickTime = x, y, now
	return false
}
