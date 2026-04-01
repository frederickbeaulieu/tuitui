package cmdbar

import (
	"charm.land/bubbles/v2/key"

	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

func errorStatusBinds(km common.KeyMap) []key.Help {
	binds := []key.Help{
		km.ClosePanel.Help(),
	}
	binds = append(binds, km.ScrollBinds()...)
	return binds
}
