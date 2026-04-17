package files

import (
	"charm.land/bubbles/v2/key"

	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

type KeyMap struct {
	common.KeyMap
	ToggleAllFiles key.Binding
	Filter         key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		KeyMap: common.DefaultKeyMap(),
		ToggleAllFiles: key.NewBinding(
			key.WithKeys("z"),
			key.WithHelp("z", "all files"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
	}
}

func (km KeyMap) StatusBinds(showAll, filtering, hasFilter bool) []key.Help {
	if filtering {
		return []key.Help{
			{Key: "esc", Desc: "cancel"},
			{Key: "enter", Desc: "apply"},
		}
	}

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
	binds = append(binds, km.Edit.Help())

	if hasFilter {
		binds = append(binds, key.Help{Key: "esc", Desc: "clear filter"})
	} else {
		binds = append(binds, km.Filter.Help())
	}

	return binds
}
