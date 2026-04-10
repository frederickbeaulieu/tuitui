package common

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// RenderPanel renders content inside a bordered panel with a title.
// width and height are the total outer dimensions including borders.
func RenderPanel(title, content string, width, height int, focused bool) string {
	var borderStyle, titleStyle lipgloss.Style
	if focused {
		borderStyle = PanelActive
		titleStyle = PanelTitle
	} else {
		borderStyle = PanelInactive
		titleStyle = PanelTitleInactive
	}

	width = max(width, 3)
	height = max(height, 3)

	rendered := borderStyle.
		Width(width).
		Height(height).
		MaxHeight(height).
		Render(content)

	if title != "" {
		titleRendered := titleStyle.Render(title)
		lines := strings.Split(rendered, "\n")
		if len(lines) > 0 {
			lines[0] = overlayTitle(lines[0], titleRendered)
			rendered = strings.Join(lines, "\n")
		}
	}

	return rendered
}

// overlayTitle places a title string over the top border of a panel,
// preserving ANSI styling on the border line.
func overlayTitle(borderLine, title string) string {
	titleVisualWidth := ansi.StringWidth(title)
	if titleVisualWidth == 0 {
		return borderLine
	}

	runes := []rune(borderLine)
	var result []rune
	var borderColor []rune
	pos := 0
	visPos := 0

	// Copy border up to visual position 2, capturing the border color sequence.
	for pos < len(runes) && visPos < 2 {
		if runes[pos] == '\x1b' {
			start := pos
			for pos < len(runes) {
				result = append(result, runes[pos])
				if runes[pos] != '\x1b' && ((runes[pos] >= 'A' && runes[pos] <= 'Z') || (runes[pos] >= 'a' && runes[pos] <= 'z')) {
					pos++
					break
				}
				pos++
			}
			borderColor = runes[start:pos]
			continue
		}
		result = append(result, runes[pos])
		visPos++
		pos++
	}

	// Insert the styled title, then re-apply the border color.
	result = append(result, []rune(title)...)
	result = append(result, borderColor...)

	// Skip border runes for the visual width of the title.
	skipped := 0
	for pos < len(runes) && skipped < titleVisualWidth {
		if runes[pos] == '\x1b' {
			for pos < len(runes) {
				if runes[pos] != '\x1b' && ((runes[pos] >= 'A' && runes[pos] <= 'Z') || (runes[pos] >= 'a' && runes[pos] <= 'z')) {
					pos++
					break
				}
				pos++
			}
			continue
		}
		skipped++
		pos++
	}

	// Copy the rest of the border line.
	if pos < len(runes) {
		result = append(result, runes[pos:]...)
	}

	return string(result)
}

// HighlightLine applies a background highlight to a line while preserving its
// existing foreground ANSI colors. It re-injects the background after every
// ANSI reset so the highlight persists across the entire line.
func HighlightLine(line string, width int) string {
	plainLen := ansi.StringWidth(line)
	if plainLen < width {
		line = line + strings.Repeat(" ", width-plainLen)
	}

	const bgSet = "\x1b[48;2;86;95;137m" // ColorOverlay RGB

	highlighted := bgSet +
		strings.ReplaceAll(
			strings.ReplaceAll(line, "\x1b[0m", "\x1b[0m"+bgSet),
			"\x1b[m", "\x1b[m"+bgSet,
		) + "\x1b[0m"

	return highlighted
}
