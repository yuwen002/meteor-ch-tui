package ch

import (
	"context"
	"fmt"
	"strings"
)

// CatalogLimit 限制侧边栏一次加载的库/表数量，避免超大实例拖慢界面。
const CatalogLimit = 1000

// Databases 返回当前账号可见的数据库名及是否发生截断。
func (c *Client) Databases(ctx context.Context) ([]string, bool, error) {
	res, err := c.Query(ctx, "SELECT name FROM system.databases ORDER BY name", CatalogLimit)
	if err != nil {
		return nil, false, err
	}
	out := make([]string, 0, len(res.Rows))
	for _, row := range res.Rows {
		if len(row) == 0 {
			continue
		}
		if name := cellString(row[0]); name != "" {
			out = append(out, name)
		}
	}
	return out, res.Truncated, nil
}

// Tables 返回指定库下的表名。
func (c *Client) Tables(ctx context.Context, db string) ([]string, error) {
	sql := fmt.Sprintf("SELECT name FROM system.tables WHERE database = %s ORDER BY name", quoteSQL(db))
	res, err := c.Query(ctx, sql, CatalogLimit)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(res.Rows))
	for _, row := range res.Rows {
		if len(row) == 0 {
			continue
		}
		if name := cellString(row[0]); name != "" {
			out = append(out, name)
		}
	}
	return out, nil
}

var viewEngines = map[string]struct{}{
	"View":             {},
	"MaterializedView": {},
	"LiveView":         {},
	"WindowView":       {},
}

func isViewEngine(engine string) bool {
	_, ok := viewEngines[engine]
	return ok
}

// TablesAndViews 一次读取指定库下的表与视图，避免为侧边栏多做一轮串行查询。
func (c *Client) TablesAndViews(ctx context.Context, db string) (tables, views []string, truncated bool, err error) {
	sql := fmt.Sprintf("SELECT name, engine FROM system.tables WHERE database = %s ORDER BY name", quoteSQL(db))
	res, err := c.Query(ctx, sql, CatalogLimit*2)
	if err != nil {
		return nil, nil, false, err
	}
	tables, views = splitTablesAndViews(res.Rows)
	return tables, views, res.Truncated, nil
}

func splitTablesAndViews(rows [][]any) (tables, views []string) {
	tables = make([]string, 0, len(rows))
	views = make([]string, 0)
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		name := cellString(row[0])
		if name == "" {
			continue
		}
		if isViewEngine(cellString(row[1])) {
			views = append(views, name)
			continue
		}
		tables = append(tables, name)
	}
	return tables, views
}

func cellString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return FormatValue(v)
}

// quoteSQL 生成 ClickHouse 字符串字面量，避免库名中的引号/反斜杠破坏查询。
func quoteSQL(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "'", `\'`)
	return "'" + s + "'"
}
