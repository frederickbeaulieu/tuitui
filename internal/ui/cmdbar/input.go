package cmdbar

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/google/shlex"
)

func (m Model) InputView() string {
	if !m.active {
		return ""
	}
	return m.input.View()
}

func (m Model) handleResult(msg CmdResultMsg) (Model, tea.Cmd) {
	m.active = false
	m.input.Blur()
	m.scroll = 0
	if msg.Err != nil {
		m.lines = strings.Split(msg.Err.Error(), "\n")
		m.showingError = true
		return m, nil
	}
	m.showingError = false
	return m, func() tea.Msg { return CmdCloseMsg{} }
}

func (m Model) handleInputKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if key.Matches(msg, m.keymap.Close) {
		return m.dismiss()
	}

	if key.Matches(msg, m.keymap.Submit) {
		cmd := strings.TrimSpace(m.input.Value())
		if cmd == "" {
			return m.dismiss()
		}
		runner := m.runner
		return m, func() tea.Msg {
			args, err := shlex.Split(cmd)
			if err != nil {
				return CmdResultMsg{Err: err}
			}
			_, err = runner.Run(args...)
			return CmdResultMsg{Err: err}
		}
	}

	var c tea.Cmd
	m.input, c = m.input.Update(msg)
	return m, c
}

func (m Model) dismiss() (Model, tea.Cmd) {
	m.active = false
	m.input.Blur()
	return m, func() tea.Msg { return CmdCloseMsg{} }
}
