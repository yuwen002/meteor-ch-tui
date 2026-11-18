package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/yuwen002/meteor-ch-tui/internal/configstore"
)

func newImportTestModel(t *testing.T) *Model {
	t.Helper()
	cs := &configstore.ConfigFile{
		Connections: []configstore.SavedConnection{
			{Name: "old", Host: "old.local", Port: 9000},
		},
	}
	m := NewModel(cs, false)
	m.width = 100
	m.height = 40
	m.mgrAddModal = false
	return m
}

func findCell(t *testing.T, plain, needle string) (int, int) {
	t.Helper()
	idx := strings.Index(plain, needle)
	if idx < 0 {
		t.Fatalf("rendered view does not contain %q", needle)
	}
	before := plain[:idx]
	lineStart := strings.LastIndex(before, "\n") + 1
	return ansi.StringWidth(before[lineStart:]), strings.Count(before, "\n")
}

func leftClick(x, y int) tea.MouseMsg {
	return tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      x,
		Y:      y,
	}
}

func isolateTestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}

func TestCommitMgrImportPersistsOnSuccess(t *testing.T) {
	isolateTestHome(t)

	m := newImportTestModel(t)
	m.openMgrImportModal()
	m.mgrImportText.SetValue(`{"connections":[{"name":"new","host":"new.local","port":9001}]}`)

	m.commitMgrImport()

	if m.mgrImportModal {
		t.Fatal("import modal should close after a successful save")
	}
	if len(m.cs.Connections) != 2 {
		t.Fatalf("connections = %+v, want 2 entries", m.cs.Connections)
	}
	if !m.statusOK || !strings.Contains(m.status, "已导入 1 个连接") {
		t.Fatalf("status = %q, ok = %v", m.status, m.statusOK)
	}
	data, err := os.ReadFile(configstore.DefaultPath())
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}
	if !strings.Contains(string(data), "new.local") {
		t.Fatalf("saved config does not contain imported connection: %s", data)
	}
}

func TestCommitMgrImportSaveFailureKeepsOriginalState(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", blocker)
	t.Setenv("USERPROFILE", blocker)

	m := newImportTestModel(t)
	m.openMgrImportModal()
	m.mgrImportText.SetValue(`{"connections":[{"name":"new","host":"new.local","port":9001}]}`)

	m.commitMgrImport()

	if len(m.cs.Connections) != 1 || m.cs.Connections[0].Name != "old" {
		t.Fatalf("connections mutated after failed save: %+v", m.cs.Connections)
	}
	if !m.mgrImportModal {
		t.Fatal("import modal should stay open after a failed save")
	}
	if m.statusOK || !strings.Contains(m.status, "保存失败") {
		t.Fatalf("status = %q, ok = %v", m.status, m.statusOK)
	}
}

func TestMgrImportFooterHotspots(t *testing.T) {
	m := newImportTestModel(t)
	m.openMgrImportModal()

	plain := ansi.Strip(m.View())
	x, y := findCell(t, plain, "[Ctrl+B] 选择文件")
	m.handleMouse(leftClick(x, y))
	if !m.fileBrowserOpen {
		t.Fatal("clicking [Ctrl+B] 选择文件 should open the file browser")
	}

	m.closeFileBrowser()
	plain = ansi.Strip(m.View())
	x, y = findCell(t, plain, "[Esc] 关闭")
	m.handleMouse(leftClick(x, y))
	if m.mgrImportModal {
		t.Fatal("clicking [Esc] 关闭 should close the import modal")
	}
}
