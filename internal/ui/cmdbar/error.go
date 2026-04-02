package cmdbar

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

func (m Model) ErrorView() string {
	if !m.showingError || len(m.lines) == 0 {
		return ""
	}
	visible := m.height
	if visible <= 0 {
		visible = 40
	}
	end := min(m.scroll+visible, len(m.lines))
	return strings.Join(m.lines[m.scroll:end], "\n")
}

func (m Model) ErrorTitle() string {
	return "Error"
}

func (m Model) ErrorStatusBinds() []key.Help {
	return errorStatusBinds(m.keymap)
}

func (m Model) handleErrorViewerKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if key.Matches(msg, m.keymap.ClosePanel) {
		m.showingError = false
		return m, func() tea.Msg { return CmdCloseMsg{} }
	}

	if newOffset, ok := m.keymap.HandleScroll(msg, m.scroll, len(m.lines)-1, m.height/2); ok {
		m.scroll = newOffset
	}
	return m, nil
}

func errorStatusBinds(km common.KeyMap) []key.Help {
	binds := []key.Help{
		km.ClosePanel.Help(),
	}
	binds = append(binds, km.ScrollBinds()...)
	return binds
}
