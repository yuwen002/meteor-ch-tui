package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestHelpOverlayFromManager(t *testing.T) {
	m := newImportTestModel(t)
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	if !m.helpOpen {
		t.Fatal("? should open the help overlay")
	}

	plain := ansi.Strip(m.View())
	for _, needle := range []string{"作者：浮笙", "yuwen002@163.com", "[Esc] 关闭"} {
		findCell(t, plain, needle)
	}

	x, y := findCell(t, plain, "[Esc] 关闭")
	m.handleMouse(leftClick(x, y))
	if m.helpOpen {
		t.Fatal("clicking [Esc] 关闭 should close the help overlay")
	}
}

func TestHelpOverlayFromMainWithF1(t *testing.T) {
	m := newImportTestModel(t)
	m.appMode = appMain
	m.initMainMode()

	m.handleKey(tea.KeyMsg{Type: tea.KeyF1})
	if !m.helpOpen {
		t.Fatal("F1 should open the help overlay")
	}
	m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if m.helpOpen {
		t.Fatal("Esc should close the help overlay")
	}
}
