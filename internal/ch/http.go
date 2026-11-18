package ch

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// httpClient 通过 ClickHouse 的 HTTP 接口执行查询。
// 驱动内置的新版 HTTP 模式会附带 client_protocol_version 等设置，
// 老版本服务端（如 22.x）会直接拒绝，因此这里使用最小的 HTTP 协议实现。
type httpClient struct {
	base   string
	user   string
	pass   string
	db     string
	client *http.Client
}

func newHTTPClient(cfg Config, addr string) *httpClient {
	return &httpClient{
		base:   cfg.Scheme() + "://" + addr + "/",
		user:   cfg.Username,
		pass:   cfg.Password,
		db:     cfg.Database,
		client: &http.Client{Timeout: 70 * time.Second},
	}
}

func (h *httpClient) Close() {
	if h != nil && h.client != nil {
		h.client.CloseIdleConnections()
	}
}

var formatClause = regexp.MustCompile(`(?i)\bFORMAT\s+[A-Za-z0-9_]+\s*$`)

func withTSVFormat(sql string) string {
	s := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(sql), ";"))
	if formatClause.MatchString(s) {
		return s
	}
	return s + "\nFORMAT TabSeparatedWithNamesAndTypes"
}

// Query 以 TabSeparatedWithNamesAndTypes 格式取回结果，保持与 Native 路径相同的 Result 结构。
func (h *httpClient) Query(ctx context.Context, sql string, limit int) (*Result, error) {
	start := time.Now()
	q := url.Values{}
	if h.db != "" {
		q.Set("database", h.db)
	}
	q.Set("max_execution_time", "60")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		h.base+"?"+q.Encode(), strings.NewReader(withTSVFormat(sql)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-ClickHouse-User", h.user)
	req.Header.Set("X-ClickHouse-Key", h.pass)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	res := &Result{}
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 64*1024), 8*1024*1024)
	lineNo := 0
	for sc.Scan() {
		line := sc.Text()
		lineNo++
		switch lineNo {
		case 1:
			res.Columns = splitTSV(line)
			continue
		case 2:
			res.Types = splitTSV(line)
			continue
		}
		if limit > 0 && len(res.Rows) >= limit {
			res.Truncated = true
			break
		}
		res.Rows = append(res.Rows, tsvRow(line, len(res.Columns)))
	}
	if err := sc.Err(); err != nil {
		return res, err
	}
	res.Elapsed = time.Since(start)
	return res, nil
}

func splitTSV(line string) []string {
	parts := strings.Split(line, "\t")
	for i := range parts {
		parts[i] = unescapeTSV(parts[i])
	}
	return parts
}

func tsvRow(line string, cols int) []any {
	parts := splitTSV(line)
	row := make([]any, cols)
	for i := 0; i < cols && i < len(parts); i++ {
		if parts[i] == `\N` {
			continue
		}
		row[i] = parts[i]
	}
	return row
}

func unescapeTSV(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 't':
			b.WriteByte('\t')
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case '0':
			b.WriteByte(0)
		case '\\':
			b.WriteByte('\\')
		default:
			b.WriteByte('\\')
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
