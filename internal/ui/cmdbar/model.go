// Package cmdbar implements a command bar for running jj commands.
package cmdbar

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/frederickbeaulieu/tuitui/internal/jj"
	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

type CmdCloseMsg struct{}

type CmdResultMsg struct {
	Err error
}

type Model struct {
	runner      *jj.Runner
	input       textinput.Model
	keymap      common.KeyMap
	active      bool
	completion  completionState
	pinDropdown bool

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
	ti.SetVirtualCursor(false)
	return Model{runner: runner, input: ti, keymap: common.DefaultKeyMap()}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.input.SetWidth(width - 1)
}

func (m Model) Active() bool        { return m.active }
func (m Model) ShowingError() bool  { return m.showingError }
func (m Model) Cursor() *tea.Cursor { return m.input.Cursor() }

// CompletionHeight returns the total height (including borders) of the
// suggestions panel. Returns 0 when hidden or when there isn't enough
// vertical space. availableHeight is the terminal height minus the bottom bar.
func (m Model) CompletionHeight(availableHeight int) int {
	if !m.completion.visible || !m.completion.showDropdown || len(m.completion.items) == 0 {
		return 0
	}
	n := len(m.completion.items)
	if n > maxVisibleCompletions {
		n = maxVisibleCompletions
	}
	totalHeight := n + 2

	maxAllowed := availableHeight - minPanelHeight
	if maxAllowed < 3 {
		return 0
	}
	if totalHeight > maxAllowed {
		totalHeight = maxAllowed
	}
	return totalHeight
}

func (m *Model) Activate() tea.Cmd {
	m.active = true
	m.showingError = false
	m.input.SetValue("")
	m.completion = completionState{selected: -1, showDropdown: m.pinDropdown}
	focusCmd := m.input.Focus()
	m.completion.tag++
	completeCmd := requestCompletion(m.runner, "", m.completion.tag)
	return tea.Batch(focusCmd, completeCmd)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case CmdResultMsg:
		return m.handleResult(msg)
	case CompletionMsg:
		if msg.Tag == m.completion.tag {
			m.completion.items = msg.Items
			m.completion.selected = -1
			m.completion.visible = len(msg.Items) > 0
		}
		return m, nil
	case completionTickMsg:
		if msg.Tag == m.completion.tag {
			return m, requestCompletion(m.runner, msg.Input, msg.Tag)
		}
		return m, nil
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
