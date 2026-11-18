package ui

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/xuri/excelize/v2"
)

func writeTwoSheetXLSXForUI(t *testing.T, path string) {
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

func TestSheetPickerFromXLSXFile(t *testing.T) {
	isolateTestHome(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "multi.xlsx")
	writeTwoSheetXLSXForUI(t, path)

	m := newImportTestModel(t)
	m.openMgrImportModal()
	m.lastImportDir = dir
	m.openFileBrowser()
	for i, item := range m.fileBrowserItems {
		if item.name == "multi.xlsx" {
			m.fileBrowserSel = i
			break
		}
	}
	m.selectFileBrowser()

	if m.fileBrowserOpen {
		t.Fatal("file browser should close after selecting xlsx")
	}
	if !m.sheetPickerOpen {
		t.Fatal("multi-sheet xlsx should open the sheet picker")
	}

	plain := ansi.Strip(m.View())
	x, y := findCell(t, plain, "Second")
	m.handleMouse(leftClick(x, y))

	if m.sheetPickerOpen {
		t.Fatal("sheet picker should close after selecting a sheet")
	}
	value := m.mgrImportText.Value()
	if !strings.Contains(value, "second.local") || strings.Contains(value, "first.local") {
		t.Fatalf("import text = %q", value)
	}
}

func TestSheetPickerCancelHotspot(t *testing.T) {
	m := newImportTestModel(t)
	m.openMgrImportModal()
	m.openSheetPicker("book.xlsx", []string{"Sheet1", "Second"})

	plain := ansi.Strip(m.View())
	x, y := findCell(t, plain, "[Esc] 取消")
	m.handleMouse(leftClick(x, y))

	if m.sheetPickerOpen {
		t.Fatal("clicking [Esc] 取消 should close the sheet picker")
	}
}
