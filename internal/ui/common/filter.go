package common

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// FilterBar manages a filter text input with show/hide state.
// Panels embed this and call ApplyFilter when the value changes.
type FilterBar struct {
	Input     textinput.Model
	Filtering bool
}

// NewFilterBar creates a FilterBar with the standard "/" prompt style.
func NewFilterBar() FilterBar {
	fi := textinput.New()
	fi.Prompt = "/"
	s := fi.Styles()
	s.Focused.Prompt = lipgloss.NewStyle().Foreground(ColorMauve).Bold(true)
	s.Focused.Text = lipgloss.NewStyle().Foreground(ColorText)
	fi.SetStyles(s)
	return FilterBar{Input: fi}
}

// Value returns the current filter text.
func (f *FilterBar) Value() string { return f.Input.Value() }

// Active returns true when the filter bar should be visible.
func (f *FilterBar) Active() bool {
	return f.Filtering || f.Input.Value() != ""
}

// StartFilter activates the filter input and returns a focus command.
func (f *FilterBar) StartFilter() tea.Cmd {
	f.Filtering = true
	f.Input.Reset()
	return f.Input.Focus()
}

// ClearFilter resets the filter text without changing filtering state.
func (f *FilterBar) ClearFilter() {
	f.Input.Reset()
}

// Reset fully resets the filter bar to inactive state.
func (f *FilterBar) Reset() {
	f.Input.Reset()
	f.Filtering = false
}

// HandleKey handles common filter input keys (esc to cancel, enter to apply).
// Returns (updated FilterBar, tea.Cmd, handled bool).
// When handled is true, the caller should run its applyFilter logic.
func (f *FilterBar) HandleKey(msg tea.KeyPressMsg, close key.Binding) (tea.Cmd, bool) {
	if !f.Filtering {
		return nil, false
	}

	switch {
	case key.Matches(msg, close):
		f.Input.Reset()
		f.Filtering = false
		f.Input.Blur()
		return nil, true

	case msg.Code == tea.KeyEnter:
		f.Filtering = false
		f.Input.Blur()
		return nil, true

	default:
		var cmd tea.Cmd
		f.Input, cmd = f.Input.Update(msg)
		return cmd, true
	}
}

// Update forwards non-key messages to the text input when filtering is active.
func (f *FilterBar) Update(msg tea.Msg) tea.Cmd {
	if !f.Filtering {
		return nil
	}
	var cmd tea.Cmd
	f.Input, cmd = f.Input.Update(msg)
	return cmd
}
