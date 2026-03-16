package common

import "charm.land/bubbles/v2/key"

// KeyMap defines the shared keybindings used across all panels.
type KeyMap struct {
	Quit     key.Binding
	Open     key.Binding
	Back     key.Binding
	Up       key.Binding
	Down     key.Binding
	HalfUp   key.Binding
	HalfDown key.Binding
	Top      key.Binding
	Bottom   key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
		),
		Open: key.NewBinding(
			key.WithKeys("l"),
		),
		Back: key.NewBinding(
			key.WithKeys("h"),
		),
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
		),
		HalfUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
		),
		HalfDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
		),
		Top: key.NewBinding(
			key.WithKeys("g"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
		),
	}
}
