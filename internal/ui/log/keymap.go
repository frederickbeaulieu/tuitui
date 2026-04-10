package log

import (
	"charm.land/bubbles/v2/key"

	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

type KeyMap struct {
	common.KeyMap
	ToggleRevisions key.Binding
	Filter          key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		KeyMap: common.DefaultKeyMap(),
		ToggleRevisions: key.NewBinding(
			key.WithKeys("z"),
			key.WithHelp("z", "all revisions"),
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

	toggle := km.ToggleRevisions.Help()
	if showAll {
		toggle.Desc = "current tree"
	}

	binds := []key.Help{
		km.Open.Help(),
	}
	binds = append(binds, km.NavigationBinds()...)
	binds = append(binds, toggle)

	if hasFilter {
		binds = append(binds, key.Help{Key: "esc", Desc: "clear filter"})
	} else {
		binds = append(binds, km.Filter.Help())
	}

	return binds
}
