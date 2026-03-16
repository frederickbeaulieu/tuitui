package app

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/frederickbeaulieu/tuitui/internal/jj"
	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
	"github.com/frederickbeaulieu/tuitui/internal/ui/diff"
	"github.com/frederickbeaulieu/tuitui/internal/ui/files"
	logpanel "github.com/frederickbeaulieu/tuitui/internal/ui/log"
)

const statusBarHeight = 1

type mode int

const (
	modeLog mode = iota
	modeFiles
	modeDiff
)

// Model is the root Bubble Tea model.
type Model struct {
	log    logpanel.Model
	files  files.Model
	diff   diff.Model
	mode   mode
	width  int
	height int
	keymap common.KeyMap
}

func New(runner *jj.Runner, watcher *jj.RepoWatcher) Model {
	return Model{
		log:    logpanel.New(runner, watcher),
		files:  files.New(runner),
		diff:   diff.New(runner),
		mode:   modeLog,
		keymap: common.DefaultKeyMap(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.log.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layoutPanels()
		return m, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keymap.Quit):
			return m, tea.Quit
		}

	case logpanel.LogSelectMsg:
		m.mode = modeFiles
		m.log.Blur()
		m.files.Focus()
		m.diff.Blur()
		m.layoutPanels()
		cmd := m.files.SetRevision(msg.ChangeID)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case files.FileSelectedMsg:
		m.mode = modeDiff
		m.log.Blur()
		m.files.Blur()
		m.diff.Focus()
		m.layoutPanels()
		cmd := m.diff.SetRevisionFile(msg.ChangeID, msg.Path)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case files.FilesCloseMsg:
		m.mode = modeLog
		m.log.Focus()
		m.files.Blur()
		m.diff.Blur()
		m.layoutPanels()
		return m, nil

	case diff.DiffCloseMsg:
		m.mode = modeFiles
		m.log.Blur()
		m.files.Focus()
		m.diff.Blur()
		m.layoutPanels()
		return m, nil

	case logpanel.CursorChangedMsg:
		if m.mode == modeFiles && msg.ChangeID != "" {
			cmd := m.files.SetRevision(msg.ChangeID)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case logpanel.RepoChangedMsg:
		if m.mode == modeFiles || m.mode == modeDiff {
			cmd := m.files.Refresh()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		if m.mode == modeDiff {
			cmd := m.diff.Refresh()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	var cmd tea.Cmd

	m.log, cmd = m.log.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	if m.mode == modeFiles || m.mode == modeDiff {
		m.files, cmd = m.files.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if m.mode == modeDiff {
		m.diff, cmd = m.diff.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() tea.View {
	panelHeight := m.panelHeight()
	var content string

	switch m.mode {
	case modeLog:
		logContent := m.log.View()
		logPanel := common.RenderPanel("Log", logContent, m.width, panelHeight, true)
		statusBar := m.renderStatusBar()
		content = lipgloss.JoinVertical(lipgloss.Left, logPanel, statusBar)

	case modeFiles:
		logWidth, filesWidth := m.splitWidths()
		logContent := m.log.View()
		filesContent := m.files.View()
		logPanel := common.RenderPanel("Log", logContent, logWidth, panelHeight, false)
		filesPanel := common.RenderPanel("Files", filesContent, filesWidth, panelHeight, true)
		panels := lipgloss.JoinHorizontal(lipgloss.Top, logPanel, filesPanel)
		statusBar := m.renderStatusBar()
		content = lipgloss.JoinVertical(lipgloss.Left, panels, statusBar)

	case modeDiff:
		diffContent := m.diff.View()
		diffPanel := common.RenderPanel("Diff", diffContent, m.width, panelHeight, true)
		statusBar := m.renderStatusBar()
		content = lipgloss.JoinVertical(lipgloss.Left, diffPanel, statusBar)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m Model) splitWidths() (int, int) {
	logWidth := max(m.width*3/5, 30)
	filesWidth := max(m.width-logWidth, 20)
	return logWidth, filesWidth
}

func (m Model) panelHeight() int {
	return max(m.height-statusBarHeight, 3)
}

func (m *Model) layoutPanels() {
	panelHeight := m.panelHeight()

	switch m.mode {
	case modeLog:
		m.log.SetSize(m.width-2, panelHeight-2)
	case modeFiles:
		logWidth, filesWidth := m.splitWidths()
		m.log.SetSize(logWidth-2, panelHeight-2)
		m.files.SetSize(filesWidth-2, panelHeight-2)
	case modeDiff:
		m.diff.SetSize(m.width-2, panelHeight-2)
	}
}

func (m Model) renderStatusBar() string {
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

	type bind struct {
		key  string
		desc string
	}

	binds := []bind{
		{"q", "quit"},
	}

	switch m.mode {
	case modeLog:
		binds = append(binds,
			bind{"l", "open"},
			bind{"j/k", "navigate"},
			bind{"C-d/C-u", "half page"},
			bind{"g/G", "top/bottom"},
		)

	case modeFiles:
		binds = append(binds,
			bind{"h", "back"},
			bind{"l", "open"},
			bind{"j/k", "navigate"},
			bind{"C-d/C-u", "half page"},
			bind{"g/G", "top/bottom"},
		)

	case modeDiff:
		contextLabel := "full file"
		if m.diff.ShowFullFile() {
			contextLabel = "changes only"
		}
		binds = append(binds,
			bind{"h", "back"},
			bind{"j/k", "scroll"},
			bind{"C-d/C-u", "half page"},
			bind{"g/G", "top/bottom"},
			bind{"s", "split/inline"},
			bind{"z", contextLabel},
		)
	}

	var parts []string
	for i, b := range binds {
		part := keyStyle.Render(b.key) + descStyle.Render(" "+b.desc)
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
