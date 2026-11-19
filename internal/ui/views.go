package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
)

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "初始化中..."
	}

	if m.appMode == appManager {
		out := m.managerView()
		if m.mgrBatchMoveOpen {
			out = m.renderBatchMoveModal(out)
		}
		if m.confirm.kind != confirmNone {
			out = m.renderConfirmModal(out)
		}
		if m.dirBrowserOpen {
			out = m.renderDirBrowser(out)
		}
		if m.fileBrowserOpen {
			out = m.renderFileBrowser(out)
		}
		if m.sheetPickerOpen {
			out = m.renderSheetPicker(out)
		}
		if m.helpOpen {
			out = m.renderHelp(out)
		}
		return out
	}

	header := m.headerView()
	body := lipgloss.JoinHorizontal(lipgloss.Top, m.sidebarView(), m.rightView())
	status := m.statusView()

	out := lipgloss.JoinVertical(lipgloss.Left, header, body, status)
	if m.rowModal {
		out = m.renderRowModal(out)
	}
	if m.columnModalOpen {
		out = m.renderColumnModal(out)
	}
	if m.cfgModal {
		out = m.renderCfgModal(out)
	}
	if m.confirm.kind != confirmNone {
		out = m.renderConfirmModal(out)
	}
	if m.historyModal {
		out = m.renderHistoryModal(out)
	}
	if m.resultExportModal {
		out = m.renderResultExportModal(out)
	}
	if m.ctxMenuOpen {
		out = m.renderContextMenu(out)
	}
	if m.dirBrowserOpen {
		out = m.renderDirBrowser(out)
	}
	if m.fileBrowserOpen {
		out = m.renderFileBrowser(out)
	}
	if m.sheetPickerOpen {
		out = m.renderSheetPicker(out)
	}
	if m.helpOpen {
		out = m.renderHelp(out)
	}
	return out
}

// ============ Manager 视图 ============

func (m *Model) managerView() string {
	header := m.managerHeader()
	list := m.renderMgrList(header)
	status := m.statusView()

	out := lipgloss.JoinVertical(lipgloss.Left, header, list, status)
	if m.mgrGroupModal {
		out = m.renderGroupModal(out)
	}
	if m.mgrAddModal {
		out = m.renderMgrAddModal(out)
	}
	if m.mgrImportModal {
		out = m.renderMgrImportModal(out)
	}
	return out
}

func (m *Model) managerHeader() string {
	left := styleTitle.Render(" meteor-ch-tui ") +
		styleMuted.Render(" 连接管理器")

	right := styleMuted.Render(fmt.Sprintf(" 双击/Enter 连接 · 工具栏可新建/编辑/删除/分组/导入/导出配置 · %d 个连接 · Ctrl+C 退出 ", len(m.cs.Connections)))
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	line := left + strings.Repeat(" ", gap) + right
	return styleHeader.Width(m.width).Render(line)
}

