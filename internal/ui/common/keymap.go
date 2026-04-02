// Package common provides shared keybindings, styles, and UI components.
package common

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// formatKey converts a key string like "ctrl+u" into "C-u" for display.
func formatKey(k string) string {
	if strings.HasPrefix(k, "ctrl+") {
		return "C-" + strings.TrimPrefix(k, "ctrl+")
	}
	return k
}

// helpKey returns the display-friendly label for a binding's first key.
func helpKey(b key.Binding) string {
	keys := b.Keys()
	if len(keys) == 0 {
		return ""
	}
	return formatKey(keys[0])
}

// PairHelp combines two related bindings (e.g. Up/Down) into a single
// key.Help with a merged key label like "j/k" and the first binding's description.
func PairHelp(a, b key.Binding) key.Help {
	return key.Help{
		Key:  helpKey(a) + "/" + helpKey(b),
		Desc: a.Help().Desc,
	}
}

// KeyMap defines the shared keybindings used across all panels.
type KeyMap struct {
	Quit           key.Binding
	Open           key.Binding
	Back           key.Binding
	Close          key.Binding
	ClosePanel     key.Binding
	Submit         key.Binding
	Up             key.Binding
	Down           key.Binding
	HalfUp         key.Binding
	HalfDown       key.Binding
	Top            key.Binding
	Bottom         key.Binding
	Command        key.Binding
	CompleteTab    key.Binding
	CompleteNext   key.Binding
	CompletePrev   key.Binding
	CompleteToggle key.Binding
}

// DefaultKeyMap returns the shared keybindings with their help text.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Open: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "open"),
		),
		Back: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "back"),
		),
		Close: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "close"),
		),
		ClosePanel: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("q/esc", "close"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
		),
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k", "navigate"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j", "navigate"),
		),
		HalfUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("C-u", "half page"),
		),
		HalfDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("C-d", "half page"),
		),
		Top: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "top/bottom"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "top/bottom"),
		),
		Command: key.NewBinding(
			key.WithKeys(":"),
			key.WithHelp(":", "command"),
		),
		CompleteTab: key.NewBinding(
			key.WithKeys("tab"),
		),
		CompleteNext: key.NewBinding(
			key.WithKeys("ctrl+n"),
		),
		CompletePrev: key.NewBinding(
			key.WithKeys("ctrl+p"),
		),
		CompleteToggle: key.NewBinding(
			key.WithKeys("ctrl+space"),
		),
	}
}

// NavigationBinds returns paired help entries for the standard navigation keys.
func (km KeyMap) NavigationBinds() []key.Help {
	return []key.Help{
		PairHelp(km.Up, km.Down),
		PairHelp(km.HalfUp, km.HalfDown),
		PairHelp(km.Top, km.Bottom),
	}
}

// ScrollBinds returns paired help entries with "scroll" as the description.
func (km KeyMap) ScrollBinds() []key.Help {
	h := PairHelp(km.Up, km.Down)
	h.Desc = "scroll"
	return []key.Help{
		h,
		PairHelp(km.HalfUp, km.HalfDown),
		PairHelp(km.Top, km.Bottom),
	}
}

// HandleScroll matches a key against the scroll bindings and returns the
// updated position. The halfPage parameter controls how far HalfUp/HalfDown
// jump (typically height/2 for viewports, or height/4 for multi-line entries).
// Returns (position, true) if the key was handled.
func (km KeyMap) HandleScroll(msg tea.KeyPressMsg, pos, maxPos, halfPage int) (int, bool) {
	switch {
	case key.Matches(msg, km.Down):
		return Clamp(pos+1, 0, maxPos), true
	case key.Matches(msg, km.Up):
		return Clamp(pos-1, 0, maxPos), true
	case key.Matches(msg, km.HalfDown):
		return Clamp(pos+halfPage, 0, maxPos), true
	case key.Matches(msg, km.HalfUp):
		return Clamp(pos-halfPage, 0, maxPos), true
	case key.Matches(msg, km.Top):
		return 0, true
	case key.Matches(msg, km.Bottom):
		return maxPos, true
	}
	return pos, false
}

func Clamp(v, lo, hi int) int {
	return max(lo, min(v, hi))
}
