// Package cmdbar implements a vim-style command bar for running jj commands.
package cmdbar

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/shlex"

	"github.com/frederickbeaulieu/tuitui/internal/jj"
	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

// ---------------------------------------------------------------------------
// Messages
// ---------------------------------------------------------------------------

type CmdCloseMsg struct{}

type CmdResultMsg struct {
	Err error
}

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Command input
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Error viewer
// ---------------------------------------------------------------------------

func (m Model) ErrorView() string {
	if !m.showingError || len(m.lines) == 0 {
		return ""
	}
	visible := m.height
	if visible <= 0 {
		visible = 40
	}
	end := min(m.scroll+visible, len(m.lines))
	var b strings.Builder
	for i := m.scroll; i < end; i++ {
		if i > m.scroll {
			b.WriteString("\n")
		}
		b.WriteString(m.lines[i])
	}
	return b.String()
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
