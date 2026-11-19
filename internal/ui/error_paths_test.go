package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
)

func TestExportFilePathUnwritableDir(t *testing.T) {
	isolateTestHome(t)
	blocker := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := newImportTestModel(t)
	m.initExportPathInput()
	m.exportPathInput.SetValue(blocker)

	_, err := m.exportFilePath("demo.json")
	if err == nil || !strings.Contains(err.Error(), "创建保存目录失败") {
		t.Fatalf("exportFilePath() error = %v", err)
	}
}

func TestResultXLSXWriteFailure(t *testing.T) {
	target := filepath.Join(t.TempDir(), "blocked.xlsx")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}

	m := newImportTestModel(t)
	m.table.setResult(&ch.Result{
		Columns: []string{"name"},
		Rows:    [][]any{{"alpha"}},
	})
	if err := m.resultXLSX(target); err == nil {
		t.Fatal("resultXLSX() should fail when the target path is a directory")
	}
}

func TestBrokenXLSXFileBrowser(t *testing.T) {
	isolateTestHome(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.xlsx")
	if err := os.WriteFile(path, []byte("not xlsx"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := newImportTestModel(t)
	m.openMgrImportModal()
	m.lastImportDir = dir
	m.openFileBrowser()
	for i, item := range m.fileBrowserItems {
		if item.name == "broken.xlsx" {
			m.fileBrowserSel = i
			break
		}
	}
	m.selectFileBrowser()

	if m.fileBrowserOpen {
		t.Fatal("file browser should close after selecting a broken xlsx")
	}
	if m.sheetPickerOpen {
		t.Fatal("broken xlsx should not open the sheet picker")
	}
	if m.statusOK || !strings.Contains(m.status, "读取文件失败") {
		t.Fatalf("status = %q, ok = %v", m.status, m.statusOK)
	}
}

func TestLargeResultTextExport(t *testing.T) {
	const n = 10000
	rows := make([][]any, n)
	for i := range rows {
		rows[i] = []any{fmt.Sprintf("name-%d", i), i}
	}
	m := newImportTestModel(t)
	m.table.setResult(&ch.Result{
		Columns: []string{"name", "value"},
		Rows:    rows,
	})

	if got := len(m.resultRows()); got != n {
		t.Fatalf("resultRows() length = %d, want %d", got, n)
	}

	tsv, err := m.resultTSV()
	if err != nil {
		t.Fatalf("resultTSV() error = %v", err)
	}
	if got := strings.Count(tsv, "\n"); got != n+1 {
		t.Fatalf("TSV lines = %d, want %d", got, n+1)
	}

	csvData, err := m.resultCSV()
	if err != nil {
		t.Fatalf("resultCSV() error = %v", err)
	}
	if got := strings.Count(csvData, "\n"); got != n+1 {
		t.Fatalf("CSV lines = %d, want %d", got, n+1)
	}

	jsonData, err := m.resultJSON()
	if err != nil {
		t.Fatalf("resultJSON() error = %v", err)
	}
	if got := strings.Count(jsonData, `"name":`); got != n {
		t.Fatalf("JSON objects = %d, want %d", got, n)
	}
}

func TestLargeResultXLSXExport(t *testing.T) {
	const n = 3000
	rows := make([][]any, n)
	for i := range rows {
		rows[i] = []any{fmt.Sprintf("name-%d", i), i}
	}
	m := newImportTestModel(t)
	m.table.setResult(&ch.Result{
		Columns: []string{"name", "value"},
		Rows:    rows,
	})

	path := filepath.Join(t.TempDir(), "large.xlsx")
	if err := m.resultXLSX(path); err != nil {
		t.Fatalf("resultXLSX() error = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat xlsx: %v", err)
	}
	if info.Size() < 10000 {
		t.Fatalf("xlsx size = %d, unexpectedly small", info.Size())
	}
}
