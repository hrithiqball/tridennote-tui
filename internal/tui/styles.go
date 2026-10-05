package tui

import "github.com/charmbracelet/lipgloss"

var (
	indigo = lipgloss.Color("#7c85f5")
	violet = lipgloss.Color("#c084fc")
	mint   = lipgloss.Color("#6ee7b7")
	muted  = lipgloss.Color("#8a8a99")
	faint  = lipgloss.Color("#4a4a58")
	text   = lipgloss.Color("#dcddde")
	danger = lipgloss.Color("#f87171")

	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(text)
	accentStyle  = lipgloss.NewStyle().Foreground(violet).Bold(true)
	mutedStyle   = lipgloss.NewStyle().Foreground(muted)
	faintStyle   = lipgloss.NewStyle().Foreground(faint)
	errorStyle   = lipgloss.NewStyle().Foreground(danger)
	successStyle = lipgloss.NewStyle().Foreground(mint).Bold(true)

	codeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#2a2a3a")).
			Padding(0, 2)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(indigo).
			Padding(1, 4)

	keyStyle = lipgloss.NewStyle().Foreground(indigo).Bold(true)

	folderStyle   = lipgloss.NewStyle().Foreground(violet)
	fileStyle     = lipgloss.NewStyle().Foreground(text)
	cursorStyle   = lipgloss.NewStyle().Background(lipgloss.Color("#2d2f55")).Foreground(lipgloss.Color("#ffffff")).Bold(true)
	cursorDimmed  = lipgloss.NewStyle().Background(lipgloss.Color("#26262f"))
	paneFocused   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(indigo)
	paneUnfocused = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(faint)
	statusStyle   = lipgloss.NewStyle().Foreground(muted).Padding(0, 1)
)
