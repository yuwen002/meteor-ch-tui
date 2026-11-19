package ui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
)

func columnTestModel(t *testing.T) *Model {
	t.Helper()
	isolateTestHome(t)
	m := newImportTestModel(t)
	m.appMode = appMain
	m.initMainMode()
	m.focus = focusTable
	m.table.setResult(&ch.Result{
		Columns: []string{"name", "value", "extra"},
		Rows:    [][]any{{"alpha", 1, "x"}},
	})
	m.applyResultLayout(&m.table)
	return m
}

func TestColumnModalToggleAndPersist(t *testing.T) {
	m := columnTestModel(t)

	m.handleTableKey("c")
	if !m.columnModalOpen {
		t.Fatal("c should open the column modal")
	}
	m.handleColumnModalKey("down")
	m.handleColumnModalKey(" ")

	if !m.table.hidden["value"] {
		t.Fatalf("hidden columns = %v, want value hidden", m.table.hiddenList())
	}
	if got := m.table.visibleColumns(200); !reflect.DeepEqual(got, []int{0, 2}) {
		t.Fatalf("visibleColumns() = %v, want [0 2]", got)
	}

	before := m.table.widths[1]
	m.handleColumnModalKey("]")
	if m.table.widths[1] <= before {
		t.Fatalf("column width = %d, want > %d", m.table.widths[1], before)
	}
	persistedWidth := m.table.widths[1]

	m.handleColumnModalKey("esc")
	if m.columnModalOpen {
		t.Fatal("esc should close the column modal")
	}

	next := newImportTestModel(t)
	next.appMode = appMain
	next.initMainMode()
	next.focus = focusTable
	next.table.setResult(&ch.Result{
		Columns: []string{"name", "value", "extra"},
		Rows:    [][]any{{"alpha", 1, "x"}},
	})
	next.applyResultLayout(&next.table)
	if !next.table.hidden["value"] {
		t.Fatal("hidden column was not restored")
	}
	if next.table.widths[1] != persistedWidth {
		t.Fatalf("restored width = %d, want %d", next.table.widths[1], persistedWidth)
	}
}

func TestColumnModalFooterHotspots(t *testing.T) {
	m := columnTestModel(t)
	m.handleTableKey("c")
	m.handleColumnModalKey("down")
	m.handleColumnModalKey(" ")

	plain := ansi.Strip(m.View())
	x, y := findCell(t, plain, "[A] 全部显示")
	m.handleMouse(leftClick(x, y))
	if len(m.table.hiddenList()) != 0 {
		t.Fatalf("hidden columns after show-all = %v", m.table.hiddenList())
	}

	plain = ansi.Strip(m.View())
	x, y = findCell(t, plain, "[Esc] 关闭")
	m.handleMouse(leftClick(x, y))
	if m.columnModalOpen {
		t.Fatal("clicking [Esc] 关闭 should close the column modal")
	}
}

func TestHiddenColumnsDoNotAffectExport(t *testing.T) {
	m := columnTestModel(t)
	m.table.toggleColumn("extra")

	data, err := m.resultTSV()
	if err != nil {
		t.Fatalf("resultTSV() error = %v", err)
	}
	if !strings.Contains(data, "extra") {
		t.Fatalf("hidden column missing from export: %q", data)
	}
}
