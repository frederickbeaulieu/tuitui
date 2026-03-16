package common

import (
	"strings"

	"charm.land/lipgloss/v2"
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
	titleVisualWidth := VisualLen(StripAnsi(title))
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

// VisualLen returns the number of runes in s. Callers should strip ANSI first.
func VisualLen(s string) int {
	return len([]rune(s))
}

func StripAnsi(s string) string {
	result := make([]rune, 0, len(s))
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
			}
			continue
		}
		result = append(result, r)
	}
	return string(result)
}

// Truncate truncates a string to the given visual width (excluding ANSI escapes).
func Truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	visualLen := len([]rune(StripAnsi(s)))
	if visualLen <= maxWidth {
		return s
	}
	return TruncateAnsi(s, maxWidth)
}

// TruncateAnsi truncates a string containing ANSI escape sequences to the
// given visual width without splitting escape sequences.
func TruncateAnsi(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	runes := []rune(s)
	var result []rune
	visibleCount := 0
	pos := 0

	for pos < len(runes) && visibleCount < maxWidth {
		if runes[pos] == '\x1b' {
			for pos < len(runes) {
				result = append(result, runes[pos])
				if runes[pos] != '\x1b' && ((runes[pos] >= 'A' && runes[pos] <= 'Z') || (runes[pos] >= 'a' && runes[pos] <= 'z')) {
					pos++
					break
				}
				pos++
			}
			continue
		}

		result = append(result, runes[pos])
		visibleCount++
		pos++
	}

	if visibleCount >= maxWidth {
		result = append(result, []rune("\x1b[0m")...)
	}

	return string(result)
}

// HighlightLine applies a background highlight to a line while preserving its
// existing foreground ANSI colors. It re-injects the background after every
// ANSI reset so the highlight persists across the entire line.
func HighlightLine(line string, width int) string {
	plainLen := VisualLen(StripAnsi(line))
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
