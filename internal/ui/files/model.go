// Package files implements the files panel for listing changed files.
package files

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

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
	err           error
	keymap        KeyMap
}

func New(runner *jj.Runner) Model {
	return Model{
		runner: runner,
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
	return m.keymap.StatusBinds(m.showAll)
}

func (m *Model) SetRevision(changeID string) tea.Cmd {
	m.changeID = changeID
	m.cursor = 0
	m.offset = 0
	m.loading = true
	m.err = nil

	runner := m.runner
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

func (m *Model) Refresh() tea.Cmd {
	if m.changeID == "" {
		return nil
	}
	m.loading = true
	m.err = nil

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

func (m Model) SelectedFile() *jj.FileChange {
	if len(m.files) == 0 || m.cursor >= len(m.files) {
		return nil
	}
	return &m.files[m.cursor]
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FilesDataMsg:
		if msg.ChangeID == m.changeID {
			m.loading = false
			m.err = msg.Err
			m.files = msg.Files
			m.conflictFiles = msg.ConflictFiles
			if m.cursor >= len(m.files) && len(m.files) > 0 {
				m.cursor = len(m.files) - 1
			}
		}
		return m, nil

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

	maxCursor := len(m.files) - 1
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
	if len(m.files) == 0 {
		if m.showAll {
			return common.TextMuted.Render("No files")
		}
		return common.TextMuted.Render("No changes")
	}

	visible := m.height
	if visible <= 0 {
		visible = 40
	}

	end := min(m.offset+visible, len(m.files))

	var b strings.Builder
	for i := m.offset; i < end; i++ {
		if i > m.offset {
			b.WriteString("\n")
		}

		f := m.files[i]
		line := fmt.Sprintf("%s %s", common.FileStatusSymbol(f.Status), f.Path)
		if m.conflictFiles[f.Path] {
			line += " " + common.ConflictStyle.Render("(conflict)")
		}
		line = ansi.Truncate(line, m.width, "")

		if i == m.cursor {
			line = common.HighlightLine(line, m.width)
		}

		b.WriteString(line)
	}

	return b.String()
}

func (m *Model) ensureVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	visible := m.height
	if visible <= 0 {
		visible = 40
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}