func (m *Model) renderMgrList(header string) string {
	headerLines := len(strings.Split(header, "\n"))
	innerW := m.width - 4
	innerH := m.height - 7 - (headerLines - 1)
	if innerH < 5 {
		innerH = 5
	}

	groups := m.cs.Groups()
	var lines []string
	mapping := make([]mgrLine, 0, len(m.cs.Connections)+len(groups)+2)
	add := func(s string, g, c int) {
		lines = append(lines, s)
		mapping = append(mapping, mgrLine{group: g, conn: c})
	}

	titleText := " 连接列表 "
	if innerW < 52 {
		titleText = " 列表 "
	}
	title := styleTitle.Render(titleText)
	newLabel, editLabel, delLabel, groupLabel := "[+] 新建", "[E] 编辑", "[Del] 删除", "[G] 分组"
	if innerW < 72 {
		newLabel, editLabel, delLabel, groupLabel = "[+]", "[E]", "[Del]", "[G]"
	}
	newBtn := styleBtnOK.Render(newLabel)
	editBtn := styleBtnMuted.Render(editLabel)
	delBtn := styleBtnMuted.Render(delLabel)
	groupBtn := styleBtnMuted.Render(groupLabel)
	importBtn := styleBtnMuted.Render("[I] 导入")
	exportBtn := styleBtnMuted.Render("[Ctrl+E] 导出配置")
	buttons := newBtn + " " + editBtn + " " + delBtn
	if innerW >= 52 {
		buttons += " " + groupBtn
	}
	if innerW >= 90 {
		buttons += " " + importBtn + " " + exportBtn
	}
	gap := maxInt(0, innerW-lipgloss.Width(title)-lipgloss.Width(buttons)-1)
	add(padCell(title+strings.Repeat(" ", gap)+buttons, innerW), -1, -1)
	m.mgrToolRow = headerLines + 1
	m.mgrToolNewX = 1 + lipgloss.Width(title) + gap
	m.mgrToolNewW = lipgloss.Width(newBtn)
	m.mgrToolEditX = m.mgrToolNewX + m.mgrToolNewW + 1
	m.mgrToolEditW = lipgloss.Width(editBtn)
	m.mgrToolDelX = m.mgrToolEditX + m.mgrToolEditW + 1
	m.mgrToolDelW = lipgloss.Width(delBtn)
	m.mgrToolGroupX = 0
	m.mgrToolGroupW = 0
	if innerW >= 52 {
		m.mgrToolGroupX = m.mgrToolDelX + m.mgrToolDelW + 1
		m.mgrToolGroupW = lipgloss.Width(groupBtn)
	}
	m.mgrToolImportX = 0
	m.mgrToolImportW = 0
	m.mgrToolExportX = 0
	m.mgrToolExportW = 0
	if innerW >= 90 {
		m.mgrToolImportX = m.mgrToolGroupX + m.mgrToolGroupW + 1
		m.mgrToolImportW = lipgloss.Width(importBtn)
		m.mgrToolExportX = m.mgrToolImportX + m.mgrToolImportW + 1
		m.mgrToolExportW = lipgloss.Width(exportBtn)
	}
	if len(groups) == 0 {
		add(styleMuted.Render(padCell("  (暂无连接，按 + 新建) ", innerW)), -1, -1)
	} else {
		for gIdx, g := range groups {
			add(styleMuted.Render(padCell(fmt.Sprintf(" 📁 %s (%d)", g, len(m.cs.ByGroup(g))), innerW)), gIdx, -1)
			conns := m.cs.ByGroup(g)
			for cIdx, c := range conns {
				selected := gIdx == m.mgrGroupIdx && cIdx == m.mgrConnIdx
				name := c.Name
				addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
				proto := ""
				if ch.IsHTTPProtocol(c.Protocol) {
					proto = " " + ch.ProtocolLabel(c.Protocol)
				}
				dbLabel := c.Database
				if dbLabel == "" {
					dbLabel = "全部"
				}
				mark := "  "
				if m.mgrMarked[c.Name] {
					mark = "[x]"
				}
				full := fmt.Sprintf(" %s %s [%s%s] db=%s user=%s", mark, name, addr, proto, dbLabel, c.Username)
				if lipgloss.Width(full) > innerW {
					full = ansi.Truncate(full, maxInt(1, innerW-3), "") + "..."
				}
				padded := padCell(full, innerW)
				if selected {
					add(styleRowSelected.Render(padded), gIdx, cIdx)
				} else {
					add(padded, gIdx, cIdx)
				}
			}
		}
	}

	// 让选中行始终落在视口内，并记录点击换算所需的行映射与首行 Y。
	sel := 0
	for i, ml := range mapping {
		if ml.group == m.mgrGroupIdx && ml.conn == m.mgrConnIdx && ml.conn >= 0 {
			sel = i
			break
		}
	}
	top := clampInt(sel-innerH/2, 0, maxInt(0, len(lines)-innerH))
	m.mgrLines = mapping
	m.mgrListTop = top
	// 头部可能因窄屏换行，列表上边框紧跟其后，内容首行 = 头部行数 + 1。
	m.mgrListFirstLineY = headerLines + 1

	content := strings.Join(padMgrLines(lines, innerW, innerH, top), "\n")
	return styleSidebar.Width(innerW).Height(innerH).BorderForeground(colAccent).Render(content)
}

func padMgrLines(lines []string, w, h, top int) []string {
	out := make([]string, 0, h)
	for i := 0; i < h; i++ {
		idx := top + i
		if idx < len(lines) {
			out = append(out, padCell(lines[idx], w))
		} else {
			out = append(out, strings.Repeat(" ", w))
		}
	}
	return out
}

