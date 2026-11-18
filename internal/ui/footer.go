package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type footerHitRole int

const (
	footerRoleOK footerHitRole = iota
	footerRoleCancel
	footerRoleExtra
	footerRoleTest
)

type footerAlign int

const (
	footerAlignFirstLeftRestRight footerAlign = iota
	footerAlignCenter
	footerAlignLastRight
)

type footerButton struct {
	label    string
	role     footerHitRole
	hotLeft  int
	hotRight int
}

// footerLayout 生成底部按钮行，并返回每个按钮相对弹窗左边的起始列。
// 按钮 label 应当已经套用样式，热区宽度按其显示宽度计算。
func footerLayout(innerW int, align footerAlign, buttons []footerButton) (string, []int) {
	n := len(buttons)
	if n == 0 {
		return "", nil
	}
	widths := make([]int, n)
	sum := 0
	for i, b := range buttons {
		widths[i] = lipgloss.Width(b.label)
		sum += widths[i]
	}
	positions := make([]int, n)
	parts := make([]string, 0, n*2)

	switch align {
	case footerAlignCenter:
		used := sum + n - 1
		leftGap := maxInt(0, (innerW-used)/2)
		pos := 1 + leftGap
		parts = append(parts, strings.Repeat(" ", leftGap))
		for i, b := range buttons {
			positions[i] = pos
			parts = append(parts, b.label)
			pos += widths[i]
			if i < n-1 {
				parts = append(parts, " ")
				pos++
			}
		}
	case footerAlignLastRight:
		if n == 2 {
			prefixW := maxInt(0, innerW-widths[1])
			positions[0] = 1
			positions[1] = 1 + prefixW
			return padCell(buttons[0].label, prefixW) + buttons[1].label, positions
		}
		fallthrough
	default:
		gap := maxInt(0, innerW-sum-(n-1))
		pos := 1
		for i, b := range buttons {
			positions[i] = pos
			parts = append(parts, b.label)
			pos += widths[i]
			if i == n-1 {
				break
			}
			if i == 0 {
				parts = append(parts, strings.Repeat(" ", gap))
				pos += gap
			} else {
				parts = append(parts, " ")
				pos++
			}
		}
	}
	return strings.Join(parts, ""), positions
}

// applyFooter 把底部按钮行对应的热区写入弹窗矩形。
// footerContentRow 是按钮行在弹窗内容中的 0-based 下标；边框行偏移在这里统一加 1。
func applyFooter(rect *modalRect, footerContentRow int, buttons []footerButton, positions []int) {
	rect.footerRow = footerContentRow + 1
	for i, b := range buttons {
		if i >= len(positions) {
			return
		}
		setFooterHit(rect, b.role, positions[i]-b.hotLeft, lipgloss.Width(b.label)+b.hotLeft+b.hotRight)
	}
}

func setFooterHit(rect *modalRect, role footerHitRole, x, w int) {
	switch role {
	case footerRoleOK:
		rect.okX, rect.okW = x, w
	case footerRoleCancel:
		rect.cancelX, rect.cancelW = x, w
	case footerRoleExtra:
		rect.extraX, rect.extraW = x, w
	case footerRoleTest:
		rect.testX, rect.testW = x, w
	}
}

func footerHitRange(rect modalRect, role footerHitRole) (int, int) {
	switch role {
	case footerRoleOK:
		return rect.okX, rect.okW
	case footerRoleCancel:
		return rect.cancelX, rect.cancelW
	case footerRoleExtra:
		return rect.extraX, rect.extraW
	case footerRoleTest:
		return rect.testX, rect.testW
	}
	return 0, 0
}

// hitFooter 判断相对坐标是否落在指定按钮角色上。
// 调用方显式传入本弹窗底部实际存在的角色，避免把非底部字段误当成按钮。
func hitFooter(relX, relY int, rect modalRect, roles ...footerHitRole) (footerHitRole, bool) {
	if relY != rect.footerRow {
		return 0, false
	}
	for _, role := range roles {
		x, w := footerHitRange(rect, role)
		if w > 0 && relX >= x && relX < x+w {
			return role, true
		}
	}
	return 0, false
}
