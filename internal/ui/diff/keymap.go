package diff

import (
	"charm.land/bubbles/v2/key"

	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

type KeyMap struct {
	common.NavigationKeyMap
	Close         key.Binding
	Refresh       key.Binding
	ToggleLayout  key.Binding
	ToggleContext key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		NavigationKeyMap: common.DefaultNavigationKeyMap(),
		Close: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "back"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		ToggleLayout: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "toggle split/inline"),
		),
		ToggleContext: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "toggle full file/changes"),
		),
	}
}
