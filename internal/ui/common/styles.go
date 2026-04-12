package common

import (
	"charm.land/lipgloss/v2"
)

// ANSI palette colors — follow the terminal theme automatically.
var (
	ColorMauve  = lipgloss.ANSIColor(13) // bright magenta
	ColorRed    = lipgloss.ANSIColor(1)  // red
	ColorYellow = lipgloss.ANSIColor(3)  // yellow
	ColorGreen  = lipgloss.ANSIColor(2)  // green
	ColorTeal   = lipgloss.ANSIColor(14) // bright cyan
	ColorBlue   = lipgloss.ANSIColor(4)  // blue

	ColorText    = lipgloss.ANSIColor(7)  // white
	ColorSubtext = lipgloss.ANSIColor(15) // bright white
	ColorOverlay = lipgloss.ANSIColor(8)  // bright black
	ColorSurface = lipgloss.ANSIColor(0)  // black
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
	Foreground(ColorRed)

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
