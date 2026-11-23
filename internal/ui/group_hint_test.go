package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/yuwen002/meteor-ch-tui/internal/configstore"
)

func newGroupHintModel(t *testing.T) *Model {
	t.Helper()
	cs := &configstore.ConfigFile{
		Connections: []configstore.SavedConnection{
			{Name: "a", Group: "生产", Host: "a.local"},
			{Name: "b", Group: "生产", Host: "b.local"},
			{Name: "c", Group: "测试", Host: "c.local"},
		},
	}
	m := NewModel(cs, false)
	m.width = 100
	m.height = 40
	m.appMode = appManager
	return m
}

func openGroupPickerAtGroupField(t *testing.T) *Model {
	t.Helper()
	m := newGroupHintModel(t)
	m.openAddModal()
	m.mgrAddFieldIdx = 1 // 分组字段
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlG})
	if !m.mgrGroupPickerOpen {
		t.Fatal("Ctrl+G should open the group picker")
	}
	return m
}

func TestGroupPickerOpensAndListsGroups(t *testing.T) {
	m := openGroupPickerAtGroupField(t)
	plain := ansi.Strip(m.View())
	if !strings.Contains(plain, "选择分组") {
		t.Fatalf("picker title missing:\n%s", plain)
	}
	for _, needle := range []string{"生产", "测试", "(2)", "(1)"} {
		if !strings.Contains(plain, needle) {
			t.Fatalf("picker missing %q:\n%s", needle, plain)
		}
	}
}

func TestGroupPickerNotShownInlineBelowForm(t *testing.T) {
	m := newGroupHintModel(t)
	m.openAddModal()
	// 未打开弹层时，表单内不应再内嵌"已有分组"快捷行。
	plain := ansi.Strip(m.View())
	if strings.Contains(plain, "已有分组") {
		t.Fatalf("inline group hint should be removed:\n%s", plain)
	}
}

func TestGroupPickerKeyboardSelectAndConfirm(t *testing.T) {
	m := openGroupPickerAtGroupField(t)
	vals := m.groupPickerValues()

	m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	if m.mgrGroupPickerSel != 1 {
		t.Fatalf("sel = %d, want 1", m.mgrGroupPickerSel)
	}
	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if m.mgrGroupPickerOpen {
		t.Fatal("picker should close after confirm")
	}
	if got := m.mgrAddGroup.Value(); got != vals[1] {
		t.Fatalf("group = %q, want %q", got, vals[1])
	}
}

func TestGroupPickerEscKeepsValue(t *testing.T) {
	m := openGroupPickerAtGroupField(t)
	m.mgrAddGroup.SetValue("自定义")
	m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if m.mgrGroupPickerOpen {
		t.Fatal("picker should close on Esc")
	}
	if got := m.mgrAddGroup.Value(); got != "自定义" {
		t.Fatalf("value changed after cancel: %q", got)
	}
}

func TestGroupPickerMouseClickFills(t *testing.T) {
	m := openGroupPickerAtGroupField(t)
	_ = m.View() // 填充热区
	r := m.mgrGroupPickerRect
	if r.w == 0 || r.h == 0 {
		t.Fatal("picker rect not populated")
	}
	vals := m.groupPickerValues()
	idx := 0
	for i, v := range vals {
		if v == "测试" {
			idx = i
			break
		}
	}
	m.handleMouse(tea.MouseMsg{
		Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
		X: r.left + 4, Y: r.top + r.listStart + idx,
	})
	if m.mgrGroupPickerOpen {
		t.Fatal("picker should close after click")
	}
	if got := m.mgrAddGroup.Value(); got != "测试" {
		t.Fatalf("group after click = %q, want 测试", got)
	}
}

func TestGroupPickerDefaultMapsToEmpty(t *testing.T) {
	cs := &configstore.ConfigFile{
		Connections: []configstore.SavedConnection{{Name: "x", Host: "x.local", Group: ""}},
	}
	m := NewModel(cs, false)
	m.width = 100
	m.height = 40
	m.appMode = appManager
	m.openAddModal()
	m.mgrAddFieldIdx = 1
	m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlG})

	vals := m.groupPickerValues()
	// 默认组位于 vals 中，其值为空字符串。
	found := false
	for _, v := range vals {
		if v == "" {
			found = true
		}
	}
	if !found {
		t.Fatalf("default group not mapped to empty: %+v", vals)
	}
	plain := ansi.Strip(m.View())
	if !strings.Contains(plain, configstore.DefaultGroupName) {
		t.Fatalf("default group not shown in picker:\n%s", plain)
	}
}
