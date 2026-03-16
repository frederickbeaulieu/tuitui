package diff

import (
	"charm.land/bubbles/v2/key"

	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

type KeyMap struct {
	common.KeyMap
	ToggleLayout  key.Binding
	ToggleContext key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		KeyMap: common.DefaultKeyMap(),
		ToggleLayout: key.NewBinding(
			key.WithKeys("s"),
		),
		ToggleContext: key.NewBinding(
			key.WithKeys("z"),
		),
	}
}
