package app

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

func (m Model) View() tea.View {
	panelHeight := m.panelHeight()

	if m.cmdbar.ShowingError() {
		return m.viewError(panelHeight)
	}

	bottomBar := m.viewBottomBar()
	suggestionsPanel, completionHeight := m.viewSuggestions()
	mainPanel := m.viewMainPanel(panelHeight)

	var content string
	if suggestionsPanel != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, mainPanel, suggestionsPanel, bottomBar)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left, mainPanel, bottomBar)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	if m.cmdbar.Active() {
		if c := m.cmdbar.Cursor(); c != nil {
			c.Position.Y += panelHeight + completionHeight
			v.Cursor = c
		}
	}
	return v
}

func (m Model) viewError(panelHeight int) tea.View {
	errorContent := m.cmdbar.ErrorView()
	errorPanel := common.RenderPanel(m.cmdbar.ErrorTitle(), errorContent, m.width, panelHeight, true)
	statusBar := m.renderStatusBarWith(m.cmdbar.ErrorStatusBinds())
	content := lipgloss.JoinVertical(lipgloss.Left, errorPanel, statusBar)
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m Model) viewBottomBar() string {
	if m.cmdbar.Active() {
		statusBar := m.renderStatusBarWith(m.cmdbar.InputStatusBinds())
		return lipgloss.JoinVertical(lipgloss.Left, m.cmdbar.InputView(), statusBar)
	}
	return m.viewStatusBar()
}

func (m Model) viewSuggestions() (string, int) {
	bottomBarLines := statusBarHeight
	if m.cmdbar.Active() {
		bottomBarLines = cmdbarHeight
	}
	availableHeight := m.height - bottomBarLines

	completionHeight := m.cmdbar.CompletionHeight(availableHeight)
	if completionHeight == 0 {
		return "", 0
	}

	content := m.cmdbar.CompletionView(completionHeight - 2)
	panel := common.RenderPanel("", content, m.width, completionHeight, true)
	return panel, completionHeight
}

func (m Model) viewMainPanel(panelHeight int) string {
	switch m.mode {
	case modeLog:
		logContent := m.log.View()
		return common.RenderPanel("Log", logContent, m.width, panelHeight, true)
	case modeFiles:
		logWidth, filesWidth := m.splitWidths()
		logPanel := common.RenderPanel("Log", m.log.View(), logWidth, panelHeight, false)
		filesPanel := common.RenderPanel("Files", m.files.View(), filesWidth, panelHeight, true)
		return lipgloss.JoinHorizontal(lipgloss.Top, logPanel, filesPanel)
	case modeDiff:
		diffContent := m.diff.View()
		return common.RenderPanel("Diff", diffContent, m.width, panelHeight, true)
	default:
		return ""
	}
}

func (m Model) viewStatusBar() string {
	binds := []key.Help{m.keymap.Quit.Help()}
	switch m.mode {
	case modeLog:
		binds = append(binds, m.log.StatusBinds()...)
	case modeFiles:
		binds = append(binds, m.files.StatusBinds()...)
	case modeDiff:
		binds = append(binds, m.diff.StatusBinds()...)
	}
	binds = append(binds, m.keymap.Command.Help())
	return m.renderStatusBarWith(binds)
}

func (m Model) renderStatusBarWith(binds []key.Help) string {
	keyStyle := lipgloss.NewStyle().
		Background(common.ColorSurface).
		Foreground(common.ColorMauve).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Background(common.ColorSurface).
		Foreground(common.ColorSubtext)

	sepStyle := lipgloss.NewStyle().
		Background(common.ColorSurface).
		Foreground(common.ColorOverlay)

	sep := sepStyle.Render("  |  ")

	var parts []string
	for i, b := range binds {
		part := keyStyle.Render(b.Key) + descStyle.Render(" "+b.Desc)
		parts = append(parts, part)
		if i < len(binds)-1 {
			parts = append(parts, sep)
		}
	}

	bar := strings.Join(parts, "")

	barPlain := common.StripAnsi(bar)
	barVisualLen := common.VisualLen(barPlain)
	if barVisualLen < m.width {
		padding := lipgloss.NewStyle().
			Background(common.ColorSurface).
			Width(m.width - barVisualLen).
			Render("")
		bar = bar + padding
	}

	return bar
}
