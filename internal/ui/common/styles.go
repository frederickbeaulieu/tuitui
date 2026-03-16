package common

import (
	"charm.land/lipgloss/v2"
)

// Tokyo Night color palette.
var (
	ColorMauve  = lipgloss.Color("#bb9af7")
	ColorRed    = lipgloss.Color("#f7768e")
	ColorYellow = lipgloss.Color("#e0af68")
	ColorGreen  = lipgloss.Color("#9ece6a")
	ColorTeal   = lipgloss.Color("#73daca")
	ColorBlue   = lipgloss.Color("#7aa2f7")

	ColorText    = lipgloss.Color("#c0caf5")
	ColorSubtext = lipgloss.Color("#a9b1d6")
	ColorOverlay = lipgloss.Color("#565f89")
	ColorSurface = lipgloss.Color("#24283b")
)

var (
	PanelActive = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorMauve)

	PanelInactive = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorOverlay)

	PanelTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorMauve).
			PaddingLeft(1).
			PaddingRight(1)

	PanelTitleInactive = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorOverlay).
				PaddingLeft(1).
				PaddingRight(1)
)

var (
	TextDim = lipgloss.NewStyle().
		Foreground(ColorSubtext)

	TextMuted = lipgloss.NewStyle().
			Foreground(ColorOverlay)
)

var ConflictStyle = lipgloss.NewStyle().
	Foreground(ColorRed).
	Bold(true)

var (
	FileAdded    = lipgloss.NewStyle().Foreground(ColorGreen)
	FileModified = lipgloss.NewStyle().Foreground(ColorYellow)
	FileDeleted  = lipgloss.NewStyle().Foreground(ColorRed)
	FileRenamed  = lipgloss.NewStyle().Foreground(ColorBlue)
)

// FileStatusSymbol returns a styled status symbol for a file change.
func FileStatusSymbol(status string) string {
	switch status {
	case "A":
		return FileAdded.Render("A")
	case "M":
		return FileModified.Render("M")
	case "D":
		return FileDeleted.Render("D")
	case "R":
		return FileRenamed.Render("R")
	default:
		return TextDim.Render(status)
	}
}
