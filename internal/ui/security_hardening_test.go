package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func openExportModalForTest(t *testing.T) *Model {
	t.Helper()
	m := newImportTestModel(t)
	m.openMgrExportModal()
	return m
}

func TestExportPasswordsRequiresConfirm(t *testing.T) {
	m := openExportModalForTest(t)
	if m.mgrExportPasswords {
		t.Fatal("passwords should start disabled in export modal")
	}

	m.toggleMgrImportOption()
	if m.confirm.kind != confirmExportPasswords {
		t.Fatalf("confirm kind = %v, want confirmExportPasswords", m.confirm.kind)
	}
	if m.mgrExportPasswords {
		t.Fatal("passwords should stay off before confirming")
	}

	plain := ansi.Strip(m.View())
	for _, needle := range []string{"安全提醒", "确定导出包含明文密码的连接配置吗？", "请确认导出用途"} {
		if !strings.Contains(plain, needle) {
			t.Fatalf("confirm modal missing %q:\n%s", needle, plain)
		}
	}
}

func TestExportPasswordsConfirmAcceptEnablesAndPreviews(t *testing.T) {
	isolateTestHome(t)
	m := openExportModalForTest(t)

	m.toggleMgrImportOption()
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if m.confirm.kind != confirmNone {
		t.Fatal("confirm should be cleared after accepting")
	}
	if !m.mgrExportPasswords {
		t.Fatal("passwords should be enabled after accepting")
	}
	if strings.Contains(m.mgrImportText.Value(), "<redacted>") {
		t.Fatalf("preview still redacted after enabling passwords: %q", m.mgrImportText.Value())
	}
}

func TestExportPasswordsConfirmCancelKeepsDisabled(t *testing.T) {
	m := openExportModalForTest(t)

	m.toggleMgrImportOption()
	m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if m.confirm.kind != confirmNone {
		t.Fatal("confirm should be cleared after cancel")
	}
	if m.mgrExportPasswords {
		t.Fatal("passwords should stay disabled after cancel")
	}
}

func TestExportPasswordsCanDisableWithoutConfirm(t *testing.T) {
	m := openExportModalForTest(t)
	m.mgrExportPasswords = true

	m.toggleMgrImportOption()
	if m.confirm.kind != confirmNone {
		t.Fatal("turning passwords off should not require confirm")
	}
	if m.mgrExportPasswords {
		t.Fatal("passwords should be disabled")
	}
}

func TestExportPasswordsConfirmFooterHotspot(t *testing.T) {
	isolateTestHome(t)
	m := openExportModalForTest(t)
	m.toggleMgrImportOption()

	plain := ansi.Strip(m.View())
	x, y := findCell(t, plain, "[Enter/Y] 确认")
	m.handleMouse(leftClick(x, y))
	if !m.mgrExportPasswords {
		t.Fatal("clicking confirm button should enable passwords")
	}
}
