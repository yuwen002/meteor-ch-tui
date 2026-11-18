package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) initExportPathInput() {
	ti := textinput.New()
	ti.Placeholder = "保存目录（~ / $VAR / %VAR%，留空=当前目录）"
	ti.SetValue(m.lastExportDir)
	ti.CharLimit = 0
	ti.Width = 48
	m.exportPathInput = ti
	m.exportPathFocused = false
}

func (m *Model) exportPathLine(innerW int) (string, int, int, int, int) {
	mark := "  "
	if m.exportPathFocused {
		mark = "> "
	}
	label := mark + "保存目录: "
	browse := styleBtnMuted.Render("[Ctrl+B] 浏览")
	inputW := maxInt(10, innerW-lipgloss.Width(label)-lipgloss.Width(browse)-3)
	m.exportPathInput.Width = inputW
	view := m.exportPathInput.View()
	line := label + view + " " + browse
	return padCell(line, innerW), lipgloss.Width(label), lipgloss.Width(view),
		lipgloss.Width(label) + inputW + 1, lipgloss.Width(browse)
}

func (m *Model) focusExportPath() tea.Cmd {
	m.exportPathFocused = true
	return m.exportPathInput.Focus()
}

func (m *Model) blurExportPath() {
	m.exportPathFocused = false
	m.exportPathInput.Blur()
}

func expandWindowsEnv(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '%' {
			if rel := strings.IndexByte(s[i+1:], '%'); rel >= 0 {
				end := i + 1 + rel
				name := s[i+1 : end]
				if name != "" {
					if value, ok := os.LookupEnv(name); ok {
						b.WriteString(value)
						i = end + 1
						continue
					}
				}
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func expandPathInput(value string) (string, error) {
	s := strings.TrimSpace(value)
	if s == "" {
		return "", nil
	}
	if s == "~" || strings.HasPrefix(s, "~/") || strings.HasPrefix(s, "~\\") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("无法获取用户目录: %w", err)
		}
		if s == "~" {
			s = home
		} else {
			s = filepath.Join(home, s[2:])
		}
	}
	s = os.ExpandEnv(s)
	if strings.Contains(s, "%") {
		s = expandWindowsEnv(s)
	}
	return filepath.Clean(s), nil
}

func (m *Model) exportFilePath(name string) (string, error) {
	dir, err := expandPathInput(m.exportPathInput.Value())
	if err != nil {
		return "", fmt.Errorf("解析保存目录失败: %w", err)
	}
	if dir == "" {
		return name, nil
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建保存目录失败: %w", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		return "", fmt.Errorf("检查保存目录失败: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("保存位置不是目录: %s", dir)
	}
	m.setLastExportDir(dir)
	return filepath.Join(dir, name), nil
}