func (m *Model) renderMgrAddModal(bg string) string {
	boxW := 58
	if boxW > m.width-8 {
		boxW = m.width - 8
	}
	innerW := boxW - 2

	titleTxt := " 新建连接 "
	if m.mgrEditModal {
		titleTxt = " 编辑连接 "
	}

	lines := make([]string, 0, 12)
	title := styleTitle.Render(titleTxt)
	hint := styleMuted.Render(" Tab 下一项 · Shift+Tab 上一项 · Enter 保存 · Esc 取消 ")
	titleLine := padCell(title, maxInt(0, innerW-lipgloss.Width(hint))) + hint
	lines = append(lines, titleLine)
	lines = append(lines, "")

	protoContentRow := -1
	for i, info := range mgrAddFieldsInfo {
		ti := info.ti(m)
		label := styleMuted.Render(padCell(info.name, 8))
		if info.name == "协议" {
			prompt := "  "
			if i == m.mgrAddFieldIdx {
				prompt = "▸ "
			}
			opt := styleBtnOK.Render(ch.ProtocolLabel(ti.Value()))
			note := styleWarnText.Render(protocolPortNote(ti.Value(), m.mgrAddPort.Value()))
			lines = append(lines, label+prompt+opt+styleMuted.Render("  ←/→ 切换")+note)
			protoContentRow = len(lines) - 1
			continue
		}
		if i == m.mgrAddFieldIdx {
			ti.Prompt = "▸ "
		} else {
			ti.Prompt = "  "
		}
		ti.Width = innerW - 10
		fieldView := ti.View()
		lines = append(lines, label+fieldView)
	}

	lines = append(lines, "")
	result := ""
	if m.mgrTesting {
		result = styleWarnText.Render(" 正在测试连接...")
	} else if m.mgrTestResult != "" {
		if m.mgrTestOK {
			result = styleOkText.Render(" " + m.mgrTestResult)
		} else {
			result = styleErrText.Render(" " + m.mgrTestResult)
		}
	}
	lines = append(lines, padCell(result, innerW))

	addBtns := []footerButton{
		{label: styleBtnMuted.Render("[Ctrl+T] 测试"), role: footerRoleTest},
		{label: styleBtnOK.Render("[Enter] 保存"), role: footerRoleOK},
		{label: styleBtnMuted.Render("[Esc] 取消"), role: footerRoleCancel},
	}
	footer, footerPositions := footerLayout(innerW, footerAlignFirstLeftRestRight, addBtns)
	footerContentRow := len(lines)
	lines = append(lines, footer)

	for i := range lines {
		lines[i] = padCell(lines[i], innerW)
	}
	content := strings.Join(lines, "\n")
	box := styleModal.Width(innerW).Height(len(lines)).Render(content)
	out, rect := overlayRect(bg, box, m.width)
	rect.fieldStart = 3
	rect.fieldCount = len(mgrAddFieldsInfo)
	rect.protocolRow = protoContentRow + 1
	applyFooter(&rect, footerContentRow, addBtns, footerPositions)
	m.addModalRect = rect
	return out
}

func (m *Model) renderMgrImportModal(bg string) string {
	boxW := maxInt(40, m.width-8)
	if boxW > 120 {
		boxW = 120
	}
	innerW := boxW - 2

	lines := make([]string, 0, 20)
	title := styleTitle.Render(" 导入连接配置 ")
	hint := styleMuted.Render(" 粘贴 JSON/CSV/TSV · Ctrl+B 选择文件 · Tab 覆盖开关 · Ctrl+O 导入 · Esc 关闭 ")
	formatRow, listCount := 0, 0
	if m.mgrExportMode {
		title = styleTitle.Render(" 导出连接配置 ")
		hint = styleMuted.Render(" ←/→ 格式 · Tab 密码 · Ctrl+P 目录 · Ctrl+B 浏览 · Enter 保存 · Ctrl+O 复制 · Esc 关闭 ")
	}
	titleLine := padCell(title, maxInt(0, innerW-lipgloss.Width(hint))) + hint
	lines = append(lines, titleLine)
	lines = append(lines, "")

	m.mgrImportText.SetWidth(innerW - 2)
	body := clipLines(m.mgrImportText.View(), 12)
	body = padToHeight(body, 12, innerW-2)
	fieldStart := len(lines) + 1
	fieldCount := 12
	for _, bodyLine := range strings.Split(body, "\n") {
		lines = append(lines, bodyLine)
	}
	lines = append(lines, "")

	if m.mgrExportMode {
		choice := configExportChoices[m.mgrExportFormat]
		formatText := styleBtnOK.Render(fmt.Sprintf(" [←/→] 格式: %s (.%s) ", choice.label, choice.ext))
		formatRow = len(lines) + 1
		listCount = 1
		lines = append(lines, formatText)
		lines = append(lines, "")
	}

	toggleText := "[Tab] 同名覆盖：关"
	toggleStyle := styleBtnMuted
	if m.mgrImportOverwrite {
		toggleText = "[Tab] 同名覆盖：开"
		toggleStyle = styleBtnOK
	}
	if m.mgrExportMode {
		toggleText = "[Tab] 包含密码：关"
		toggleStyle = styleBtnMuted
		if m.mgrExportPasswords {
			toggleText = "[Tab] 包含密码：开"
			toggleStyle = styleBtnOK
		}
	}
	toggle := toggleStyle.Render(toggleText)
	toggleRow := len(lines) + 1
	lines = append(lines, toggle)
	lines = append(lines, "")

	pathRow, pathX, pathW, browseX, browseW := 0, 0, 0, 0, 0
	if m.mgrExportMode {
		pathLine, x, w, bx, bw := m.exportPathLine(innerW)
		pathRow = len(lines) + 1
		pathX, pathW, browseX, browseW = x, w, bx, bw
		lines = append(lines, pathLine)
		lines = append(lines, "")
	}

	okLabel := "[Ctrl+O] 导入"
	if m.mgrExportMode {
		okLabel = "[Enter] 保存文件"
	}
	ok := styleBtnOK.Render(okLabel)
	cancel := styleBtnMuted.Render("[Esc] 关闭")
	extraLabel := "[Ctrl+B] 选择文件"
	if m.mgrExportMode {
		extraLabel = "[Ctrl+O] 复制"
	}
	importBtns := []footerButton{
		{label: ok, role: footerRoleOK},
		{label: styleBtnMuted.Render(extraLabel), role: footerRoleExtra},
		{label: cancel, role: footerRoleCancel},
	}
	footer, footerPositions := footerLayout(innerW, footerAlignFirstLeftRestRight, importBtns)
	footerContentRow := len(lines)
	lines = append(lines, footer)

	content := strings.Join(lines, "\n")
	box := styleModal.Width(innerW).Height(len(lines)).Render(content)
	out, rect := overlayRect(bg, box, m.width)
	rect.fieldStart = fieldStart
	rect.fieldCount = fieldCount
	rect.protocolRow = toggleRow
	rect.listStart = formatRow
	rect.listCount = listCount
	rect.pathRow = pathRow
	rect.pathX = pathX
	rect.pathW = pathW
	rect.browseX = browseX
	rect.browseW = browseW
	rect.testX = 1
	rect.testW = lipgloss.Width(toggle)
	applyFooter(&rect, footerContentRow, importBtns, footerPositions)
	m.mgrImportRect = rect
	return out
}

