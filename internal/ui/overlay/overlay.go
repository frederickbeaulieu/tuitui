// Package overlay composites a floating widget on top of a base view.
//
// The implementation is a line-based string overlay: it splits the base
// into visual lines, and for each line covered by the overlay it
// replaces the overlay's column range with the overlay's content using
// ANSI-aware slicing. No dimming, no transparency — the overlay is
// opaque in the cells it occupies.
//
// The package is deliberately minimal and UI-agnostic: it knows nothing
// about prompts, menus, or dialogs. Keyboard routing, focus and state
// are the caller's responsibility.
package overlay

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Layer represents a single overlay layer with its content and desired
// position relative to the base canvas.
type Layer struct {
	content string
	width   int
	height  int
	anchor  Anchor
}

// Anchor describes how a layer should be positioned relative to the
// canvas it is composited onto.
type Anchor int

const (
	// Center anchors the layer at the middle of the canvas.
	Center Anchor = iota
)

// Centered wraps content in a layer that will be drawn at the center of
// the base canvas. Width and height are inferred from the rendered
// content; the caller is responsible for producing a fully styled,
// box-sized string (e.g. via lipgloss.Style.Render).
func Centered(content string) *Layer {
	w, h := measure(content)
	return &Layer{
		content: content,
		width:   w,
		height:  h,
		anchor:  Center,
	}
}

// Composite draws layers on top of base. base is the full-screen
// rendered view (baseWidth x baseHeight). The returned string has the
// overlay painted over the corresponding rows/columns of base while
// leaving the rest of base untouched.
//
// If no layers are provided base is returned unchanged.
func Composite(base string, baseWidth, baseHeight int, layers ...*Layer) string {
	if len(layers) == 0 {
		return base
	}
	lines := strings.Split(base, "\n")
	for _, l := range layers {
		x, y := position(l, baseWidth, baseHeight)
		overlayLines := strings.Split(l.content, "\n")
		for i, ol := range overlayLines {
			row := y + i
			if row < 0 || row >= len(lines) {
				continue
			}
			lines[row] = spliceLine(lines[row], ol, x, l.width, baseWidth)
		}
	}
	return strings.Join(lines, "\n")
}

// spliceLine replaces the column range [x, x+width) of base with
// overlayLine. Both inputs may contain ANSI escape sequences; slicing
// is performed in visual-cell space using ansi.Truncate/TruncateLeft.
//
// The trailing portion of base after the overlay is restored with a
// preceding ANSI reset so the overlay's styles don't leak past it.
func spliceLine(base, overlayLine string, x, width, baseWidth int) string {
	if x < 0 {
		x = 0
	}
	// Left part of base: columns [0, x).
	left := ansi.Truncate(base, x, "")
	leftWidth := ansi.StringWidth(left)
	// Pad left with spaces if base was shorter than x.
	if leftWidth < x {
		left = left + strings.Repeat(" ", x-leftWidth)
	}

	// Right part of base: columns [x+width, baseWidth).
	rightStart := x + width
	right := ""
	if rightStart < baseWidth {
		right = ansi.TruncateLeft(base, rightStart, "")
	}

	// Reset before the right part to stop overlay style bleed.
	const reset = "\x1b[0m"
	return left + overlayLine + reset + right
}

func position(l *Layer, canvasW, canvasH int) (int, int) {
	switch l.anchor {
	case Center:
		x := (canvasW - l.width) / 2
		y := (canvasH - l.height) / 2
		return clampNonNeg(x), clampNonNeg(y)
	}
	return 0, 0
}

func clampNonNeg(v int) int {
	if v < 0 {
		return 0
	}
	return v
}

// measure returns the visual width (widest line) and height (number of
// lines) of a rendered string, ignoring ANSI escape sequences.
func measure(s string) (int, int) {
	if s == "" {
		return 0, 0
	}
	lines := strings.Split(s, "\n")
	maxW := 0
	for _, l := range lines {
		if w := ansi.StringWidth(l); w > maxW {
			maxW = w
		}
	}
	return maxW, len(lines)
}
