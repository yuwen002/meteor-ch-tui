package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
)

func TestConfigExportAllFormatsSave(t *testing.T) {
	isolateTestHome(t)
	formats := []struct {
		format configExportFormat
		ext    string
	}{
		{configExportJSON, "json"},
		{configExportCSV, "csv"},
		{configExportTSV, "tsv"},
		{configExportXLSX, "xlsx"},
	}
	for _, tc := range formats {
		t.Run(tc.ext, func(t *testing.T) {
			dir := t.TempDir()
			m := newImportTestModel(t)
			m.mgrExportMode = true
			m.mgrExportPasswords = false
			m.mgrExportFormat = tc.format
			m.initExportPathInput()
			m.exportPathInput.SetValue(dir)

			m.saveMgrExport()

			if !m.statusOK {
				t.Fatalf("saveMgrExport() status = %q", m.status)
			}
			matches, err := filepath.Glob(filepath.Join(dir, "connections_*."+tc.ext))
			if err != nil || len(matches) != 1 {
				t.Fatalf("export files = %v, err = %v", matches, err)
			}
			data, err := os.ReadFile(matches[0])
			if err != nil || len(data) == 0 {
				t.Fatalf("read export: %v, size = %d", err, len(data))
			}
			if tc.ext != "xlsx" && !strings.Contains(string(data), "old.local") {
				t.Fatalf("export content = %q", data)
			}
		})
	}
}

func TestResultExportAllFormatsSave(t *testing.T) {
	isolateTestHome(t)
	formats := []struct {
		sel int
		ext string
	}{
		{0, "tsv"},
		{1, "csv"},
		{2, "json"},
		{3, "xlsx"},
	}
	for _, tc := range formats {
		t.Run(tc.ext, func(t *testing.T) {
			dir := t.TempDir()
			m := newImportTestModel(t)
			m.appMode = appMain
			m.initMainMode()
			m.table.setResult(&ch.Result{
				Columns: []string{"name", "value"},
				Rows:    [][]any{{"alpha", 1}, {"beta", 2}},
			})
			m.initExportPathInput()
			m.exportPathInput.SetValue(dir)
			m.resultExportSel = tc.sel

			m.executeResultExport()

			if !m.statusOK {
				t.Fatalf("executeResultExport() status = %q", m.status)
			}
			matches, err := filepath.Glob(filepath.Join(dir, "query_result_*."+tc.ext))
			if err != nil || len(matches) != 1 {
				t.Fatalf("export files = %v, err = %v", matches, err)
			}
			data, err := os.ReadFile(matches[0])
			if err != nil || len(data) == 0 {
				t.Fatalf("read export: %v, size = %d", err, len(data))
			}
			if tc.ext != "xlsx" && !strings.Contains(string(data), "alpha") {
				t.Fatalf("export content = %q", data)
			}
		})
	}
}

func TestDirBrowserMouseSelectDirectory(t *testing.T) {
	isolateTestHome(t)
	base := t.TempDir()
	child := filepath.Join(base, "child")
	if err := os.MkdirAll(child, 0755); err != nil {
		t.Fatal(err)
	}

	m := newImportTestModel(t)
	m.initExportPathInput()
	m.exportPathInput.SetValue(base)
	m.openDirBrowser()

	plain := ansi.Strip(m.View())
	x, y := findCell(t, plain, "child")
	m.handleMouse(leftClick(x, y))
	if m.dirBrowserPath != child {
		t.Fatalf("dirBrowserPath = %q, want %q", m.dirBrowserPath, child)
	}

	plain = ansi.Strip(m.View())
	x, y = findCell(t, plain, "[Ctrl+S] 选择当前")
	m.handleMouse(leftClick(x, y))
	if m.dirBrowserOpen {
		t.Fatal("directory browser should close after selecting current directory")
	}
	if got := m.exportPathInput.Value(); got != child {
		t.Fatalf("export path = %q, want %q", got, child)
	}
	if m.lastExportDir != child {
		t.Fatalf("lastExportDir = %q, want %q", m.lastExportDir, child)
	}
}

func TestResultExportCancelHotspot(t *testing.T) {
	m := newImportTestModel(t)
	m.appMode = appMain
	m.initMainMode()
	m.table.setResult(&ch.Result{
		Columns: []string{"name"},
		Rows:    [][]any{{"alpha"}},
	})
	m.initExportPathInput()
	m.exportPathInput.SetValue(t.TempDir())
	m.openResultExportModal()

	plain := ansi.Strip(m.View())
	x, y := findCell(t, plain, "[Esc] 取消")
	m.handleMouse(leftClick(x, y))
	if m.resultExportModal {
		t.Fatal("clicking [Esc] 取消 should close the result export modal")
	}
}

func TestConfigExportCancelHotspot(t *testing.T) {
	m := newImportTestModel(t)
	m.openMgrExportModal()

	plain := ansi.Strip(m.View())
	x, y := findCell(t, plain, "[Esc] 关闭")
	m.handleMouse(leftClick(x, y))
	if m.mgrImportModal {
		t.Fatal("clicking [Esc] 关闭 should close the config export modal")
	}
}

func TestDirectoriesPersistAcrossModels(t *testing.T) {
	home := isolateTestHome(t)
	exportDir := filepath.Join(home, "exports")
	importDir := filepath.Join(home, "imports")
	for _, dir := range []string{exportDir, importDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	m := newImportTestModel(t)
	m.setLastExportDir(exportDir)
	m.setLastImportDir(importDir)

	next := newImportTestModel(t)
	if next.lastExportDir != exportDir || next.lastImportDir != importDir {
		t.Fatalf("loaded dirs = (%q, %q), want (%q, %q)", next.lastExportDir, next.lastImportDir, exportDir, importDir)
	}
	next.initExportPathInput()
	if got := next.exportPathInput.Value(); got != exportDir {
		t.Fatalf("export path input = %q, want %q", got, exportDir)
	}
	next.openFileBrowser()
	if got := next.fileBrowserPath; got != importDir {
		t.Fatalf("file browser path = %q, want %q", got, importDir)
	}
}
