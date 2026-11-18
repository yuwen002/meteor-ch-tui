package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestEndToEndSmoke(t *testing.T) {
	home := isolateTestHome(t)
	m := newImportTestModel(t)

	m.handleKey(tea.KeyMsg{Type: tea.KeyTab})
	if !m.mgrImportModal {
		t.Fatal("tab did not open the import modal")
	}
	m.mgrImportText.SetValue("名称,主机,端口\nnew,new.local,9001\n")
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlO})
	if m.mgrImportModal || len(m.cs.Connections) != 2 {
		t.Fatalf("import failed: modal=%v connections=%+v", m.mgrImportModal, m.cs.Connections)
	}

	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlE})
	if !m.mgrImportModal || !m.mgrExportMode {
		t.Fatal("ctrl+e did not open the config export modal")
	}
	m.exportPathInput.SetValue(home)
	for i := 0; i < 3; i++ {
		m.handleKey(tea.KeyMsg{Type: tea.KeyRight})
	}
	if m.mgrExportFormat != configExportXLSX {
		t.Fatalf("export format = %d, want xlsx", m.mgrExportFormat)
	}
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if m.mgrImportModal || !m.statusOK {
		t.Fatalf("xlsx export failed: modal=%v status=%q", m.mgrImportModal, m.status)
	}
	exports, err := filepath.Glob(filepath.Join(home, "connections_*.xlsx"))
	if err != nil || len(exports) != 1 {
		t.Fatalf("xlsx exports = %v, err = %v", exports, err)
	}

	m.lastImportDir = home
	m.handleKey(tea.KeyMsg{Type: tea.KeyTab})
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlB})
	if !m.fileBrowserOpen {
		t.Fatal("ctrl+b did not open the file browser")
	}
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	if !m.fileBrowserFilterFocused {
		t.Fatal("/ did not focus the file browser filter")
	}
	for _, r := range "connections" {
		m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if len(m.fileBrowserItems) != 2 || !strings.HasPrefix(m.fileBrowserItems[1].name, "connections_") {
		t.Fatalf("filtered items = %+v", m.fileBrowserItems)
	}
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if m.fileBrowserOpen || m.sheetPickerOpen {
		t.Fatalf("file still open: browser=%v sheetPicker=%v", m.fileBrowserOpen, m.sheetPickerOpen)
	}
	if value := m.mgrImportText.Value(); !strings.Contains(value, "new.local") {
		t.Fatalf("import text = %q", value)
	}

	if _, err := os.Stat(exports[0]); err != nil {
		t.Fatalf("export file missing: %v", err)
	}
}
