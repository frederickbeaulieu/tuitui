package cmdbar

import (
	"charm.land/bubbles/v2/key"

	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

var (
	keyClose = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "close"),
	)
	keyCloseError = key.NewBinding(
		key.WithKeys("escape", "q"),
		key.WithHelp("q/esc", "close"),
	)
	keySubmit = key.NewBinding(
		key.WithKeys("enter"),
	)
)

func errorStatusBinds(km common.KeyMap) []key.Help {
	binds := []key.Help{
		keyCloseError.Help(),
	}
	binds = append(binds, km.ScrollBinds()...)
	return binds
}
