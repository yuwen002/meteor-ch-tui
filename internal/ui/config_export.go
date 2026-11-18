package ui

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuwen002/meteor-ch-tui/internal/configstore"
)

type configExportFormat int

const (
	configExportJSON configExportFormat = iota
	configExportCSV
	configExportTSV
	configExportXLSX
)

type configExportChoice struct {
	label string
	ext   string
}

var configExportChoices = []configExportChoice{
	{label: "JSON", ext: "json"},
	{label: "CSV", ext: "csv"},
	{label: "TSV", ext: "tsv"},
	{label: "Excel", ext: "xlsx"},
}

func (m *Model) cycleMgrExportFormat(delta int) {
	n := len(configExportChoices)
	m.mgrExportFormat = configExportFormat((int(m.mgrExportFormat) + delta%n + n) % n)
	m.refreshMgrExportPreview()
}

func (m *Model) refreshMgrExportPreview() {
	if !m.mgrExportMode {
		return
	}
	data, err := m.configExportText()
	if err != nil {
		data = "Excel 格式请按 Enter 保存为文件"
	}
	m.mgrImportText.SetValue(data)
}

func (m *Model) connectionExportColumns() []string {
	return []string{"名称", "分组", "协议", "主机", "端口", "用户名", "密码", "数据库"}
}

func (m *Model) connectionPasswordForExport() string {
	if m.mgrExportPasswords {
		return ""
	}
	return configstore.RedactedPassword
}

func (m *Model) connectionExportRows() [][]any {
	rows := make([][]any, 0, len(m.cs.Connections))
	password := m.connectionPasswordForExport()
	for _, conn := range m.cs.Connections {
		pwd := conn.Password
		if password != "" {
			pwd = password
		}
		rows = append(rows, []any{
			conn.Name,
			conn.Group,
			conn.Protocol,
			conn.Host,
			conn.Port,
			conn.Username,
			pwd,
			conn.Database,
		})
	}
	return rows
}

func (m *Model) configExportText() (string, error) {
	switch m.mgrExportFormat {
	case configExportJSON:
		return m.cs.ExportJSON(m.mgrExportPasswords)
	case configExportCSV:
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		if err := w.Write(m.connectionExportColumns()); err != nil {
			return "", err
		}
		for _, row := range m.connectionExportRows() {
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
	case configExportTSV:
		var b strings.Builder
		b.WriteString(strings.Join(m.connectionExportColumns(), "\t"))
		b.WriteByte('\n')
		for _, row := range m.connectionExportRows() {
			for i, v := range row {
				if i > 0 {
					b.WriteByte('\t')
				}
				b.WriteString(escapeTSV(tsvValue(v)))
			}
			b.WriteByte('\n')
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("Excel 格式请按 Enter 保存为文件")
	}
}

func (m *Model) configExportData() (string, string, error) {
	choice := configExportChoices[m.mgrExportFormat]
	data, err := m.configExportText()
	return data, choice.ext, err
}

func (m *Model) copyMgrExport() (tea.Model, tea.Cmd) {
	data, _, err := m.configExportData()
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
	m.mgrImportText.SetValue(data)
	choice := configExportChoices[m.mgrExportFormat]
	m.status = fmt.Sprintf("已复制 %d 个连接配置到剪贴板（%s）", len(m.cs.Connections), choice.label)
	m.statusOK = true
	return m, nil
}

func (m *Model) saveMgrExport() (tea.Model, tea.Cmd) {
	choice := configExportChoices[m.mgrExportFormat]
	name := "connections_" + time.Now().Format("20060102_150405") + "." + choice.ext
	path, err := m.exportFilePath(name)
	if err != nil {
		m.status = "导出失败: " + err.Error()
		m.statusOK = false
		return m, nil
	}
	if m.mgrExportFormat == configExportXLSX {
		err = writeXLSX("连接配置", m.connectionExportColumns(), m.connectionExportRows(), path)
	} else {
		var data string
		data, _, err = m.configExportData()
		if err == nil {
			err = os.WriteFile(path, []byte(data), 0644)
		}
	}
	m.closeMgrImportModal()
	if err != nil {
		m.status = "导出失败: " + err.Error()
		m.statusOK = false
		return m, nil
	}
	m.status = fmt.Sprintf("已导出 %d 个连接到 %s", len(m.cs.Connections), path)
	m.statusOK = true
	return m, nil
}

func (m *Model) openMgrExportModal() {
	m.mgrExportPasswords = false
	m.mgrExportFormat = configExportJSON
	data, err := m.cs.ExportJSON(false)
	if err != nil {
		m.status = "导出失败: " + err.Error()
		m.statusOK = false
		return
	}
	m.mgrImportModal = true
	m.mgrExportMode = true
	m.initExportPathInput()
	ta := textarea.New()
	ta.SetValue(data)
	ta.SetHeight(12)
	ta.CharLimit = 0
	m.mgrImportText = ta
	m.mgrImportText.Focus()
	if err := clipboard.WriteAll(data); err != nil {
		m.status = "导出内容已显示；剪贴板不可用 (" + err.Error() + ")，可在弹窗中复制"
		m.statusOK = false
		return
	}
	m.status = "已复制 JSON 到剪贴板；可切换格式，Enter 保存文件，Esc 关闭"
	m.statusOK = true
}
