package ui

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
)

type resultExportFormat int

const (
	exportTSV resultExportFormat = iota
	exportCSV
	exportJSON
	exportXLSX
)

type resultExportChoice struct {
	label  string
	ext    string
	desc   string
	format resultExportFormat
}

var resultExportChoices = []resultExportChoice{
	{label: "TSV", ext: "tsv", desc: "适合粘贴到表格软件", format: exportTSV},
	{label: "CSV", ext: "csv", desc: "标准逗号分隔文件", format: exportCSV},
	{label: "JSON", ext: "json", desc: "每行一个对象，保留类型", format: exportJSON},
	{label: "Excel", ext: "xlsx", desc: "带表头样式的 Excel 工作簿", format: exportXLSX},
}

func tsvValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ch.FormatValue(v)
}

func escapeTSV(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\t", "\\t")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func (m *Model) resultRows() [][]any {
	res := m.table.res
	if res == nil {
		return nil
	}
	rows := make([][]any, 0, len(res.Rows))
	for _, src := range res.Rows {
		row := make([]any, len(res.Columns))
		for i := range row {
			if i < len(src) {
				row[i] = src[i]
			}
		}
		rows = append(rows, row)
	}
	return rows
}

func (m *Model) resultTSV() (string, error) {
	if m.table.res == nil {
		return "", fmt.Errorf("当前没有查询结果")
	}
	var b strings.Builder
	for i, col := range m.table.res.Columns {
		if i > 0 {
			b.WriteByte('\t')
		}
		b.WriteString(escapeTSV(col))
	}
	b.WriteByte('\n')
	for _, row := range m.resultRows() {
		for i := range row {
			if i > 0 {
				b.WriteByte('\t')
			}
			b.WriteString(escapeTSV(tsvValue(row[i])))
		}
		b.WriteByte('\n')
	}
	return b.String(), nil
}

func (m *Model) resultCSV() (string, error) {
	if m.table.res == nil {
		return "", fmt.Errorf("当前没有查询结果")
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(m.table.res.Columns); err != nil {
		return "", err
	}
	for _, row := range m.resultRows() {
		values := make([]string, len(row))
		for i, v := range row {
			values[i] = tsvValue(v)
		}
		if err := w.Write(values); err != nil {
			return "", err
		}
	}
	w.Flush()
	return buf.String(), w.Error()
}

func jsonValue(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case string:
		return x
	case []byte:
		return string(x)
	case bool:
		return x
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return x
	default:
		return fmt.Sprint(x)
	}
}

func (m *Model) resultJSON() (string, error) {
	if m.table.res == nil {
		return "", fmt.Errorf("当前没有查询结果")
	}
	rows := m.resultRows()
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		obj := make(map[string]any, len(m.table.res.Columns))
		for i, col := range m.table.res.Columns {
			obj[col] = jsonValue(row[i])
		}
		out = append(out, obj)
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func xlsxValue(v any) any {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case []byte:
		return string(x)
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return x
	default:
		return fmt.Sprint(x)
	}
}

func (m *Model) resultXLSX(path string) error {
	if m.table.res == nil {
		return fmt.Errorf("当前没有查询结果")
	}
	return writeXLSX("查询结果", m.table.res.Columns, m.resultRows(), path)
}

func (m *Model) copyResultTSV() (tea.Model, tea.Cmd) {
	data, err := m.resultTSV()
	if err != nil {
		m.status = "复制失败: " + err.Error()
		m.statusOK = false
		return m, nil
	}
	if err := clipboard.WriteAll(data); err != nil {
		m.status = "复制失败: " + err.Error()
		m.statusOK = false
		return m, nil
	}
	m.status = fmt.Sprintf("已复制 %d 行 / %d 列查询结果", len(m.table.res.Rows), len(m.table.res.Columns))
	m.statusOK = true
	return m, nil
}

func (m *Model) openResultExportModal() {
	if m.table.res == nil {
		m.status = "当前没有查询结果，无法导出"
		m.statusOK = false
		return
	}
	m.resultExportModal = true
	m.resultExportSel = 0
	m.initExportPathInput()
}

