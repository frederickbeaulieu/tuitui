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
			key.WithHelp("s", "split/inline"),
		),
		ToggleContext: key.NewBinding(
			key.WithKeys("z"),
			key.WithHelp("z", "full file"),
		),
	}
}

func (km KeyMap) StatusBinds(showFullFile bool, plainFile bool) []key.Help {
	binds := []key.Help{
		km.Back.Help(),
	}
	binds = append(binds, km.ScrollBinds()...)
	if !plainFile {
		context := km.ToggleContext.Help()
		if showFullFile {
			context.Desc = "changes only"
		}
		binds = append(binds,
			km.ToggleLayout.Help(),
			context,
		)
	}
	return binds
}