func (m *Model) renderGroupModal(bg string) string {
	boxW := 68
	if boxW > m.width-8 {
		boxW = m.width - 8
	}
	if boxW < 40 {
		boxW = 40
	}
	innerW := boxW - 2
	groups := m.cs.Groups()

	lines := make([]string, 0, len(groups)+8)
	title := styleTitle.Render(" 分组管理 ")
	hint := styleMuted.Render(" ↑↓ 选择 · Ctrl+K/J 排序 · Esc 取消 ")
	titleLine := padCell(title, maxInt(0, innerW-lipgloss.Width(hint))) + hint
	lines = append(lines, titleLine)
	lines = append(lines, "")

	listStart := len(lines) + 1
	for i, g := range groups {
		text := padCell(fmt.Sprintf(" %s (%d)", g, len(m.cs.ByGroup(g))), innerW)
		if i == m.mgrGroupSel {
			lines = append(lines, styleRowSelected.Render(text))
		} else {
			lines = append(lines, text)
		}
	}
	if len(groups) == 0 {
		lines = append(lines, styleMuted.Render(padCell(" (暂无分组)", innerW)))
	}
	listCount := len(groups)

	lines = append(lines, "")
	fieldStart := len(lines) + 1
	m.mgrGroupInput.Width = maxInt(8, innerW-8)
	lines = append(lines, styleMuted.Render(padCell(" 名称 ", 7))+m.mgrGroupInput.View())
	lines = append(lines, "")

	renameLabel, newLabel, delLabel, cancelLabel := "[Ctrl+R] 重命名", "[Enter] 新建", "[Ctrl+D] 删除", "[Esc] 取消"
	if innerW < 62 {
		renameLabel, newLabel, delLabel, cancelLabel = "[R] 重命名", "[Enter] 新建", "[D] 删除", "[Esc] 取消"
	}
	groupBtns := []footerButton{
		{label: styleBtnMuted.Render(renameLabel), role: footerRoleTest},
		{label: styleBtnOK.Render(newLabel), role: footerRoleOK},
		{label: styleBtnMuted.Render(delLabel), role: footerRoleCancel},
		{label: styleBtnMuted.Render(cancelLabel), role: footerRoleExtra},
	}
	footer, footerPositions := footerLayout(innerW, footerAlignFirstLeftRestRight, groupBtns)
	footerContentRow := len(lines)
	lines = append(lines, footer)

	for i := range lines {
		lines[i] = padCell(lines[i], innerW)
	}
	content := strings.Join(lines, "\n")
	box := styleModal.Width(innerW).Height(len(lines)).Render(content)
	out, rect := overlayRect(bg, box, m.width)
	rect.listStart = listStart
	rect.listCount = listCount
	rect.fieldStart = fieldStart
	rect.fieldCount = 1
	// testX/testW = 重命名，okX/okW = 新建，cancelX/cancelW = 删除，extraX/extraW = 取消。
	applyFooter(&rect, footerContentRow, groupBtns, footerPositions)
	m.mgrGroupRect = rect
	return out
}

