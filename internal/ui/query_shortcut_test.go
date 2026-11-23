package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSQLExecuteShortcutDoesNotUseAltEnter(t *testing.T) {
	m := newImportTestModel(t)
	m.initMainMode()
	m.appMode = appMain
	m.focus = focusEditor
	m.editor.SetValue("SELECT 1")

	// alt+enter 不再执行 SQL：状态不应变为"执行中"。
	altEnter := tea.KeyMsg{Type: tea.KeyEnter, Alt: true}
	if altEnter.String() != "alt+enter" {
		t.Fatalf("alt+enter key string = %q, want alt+enter", altEnter.String())
	}
	m.handleKey(altEnter)
	if m.status == "执行中..." {
		t.Fatal("alt+enter should not execute SQL")
	}

	// ctrl+s 仍可执行（无 client 时 report "尚未连接 ClickHouse"，但说明已进入执行逻辑）。
	m.editor.SetValue("SELECT 1")
	m.status = ""
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.status != "执行中..." && m.status != "尚未连接 ClickHouse" {
		t.Fatalf("ctrl+s should execute SQL, status = %q", m.status)
	}
}
