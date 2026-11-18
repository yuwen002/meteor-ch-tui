package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/xuri/excelize/v2"
)

func writeImportXLSXForTest(t *testing.T, path string) {
	t.Helper()
	f := excelize.NewFile()
	sheet := f.GetSheetList()[0]
	rows := [][]string{
		{"名称", "主机", "端口"},
		{"from-xlsx", "xlsx.local", "9101"},
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
}

func TestFileBrowserMouseSelectsJSON(t *testing.T) {
	isolateTestHome(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "conns.json")
	data := `{"connections":[{"name":"from-file","host":"file.local","port":9100}]}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	m := newImportTestModel(t)
	m.openMgrImportModal()
	m.lastImportDir = dir
	m.openFileBrowser()

	plain := ansi.Strip(m.View())
	x, y := findCell(t, plain, "conns.json")
	m.handleMouse(leftClick(x, y))

	if m.fileBrowserOpen {
		t.Fatal("file browser should close after selecting a file")
	}
	if !strings.Contains(m.mgrImportText.Value(), "from-file") {
		t.Fatalf("import text = %q", m.mgrImportText.Value())
	}
	if m.lastImportDir != dir {
		t.Fatalf("lastImportDir = %q, want %q", m.lastImportDir, dir)
	}
}

func TestFileBrowserLoadsXLSX(t *testing.T) {
	isolateTestHome(t)
	dir := t.TempDir()
	writeImportXLSXForTest(t, filepath.Join(dir, "conns.xlsx"))

	m := newImportTestModel(t)
	m.openMgrImportModal()
	m.lastImportDir = dir
	m.openFileBrowser()

	idx := -1
	for i, item := range m.fileBrowserItems {
		if item.name == "conns.xlsx" {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatal("file browser did not list conns.xlsx")
	}
	m.fileBrowserSel = idx
	m.selectFileBrowser()

	if m.fileBrowserOpen {
		t.Fatal("file browser should close after selecting a file")
	}
	if m.sheetPickerOpen {
		t.Fatal("single-sheet xlsx should not open the sheet picker")
	}
	value := m.mgrImportText.Value()
	if !strings.Contains(value, "from-xlsx") || !strings.Contains(value, "xlsx.local") {
		t.Fatalf("xlsx import text = %q", value)
	}
}

func TestFileBrowserCancelHotspot(t *testing.T) {
	m := newImportTestModel(t)
	m.openMgrImportModal()
	m.lastImportDir = t.TempDir()
	m.openFileBrowser()

	plain := ansi.Strip(m.View())
	x, y := findCell(t, plain, "[Esc] 取消")
	m.handleMouse(leftClick(x, y))

	if m.fileBrowserOpen {
		t.Fatal("clicking [Esc] 取消 should close the file browser")
	}
}
