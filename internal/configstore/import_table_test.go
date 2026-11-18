package configstore

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
)

func TestImportAuto(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{"json", `{"connections":[{"name":"a","host":"a.local","port":9000}]}`, "a"},
		{"csv", "名称,主机,端口\nb,b.local,9001\n", "b"},
		{"tsv", "名称\t主机\t端口\nc\tc.local\t9002\n", "c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := ImportAuto(tt.data)
			if err != nil {
				t.Fatalf("ImportAuto() error = %v", err)
			}
			if len(cfg.Connections) != 1 || cfg.Connections[0].Name != tt.want {
				t.Fatalf("ImportAuto() = %+v, want connection %q", cfg.Connections, tt.want)
			}
		})
	}

	if _, err := ImportAuto("无法识别"); err == nil {
		t.Fatal("ImportAuto() accepted unrecognized input")
	}
}

func TestImportTableAliasesAndValidation(t *testing.T) {
	cfg, err := ImportTable(
		[]string{"name", "主机", "PORT", "user", "group", "protocol"},
		[][]string{{"web", "web.local", "8123", "ops", "站点", "http"}},
	)
	if err != nil {
		t.Fatalf("ImportTable() error = %v", err)
	}
	got := cfg.Connections[0]
	if got.Host != "web.local" || got.Port != 8123 || got.Username != "ops" || got.Group != "站点" {
		t.Fatalf("ImportTable() connection = %+v", got)
	}
	if got.Protocol != ch.ProtocolHTTP {
		t.Fatalf("ImportTable() protocol = %q, want %q", got.Protocol, ch.ProtocolHTTP)
	}

	if _, err := ImportTable([]string{"名称", "端口"}, [][]string{{"a", "9000"}}); err == nil || !strings.Contains(err.Error(), "host") {
		t.Fatalf("ImportTable() missing host error = %v", err)
	}
	if _, err := ImportTable(
		[]string{"名称", "主机", "端口"},
		[][]string{{"dup", "a.local", "9000"}, {"dup", "b.local", "9001"}},
	); err == nil || !strings.Contains(err.Error(), "重复") {
		t.Fatalf("ImportTable() duplicate error = %v", err)
	}
	if _, err := ImportTable(
		[]string{"名称", "主机", "端口"},
		[][]string{{"bad", "a.local", "not-a-port"}},
	); err == nil || !strings.Contains(err.Error(), "端口无效") {
		t.Fatalf("ImportTable() invalid port error = %v", err)
	}
}

func writeMultiSheetXLSXForTest(t *testing.T, path string) {
	t.Helper()
	f := excelize.NewFile()
	first := f.GetSheetList()[0]
	if _, err := f.NewSheet("Second"); err != nil {
		t.Fatal(err)
	}
	writeRows := func(sheet string, rows [][]string) {
		for r, row := range rows {
			for c, value := range row {
				cell, err := excelize.CoordinatesToCellName(c+1, r+1)
				if err != nil {
					t.Fatal(err)
				}
				if err := f.SetCellValue(sheet, cell, value); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	writeRows(first, [][]string{
		{"名称", "主机", "端口"},
		{"first", "first.local", "9101"},
	})
	writeRows("Second", [][]string{
		{"名称", "主机", "端口"},
		{"second", "second.local", "9102"},
	})
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestXLSXSheetSelection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "multi.xlsx")
	writeMultiSheetXLSXForTest(t, path)

	sheets, err := XLSXSheets(path)
	if err != nil {
		t.Fatalf("XLSXSheets() error = %v", err)
	}
	if len(sheets) != 2 || sheets[0] != "Sheet1" || sheets[1] != "Second" {
		t.Fatalf("XLSXSheets() = %v", sheets)
	}

	cfg, err := ImportXLSXSheet(path, "Second")
	if err != nil {
		t.Fatalf("ImportXLSXSheet() error = %v", err)
	}
	if len(cfg.Connections) != 1 || cfg.Connections[0].Name != "second" || cfg.Connections[0].Host != "second.local" {
		t.Fatalf("ImportXLSXSheet() = %+v", cfg.Connections)
	}

	first, err := ImportXLSX(path)
	if err != nil {
		t.Fatalf("ImportXLSX() error = %v", err)
	}
	if len(first.Connections) != 1 || first.Connections[0].Name != "first" {
		t.Fatalf("ImportXLSX() = %+v", first.Connections)
	}
}

func TestImportXLSX(t *testing.T) {
	path := filepath.Join(t.TempDir(), "connections.xlsx")
	f := excelize.NewFile()
	sheet := f.GetSheetList()[0]
	rows := [][]string{
		{"名称", "主机", "端口", "用户名"},
		{"excel", "excel.local", "9003", "alice"},
	}
	for r, row := range rows {
		for c, value := range row {
			cell, err := excelize.CoordinatesToCellName(c+1, r+1)
			if err != nil {
				t.Fatal(err)
			}
			if err := f.SetCellValue(sheet, cell, value); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := ImportXLSX(path)
	if err != nil {
		t.Fatalf("ImportXLSX() error = %v", err)
	}
	if len(cfg.Connections) != 1 {
		t.Fatalf("ImportXLSX() connections = %+v", cfg.Connections)
	}
	got := cfg.Connections[0]
	if got.Name != "excel" || got.Host != "excel.local" || got.Port != 9003 || got.Username != "alice" {
		t.Fatalf("ImportXLSX() connection = %+v", got)
	}

	empty := filepath.Join(t.TempDir(), "empty.xlsx")
	ef := excelize.NewFile()
	if err := ef.SaveAs(empty); err != nil {
		t.Fatal(err)
	}
	if err := ef.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportXLSX(empty); err == nil {
		t.Fatal("ImportXLSX() accepted an empty workbook")
	}
}
