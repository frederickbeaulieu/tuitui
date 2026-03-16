package files

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/frederickbeaulieu/tuitui/internal/jj"
	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

type FilesDataMsg struct {
	ChangeID string
	Files    []jj.FileChange
	Err      error
}

type FileSelectedMsg struct {
	ChangeID string
	Path     string
}

type FilesCloseMsg struct{}

type Model struct {
	runner   *jj.Runner
	changeID string
	files    []jj.FileChange
	cursor   int
	offset   int // scroll offset in file entries
	width    int
	height   int
	focused  bool
	loading  bool
	err      error
	keymap   common.KeyMap
}

func New(runner *jj.Runner) Model {
	return Model{
		runner: runner,
		keymap: common.DefaultKeyMap(),
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) Focus() { m.focused = true }

func (m *Model) Blur() { m.focused = false }

func (m Model) Focused() bool { return m.focused }

func (m *Model) SetRevision(changeID string) tea.Cmd {
	m.changeID = changeID
	m.cursor = 0
	m.offset = 0
	m.loading = true
	m.err = nil

	runner := m.runner
	return func() tea.Msg {
		files, err := runner.ChangedFiles(changeID)
		return FilesDataMsg{ChangeID: changeID, Files: files, Err: err}
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
	return func() tea.Msg {
		files, err := runner.ChangedFiles(changeID)
		return FilesDataMsg{ChangeID: changeID, Files: files, Err: err}
	}
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
	case key.Matches(msg, m.keymap.Up):
		if m.cursor > 0 {
			m.cursor--
			m.ensureVisible()
		}
		return m, nil

	case key.Matches(msg, m.keymap.Down):
		if m.cursor < len(m.files)-1 {
			m.cursor++
			m.ensureVisible()
		}
		return m, nil

	case key.Matches(msg, m.keymap.HalfUp):
		m.cursor -= m.height / 2
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.ensureVisible()
		return m, nil

	case key.Matches(msg, m.keymap.HalfDown):
		m.cursor += m.height / 2
		if m.cursor >= len(m.files) {
			m.cursor = len(m.files) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.ensureVisible()
		return m, nil

	case key.Matches(msg, m.keymap.Top):
		m.cursor = 0
		m.offset = 0
		return m, nil

	case key.Matches(msg, m.keymap.Bottom):
		if len(m.files) > 0 {
			m.cursor = len(m.files) - 1
			m.ensureVisible()
		}
		return m, nil

	case key.Matches(msg, m.keymap.Open):
		if f := m.SelectedFile(); f != nil {
			return m, func() tea.Msg {
				return FileSelectedMsg{ChangeID: m.changeID, Path: f.Path}
			}
		}
		return m, nil

	case key.Matches(msg, m.keymap.Back):
		return m, func() tea.Msg {
			return FilesCloseMsg{}
		}
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
		line = common.Truncate(line, m.width)

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
