package ui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func TestFileBrowserFilter(t *testing.T) {
	isolateTestHome(t)
	dir := t.TempDir()
	for _, name := range []string{"alpha.json", "beta.csv"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	m := newImportTestModel(t)
	m.lastImportDir = dir
	m.openFileBrowser()
	if len(m.fileBrowserItems) != 3 {
		t.Fatalf("items = %+v, want parent + 2 files", m.fileBrowserItems)
	}

	m.handleFileBrowserKey(tea.KeyMsg{}, "/")
	if !m.fileBrowserFilterFocused {
		t.Fatal("/ should focus the filter input")
	}
	m.fileBrowserFilter.SetValue("alp")
	m.loadFileBrowser()
	if len(m.fileBrowserItems) != 2 || m.fileBrowserItems[1].name != "alpha.json" {
		t.Fatalf("filtered items = %+v", m.fileBrowserItems)
	}

	m.handleFileBrowserKey(tea.KeyMsg{}, "esc")
	if m.fileBrowserFilterFocused {
		t.Fatal("esc should blur the filter input")
	}
}

func TestFileBrowserSortToggle(t *testing.T) {
	isolateTestHome(t)
	dir := t.TempDir()
	a := filepath.Join(dir, "a.json")
	b := filepath.Join(dir, "b.csv")
	for _, path := range []string{a, b} {
		if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	old := time.Now().Add(-2 * time.Hour)
	recent := time.Now().Add(-time.Hour)
	if err := os.Chtimes(a, old, old); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(b, recent, recent); err != nil {
		t.Fatal(err)
	}

	m := newImportTestModel(t)
	m.lastImportDir = dir
	m.openFileBrowser()
	if got := m.fileBrowserItems[1].name; got != "a.json" {
		t.Fatalf("name sort first file = %q, want a.json", got)
	}

	m.handleFileBrowserKey(tea.KeyMsg{}, "s")
	if got := m.fileBrowserItems[1].name; got != "b.csv" {
		t.Fatalf("time sort first file = %q, want b.csv", got)
	}
}

func TestFileBrowserSortModePersists(t *testing.T) {
	isolateTestHome(t)
	m := newImportTestModel(t)
	m.setFileBrowserSortMode(1)

	next := newImportTestModel(t)
	if next.fileBrowserSortMode != 1 {
		t.Fatalf("sort mode = %d, want 1", next.fileBrowserSortMode)
	}
}

func TestFileBrowserFilterAndSortHotspots(t *testing.T) {
	isolateTestHome(t)
	m := newImportTestModel(t)
	m.lastImportDir = t.TempDir()
	m.openFileBrowser()

	plain := ansi.Strip(m.View())
	x, y := findCell(t, plain, "搜索:")
	m.handleMouse(leftClick(x+lipgloss.Width("搜索: ")+1, y))
	if !m.fileBrowserFilterFocused {
		t.Fatal("clicking the filter input should focus it")
	}

	plain = ansi.Strip(m.View())
	x, y = findCell(t, plain, "[S] 名称")
	m.handleMouse(leftClick(x, y))
	if m.fileBrowserSortMode != 1 {
		t.Fatalf("sort mode = %d, want 1", m.fileBrowserSortMode)
	}
}