func (m *Model) closeResultExportModal() {
	m.resultExportModal = false
	m.resultExportRect = modalRect{}
	m.blurExportPath()
}

func (m *Model) moveResultExport(delta int) {
	m.resultExportSel = clampInt(m.resultExportSel+delta, 0, len(resultExportChoices)-1)
}

func (m *Model) executeResultExport() (tea.Model, tea.Cmd) {
	if m.resultExportSel < 0 || m.resultExportSel >= len(resultExportChoices) {
		m.closeResultExportModal()
		return m, nil
	}
	choice := resultExportChoices[m.resultExportSel]
	name := "query_result_" + time.Now().Format("20060102_150405") + "." + choice.ext
	path, err := m.exportFilePath(name)
	if err != nil {
		m.status = "导出失败: " + err.Error()
		m.statusOK = false
		return m, nil
	}
	switch choice.format {
	case exportTSV:
		var data string
		data, err = m.resultTSV()
		if err == nil {
			err = os.WriteFile(path, []byte(data), 0644)
		}
	case exportCSV:
		var data string
		data, err = m.resultCSV()
		if err == nil {
			err = os.WriteFile(path, []byte(data), 0644)
		}
	case exportJSON:
		var data string
		data, err = m.resultJSON()
		if err == nil {
			err = os.WriteFile(path, []byte(data), 0644)
		}
	case exportXLSX:
		err = m.resultXLSX(path)
	}
	m.closeResultExportModal()
	if err != nil {
		m.status = "导出失败: " + err.Error()
		m.statusOK = false
		return m, nil
	}
	m.status = fmt.Sprintf("已导出 %d 行到 %s (%s)", len(m.table.res.Rows), path, strings.ToUpper(choice.ext))
	m.statusOK = true
	return m, nil
}

func (m *Model) exportResultTSV() (tea.Model, tea.Cmd) {
	m.resultExportSel = 0
	return m.executeResultExport()
}

func (m *Model) handleResultExportKey(msg tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	if m.exportPathFocused {
		switch key {
		case "esc", "enter":
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
		m.closeResultExportModal()
	case "ctrl+p":
		return m, m.focusExportPath()
	case "ctrl+b":
		m.openDirBrowser()
		return m, nil
	case "up", "k":
		m.moveResultExport(-1)
	case "down", "j":
		m.moveResultExport(1)
	case "enter":
		return m.executeResultExport()
	}
	return m, nil
}

func (m *Model) handleResultExportMouse(msg tea.MouseMsg, isClick, isWheelUp, isWheelDown bool) (tea.Model, tea.Cmd) {
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonRight {
		m.closeResultExportModal()
		return m, nil
	}
	if isWheelUp || isWheelDown {
		if isWheelUp {
			m.moveResultExport(-1)
		} else {
			m.moveResultExport(1)
		}
		return m, nil
	}
	if !isClick {
		return m, nil
	}
	r := m.resultExportRect
	if r.w == 0 || r.h == 0 || msg.X < r.left || msg.X >= r.left+r.w || msg.Y < r.top || msg.Y >= r.top+r.h {
		return m, nil
	}
	relX, relY := msg.X-r.left, msg.Y-r.top
	if r.browseW > 0 && relY == r.pathRow &&
		relX >= r.browseX && relX < r.browseX+r.browseW {
		m.blurExportPath()
		m.openDirBrowser()
		return m, nil
	}
	if r.pathRow > 0 && relY == r.pathRow &&
		r.pathW > 0 && relX >= r.pathX && relX < r.pathX+r.pathW {
		return m, m.focusExportPath()
	}
	if role, ok := hitFooter(relX, relY, r, footerRoleOK, footerRoleCancel); ok {
		switch role {
		case footerRoleOK:
			return m.executeResultExport()
		case footerRoleCancel:
			m.closeResultExportModal()
		}
		return m, nil
	}
	idx := relY - r.listStart
	if idx >= 0 && idx < r.listCount {
		m.resultExportSel = idx
	}
	return m, nil
}