func (m *Model) renderResultExportModal(bg string) string {
	boxW := 68
	if boxW > m.width-8 {
		boxW = maxInt(40, m.width-8)
	}
	innerW := boxW - 2

	title := styleTitle.Render(" 导出查询结果 ")
	hint := styleMuted.Render(" ↑↓ 选择 · Ctrl+P 输入目录 · Ctrl+B 浏览 · Enter 导出 · Esc 取消 ")
	lines := make([]string, 0, len(resultExportChoices)+5)
	lines = append(lines, padCell(title, maxInt(0, innerW-lipgloss.Width(hint)))+hint)
	lines = append(lines, "")
	listStart := 3
	for i, choice := range resultExportChoices {
		label := fmt.Sprintf(" %s  .%s  %s", choice.label, choice.ext, choice.desc)
		text := padCell(label, innerW)
		if i == m.resultExportSel {
			lines = append(lines, styleRowSelected.Render(text))
		} else {
			lines = append(lines, text)
		}
	}
	listCount := len(resultExportChoices)
	lines = append(lines, "")

	pathLine, pathX, pathW, browseX, browseW := m.exportPathLine(innerW)
	pathRow := len(lines) + 1
	lines = append(lines, pathLine)
	lines = append(lines, "")

	resultBtns := []footerButton{
		{label: styleBtnOK.Render("[Enter] 导出"), role: footerRoleOK},
		{label: styleBtnMuted.Render("[Esc] 取消"), role: footerRoleCancel},
	}
	footer, footerPositions := footerLayout(innerW, footerAlignFirstLeftRestRight, resultBtns)
	footerContentRow := len(lines)
	lines = append(lines, footer)

	content := strings.Join(lines, "\n")
	box := styleModal.Width(innerW).Height(len(lines)).Render(content)
	out, rect := overlayRect(bg, box, m.width)
	rect.listStart = listStart
	rect.listCount = listCount
	rect.pathRow = pathRow
	rect.pathX = pathX
	rect.pathW = pathW
	rect.browseX = browseX
	rect.browseW = browseW
	applyFooter(&rect, footerContentRow, resultBtns, footerPositions)
	m.resultExportRect = rect
	return out
}

func (m *Model) renderHistoryModal(bg string) string {
	boxW := minInt(100, m.width-8)
	if boxW < 40 {
		boxW = maxInt(30, m.width-4)
	}
	innerW := boxW - 2

	title := styleTitle.Render(fmt.Sprintf(" 查询历史 (%d) ", len(m.queryHistory)))
	hint := styleMuted.Render(" ↑↓ 选择 · Enter 使用 · Esc 关闭 ")
	lines := make([]string, 0, len(m.queryHistory)+5)
	lines = append(lines, padCell(title, maxInt(0, innerW-lipgloss.Width(hint)))+hint)
	lines = append(lines, "")
	listStart := 3
	for i, sql := range m.queryHistory {
		oneLine := strings.ReplaceAll(sql, "\n", " ")
		text := padCell(fmt.Sprintf(" %2d  %s", i+1, ansi.Truncate(oneLine, maxInt(8, innerW-8), "~")), innerW)
		if i == m.historySel {
			lines = append(lines, styleRowSelected.Render(text))
		} else {
			lines = append(lines, text)
		}
	}
	listCount := len(m.queryHistory)
	lines = append(lines, "")

	historyBtns := []footerButton{
		{label: styleBtnOK.Render("[Enter] 使用"), role: footerRoleOK},
		{label: styleBtnMuted.Render("[Esc] 关闭"), role: footerRoleCancel},
	}
	footer, footerPositions := footerLayout(innerW, footerAlignFirstLeftRestRight, historyBtns)
	footerContentRow := len(lines)
	lines = append(lines, footer)

	content := strings.Join(lines, "\n")
	box := styleModal.Width(innerW).Height(len(lines)).Render(content)
	out, rect := overlayRect(bg, box, m.width)
	rect.listStart = listStart
	rect.listCount = listCount
	applyFooter(&rect, footerContentRow, historyBtns, footerPositions)
	m.historyRect = rect
	return out
}

func (m *Model) renderContextMenu(bg string) string {
	if len(m.ctxMenuItems) == 0 {
		return bg
	}
	boxW := 0
	for _, item := range m.ctxMenuItems {
		if w := ansi.StringWidth(item.label); w > boxW {
			boxW = w
		}
	}
	boxW = clampInt(boxW+4, 20, maxInt(20, m.width-4))
	innerW := boxW - 2

	lines := make([]string, 0, len(m.ctxMenuItems)+2)
	lines = append(lines, padCell(" 操作 ", innerW))
	lines = append(lines, "")
	for i, item := range m.ctxMenuItems {
		text := padCell(" "+item.label, innerW)
		if i == m.ctxMenuSel {
			lines = append(lines, styleRowSelected.Render(text))
		} else {
			lines = append(lines, text)
		}
	}
	content := strings.Join(lines, "\n")
	box := styleModal.Width(innerW).Height(len(lines)).Render(content)
	out, rect := overlayRectAt(bg, box, m.width, m.ctxMenuX, m.ctxMenuY)
	rect.listStart = 3
	rect.listCount = len(m.ctxMenuItems)
	m.ctxMenuRect = rect
	return out
}

