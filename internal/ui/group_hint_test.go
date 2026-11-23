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

func TestAddModalShowsExistingGroups(t *testing.T) {
	m := newGroupHintModel(t)
	m.openAddModal()

	plain := ansi.Strip(m.View())
	if !strings.Contains(plain, "已有分组") {
		t.Fatalf("add modal missing group hint:\n%s", plain)
	}
	// 生产、测试，以及默认组（有连接时才会出现，这里连接都属具名组，默认组应不存在）。
	for _, needle := range []string{"生产", "测试"} {
		if !strings.Contains(plain, needle) {
			t.Fatalf("add modal missing group %q:\n%s", needle, plain)
		}
	}
	if !strings.Contains(plain, "(2)") {
		t.Fatalf("add modal missing group count (2):\n%s", plain)
	}
}

func TestAddModalGroupHintHotspots(t *testing.T) {
	m := newGroupHintModel(t)
	m.openAddModal()
	_ = m.View() // 触发渲染以填充热区

	r := m.addModalRect
	if r.groupRow <= 0 || len(r.groupXs) == 0 {
		t.Fatalf("group hint hotspots not populated: %+v", r)
	}
	// 点击第一个分组应填入其名称（Groups() 对无 GroupOrder 的分组按名称排序）。
	first := r.groupVals[0]
	m.handleMouse(tea.MouseMsg{
		Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
		X: r.left + r.groupXs[0] + 1, Y: r.top + r.groupRow,
	})
	if got := m.mgrAddGroup.Value(); got != first {
		t.Fatalf("group after click = %q, want %q", got, first)
	}
}

func TestAddModalGroupCycleByKeyboard(t *testing.T) {
	m := newGroupHintModel(t)
	m.openAddModal()
	m.mgrAddFieldIdx = 1 // 分组字段
	vals := m.mgrAddGroupValues()

	m.handleKey(tea.KeyMsg{Type: tea.KeyRight})
	if got := m.mgrAddGroup.Value(); got != vals[0] {
		t.Fatalf("group after first right = %q, want %q", got, vals[0])
	}
	m.handleKey(tea.KeyMsg{Type: tea.KeyRight})
	if got := m.mgrAddGroup.Value(); got != vals[1] {
		t.Fatalf("group after second right = %q, want %q", got, vals[1])
	}
	m.handleKey(tea.KeyMsg{Type: tea.KeyLeft})
	if got := m.mgrAddGroup.Value(); got != vals[0] {
		t.Fatalf("group after left = %q, want %q", got, vals[0])
	}
}

func TestAddModalGroupHintHighlightsCurrent(t *testing.T) {
	m := newGroupHintModel(t)
	m.openAddModal()
	vals := m.mgrAddGroupValues()
	// 选中第一个分组，当前项应以高亮按钮样式渲染。
	m.mgrAddGroup.SetValue(vals[0])

	view := m.View()
	plain := ansi.Strip(view)
	if !strings.Contains(plain, "已有分组") {
		t.Fatalf("group hint missing:\n%s", plain)
	}
	// 高亮样式带背景色，原始 View 中该分组文本出现次数应多于纯文本一次
	// （styleBtnOK 与 styleBtnMuted 的差异通过 ANSI 转义体现）。
	selectedGroup := m.cs.Groups()[0]
	if !strings.Contains(plain, selectedGroup) {
		t.Fatalf("selected group %q not rendered:\n%s", selectedGroup, plain)
	}
	if strings.Contains(view, styleBtnOK.Render(" "+selectedGroup+" ")) == false {
		t.Fatalf("selected group not highlighted with styleBtnOK:\n%s", view)
	}
}

func TestAddModalGroupHintIncludesDefault(t *testing.T) {
	cs := &configstore.ConfigFile{
		Connections: []configstore.SavedConnection{{Name: "x", Host: "x.local", Group: ""}},
	}
	m := NewModel(cs, false)
	m.width = 100
	m.height = 40
	m.appMode = appManager
	m.openAddModal()

	plain := ansi.Strip(m.View())
	if !strings.Contains(plain, configstore.DefaultGroupName) {
		t.Fatalf("default group not shown:\n%s", plain)
	}
	// 点击默认组应填入空字符串，避免把"(默认)"当字面量存进配置。
	r := m.addModalRect
	for i := range r.groupVals {
		if r.groupVals[i] == "" {
			m.handleMouse(tea.MouseMsg{
				Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
				X: r.left + r.groupXs[i] + 1, Y: r.top + r.groupRow,
			})
			if got := m.mgrAddGroup.Value(); got != "" {
				t.Fatalf("default group filled as %q, want empty", got)
			}
			return
		}
	}
	t.Fatal("default group option not found in hotspots")
}
