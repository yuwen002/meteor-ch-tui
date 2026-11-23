package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
)

func (m *Model) tableView() string {
	innerW := m.rightWidth() - 2
	innerH := m.tablePaneHeight() - 2

	var content string
	title := styleTitle.Render("结果")
	if m.mode == modeSlow {
		title = styleTitle.Render("活跃查询")
	}
	t := m.activeTable()
	if t.res == nil {
		msg := styleMuted.Render("尚无结果，输入 SQL 后按 Ctrl+S 或 F5 执行")
		if m.mode == modeSlow && m.slowErr != nil {
			msg = styleErrText.Render("读取失败: " + m.slowErr.Error())
		}
		content = padCell(title, innerW) + "\n" + padCell(msg, innerW)
	} else {
		content = m.renderTable(t, innerW, innerH, title)
	}
	content = padToHeight(content, innerH, innerW)

	style := stylePane
	if m.focus == focusTable {
		style = stylePaneFocus
	}
	return style.Width(innerW).Height(innerH).Render(content)
}

func (m *Model) renderTable(t *tableState, innerW, innerH int, title string) string {
	rowsVisible := innerH - 2
	if rowsVisible < 1 {
		rowsVisible = 1
	}

	idx := t.visibleColumns(innerW)
	meta := styleMuted.Render(fmt.Sprintf(" %d 行 · 列 %d/%d · 起始列 %d",
		len(t.res.Rows), len(idx), len(t.res.Columns), t.startCol+1))
	titleLine := padCell(title, maxInt(0, innerW-lipgloss.Width(meta))) + meta

	headerLine := styleTableHeader.Render(t.headerLine(idx))
	lines := make([]string, 0, rowsVisible+2)
	lines = append(lines, padCell(titleLine, innerW))
	lines = append(lines, headerLine)

	for i := 0; i < rowsVisible; i++ {
		r := t.topRow + i
		if r >= len(t.res.Rows) {
			lines = append(lines, strings.Repeat(" ", innerW))
			continue
		}
		line := t.rowLine(r, idx)
		if r == t.cursor {
			lines = append(lines, styleRowSelected.Width(innerW).Render(padCell(line, innerW)))
		} else {
			lines = append(lines, padCell(line, innerW))
		}
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderRowModal(bg string) string {
	t := &m.table
	if t.res == nil || t.cursor >= len(t.res.Rows) {
		return bg
	}

	boxW := m.width - 8
	if boxW < 30 {
		boxW = m.width - 4
	}
	if boxW < 20 {
		boxW = 20
	}
	innerW := boxW - 2
	page := m.modalPageSize()
	innerH := page

	row := t.res.Rows[t.cursor]
	title := styleTitle.Render(fmt.Sprintf(" 行 #%d · %d 列", t.cursor+1, len(t.res.Columns)))
	hint := styleMuted.Render(" ↑↓ 滚动 · Esc/v 关闭 ")
	titleLine := padCell(title, maxInt(0, innerW-lipgloss.Width(hint))) + hint

	keyW := 0
	for _, c := range t.res.Columns {
		if w := runeWidth(c); w > keyW {
			keyW = w
		}
	}
	keyW = clampInt(keyW, 6, 24)

	lines := make([]string, 0, len(t.res.Columns))
	for i, c := range t.res.Columns {
		val := ""
		if i < len(row) {
			val = ch.FormatValue(row[i])
		}
		key := styleMuted.Render(padCell(c, keyW))
		lines = append(lines, key+" "+styleOkText.Render(val))
	}

	start := m.rowScroll
	if start > len(lines) {
		start = len(lines)
	}
	end := start + innerH
	if end > len(lines) {
		end = len(lines)
	}
	visible := lines[start:end]

	contentLines := make([]string, 0, innerH+2)
	contentLines = append(contentLines, padCell(titleLine, innerW))
	contentLines = append(contentLines, "")
	for i := 0; i < innerH; i++ {
		if i < len(visible) {
			contentLines = append(contentLines, visible[i])
		} else {
			contentLines = append(contentLines, "")
		}
	}
	content := padToHeight(strings.Join(contentLines, "\n"), innerH+2, innerW)

	box := styleModal.Width(innerW).Height(innerH + 2).Render(content)
	return overlayAt(bg, box, m.width)
}