func (m *Model) renderConfirmModal(bg string) string {
	boxW := 52
	if boxW > m.width-8 {
		boxW = maxInt(28, m.width-8)
	}
	innerW := boxW - 2

	title := styleErrText.Render(" 确认删除 ")
	hint := styleMuted.Render(" 此操作不可撤销 ")
	lines := make([]string, 0, 8)
	lines = append(lines, padCell(title, maxInt(0, innerW-lipgloss.Width(hint)))+hint)
	lines = append(lines, "")

	object := "连接"
	if m.confirm.kind == confirmDeleteGroup {
		object = "分组"
	}
	question := fmt.Sprintf("确定删除%s「%s」吗？", object, m.confirm.name)
	if m.confirm.kind == confirmBatchDeleteConns {
		question = fmt.Sprintf("确定删除选中的 %d 个连接吗？", len(m.confirm.names))
	}
	lines = append(lines, styleWarnText.Render(padCell(" "+question, innerW)))
	if m.confirm.info != "" {
		lines = append(lines, styleMuted.Render(padCell(" "+m.confirm.info, innerW)))
	}
	lines = append(lines, "")

	confirmBtns := []footerButton{
		{label: styleBtnDanger.Render("[Enter/Y] 确认"), role: footerRoleOK},
		{label: styleBtnMuted.Render("[Esc/N] 取消"), role: footerRoleCancel},
	}
	footer, footerPositions := footerLayout(innerW, footerAlignCenter, confirmBtns)
	footerContentRow := len(lines)
	lines = append(lines, footer)

	for i := range lines {
		lines[i] = padCell(lines[i], innerW)
	}
	content := strings.Join(lines, "\n")
	box := styleModal.Width(innerW).Height(len(lines)).Render(content)
	out, rect := overlayRect(bg, box, m.width)
	applyFooter(&rect, footerContentRow, confirmBtns, footerPositions)
	m.confirmRect = rect
	return out
}

// ============ Main 视图 ============

func (m *Model) headerView() string {
	cluster := "-"
	if m.curNode != nil {
		cluster = m.curNode.Cluster
	}
	node := "-"
	if m.curNode != nil {
		node = m.curNode.Label()
	}
	db := m.currentDB
	if db == "" {
		db = m.cfg.EffectiveDatabase()
	}
	dbCount := ""
	if len(m.databases) > 0 {
		dbCount = fmt.Sprintf(" / %d 库", len(m.databases))
	}
	left := styleTitle.Render(" meteor-ch-tui ") +
		styleMuted.Render(" cluster ") + styleOkText.Render(cluster) +
		styleMuted.Render(" node ") + styleOkText.Render(node) +
		styleMuted.Render(" db ") + styleOkText.Render(db) + styleMuted.Render(dbCount)

	if m.loading {
		left += styleWarnText.Render("  ⏳")
	}

	back := styleBtnMuted.Render("返回")
	right := styleMuted.Render(fmt.Sprintf(" %s · 右键/Ctrl+Q 返回 ", m.focusName())) + back
	leftW := m.width - lipgloss.Width(right)
	if leftW < 0 {
		leftW = 0
	}
	line := padCell(left, leftW) + right
	line = padCell(line, m.width)
	m.backBtnX = m.width - lipgloss.Width(back)
	if m.backBtnX < 0 {
		m.backBtnX = 0
	}
	m.backBtnW = lipgloss.Width(back)
	return styleHeader.Width(m.width).Render(line)
}

func (m *Model) focusName() string {
	switch m.focus {
	case focusSidebar:
		return "侧边栏"
	case focusEditor:
		return "SQL 输入"
	default:
		if m.mode == modeSlow {
			return "活跃查询"
		}
		return "结果表格"
	}
}

func (m *Model) statusView() string {
	text := m.status
	if text == "" {
		text = "Tab 切换焦点 | 侧边栏 1=库 2=视图 3=节点 · ↑↓/PgUp/PgDn/滚轮 · Enter 预览 | SQL Alt+Enter 执行 · Ctrl+Y 复制 · Ctrl+E 格式导出 · Ctrl+H 历史 | v 行详情 · Ctrl+R 活跃查询 · 右键菜单 · Ctrl+Q 返回 · Ctrl+C 退出"
	}
	var styled string
	if m.statusOK {
		styled = styleOkText.Render(" " + text)
	} else {
		styled = styleErrText.Render(" " + text)
	}
	styled = ansi.Truncate(styled, m.width, "")
	return styleStatus.Width(m.width).Render(styled)
}

func sidebarScrollGlyph(row, top, available, total int) string {
	if total <= available || available <= 0 {
		return " "
	}
	thumbSize := maxInt(1, available*available/total)
	if thumbSize > available {
		thumbSize = available
	}
	thumbTop := top * (available - thumbSize) / (total - available)
	if row >= thumbTop && row < thumbTop+thumbSize {
		return "█"
	}
	return "│"
}

