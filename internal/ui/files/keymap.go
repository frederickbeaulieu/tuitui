package files

import (
	"charm.land/bubbles/v2/key"

	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

type KeyMap struct {
	common.NavigationKeyMap
	Select key.Binding
	Back   key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		NavigationKeyMap: common.DefaultNavigationKeyMap(),
		Select: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "open diff"),
		),
		Back: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "back"),
		),
	}
}
