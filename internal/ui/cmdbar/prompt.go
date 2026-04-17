package cmdbar

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

// PromptResultMsg is emitted when the user answers a yes/no prompt.
// Token is the value passed to StartPrompt and lets callers correlate
// results with the prompt they issued.
type PromptResultMsg struct {
	Token    string
	Accepted bool
}

// StartPrompt puts the cmdbar into a blocking yes/no confirm state.
// The bar displays `message (y/N)`. Pressing y or Y emits
// PromptResultMsg{Accepted: true}; anything else (n, esc, enter, …)
// emits PromptResultMsg{Accepted: false}.
//
// Token is echoed back in the result so the caller can match it to
// the original request without keeping per-prompt state in the cmdbar.
func (m *Model) StartPrompt(token, message string) tea.Cmd {
	m.prompting = true
	m.active = false
	m.showingError = false
	m.promptToken = token
	m.promptMessage = message
	return nil
}

// Prompting reports whether the cmdbar is currently awaiting a y/n answer.
func (m Model) Prompting() bool { return m.prompting }

// PromptView renders the prompt line. Empty when not prompting.
func (m Model) PromptView() string {
	if !m.prompting {
		return ""
	}
	prefix := lipgloss.NewStyle().
		Foreground(common.ColorMauve).
		Bold(true).
		Render("? ")
	msg := m.promptMessage
	hint := lipgloss.NewStyle().
		Foreground(common.ColorOverlay).
		Render(" (y/N)")
	return prefix + msg + hint
}

// PromptStatusBinds are the hints shown while prompting.
func (m Model) PromptStatusBinds() []key.Help {
	return []key.Help{
		{Key: "y", Desc: "yes"},
		{Key: "n/esc", Desc: "no"},
	}
}

func (m Model) handlePromptKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	accepted := false
	switch msg.String() {
	case "y", "Y":
		accepted = true
	}
	token := m.promptToken
	m.prompting = false
	m.promptToken = ""
	m.promptMessage = ""
	return m, func() tea.Msg {
		return PromptResultMsg{Token: token, Accepted: accepted}
	}
}
