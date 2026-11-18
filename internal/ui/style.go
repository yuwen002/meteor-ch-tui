package ui

import "github.com/charmbracelet/lipgloss"

var (
	colAccent  = lipgloss.Color("39")
	colMuted   = lipgloss.Color("241")
	colFg      = lipgloss.Color("252")
	colWarn    = lipgloss.Color("214")
	colErr     = lipgloss.Color("203")
	colOk      = lipgloss.Color("42")
	colBorder  = lipgloss.Color("240")
	colSelect  = lipgloss.Color("229")
	colSelectB = lipgloss.Color("24")

	styleHeader = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(colFg)

	styleStatus = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(colMuted)

	styleSidebar = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colBorder)

	stylePane = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colBorder)

	stylePaneFocus = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colAccent)

	styleTitle = lipgloss.NewStyle().
			Foreground(colAccent).
			Bold(true)

	styleMuted = lipgloss.NewStyle().Foreground(colMuted)

	styleErrText = lipgloss.NewStyle().Foreground(colErr)

	styleOkText = lipgloss.NewStyle().Foreground(colOk)

	styleWarnText = lipgloss.NewStyle().Foreground(colWarn)

	styleBtnOK = lipgloss.NewStyle().
			Foreground(colOk).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	styleBtnMuted = lipgloss.NewStyle().
			Foreground(colMuted).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	styleBtnDanger = lipgloss.NewStyle().
			Foreground(colErr).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	styleTableHeader = lipgloss.NewStyle().
				Foreground(colAccent).
				Bold(true)

	styleRowSelected = lipgloss.NewStyle().
				Background(colSelectB).
				Foreground(colSelect)

	styleModal = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colAccent).
			Background(lipgloss.Color("235"))
)
