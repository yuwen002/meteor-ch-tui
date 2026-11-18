package ch

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// 支持的连接协议。
const (
	ProtocolNative = "native"
	ProtocolHTTP   = "http"
	ProtocolHTTPS  = "https"
)

// NormalizeProtocol 把任意输入规范化为受支持的协议值，空值视为 native（兼容旧配置）。
func NormalizeProtocol(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case ProtocolHTTP:
		return ProtocolHTTP
	case ProtocolHTTPS:
		return ProtocolHTTPS
	default:
		return ProtocolNative
	}
}

// IsHTTPProtocol 判断协议是否为 HTTP/HTTPS。
func IsHTTPProtocol(s string) bool {
	p := NormalizeProtocol(s)
	return p == ProtocolHTTP || p == ProtocolHTTPS
}

// ProtocolLabel 返回界面展示用的协议名。
func ProtocolLabel(s string) string {
	switch NormalizeProtocol(s) {
	case ProtocolHTTP:
		return "HTTP"
	case ProtocolHTTPS:
		return "HTTPS"
	default:
		return "Native"
	}
}

// Config 描述到某个 ClickHouse 节点的连接参数。
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Protocol string
}

// IsHTTP 报告该配置是否使用 HTTP 接口。
func (c Config) IsHTTP() bool { return IsHTTPProtocol(c.Protocol) }

// Scheme 返回 HTTP 族协议使用的 URL scheme。
func (c Config) Scheme() string {
	if NormalizeProtocol(c.Protocol) == ProtocolHTTPS {
		return "https"
	}
	return "http"
}

// Addr 返回 "host:port" 形式的连接地址。
func (c Config) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

// EffectiveDatabase 返回实际建立会话时使用的库名。
// 留空表示「不指定库」：沿用服务端默认库，但账号可见的全部库依然可以浏览。
func (c Config) EffectiveDatabase() string {
	if strings.TrimSpace(c.Database) == "" {
		return "default"
	}
	return c.Database
}

// Node 描述 system.clusters 中的一个节点。
type Node struct {
	Cluster  string
	Shard    string
	Replica  string
	HostName string
	Address  string
	Port     int
}

func (n Node) Addr() string {
	host := n.Address
	if host == "" {
		host = n.HostName
	}
	port := n.Port
	if port == 0 {
		port = 9000
	}
	return net.JoinHostPort(host, strconv.Itoa(port))
}

// Label 返回侧边栏展示的节点标签。
func (n Node) Label() string {
	return fmt.Sprintf("s%s/r%s  %s", n.Shard, n.Replica, n.HostName)
}

// Result 保存一次查询的结果集与元数据。
type Result struct {
	Columns   []string
	Types     []string
	Rows      [][]any
	Elapsed   time.Duration
	Truncated bool
}

// Client 封装一个 ClickHouse 连接（Native 或 HTTP）。
type Client struct {
	conn driver.Conn
	http *httpClient
	addr string
	cfg  Config
}

// Dial 建立到指定地址的连接并完成握手校验。
func Dial(ctx context.Context, cfg Config, addr string) (*Client, error) {
	if cfg.IsHTTP() {
		h := newHTTPClient(cfg, addr)
		if _, err := h.Query(ctx, "SELECT 1", 1); err != nil {
			h.Close()
			return nil, err
		}
		return &Client{http: h, addr: addr, cfg: cfg}, nil
	}

	opts := &clickhouse.Options{
		Addr:        []string{addr},
		Auth:        clickhouse.Auth{Database: cfg.EffectiveDatabase(), Username: cfg.Username, Password: cfg.Password},
		DialTimeout: 5 * time.Second,
		ReadTimeout: 60 * time.Second,
		Compression: &clickhouse.Compression{Method: clickhouse.CompressionLZ4},
		Settings:    clickhouse.Settings{"max_execution_time": 60},
	}
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(ctx); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &Client{conn: conn, addr: addr, cfg: cfg}, nil
}

// Close 关闭连接。
func (c *Client) Close() {
	if c == nil {
		return
	}
	if c.http != nil {
		c.http.Close()
		return
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

// Addr 返回该连接实际使用的地址。
func (c *Client) Addr() string {
	if c == nil {
		return ""
	}
	return c.addr
}

type anyScanner struct{ v any }

func (a *anyScanner) Scan(src any) error {
	a.v = src
	return nil
}

func scanDest(types []driver.ColumnType) ([]any, []reflect.Value) {
	dest := make([]any, len(types))
	holders := make([]reflect.Value, len(types))
	for i, ct := range types {
		st := ct.ScanType()
		if st == nil || st.Kind() == reflect.Interface {
			dest[i] = &anyScanner{}
			holders[i] = reflect.ValueOf(dest[i]).Elem()
			continue
		}
		holders[i] = reflect.New(st)
		dest[i] = holders[i].Interface()
	}
	return dest, holders
}

// Query 执行 SQL，返回最多 limit 行（limit<=0 表示不限制）。
func (c *Client) Query(ctx context.Context, sql string, limit int) (*Result, error) {
	if c == nil {
		return nil, fmt.Errorf("连接已关闭")
	}
	if c.http != nil {
		return c.http.Query(ctx, sql, limit)
	}
	start := time.Now()
	rows, err := c.conn.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	cols := rows.Columns()
	types := rows.ColumnTypes()
	res := &Result{Columns: cols}
	for _, t := range types {
		res.Types = append(res.Types, t.DatabaseTypeName())
	}

	for rows.Next() {
		if limit > 0 && len(res.Rows) >= limit {
			res.Truncated = true
			break
		}
		dest, holders := scanDest(types)
		if err := rows.Scan(dest...); err != nil {
			return res, err
		}
		row := make([]any, len(cols))
		for i := range row {
			if s, ok := dest[i].(*anyScanner); ok {
				row[i] = s.v
				continue
			}
			row[i] = holders[i].Elem().Interface()
		}
		res.Rows = append(res.Rows, row)
	}
	if err := rows.Err(); err != nil {
		return res, err
	}
	res.Elapsed = time.Since(start)
	return res, nil
}

func atoi(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

var _ sql.Scanner = (*anyScanner)(nil)
