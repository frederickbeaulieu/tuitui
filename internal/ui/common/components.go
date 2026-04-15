package common

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// RenderPanel renders content inside a bordered panel with a title.
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

	// Copy border up to visual position 2, capturing the color sequence.
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

	// Insert title, restore border color.
	result = append(result, []rune(title)...)
	result = append(result, borderColor...)

	// Skip border runes covered by the title.
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

	// Copy the rest.
	if pos < len(runes) {
		result = append(result, runes[pos:]...)
	}

	return string(result)
}

// HighlightRow applies a background highlight to a row, re-injecting the
// background after every ANSI reset. Bright-black foreground is boosted to
// white for contrast.
func HighlightRow(line string, width int) string {
	plainLen := ansi.StringWidth(line)
	if plainLen < width {
		line = line + strings.Repeat(" ", width-plainLen)
	}

	bgStyle := ansi.NewStyle().BackgroundColor(ansi.BrightBlack)
	bgSet := bgStyle.String()
	brightBlackFg := ansi.NewStyle().ForegroundColor(ansi.BrightBlack).String()
	brightBlackFgExt := ansi.NewStyle().ForegroundColor(ansi.ExtendedColor(8)).String()
	whiteFg := ansi.NewStyle().ForegroundColor(ansi.White).String()

	// Boost bright-black foreground to white for contrast.
	line = strings.ReplaceAll(line, brightBlackFg, whiteFg)
	line = strings.ReplaceAll(line, brightBlackFgExt, whiteFg)

	highlighted := bgSet +
		strings.ReplaceAll(
			strings.ReplaceAll(line, "\x1b[0m", "\x1b[0m"+bgSet),
			ansi.ResetStyle, ansi.ResetStyle+bgSet,
		) + ansi.ResetStyle

	return highlighted
}

// HighlightMatches applies mauve/bold styling to matched character positions
// in ANSI-colored lines. Positions are indices into the joined plain text
// of all lines (separated by " ").
func HighlightMatches(lines []string, matchedPositions []int) []string {
	if len(matchedPositions) == 0 || len(lines) == 0 {
		return lines
	}

	matchSet := make(map[int]bool, len(matchedPositions))
	for _, p := range matchedPositions {
		matchSet[p] = true
	}

	highlightStart := ansi.NewStyle().ForegroundColor(ansi.ExtendedColor(13)).Bold().String()
	highlightEnd := ansi.ResetStyle

	result := make([]string, len(lines))
	visPos := 0

	for lineIdx, line := range lines {
		if lineIdx > 0 {
			visPos++ // account for " " separator
		}

		var b strings.Builder
		var savedStyle string
		var state byte
		p := ansi.NewParser()

		for len(line) > 0 {
			seq, width, n, newState := ansi.DecodeSequence(line, state, p)
			state = newState

			if width == 0 {
				b.WriteString(seq)
				if ansi.HasCsiPrefix(seq) {
					if isSGRReset(p) {
						savedStyle = ""
					} else {
						savedStyle += seq
					}
				}
			} else {
				// Printable grapheme.
				if matchSet[visPos] {
					b.WriteString(highlightStart)
					b.WriteString(seq)
					b.WriteString(highlightEnd)
					if savedStyle != "" {
						b.WriteString(savedStyle)
					}
				} else {
					b.WriteString(seq)
				}
				visPos++
			}
			line = line[n:]
		}

		result[lineIdx] = b.String()
	}

	return result
}

// isSGRReset returns true if the parser just decoded an SGR reset
// sequence (\x1b[m or \x1b[0m).
func isSGRReset(p *ansi.Parser) bool {
	if p.Command()&0xff != 'm' {
		return false
	}
	params := p.Params()
	return len(params) == 0 || (len(params) == 1 && params[0] == 0)
}