func (m *Model) sidebarView() string {
	innerW := m.sidebarWidth() - 2
	innerH := m.bodyHeight() - 2
	contentW := maxInt(8, innerW-1)
	lines, sel, items := m.sidebarLines(contentW)
	m.sidebarItems = items

	available := innerH - 1
	if available < 1 {
		available = 1
	}
	top := clampInt(sel-available/2, 0, maxInt(0, len(lines)-available))
	vis := []string(nil)
	if top < len(lines) {
		end := top + available
		if end > len(lines) {
			end = len(lines)
		}
		vis = lines[top:end]
	}

	dbLabel := fmt.Sprintf("数据库 %d", len(m.databases))
	if contentW < 28 {
		dbLabel = fmt.Sprintf("库 %d", len(m.databases))
	}
	viewLabel := fmt.Sprintf("视图 %d", len(m.views))
	nodeLabel := fmt.Sprintf("节点 %d", len(m.nodes))
	dbTab := styleBtnMuted.Render(dbLabel)
	viewTab := styleBtnMuted.Render(viewLabel)
	nodeTab := styleBtnMuted.Render(nodeLabel)
	switch m.sidebarTab {
	case tabDatabases:
		dbTab = styleBtnOK.Render(dbLabel)
	case tabViews:
		viewTab = styleBtnOK.Render(viewLabel)
	case tabNodes:
		nodeTab = styleBtnOK.Render(nodeLabel)
	}
	m.sidebarTabDBX = 1
	m.sidebarTabDBW = lipgloss.Width(dbTab)
	m.sidebarTabViewX = m.sidebarTabDBX + m.sidebarTabDBW + 1
	m.sidebarTabViewW = lipgloss.Width(viewTab)
	m.sidebarTabNodeX = m.sidebarTabViewX + m.sidebarTabViewW + 1
	m.sidebarTabNodeW = lipgloss.Width(nodeTab)
	m.sidebarTabRowY = 2
	m.sidebarFirstLineY = 3
	m.sidebarLineTop = top

	rendered := make([]string, 0, innerH)
	rendered = append(rendered, padCell(dbTab+" "+viewTab+" "+nodeTab, contentW)+" ")
	for i := 0; i < available; i++ {
		body := strings.Repeat(" ", contentW)
		if i < len(vis) {
			body = padCell(vis[i], contentW)
		}
		rendered = append(rendered, body+sidebarScrollGlyph(i, top, available, len(lines)))
	}

	content := strings.Join(rendered, "\n")
	st := styleSidebar
	if m.focus == focusSidebar {
		st = styleSidebar.BorderForeground(colAccent)
	}
	return st.Width(innerW).Height(innerH).Render(content)
}

func (m *Model) sidebarLines(innerW int) ([]string, int, []sidebarItem) {
	lines := make([]string, 0, len(m.databases)+len(m.tables)+len(m.views)+len(m.nodes)+6)
	items := make([]sidebarItem, 0, len(m.databases)+len(m.tables)+len(m.views)+len(m.nodes))
	sel := 0
	dbNote, objectNote := "", ""
	if m.databasesTruncated {
		dbNote = fmt.Sprintf(" · 仅显示前 %d 项", ch.CatalogLimit)
	}
	if m.objectsTruncated {
		objectNote = fmt.Sprintf(" · 仅显示前 %d 项(表+视图)", ch.CatalogLimit*2)
	}

	switch m.sidebarTab {
	case tabDatabases:
		lines = append(lines, styleMuted.Render(padCell(fmt.Sprintf(" 全部库 (%d) · 点击切换%s", len(m.databases), dbNote), innerW)))
		if len(m.databases) == 0 {
			lines = append(lines, styleMuted.Render(padCell(" (未读取到数据库)", innerW)))
		}
		for i, db := range m.databases {
			mark := "  "
			if db == m.currentDB {
				mark = "● "
			}
			text := padCell(" "+mark+db, innerW)
			items = append(items, sidebarItem{kind: itemDatabase, idx: i, row: len(lines)})
			switch {
			case len(items)-1 == m.sidebarIdx:
				sel = len(lines)
				lines = append(lines, styleRowSelected.Render(text))
			case db == m.currentDB:
				lines = append(lines, styleOkText.Render(text))
			default:
				lines = append(lines, text)
			}
		}
		lines = append(lines, padCell("", innerW))
		lines = append(lines, styleMuted.Render(padCell(fmt.Sprintf(" 表 · %s (%d)%s", m.currentDB, len(m.tables), objectNote), innerW)))
		if len(m.tables) == 0 {
			lines = append(lines, styleMuted.Render(padCell(" (该库暂无表)", innerW)))
		}
		for i, t := range m.tables {
			text := padCell("   "+t, innerW)
			items = append(items, sidebarItem{kind: itemTable, idx: i, row: len(lines)})
			if len(items)-1 == m.sidebarIdx {
				sel = len(lines)
				lines = append(lines, styleRowSelected.Render(text))
			} else {
				lines = append(lines, text)
			}
		}
		return lines, sel, items

	case tabViews:
		lines = append(lines, styleMuted.Render(padCell(fmt.Sprintf(" 视图 · %s (%d)%s", m.currentDB, len(m.views), objectNote), innerW)))
		if len(m.views) == 0 {
			lines = append(lines, styleMuted.Render(padCell(" (该库暂无视图)", innerW)))
			return lines, 0, items
		}
		for i, v := range m.views {
			text := padCell("   "+v, innerW)
			items = append(items, sidebarItem{kind: itemView, idx: i, row: len(lines)})
			if len(items)-1 == m.sidebarIdx {
				sel = len(lines)
				lines = append(lines, styleRowSelected.Render(text))
			} else {
				lines = append(lines, text)
			}
		}
		return lines, sel, items

	default:
		if len(m.nodes) == 0 {
			lines = append(lines, styleMuted.Render(padCell(" (无节点数据)", innerW)))
			return lines, 0, items
		}
		lastCluster := ""
		for i, n := range m.nodes {
			if n.Cluster != lastCluster {
				lastCluster = n.Cluster
				lines = append(lines, styleMuted.Render(padCell(" ◆ "+lastCluster, innerW)))
			}
			text := padCell(fmt.Sprintf("   s%s/r%s %s", n.Shard, n.Replica, n.HostName), innerW)
			items = append(items, sidebarItem{kind: itemNode, idx: i, row: len(lines)})
			if len(items)-1 == m.sidebarIdx {
				sel = len(lines)
				lines = append(lines, styleRowSelected.Render(text))
			} else {
				lines = append(lines, text)
			}
		}
		return lines, sel, items
	}
}

