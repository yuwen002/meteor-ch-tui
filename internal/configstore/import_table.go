package configstore

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
)

func normalizeImportedConfig(cfg *ConfigFile) (*ConfigFile, error) {
	if cfg.Connections == nil {
		cfg.Connections = []SavedConnection{}
	}
	seen := map[string]bool{}
	for i := range cfg.Connections {
		conn := &cfg.Connections[i]
		conn.Name = strings.TrimSpace(conn.Name)
		conn.Host = strings.TrimSpace(conn.Host)
		conn.Group = strings.TrimSpace(conn.Group)
		if conn.Name == "" {
			return nil, fmt.Errorf("第 %d 个连接名称不能为空", i+1)
		}
		if seen[conn.Name] {
			return nil, fmt.Errorf("第 %d 个连接名称重复: %s", i+1, conn.Name)
		}
		seen[conn.Name] = true
		if conn.Host == "" {
			return nil, fmt.Errorf("连接 %s 的 Host 不能为空", conn.Name)
		}
		if conn.Port < 1 || conn.Port > 65535 {
			return nil, fmt.Errorf("连接 %s 的 Port 必须在 1-65535", conn.Name)
		}
		protocol := strings.ToLower(strings.TrimSpace(conn.Protocol))
		if protocol != "" && protocol != ch.ProtocolNative && protocol != ch.ProtocolHTTP && protocol != ch.ProtocolHTTPS {
			return nil, fmt.Errorf("连接 %s 的协议无效: %s", conn.Name, conn.Protocol)
		}
		conn.Protocol = ch.NormalizeProtocol(protocol)
	}
	return cfg, nil
}

var importHeaderAliases = map[string]string{
	"名称":       "name",
	"name":     "name",
	"分组":       "group",
	"group":    "group",
	"协议":       "protocol",
	"protocol": "protocol",
	"主机":       "host",
	"host":     "host",
	"端口":       "port",
	"port":     "port",
	"用户名":      "username",
	"user":     "username",
	"username": "username",
	"密码":       "password",
	"password": "password",
	"数据库":      "database",
	"database": "database",
}

func importHeader(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return importHeaderAliases[s]
}

func ImportTable(columns []string, rows [][]string) (*ConfigFile, error) {
	index := map[string]int{}
	for i, col := range columns {
		if key := importHeader(col); key != "" {
			index[key] = i
		}
	}
	for _, required := range []string{"name", "host", "port"} {
		if _, ok := index[required]; !ok {
			return nil, fmt.Errorf("缺少必需列: %s", required)
		}
	}
	get := func(row []string, key string) string {
		i, ok := index[key]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	cfg := &ConfigFile{Connections: []SavedConnection{}}
	for i, row := range rows {
		portText := get(row, "port")
		port, err := strconv.Atoi(portText)
		if err != nil {
			return nil, fmt.Errorf("第 %d 行端口无效: %s", i+2, portText)
		}
		cfg.Connections = append(cfg.Connections, SavedConnection{
			Name:     get(row, "name"),
			Group:    get(row, "group"),
			Host:     get(row, "host"),
			Port:     port,
			Username: get(row, "username"),
			Password: get(row, "password"),
			Database: get(row, "database"),
			Protocol: get(row, "protocol"),
		})
	}
	return normalizeImportedConfig(cfg)
}

func importCSV(data string, comma rune, unescape bool) (*ConfigFile, error) {
	r := csv.NewReader(strings.NewReader(data))
	r.Comma = comma
	r.LazyQuotes = true
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("表格解析失败: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("导入内容为空")
	}
	if unescape {
		for i := range rows {
			for j := range rows[i] {
				rows[i][j] = unescapeTSV(rows[i][j])
			}
		}
	}
	return ImportTable(rows[0], rows[1:])
}

func ImportCSV(data string) (*ConfigFile, error) {
	return importCSV(data, ',', false)
}

func ImportTSV(data string) (*ConfigFile, error) {
	return importCSV(data, '\t', true)
}

func ImportAuto(data string) (*ConfigFile, error) {
	trimmed := strings.TrimSpace(data)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return ImportJSON(data)
	}
	if strings.Contains(trimmed, "\t") {
		return ImportTSV(data)
	}
	if strings.Contains(trimmed, ",") {
		return ImportCSV(data)
	}
	return nil, fmt.Errorf("无法识别导入格式，请粘贴 JSON/CSV/TSV，或选择文件")
}

// XLSXSheets 返回 Excel 文件中的工作表名称。
func XLSXSheets(path string) ([]string, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("打开 Excel 失败: %w", err)
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel 中没有工作表")
	}
	return sheets, nil
}

// ImportXLSXSheet 从指定工作表导入连接配置。
func ImportXLSXSheet(path, sheet string) (*ConfigFile, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("打开 Excel 失败: %w", err)
	}
	defer f.Close()
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("读取工作表 %s 失败: %w", sheet, err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("工作表 %s 内容为空", sheet)
	}
	return ImportTable(rows[0], rows[1:])
}

// ImportXLSX 从第一个工作表导入，保留原有行为。
func ImportXLSX(path string) (*ConfigFile, error) {
	sheets, err := XLSXSheets(path)
	if err != nil {
		return nil, err
	}
	return ImportXLSXSheet(path, sheets[0])
}

func unescapeTSV(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				b.WriteByte('\n')
				i++
				continue
			case 'r':
				b.WriteByte('\r')
				i++
				continue
			case 't':
				b.WriteByte('\t')
				i++
				continue
			case '\\':
				b.WriteByte('\\')
				i++
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
