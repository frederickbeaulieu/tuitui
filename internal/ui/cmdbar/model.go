// Package cmdbar implements a vim-style command bar for running jj commands.
package cmdbar

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/frederickbeaulieu/tuitui/internal/jj"
	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

// CmdCloseMsg is sent when the command bar is dismissed.
type CmdCloseMsg struct{}

// CmdResultMsg carries the result of an executed jj command.
type CmdResultMsg struct {
	Err error
}

type Model struct {
	runner *jj.Runner
	input  textinput.Model
	keymap common.KeyMap
	active bool

	showingError bool
	scroll       int
	lines        []string
	width        int
	height       int
}

func New(runner *jj.Runner) Model {
	ti := textinput.New()
	ti.Prompt = ":" + lipgloss.NewStyle().Foreground(common.ColorOverlay).Render("jj ")
	ti.CharLimit = 256
	return Model{runner: runner, input: ti, keymap: common.DefaultKeyMap()}
}

func (m Model) Active() bool       { return m.active }
func (m Model) ShowingError() bool { return m.showingError }

func (m *Model) Activate() tea.Cmd {
	m.active = true
	m.showingError = false
	m.input.SetValue("")
	return m.input.Focus()
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.input.SetWidth(width - 1)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case CmdResultMsg:
		return m.handleResult(msg)
	case tea.KeyPressMsg:
		if m.showingError {
			return m.handleErrorViewerKey(msg)
		}
		if m.active {
			return m.handleInputKey(msg)
		}
	}

	if m.active {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}
