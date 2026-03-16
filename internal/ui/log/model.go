package log

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/frederickbeaulieu/tuitui/internal/jj"
	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

// LogDataMsg carries log entries from an async fetch.
type LogDataMsg struct {
	Entries []jj.GraphEntry
	Err     error
}

type RepoChangedMsg struct{}

type CursorChangedMsg struct {
	ChangeID string
}

type LogSelectMsg struct {
	ChangeID string
}

// Model is the log panel.
type Model struct {
	runner       *jj.Runner
	watcher      *jj.RepoWatcher
	entries      []jj.GraphEntry
	cursor       int
	offset       int
	width        int
	height       int
	focused      bool
	keymap       common.KeyMap
	err          error
	prevChangeID string
}

func New(runner *jj.Runner, watcher *jj.RepoWatcher) Model {
	return Model{
		runner:  runner,
		watcher: watcher,
		focused: true,
		keymap:  common.DefaultKeyMap(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.fetchEntries(), m.awaitRepoChange())
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) Focus() { m.focused = true }

func (m *Model) Blur() { m.focused = false }

func (m Model) Focused() bool { return m.focused }

func (m Model) SelectedChangeID() string {
	if len(m.entries) == 0 || m.cursor >= len(m.entries) {
		return ""
	}
	return m.entries[m.cursor].Commit.ChangeID
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case LogDataMsg:
		m.entries = msg.Entries
		m.err = msg.Err
		if m.cursor >= len(m.entries) && len(m.entries) > 0 {
			m.cursor = len(m.entries) - 1
		}
		return m, m.emitCursorChanged()

	case RepoChangedMsg:
		return m, tea.Batch(m.fetchEntries(), m.awaitRepoChange())

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Up):
		if m.cursor > 0 {
			m.cursor--
			m.ensureVisible()
		}
		return m, m.emitCursorChanged()

	case key.Matches(msg, m.keymap.Down):
		if m.cursor < len(m.entries)-1 {
			m.cursor++
			m.ensureVisible()
		}
		return m, m.emitCursorChanged()

	case key.Matches(msg, m.keymap.HalfUp):
		m.cursor -= m.viewportHeight() / 4
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.ensureVisible()
		return m, m.emitCursorChanged()

	case key.Matches(msg, m.keymap.HalfDown):
		m.cursor += m.viewportHeight() / 4
		if m.cursor >= len(m.entries) {
			m.cursor = len(m.entries) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.ensureVisible()
		return m, m.emitCursorChanged()

	case key.Matches(msg, m.keymap.Top):
		m.cursor = 0
		m.offset = 0
		return m, m.emitCursorChanged()

	case key.Matches(msg, m.keymap.Bottom):
		if len(m.entries) > 0 {
			m.cursor = len(m.entries) - 1
			m.ensureVisible()
		}
		return m, m.emitCursorChanged()

	case key.Matches(msg, m.keymap.Open):
		id := m.SelectedChangeID()
		if id != "" {
			return m, func() tea.Msg {
				return LogSelectMsg{ChangeID: id}
			}
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) emitCursorChanged() tea.Cmd {
	id := m.SelectedChangeID()
	if id == m.prevChangeID {
		return nil
	}
	m.prevChangeID = id
	return func() tea.Msg {
		return CursorChangedMsg{ChangeID: id}
	}
}

func (m Model) fetchEntries() tea.Cmd {
	runner := m.runner
	return func() tea.Msg {
		entries, err := runner.LogGraphEntries("")
		return LogDataMsg{Entries: entries, Err: err}
	}
}

func (m Model) awaitRepoChange() tea.Cmd {
	if m.watcher == nil {
		return nil
	}
	ch := m.watcher.C
	return func() tea.Msg {
		<-ch
		return RepoChangedMsg{}
	}
}

func (m Model) View() string {
	if m.err != nil {
		return common.ConflictStyle.Render("Error: " + m.err.Error())
	}
	if len(m.entries) == 0 {
		return common.TextMuted.Render("No commits found")
	}
	return m.renderGraph()
}

func (m Model) renderGraph() string {
	availableLines := m.viewportHeight()
	if availableLines <= 0 {
		return ""
	}

	var b strings.Builder
	linesUsed := 0

	for i := m.offset; i < len(m.entries) && linesUsed < availableLines; i++ {
		entry := m.entries[i]
		isCurrent := i == m.cursor

		for _, line := range entry.Lines {
			if linesUsed >= availableLines {
				break
			}

			displayLine := common.Truncate(line, m.width)

			if isCurrent {
				displayLine = common.HighlightLine(displayLine, m.width)
			}

			if linesUsed > 0 {
				b.WriteString("\n")
			}
			b.WriteString(displayLine)
			linesUsed++
		}
	}

	return b.String()
}

func (m Model) viewportHeight() int {
	if m.height <= 0 {
		return 40
	}
	return m.height
}

func (m *Model) ensureVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}

	linesNeeded := 0
	for i := m.offset; i <= m.cursor && i < len(m.entries); i++ {
		linesNeeded += len(m.entries[i].Lines)
	}

	available := m.viewportHeight()
	for linesNeeded > available && m.offset < m.cursor {
		linesNeeded -= len(m.entries[m.offset].Lines)
		m.offset++
	}
}
