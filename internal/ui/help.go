package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	appAuthor = "浮笙"
	appEmail  = "yuwen002@163.com"
)

func helpContentLines() []string {
	return []string{
		" 连接管理器",
		"   ↑/↓ 选择连接 · Enter 连接 · + 新建 · e 编辑 · Del 删除 · g 分组",
		"   Tab/Ctrl+I 导入配置 · Ctrl+E 导出配置",
		"",
		" 查询界面",
		"   Tab/Shift+Tab 切换焦点 · Ctrl+N 打开连接配置",
		"   Alt+Enter/Ctrl+S/F5 执行 SQL · Ctrl+R 活跃查询",
		"   Ctrl+Y 复制 TSV · Ctrl+E 导出结果 · Ctrl+H 历史",
		"   鼠标右键菜单 · Ctrl+Q 返回管理器",
		"",
		" 导入与导出",
		"   Ctrl+O 导入/复制 · Enter 保存文件 · Tab 覆盖开关/包含密码",
		"   ←/→ 切换配置导出格式 · Ctrl+P 输入保存目录 · Ctrl+B 浏览",
		"   路径支持 ~、$VAR、%VAR%；Excel 多工作表时可选择工作表",
		"",
		" 文件浏览器",
		"   ↑/↓ 选择 · Enter 进入/选择 · Backspace 返回上级",
		"   / 或 Ctrl+F 搜索 · s 切换名称/时间排序 · Esc 关闭",
		"",
		" 关于我们",
		"   作者：" + appAuthor,
		"   邮箱：" + appEmail,
	}
}

func (m *Model) openHelp() {
	m.helpOpen = true
	m.helpScroll = 0
}

func (m *Model) closeHelp() {
	m.helpOpen = false
	m.helpScroll = 0
	m.helpRect = modalRect{}
}

func helpVisibleLines() int {
	return 22
}

func (m *Model) helpMaxScroll() int {
	return maxInt(0, len(helpContentLines())-helpVisibleLines())
}

func (m *Model) moveHelp(delta int) {
	m.helpScroll = clampInt(m.helpScroll+delta, 0, m.helpMaxScroll())
}

func (m *Model) handleHelpKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc", "?", "f1":
		m.closeHelp()
	case "up", "k", "pgup":
		m.moveHelp(-1)
	case "down", "j", "pgdown":
		m.moveHelp(1)
	}
	return m, nil
}

func (m *Model) handleHelpMouse(msg tea.MouseMsg, isClick, isWheelUp, isWheelDown bool) (tea.Model, tea.Cmd) {
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonRight {
		m.closeHelp()
		return m, nil
	}
	if isWheelUp || isWheelDown {
		if isWheelUp {
			m.moveHelp(-1)
		} else {
			m.moveHelp(1)
		}
		return m, nil
	}
	if !isClick {
		return m, nil
	}
	r := m.helpRect
	if r.w == 0 || r.h == 0 || msg.X < r.left || msg.X >= r.left+r.w || msg.Y < r.top || msg.Y >= r.top+r.h {
		return m, nil
	}
	relX, relY := msg.X-r.left, msg.Y-r.top
	if _, ok := hitFooter(relX, relY, r, footerRoleCancel); ok {
		m.closeHelp()
	}
	return m, nil
}

func (m *Model) renderHelp(bg string) string {
	innerW := minInt(88, m.width-8)
	if innerW < 40 {
		innerW = maxInt(30, m.width-4)
	}

	title := styleTitle.Render(" 快捷键帮助 ")
	hint := styleMuted.Render(" ↑↓/滚轮 翻页 · Esc 关闭 ")
	lines := []string{
		padCell(title, maxInt(0, innerW-lipgloss.Width(hint))) + hint,
		"",
	}

	content := helpContentLines()
	visible := helpVisibleLines()
	start := clampInt(m.helpScroll, 0, m.helpMaxScroll())
	end := minInt(start+visible, len(content))
	for i := start; i < end; i++ {
		lines = append(lines, padCell(content[i], innerW))
	}
	for len(lines) < 2+visible {
		lines = append(lines, strings.Repeat(" ", innerW))
	}
	lines = append(lines, "")

	btns := []footerButton{
		{label: styleBtnMuted.Render("[Esc] 关闭"), role: footerRoleCancel},
	}
	footer, positions := footerLayout(innerW, footerAlignCenter, btns)
	footerContentRow := len(lines)
	lines = append(lines, footer)

	box := styleModal.Width(innerW).Height(len(lines)).Render(strings.Join(lines, "\n"))
	out, rect := overlayRect(bg, box, m.width)
	applyFooter(&rect, footerContentRow, btns, positions)
	m.helpRect = rect
	return out
}
