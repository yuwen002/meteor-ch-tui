package ui

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/yuwen002/meteor-ch-tui/internal/configstore"
)

func newBatchTestModel(t *testing.T) *Model {
	t.Helper()
	cs := &configstore.ConfigFile{
		Connections: []configstore.SavedConnection{
			{Name: "c1", Group: "g1", Host: "c1.local", Port: 9000},
			{Name: "c2", Group: "g1", Host: "c2.local", Port: 9000},
			{Name: "c3", Group: "g2", Host: "c3.local", Port: 9000},
		},
	}
	m := NewModel(cs, false)
	m.width = 100
	m.height = 40
	m.appMode = appManager
	return m
}

func batchRunes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestMarkToggleAndSelectAll(t *testing.T) {
	m := newBatchTestModel(t)
	m.mgrGroupIdx = 0
	m.mgrConnIdx = 0

	m.handleKey(batchRunes(" "))
	if m.markedCount() != 1 {
		t.Fatalf("markedCount after mark = %d, want 1", m.markedCount())
	}
	if !strings.Contains(m.status, "已选 1 个连接") {
		t.Fatalf("status = %q", m.status)
	}

	m.handleKey(batchRunes(" "))
	if m.markedCount() != 0 {
		t.Fatalf("markedCount after unmark = %d, want 0", m.markedCount())
	}

	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlA})
	if m.markedCount() != 3 {
		t.Fatalf("markedCount after ctrl+a = %d, want 3", m.markedCount())
	}

	m.clearConnMarks()
	if m.markedCount() != 0 {
		t.Fatalf("markedCount after clear = %d, want 0", m.markedCount())
	}
}

func TestMarkShownInManagerList(t *testing.T) {
	m := newBatchTestModel(t)
	m.mgrGroupIdx = 0
	m.mgrConnIdx = 0
	m.handleKey(batchRunes(" "))

	plain := ansi.Strip(m.View())
	if !strings.Contains(plain, "[x] c1") {
		t.Fatalf("marked row not shown in manager list:\n%s", plain)
	}
}

func TestBatchDeleteViaConfirm(t *testing.T) {
	isolateTestHome(t)
	m := newBatchTestModel(t)
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlA})

	m.handleKey(tea.KeyMsg{Type: tea.KeyDelete})
	if m.confirm.kind != confirmBatchDeleteConns {
		t.Fatalf("confirm kind = %v, want confirmBatchDeleteConns", m.confirm.kind)
	}
	if len(m.confirm.names) != 3 {
		t.Fatalf("confirm names = %v, want 3 entries", m.confirm.names)
	}

	m.handleKey(batchRunes("y"))
	if len(m.cs.Connections) != 0 {
		t.Fatalf("connections after batch delete = %+v, want empty", m.cs.Connections)
	}
	if m.markedCount() != 0 {
		t.Fatal("marks should be cleared after batch delete")
	}

	data, err := os.ReadFile(configstore.DefaultPath())
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}
	if strings.Contains(string(data), "c1.local") {
		t.Fatalf("deleted connection still present in config: %s", data)
	}
}

func TestBatchDeleteCancelKeepsConnections(t *testing.T) {
	isolateTestHome(t)
	m := newBatchTestModel(t)
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlA})
	m.handleKey(tea.KeyMsg{Type: tea.KeyDelete})
	m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	if m.confirm.kind != confirmNone {
		t.Fatal("confirm should be cleared after cancel")
	}
	if len(m.cs.Connections) != 3 {
		t.Fatalf("connections after cancel = %d, want 3", len(m.cs.Connections))
	}
}

func TestBatchMoveConnections(t *testing.T) {
	isolateTestHome(t)
	m := newBatchTestModel(t)

	m.mgrGroupIdx = 0
	m.mgrConnIdx = 0
	m.handleKey(batchRunes(" "))
	m.handleKey(batchRunes("m"))
	if !m.mgrBatchMoveOpen {
		t.Fatal("m should open the batch move modal")
	}

	plain := ansi.Strip(m.View())
	for _, needle := range []string{"移动连接到分组", "g2", "[Enter] 移动", "[Esc] 取消"} {
		if !strings.Contains(plain, needle) {
			t.Fatalf("batch move modal missing %q:\n%s", needle, plain)
		}
	}

	if m.batchMoveTarget() != "g1" {
		t.Fatalf("initial move target = %q, want g1", m.batchMoveTarget())
	}
	m.handleKey(batchRunes("j"))
	if m.batchMoveTarget() != "g2" {
		t.Fatalf("move target after down = %q, want g2", m.batchMoveTarget())
	}

	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if m.mgrBatchMoveOpen {
		t.Fatal("batch move modal should close after confirming")
	}
	c1 := m.cs.Find("c1")
	if c1 == nil || c1.Group != "g2" {
		t.Fatalf("c1 group after move = %+v, want g2", c1)
	}
	if m.markedCount() != 0 {
		t.Fatal("marks should be cleared after batch move")
	}

	reloaded, err := configstore.Load()
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}
	if got := reloaded.Find("c1"); got == nil || got.Group != "g2" {
		t.Fatalf("persisted c1 = %+v, want group g2", got)
	}
}

func TestBatchMoveModalFooterHotspots(t *testing.T) {
	isolateTestHome(t)
	m := newBatchTestModel(t)
	m.mgrGroupIdx = 0
	m.mgrConnIdx = 0
	m.handleKey(batchRunes(" "))
	m.handleKey(batchRunes("m"))

	plain := ansi.Strip(m.View())
	x, y := findCell(t, plain, "[Esc] 取消")
	m.handleMouse(leftClick(x, y))
	if m.mgrBatchMoveOpen {
		t.Fatal("clicking [Esc] 取消 should close the batch move modal")
	}
	if len(m.cs.Connections) != 3 {
		t.Fatal("cancelling batch move should not change connections")
	}
}
