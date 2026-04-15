// Package log implements the log panel for the revision graph.
package log

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/frederickbeaulieu/tuitui/internal/jj"
	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

const allRevset = "all()"

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

type Model struct {
	runner       *jj.Runner
	watcher      *jj.RepoWatcher
	entries      []jj.GraphEntry
	cursor       int
	offset       int
	width        int
	height       int
	focused      bool
	keymap       KeyMap
	showAll      bool
	filter       common.FilterBar
	filtered     []filteredEntry
	err          error
	prevChangeID string
}

func New(runner *jj.Runner, watcher *jj.RepoWatcher) Model {
	return Model{
		runner:  runner,
		watcher: watcher,
		focused: true,
		filter:  common.NewFilterBar(),
		keymap:  DefaultKeyMap(),
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
	entries := m.visibleEntries()
	if len(entries) == 0 || m.cursor >= len(entries) {
		return ""
	}
	return entries[m.cursor].Commit.ChangeID
}

func (m Model) ShowAll() bool { return m.showAll }

func (m Model) StatusBinds() []key.Help {
	return m.keymap.StatusBinds(m.showAll, m.filter.Filtering, m.filter.Value() != "")
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case LogDataMsg:
		m.entries = msg.Entries
		m.err = msg.Err
		m.applyFilter()
		if m.cursor >= m.visibleCount() && m.visibleCount() > 0 {
			m.cursor = m.visibleCount() - 1
		}
		return m, m.notifyCursorChanged()

	case RepoChangedMsg:
		return m, tea.Batch(m.fetchEntries(), m.awaitRepoChange())

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	if m.filter.Filtering {
		cmd := m.filter.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if m.filter.Filtering {
		return m.handleFilterKey(msg)
	}
	return m.handleNormal(msg)
}

func (m Model) handleFilterKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	cmd, handled := m.filter.HandleKey(msg, m.keymap.Close)
	if !handled {
		return m, nil
	}
	m.applyFilter()
	return m, tea.Batch(cmd, m.notifyCursorChanged())
}

func (m Model) handleNormal(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Close):
		if m.filter.Value() != "" {
			m.filter.ClearFilter()
			m.applyFilter()
			return m, m.notifyCursorChanged()
		}
		return m, nil

	case key.Matches(msg, m.keymap.Filter):
		cmd := m.filter.StartFilter()
		m.applyFilter()
		return m, cmd

	case key.Matches(msg, m.keymap.ToggleRevisions):
		m.showAll = !m.showAll
		m.cursor = 0
		m.offset = 0
		return m, m.fetchEntries()

	case key.Matches(msg, m.keymap.Open):
		id := m.SelectedChangeID()
		if id != "" {
			return m, func() tea.Msg {
				return LogSelectMsg{ChangeID: id}
			}
		}
		return m, nil
	}

	maxCursor := m.visibleCount() - 1
	if newCursor, ok := m.keymap.HandleScroll(msg, m.cursor, maxCursor, m.viewportHeight()/4); ok {
		m.cursor = newCursor
		m.ensureVisible()
		return m, m.notifyCursorChanged()
	}
	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return common.ConflictStyle.Render("Error: " + m.err.Error())
	}
	if m.hasFilter() && len(m.filtered) == 0 {
		return m.emptyView()
	}
	if m.visibleCount() == 0 {
		return m.emptyView()
	}
	return m.renderEntries()
}

func (m Model) emptyView() string {
	var msg string
	if m.filter.Value() != "" {
		msg = "No matches"
	} else {
		msg = "No commits found"
	}
	empty := common.TextMuted.Render(msg)
	if m.showFilterBar() {
		return empty + "\n" + m.filter.Input.View()
	}
	return empty
}

func (m Model) hasFilter() bool {
	return m.filter.Value() != ""
}

func (m Model) renderEntries() string {
	available := m.viewportHeight()
	if available <= 0 {
		return ""
	}
	if m.showFilterBar() {
		available--
	}

	filtering := m.hasFilter()
	entries := m.visibleEntries()
	var b strings.Builder
	linesUsed := 0

	for i := m.offset; i < len(entries) && linesUsed < available; i++ {
		isCurrent := i == m.cursor

		var lines []string
		if filtering && i < len(m.filtered) {
			lines = m.displayLines(m.filtered[i].entry)
			lines = common.HighlightMatches(lines, m.filtered[i].matchedIndexes)
		} else {
			lines = entries[i].Lines
		}

		for _, line := range lines {
			if linesUsed >= available {
				break
			}

			displayLine := ansi.Truncate(line, m.width, "")
			if isCurrent {
				displayLine = common.HighlightRow(displayLine, m.width)
			}

			if linesUsed > 0 {
				b.WriteString("\n")
			}
			b.WriteString(displayLine)
			linesUsed++
		}
	}

	if m.showFilterBar() {
		b.WriteString("\n")
		b.WriteString(m.filter.Input.View())
	}

	return b.String()
}

func (m Model) visibleEntries() []jj.GraphEntry {
	filter := m.filter.Value()
	if filter != "" && len(m.filtered) > 0 {
		entries := make([]jj.GraphEntry, len(m.filtered))
		for i, f := range m.filtered {
			entries[i] = f.entry
		}
		return entries
	}
	if filter != "" && len(m.filtered) == 0 {
		return nil
	}
	return m.entries
}

func (m Model) visibleCount() int {
	if m.filter.Value() != "" {
		return len(m.filtered)
	}
	return len(m.entries)
}

func (m Model) showFilterBar() bool {
	return m.filter.Active()
}

func (m Model) viewportHeight() int {
	return common.ViewportHeight(m.height)
}

func (m *Model) ensureVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}

	entries := m.visibleEntries()
	linesNeeded := 0
	for i := m.offset; i <= m.cursor && i < len(entries); i++ {
		linesNeeded += len(m.displayLines(entries[i]))
	}

	available := m.viewportHeight()
	if m.showFilterBar() {
		available--
	}
	for linesNeeded > available && m.offset < m.cursor {
		linesNeeded -= len(m.displayLines(entries[m.offset]))
		m.offset++
	}
}

// displayLines returns no-graph lines when filtering, graph lines otherwise.
func (m Model) displayLines(e jj.GraphEntry) []string {
	if m.hasFilter() && len(e.NoGraphLines) > 0 {
		return e.NoGraphLines
	}
	return e.Lines
}

func (m *Model) notifyCursorChanged() tea.Cmd {
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
	revset := ""
	if m.showAll {
		revset = allRevset
	}
	return func() tea.Msg {
		entries, err := runner.LogGraphEntries(revset)
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
