package log

import (
	"charm.land/bubbles/v2/key"

	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

type KeyMap struct {
	common.KeyMap
	ToggleRevisions key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		KeyMap: common.DefaultKeyMap(),
		ToggleRevisions: key.NewBinding(
			key.WithKeys("z"),
			key.WithHelp("z", "all revisions"),
		),
	}
}

func (km KeyMap) StatusBinds(showAll bool) []key.Help {
	toggle := km.ToggleRevisions.Help()
	if showAll {
		toggle.Desc = "current tree"
	}
	binds := []key.Help{
		km.Open.Help(),
	}
	binds = append(binds, km.NavigationBinds()...)
	binds = append(binds, toggle)
	return binds
}
