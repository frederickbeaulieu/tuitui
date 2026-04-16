// Package files implements the files panel for listing changed files.
package files

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/sahilm/fuzzy"

	"github.com/frederickbeaulieu/tuitui/internal/jj"
	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

type FilesDataMsg struct {
	ChangeID      string
	Files         []jj.FileChange
	ConflictFiles map[string]bool
	Err           error
}

type FileSelectedMsg struct {
	ChangeID string
	Path     string
	Changed  bool
}

type FilesCloseMsg struct{}

type filteredFile struct {
	file           jj.FileChange
	matchedIndexes []int
}

type Model struct {
	runner        *jj.Runner
	changeID      string
	files         []jj.FileChange
	conflictFiles map[string]bool
	cursor        int
	offset        int // scroll offset in file entries
	width         int
	height        int
	focused       bool
	loading       bool
	showAll       bool
	filter        common.FilterBar
	filtered      []filteredFile
	err           error
	keymap        KeyMap
}

func New(runner *jj.Runner) Model {
	return Model{
		runner: runner,
		filter: common.NewFilterBar(),
		keymap: DefaultKeyMap(),
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) Focus() { m.focused = true }

func (m *Model) Blur() { m.focused = false }

func (m Model) Focused() bool { return m.focused }

func (m Model) StatusBinds() []key.Help {
	return m.keymap.StatusBinds(m.showAll, m.filter.Filtering, m.filter.Value() != "")
}

func (m Model) SelectedFile() *jj.FileChange {
	files := m.visibleFiles()
	if len(files) == 0 || m.cursor >= len(files) {
		return nil
	}
	return &files[m.cursor]
}

func (m *Model) SetRevision(changeID string) tea.Cmd {
	m.changeID = changeID
	m.cursor = 0
	m.offset = 0
	m.loading = true
	m.err = nil
	m.filter.Reset()
	m.filtered = nil
	return m.fetchFiles()
}

func (m *Model) Refresh() tea.Cmd {
	if m.changeID == "" {
		return nil
	}
	m.loading = true
	m.err = nil
	return m.fetchFiles()
}

func (m *Model) fetchFiles() tea.Cmd {
	runner := m.runner
	changeID := m.changeID
	showAll := m.showAll
	return func() tea.Msg {
		var files []jj.FileChange
		var err error
		if showAll {
			files, err = runner.AllFiles(changeID)
			if err == nil {
				files = mergeFileStatuses(runner, changeID, files)
			}
		} else {
			files, err = runner.ChangedFiles(changeID)
		}
		conflicts := fetchConflictFiles(runner, changeID)
		return FilesDataMsg{ChangeID: changeID, Files: files, ConflictFiles: conflicts, Err: err}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FilesDataMsg:
		if msg.ChangeID == m.changeID {
			m.loading = false
			m.err = msg.Err
			m.files = msg.Files
			m.conflictFiles = msg.ConflictFiles
			m.applyFilter()
			if m.cursor >= m.visibleCount() && m.visibleCount() > 0 {
				m.cursor = m.visibleCount() - 1
			}
		}
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)

	case tea.MouseWheelMsg:
		if !m.focused {
			return m, nil
		}
		maxCursor := m.visibleCount() - 1
		if newCursor, ok := common.HandleMouseWheel(msg, m.cursor, maxCursor, 3); ok {
			m.cursor = newCursor
			m.ensureVisible()
		}
		return m, nil

	case tea.MouseClickMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleClick(msg)
	}

	if m.filter.Filtering {
		cmd := m.filter.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleClick(msg tea.MouseClickMsg) (Model, tea.Cmd) {
	if msg.Button != tea.MouseLeft {
		return m, nil
	}
	y := msg.Y - 1 // account for panel border
	if y < 0 {
		return m, nil
	}
	index := m.offset + y
	if index >= m.visibleCount() {
		return m, nil
	}

	wasAlreadySelected := index == m.cursor
	m.cursor = index
	m.ensureVisible()

	if wasAlreadySelected {
		if f := m.SelectedFile(); f != nil {
			changed := f.Status != " "
			return m, func() tea.Msg {
				return FileSelectedMsg{ChangeID: m.changeID, Path: f.Path, Changed: changed}
			}
		}
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
	return m, cmd
}

func (m Model) handleNormal(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Close):
		if m.filter.Value() != "" {
			m.filter.ClearFilter()
			m.applyFilter()
			return m, nil
		}
		return m, nil

	case key.Matches(msg, m.keymap.Filter):
		cmd := m.filter.StartFilter()
		m.applyFilter()
		return m, cmd

	case key.Matches(msg, m.keymap.ToggleAllFiles):
		m.showAll = !m.showAll
		m.cursor = 0
		m.offset = 0
		return m, m.Refresh()

	case key.Matches(msg, m.keymap.Open):
		if f := m.SelectedFile(); f != nil {
			changed := f.Status != " "
			return m, func() tea.Msg {
				return FileSelectedMsg{ChangeID: m.changeID, Path: f.Path, Changed: changed}
			}
		}
		return m, nil

	case key.Matches(msg, m.keymap.Back):
		return m, func() tea.Msg {
			return FilesCloseMsg{}
		}
	}

	maxCursor := m.visibleCount() - 1
	if newCursor, ok := m.keymap.HandleScroll(msg, m.cursor, maxCursor, m.height/2); ok {
		m.cursor = newCursor
		m.ensureVisible()
	}
	return m, nil
}

func (m Model) View() string {
	if m.changeID == "" {
		return common.TextMuted.Render("No revision selected")
	}
	if m.loading {
		return common.TextMuted.Render("Loading files...")
	}
	if m.err != nil {
		return common.ConflictStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	files := m.visibleFiles()
	if len(files) == 0 {
		return m.emptyView()
	}

	return m.renderLines(files)
}

func (m Model) emptyView() string {
	var msg string
	if m.filter.Value() != "" {
		msg = "No matches"
	} else if m.showAll {
		msg = "No files"
	} else {
		msg = "No changes"
	}
	empty := common.TextMuted.Render(msg)
	if m.showFilterBar() {
		return empty + "\n" + m.filter.Input.View()
	}
	return empty
}

func (m Model) renderLines(files []jj.FileChange) string {
	visible := common.ViewportHeight(m.height)
	if m.showFilterBar() {
		visible-- // reserve 1 line for the filter bar
	}

	end := min(m.offset+visible, len(files))
	visibleFiles := files[m.offset:end]

	var b strings.Builder
	for i, f := range visibleFiles {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(m.renderLine(m.offset+i, f))
	}

	if m.showFilterBar() {
		b.WriteString("\n")
		b.WriteString(m.filter.Input.View())
	}

	return b.String()
}

func (m Model) renderLine(index int, f jj.FileChange) string {
	matchedIndexes := m.matchedIndexesFor(index)

	var line string
	if len(matchedIndexes) > 0 {
		line = fmt.Sprintf("%s %s", common.FileStatusSymbol(f.Status), common.HighlightMatches([]string{f.Path}, matchedIndexes)[0])
	} else {
		line = fmt.Sprintf("%s %s", common.FileStatusSymbol(f.Status), f.Path)
	}
	if m.conflictFiles[f.Path] {
		line += " " + common.ConflictStyle.Render("(conflict)")
	}
	line = ansi.Truncate(line, m.width, "")

	if index == m.cursor {
		line = common.HighlightRow(line, m.width)
	}

	return line
}

func (m Model) showFilterBar() bool {
	return m.filter.Active()
}

func (m Model) visibleFiles() []jj.FileChange {
	filter := m.filter.Value()
	if filter != "" && len(m.filtered) > 0 {
		files := make([]jj.FileChange, len(m.filtered))
		for i, f := range m.filtered {
			files[i] = f.file
		}
		return files
	}
	if filter != "" && len(m.filtered) == 0 {
		return nil
	}
	return m.files
}

func (m Model) visibleCount() int {
	if m.filter.Value() != "" {
		return len(m.filtered)
	}
	return len(m.files)
}

func (m Model) matchedIndexesFor(i int) []int {
	if m.filter.Value() == "" || i >= len(m.filtered) {
		return nil
	}
	return m.filtered[i].matchedIndexes
}

func (m *Model) applyFilter() {
	filter := m.filter.Value()
	if filter == "" {
		m.filtered = nil
		m.cursor = 0
		m.offset = 0
		return
	}

	paths := make([]string, len(m.files))
	for i, f := range m.files {
		paths[i] = f.Path
	}

	matches := fuzzy.Find(filter, paths)

	m.filtered = make([]filteredFile, len(matches))
	for i, match := range matches {
		m.filtered[i] = filteredFile{
			file:           m.files[match.Index],
			matchedIndexes: match.MatchedIndexes,
		}
	}

	m.cursor = 0
	m.offset = 0
}

func (m *Model) ensureVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	visible := common.ViewportHeight(m.height)
	if m.showFilterBar() {
		visible-- // account for filter bar
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}

// mergeFileStatuses overlays change statuses (A/M/D/R) from ChangedFiles onto
// the full file list so that modified files keep their status indicator.
func mergeFileStatuses(runner *jj.Runner, changeID string, allFiles []jj.FileChange) []jj.FileChange {
	changed, err := runner.ChangedFiles(changeID)
	if err != nil || len(changed) == 0 {
		return allFiles
	}
	statusMap := make(map[string]string, len(changed))
	for _, c := range changed {
		statusMap[c.Path] = c.Status
	}
	for i, f := range allFiles {
		if s, ok := statusMap[f.Path]; ok {
			allFiles[i].Status = s
		}
	}
	return allFiles
}

// fetchConflictFiles returns a set of file paths that have conflicts in the given revision.
func fetchConflictFiles(runner *jj.Runner, changeID string) map[string]bool {
	paths, err := runner.ConflictFiles(changeID)
	if err != nil || len(paths) == 0 {
		return nil
	}
	m := make(map[string]bool, len(paths))
	for _, p := range paths {
		m[p] = true
	}
	return m
}
