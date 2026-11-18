package ch

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// FormatValue 将查询得到的任意值统一格式化为可显示字符串。
func FormatValue(v any) string {
	if v == nil {
		return "NULL"
	}
	if t, ok := v.(time.Time); ok {
		return t.Format("2006-01-02 15:04:05")
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return "NULL"
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return "NULL"
	}
	if rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return "NULL"
		}
		return FormatValue(rv.Elem().Interface())
	}
	switch rv.Kind() {
	case reflect.String:
		return sanitizeCell(rv.String())
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return sanitizeCell(string(rv.Bytes()))
		}
		return sanitizeCell(fmt.Sprint(rv.Interface()))
	default:
		return sanitizeCell(fmt.Sprint(rv.Interface()))
	}
}

// sanitizeCell 转义换行、制表符等不可见字符，避免表格渲染错乱。
func sanitizeCell(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", " / ")
	s = strings.ReplaceAll(s, "\t", "    ")
	return s
}
