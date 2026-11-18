package ui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
)

func (m *Model) connectCmd() tea.Cmd {
	cfg := m.cfg
	if m.currentDB != "" {
		cfg.Database = m.currentDB
	}
	db := cfg.EffectiveDatabase()
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		client, err := ch.Dial(ctx, cfg, cfg.Addr())
		if err != nil {
			return connectedMsg{err: err}
		}
		nodes, err := client.Clusters(ctx)
		if err != nil {
			client.Close()
			return connectedMsg{err: err}
		}
		databases, tables, views, dbTruncated, objectsTruncated, catalogErr := loadCatalog(ctx, client, db)
		return connectedMsg{
			client: client, nodes: nodes, node: pickNode(nodes, cfg.Addr()),
			databases: databases, tables: tables, views: views,
			databasesTruncated: dbTruncated, objectsTruncated: objectsTruncated,
			database: db, catalogErr: catalogErr,
		}
	}
}

func (m *Model) switchNodeCmd(n ch.Node) tea.Cmd {
	cfg := m.cfg
	db := cfg.EffectiveDatabase()
	if m.currentDB != "" {
		cfg.Database = m.currentDB
		db = m.currentDB
	}
	addr := n.Addr()
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		client, err := ch.Dial(ctx, cfg, addr)
		if err != nil {
			return connectedMsg{err: err}
		}
		nodes, err := client.Clusters(ctx)
		if err != nil {
			client.Close()
			return connectedMsg{err: err}
		}
		databases, tables, views, dbTruncated, objectsTruncated, catalogErr := loadCatalog(ctx, client, db)
		cur := n
		return connectedMsg{
			client: client, nodes: nodes, node: &cur,
			databases: databases, tables: tables, views: views,
			databasesTruncated: dbTruncated, objectsTruncated: objectsTruncated,
			database: db, catalogErr: catalogErr,
		}
	}
}

// switchDatabaseCmd 以新的默认库重建连接，保持当前节点地址不变。
func (m *Model) switchDatabaseCmd(db string) tea.Cmd {
	cfg := m.cfg
	cfg.Database = db
	addr := cfg.Addr()
	if m.client != nil && m.client.Addr() != "" {
		addr = m.client.Addr()
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		client, err := ch.Dial(ctx, cfg, addr)
		if err != nil {
			return connectedMsg{err: err}
		}
		nodes, err := client.Clusters(ctx)
		if err != nil {
			client.Close()
			return connectedMsg{err: err}
		}
		databases, tables, views, dbTruncated, objectsTruncated, catalogErr := loadCatalog(ctx, client, db)
		return connectedMsg{
			client: client, nodes: nodes, node: pickNode(nodes, addr),
			databases: databases, tables: tables, views: views,
			databasesTruncated: dbTruncated, objectsTruncated: objectsTruncated,
			database: db, catalogErr: catalogErr,
		}
	}
}

// loadCatalog 读取账号可见的全部库与当前库的表、视图；失败不阻断连接，仅提示。
func loadCatalog(ctx context.Context, client *ch.Client, db string) ([]string, []string, []string, bool, bool, error) {
	databases, dbTruncated, err := client.Databases(ctx)
	if err != nil {
		return nil, nil, nil, false, false, fmt.Errorf("读取数据库列表失败: %w", err)
	}
	tables, views, objectsTruncated, err := client.TablesAndViews(ctx, db)
	if err != nil {
		return databases, nil, nil, dbTruncated, false, fmt.Errorf("读取库 %s 的表与视图失败: %w", db, err)
	}
	return databases, tables, views, dbTruncated, objectsTruncated, nil
}

func (m *Model) runQueryCmd(sql string) tea.Cmd {
	client := m.client
	if client == nil {
		return nil
	}
	m.addQueryHistory(sql)
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 70*time.Second)
		defer cancel()
		res, err := client.Query(ctx, sql, 500)
		return queryDoneMsg{res: res, sql: sql, err: err}
	}
}

func (m *Model) slowQueryCmd() tea.Cmd {
	client := m.client
	if client == nil {
		return nil
	}
	const sql = "SELECT query_id, user, elapsed, memory_usage, read_rows, substring(query, 1, 120) AS query " +
		"FROM system.processes ORDER BY elapsed DESC LIMIT 50"
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		res, err := client.Query(ctx, sql, 0)
		return slowDoneMsg{res: res, err: err}
	}
}

func pickNode(nodes []ch.Node, addr string) *ch.Node {
	for i := range nodes {
		if nodes[i].Addr() == addr {
			return &nodes[i]
		}
	}
	if len(nodes) > 0 {
		return &nodes[0]
	}
	return nil
}

func fmtConnected(addr string, nodeCount, dbCount int) string {
	return fmt.Sprintf("已连接 %s，%d 个节点 / %d 个库", addr, nodeCount, dbCount)
}

func fmtDBSwitched(db string, tableCount, viewCount int) string {
	return fmt.Sprintf("已切换到库 %s（%d 张表 / %d 个视图）· 侧边栏可切换其它库", db, tableCount, viewCount)
}

func fmtQueryDone(res *ch.Result) string {
	s := fmt.Sprintf("返回 %d 行 / %d 列，耗时 %s",
		len(res.Rows), len(res.Columns), res.Elapsed.Round(time.Millisecond))
	if res.Truncated {
		s += "（已截断）"
	}
	return s
}

func fmtSlowDone(rows int) string {
	return fmt.Sprintf("活跃查询 %d 条（Ctrl+R 刷新，Esc 返回）", rows)
}