func (m *Model) rightView() string {
	editor := m.editorView()
	table := m.tableView()
	return lipgloss.JoinVertical(lipgloss.Left, editor, table)
}

func (m *Model) editorView() string {
	innerW := m.rightWidth() - 2
	innerH := editorPaneHeight - 2

	title := styleTitle.Render("SQL")
	if m.focus == focusEditor {
		title += styleMuted.Render("  (Alt+Enter 执行 · Tab 切换)")
	}
	body := clipLines(m.editor.View(), innerH-1)
	content := padCell(title, innerW) + "\n" + body
	content = padToHeight(content, innerH, innerW)

	st := stylePane
	if m.focus == focusEditor {
		st = stylePaneFocus
	}
	return st.Width(innerW).Height(innerH).Render(content)
}

func (m *Model) renderCfgModal(bg string) string {
	labels := []string{"Host", "Port", "Username", "Password", "Database", "协议"}

	boxW := 62
	if boxW > m.width-8 {
		boxW = m.width - 8
	}
	innerW := boxW - 2

	lines := make([]string, 0, cfgFieldCount+4)
	title := styleTitle.Render(" ClickHouse 连接配置 ")
	hint := styleMuted.Render(" Tab 下一项 · Shift+Tab 上一项 · Enter 连接 · Esc 取消 ")
	titleLine := padCell(title, maxInt(0, innerW-lipgloss.Width(hint))) + hint
	lines = append(lines, titleLine)
	lines = append(lines, "")

	protoContentRow := -1
	for i := 0; i < int(cfgFieldCount); i++ {
		ti := m.cfgFields[i]
		label := styleMuted.Render(padCell(labels[i], 10))
		if i == int(cfgFieldProtocol) {
			prompt := "  "
			if i == int(m.cfgActive) {
				prompt = "▸ "
			}
			opt := styleBtnOK.Render(ch.ProtocolLabel(ti.Value()))
			note := styleWarnText.Render(protocolPortNote(ti.Value(), m.cfgFields[cfgFieldPort].Value()))
			lines = append(lines, label+prompt+opt+styleMuted.Render("  ←/→ 切换")+note)
			protoContentRow = len(lines) - 1
			continue
		}
		if i == int(m.cfgActive) {
			ti.Prompt = "▸ "
		} else {
			ti.Prompt = "  "
		}
		ti.Width = innerW - 12
		fieldView := ti.View()
		lines = append(lines, label+fieldView)
	}

	lines = append(lines, "")
	cfgBtns := []footerButton{
		{label: styleBtnOK.Render("[Enter] 连接"), role: footerRoleOK, hotRight: 2},
		{label: styleBtnMuted.Render("[Esc] 取消"), role: footerRoleCancel, hotLeft: 2, hotRight: 2},
	}
	footer, footerPositions := footerLayout(innerW, footerAlignLastRight, cfgBtns)
	footerContentRow := len(lines)
	lines = append(lines, footer)

	for i := range lines {
		lines[i] = padCell(lines[i], innerW)
	}
	content := strings.Join(lines, "\n")
	box := styleModal.Width(innerW).Height(len(lines)).Render(content)
	out, rect := overlayRect(bg, box, m.width)
	rect.fieldStart = 3
	rect.fieldCount = int(cfgFieldCount)
	rect.protocolRow = protoContentRow + 1
	applyFooter(&rect, footerContentRow, cfgBtns, footerPositions)
	m.cfgModalRect = rect
	return out
}
