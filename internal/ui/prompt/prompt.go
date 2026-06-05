// Package prompt provides a modal yes/no confirmation dialog that
// renders as a centered overlay on top of the main view.
//
// Lifecycle
//
//  1. Caller starts a prompt: `promptModel.Start(token, message)` returns
//     a tea.Cmd that, when executed, activates the prompt.
//  2. While `Active()` the prompt consumes all key events: `y` / `Y` →
//     accepted, anything else (n, esc, enter, …) → rejected.
//  3. On resolution a ResultMsg{Token, Accepted} is emitted so the
//     caller can correlate the answer to the request that spawned it.
//
// The prompt itself is state + behavior; rendering goes through
// internal/ui/overlay so the same overlay machinery can back future
// settings menus, dialogs, pickers, etc.
package prompt

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
	"github.com/frederickbeaulieu/tuitui/internal/ui/overlay"
)

// ResultMsg is emitted when the user answers the prompt. Token echoes
// whatever string was passed to Start so the caller can distinguish
// between concurrent prompt requests.
type ResultMsg struct {
	Token    string
	Accepted bool
}

// activateMsg is the internal signal produced by Start's tea.Cmd.
type activateMsg struct {
	token   string
	message string
}

// Model holds the prompt's state. It is modal: when Active() returns
// true the app should route all key events to Update and should not
// process them elsewhere.
type Model struct {
	active  bool
	token   string
	message string
}

// New returns an inactive prompt.
func New() Model { return Model{} }

// Start returns a command that, when executed, places the prompt into
// active state with the given message and correlation token.
func Start(token, message string) tea.Cmd {
	return func() tea.Msg {
		return activateMsg{token: token, message: message}
	}
}

// Active reports whether the prompt is currently awaiting an answer.
func (m Model) Active() bool { return m.active }

// Update processes messages. When active, key events are consumed and
// resolved to a ResultMsg. The activation message flips the model on.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case activateMsg:
		m.active = true
		m.token = msg.token
		m.message = msg.message
		return m, nil
	case tea.KeyPressMsg:
		if !m.active {
			return m, nil
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	accepted := false
	switch msg.String() {
	case "y", "Y":
		accepted = true
	}
	token := m.token
	m.active = false
	m.token = ""
	m.message = ""
	return m, func() tea.Msg {
		return ResultMsg{Token: token, Accepted: accepted}
	}
}

// Overlay returns the prompt rendered as a centered overlay layer, or
// nil when not active. Width and height are the dimensions of the
// underlying view so the overlay can be sized responsively if needed.
func (m Model) Overlay(baseWidth, baseHeight int) *overlay.Layer {
	if !m.active {
		return nil
	}
	body := m.renderBody(baseWidth)
	return overlay.Centered(body)
}

func (m Model) renderBody(baseWidth int) string {
	// Target width: fit the message with padding, but never exceed half
	// the screen. Minimum width keeps the hint line readable.
	const minWidth = 30
	const maxFraction = 2 // width ≤ baseWidth / maxFraction
	wantWidth := len(m.message) + 4
	maxWidth := max(baseWidth/maxFraction, minWidth)
	w := max(min(wantWidth, maxWidth), minWidth)

	msgStyle := lipgloss.NewStyle().
		Foreground(common.ColorText)
	hintKey := lipgloss.NewStyle().
		Foreground(common.ColorMauve).
		Bold(true)
	hintDesc := lipgloss.NewStyle().
		Foreground(common.ColorSubtext)

	hint := hintKey.Render("y") + hintDesc.Render(" yes") +
		"  " +
		hintKey.Render("n/esc") + hintDesc.Render(" no")

	inner := msgStyle.Render(m.message) + "\n\n" + hint

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(common.ColorMauve).
		Padding(0, 1).
		Width(w).
		Render(inner)
}
