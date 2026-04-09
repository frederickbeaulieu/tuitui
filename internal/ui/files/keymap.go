package files

import (
	"charm.land/bubbles/v2/key"

	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

type KeyMap struct {
	common.KeyMap
	ToggleAllFiles key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		KeyMap: common.DefaultKeyMap(),
		ToggleAllFiles: key.NewBinding(
			key.WithKeys("z"),
			key.WithHelp("z", "all files"),
		),
	}
}

func (km KeyMap) StatusBinds(showAll bool) []key.Help {
	toggle := km.ToggleAllFiles.Help()
	if showAll {
		toggle.Desc = "changed only"
	}
	binds := []key.Help{
		km.Back.Help(),
		km.Open.Help(),
	}
	binds = append(binds, km.NavigationBinds()...)
	binds = append(binds, toggle)
	return binds
}
