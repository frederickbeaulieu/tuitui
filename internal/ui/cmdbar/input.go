package cmdbar

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
	"github.com/google/shlex"
)

func (m Model) InputStatusBinds() []key.Help {
	return []key.Help{
		{Key: "esc", Desc: "close"},
		{Key: "C-space", Desc: "suggestions"},
		{Key: "C-n/C-p", Desc: "navigate"},
		{Key: "tab", Desc: "complete"},
		{Key: "enter", Desc: "run"},
	}
}

func (m Model) handleResult(msg CmdResultMsg) (Model, tea.Cmd) {
	m.active = false
	m.input.Blur()
	m.completion.visible = false
	m.completion.showDropdown = false
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
	switch {
	case key.Matches(msg, m.keymap.CompleteToggle):
		return m.handleCompleteToggle()
	case key.Matches(msg, m.keymap.Close):
		return m.dismiss()
	case key.Matches(msg, m.keymap.CompleteTab):
		return m.handleCompleteTab()
	case key.Matches(msg, m.keymap.CompleteNext):
		return m.handleCompleteNext()
	case key.Matches(msg, m.keymap.CompletePrev):
		return m.handleCompletePrev()
	case key.Matches(msg, m.keymap.Submit):
		return m.handleSubmit()
	default:
		return m.handleTextInput(msg)
	}
}

func (m Model) handleCompleteToggle() (Model, tea.Cmd) {
	m.pinDropdown = !m.pinDropdown
	m.completion.showDropdown = m.pinDropdown
	if !m.completion.showDropdown {
		m.completion.selected = -1
	}
	return m, nil
}

func (m Model) handleCompleteTab() (Model, tea.Cmd) {
	if m.completion.visible && len(m.completion.items) > 0 {
		if m.completion.selected < 0 {
			m.completion.selected = 0
		}
		return m.acceptSelected()
	}
	m.completion.tag++
	return m, requestCompletion(m.runner, m.input.Value(), m.completion.tag)
}

func (m Model) handleCompleteNext() (Model, tea.Cmd) {
	if m.completion.visible && len(m.completion.items) > 0 {
		m.completion.showDropdown = true
		m.completion.selected++
		if m.completion.selected >= len(m.completion.items) {
			m.completion.selected = 0
		}
	}
	return m, nil
}

func (m Model) handleCompletePrev() (Model, tea.Cmd) {
	if m.completion.visible && len(m.completion.items) > 0 {
		m.completion.showDropdown = true
		m.completion.selected--
		if m.completion.selected < 0 {
			m.completion.selected = len(m.completion.items) - 1
		}
	}
	return m, nil
}

func (m Model) handleSubmit() (Model, tea.Cmd) {
	if m.completion.showDropdown && m.completion.selected >= 0 && m.completion.selected < len(m.completion.items) {
		return m.acceptSelected()
	}
	cmd := strings.TrimSpace(m.input.Value())
	if cmd == "" {
		return m.dismiss()
	}
	m.completion.visible = false
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

func (m Model) handleTextInput(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	prevValue := m.input.Value()
	var c tea.Cmd
	m.input, c = m.input.Update(msg)
	newValue := m.input.Value()

	if strings.HasPrefix(newValue, "jj ") {
		newValue = strings.TrimPrefix(newValue, "jj ")
		m.input.SetValue(newValue)
		m.input.SetCursor(len(newValue))
	}

	var cmds []tea.Cmd
	if c != nil {
		cmds = append(cmds, c)
	}

	if newValue != prevValue {
		m.completion.tag++
		cmds = append(cmds, requestCompletionTick(m.completion.tag, newValue))
	}

	return m, tea.Batch(cmds...)
}

func (m Model) InputView() string {
	if !m.active {
		return ""
	}

	if m.input.Position() != len([]rune(m.input.Value())) {
		return m.input.View()
	}

	ghost := ghostText(m.input.Value(), m.completion)
	if ghost == "" {
		return m.input.View()
	}

	// Disable width so View() doesn't pad with trailing spaces, then strip
	// the hidden virtual cursor character so ghost text follows with no gap.
	savedWidth := m.input.Width()
	m.input.SetWidth(0)
	view := m.input.View()
	m.input.SetWidth(savedWidth)

	viewWidth := lipgloss.Width(view)
	if viewWidth > 0 {
		view = ansi.Truncate(view, viewWidth-1, "")
	}

	ghostStyled := lipgloss.NewStyle().Foreground(common.ColorOverlay).Render(ghost)
	return view + ghostStyled
}

func (m Model) CompletionView(maxItems int) string {
	return completionView(m.completion, m.width-2, maxItems)
}

func (m Model) acceptSelected() (Model, tea.Cmd) {
	item := m.completion.items[m.completion.selected]
	newInput := acceptCompletion(m.input.Value(), item.Value)
	m.input.SetValue(newInput)
	m.input.SetCursor(len(newInput))
	m.completion.visible = false
	m.completion.items = nil
	m.completion.selected = -1

	if !m.pinDropdown {
		m.completion.showDropdown = false
	}

	m.completion.tag++
	return m, requestCompletion(m.runner, newInput, m.completion.tag)
}

func (m Model) dismiss() (Model, tea.Cmd) {
	m.active = false
	m.completion.visible = false
	m.completion.showDropdown = false
	m.completion.items = nil
	m.completion.selected = -1
	m.input.Blur()
	return m, func() tea.Msg { return CmdCloseMsg{} }
}
