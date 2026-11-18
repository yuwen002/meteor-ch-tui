package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func TestFooterLayoutAndHitRoles(t *testing.T) {
	tests := []struct {
		name  string
		align footerAlign
		btns  []footerButton
		roles []footerHitRole
	}{
		{
			name:  "two buttons left and right",
			align: footerAlignFirstLeftRestRight,
			btns: []footerButton{
				{label: "[Enter] 使用", role: footerRoleOK},
				{label: "[Esc] 关闭", role: footerRoleCancel},
			},
			roles: []footerHitRole{footerRoleOK, footerRoleCancel},
		},
		{
			name:  "three buttons",
			align: footerAlignFirstLeftRestRight,
			btns: []footerButton{
				{label: "[Ctrl+O] 导入", role: footerRoleOK},
				{label: "[Ctrl+B] 选择文件", role: footerRoleExtra},
				{label: "[Esc] 关闭", role: footerRoleCancel},
			},
			roles: []footerHitRole{footerRoleOK, footerRoleExtra, footerRoleCancel},
		},
		{
			name:  "four buttons with reused fields",
			align: footerAlignFirstLeftRestRight,
			btns: []footerButton{
				{label: "[Ctrl+R] 重命名", role: footerRoleTest},
				{label: "[Enter] 新建", role: footerRoleOK},
				{label: "[Ctrl+D] 删除", role: footerRoleCancel},
				{label: "[Esc] 取消", role: footerRoleExtra},
			},
			roles: []footerHitRole{footerRoleTest, footerRoleOK, footerRoleCancel, footerRoleExtra},
		},
		{
			name:  "centered",
			align: footerAlignCenter,
			btns: []footerButton{
				{label: "[Enter/Y] 确认", role: footerRoleOK},
				{label: "[Esc/N] 取消", role: footerRoleCancel},
			},
			roles: []footerHitRole{footerRoleOK, footerRoleCancel},
		},
		{
			name:  "last right with padded hotspots",
			align: footerAlignLastRight,
			btns: []footerButton{
				{label: "[Enter] 连接", role: footerRoleOK, hotRight: 2},
				{label: "[Esc] 取消", role: footerRoleCancel, hotLeft: 2, hotRight: 2},
			},
			roles: []footerHitRole{footerRoleOK, footerRoleCancel},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, positions := footerLayout(60, tt.align, tt.btns)
			rect := modalRect{}
			applyFooter(&rect, 0, tt.btns, positions)

			for _, btn := range tt.btns {
				x, y := findCell(t, text, btn.label)
				if y != 0 {
					t.Fatalf("footer is not one line: %q", text)
				}
				role, ok := hitFooter(x+1, y+1, rect, tt.roles...)
				if !ok || role != btn.role {
					t.Fatalf("click %q at x=%d hit (%v, %v), want %v", btn.label, x, role, ok, btn.role)
				}
			}

			if tt.name == "last right with padded hotspots" {
				x, _ := findCell(t, text, tt.btns[0].label)
				padX := x + lipgloss.Width(tt.btns[0].label) + 2
				if role, ok := hitFooter(padX, 1, rect, tt.roles...); !ok || role != footerRoleOK {
					t.Fatalf("padded OK area hit (%v, %v)", role, ok)
				}
				cx, _ := findCell(t, text, tt.btns[1].label)
				if role, ok := hitFooter(cx, 1, rect, tt.roles...); !ok || role != footerRoleCancel {
					t.Fatalf("padded cancel area hit (%v, %v)", role, ok)
				}
			}
		})
	}
}

func TestHistoryModalFooterHotspots(t *testing.T) {
	m := newImportTestModel(t)
	m.queryHistory = []string{"SELECT 1"}
	m.historySel = 0
	bg := strings.Repeat(strings.Repeat(" ", m.width)+"\n", m.height)

	out := m.renderHistoryModal(bg)
	plain := ansi.Strip(out)

	for _, tc := range []struct {
		label string
		role  footerHitRole
	}{
		{"[Enter] 使用", footerRoleOK},
		{"[Esc] 关闭", footerRoleCancel},
	} {
		x, y := findCell(t, plain, tc.label)
		role, ok := hitFooter(x-m.historyRect.left, y-m.historyRect.top, m.historyRect, footerRoleOK, footerRoleCancel)
		if !ok || role != tc.role {
			t.Fatalf("click %q at (%d,%d) hit (%v, %v), want %v", tc.label, x, y, role, ok, tc.role)
		}
	}
}
