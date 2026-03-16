package diff

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/frederickbeaulieu/tuitui/internal/jj"
	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

type DiffContentMsg struct {
	ChangeID string
	FilePath string
	Lines    []string
	Err      error
}

type DiffCloseMsg struct{}

type layout int

const (
	inline     layout = iota // unified/inline diff (default)
	sideBySide               // side-by-side diff
)

// Model is the diff panel.
type Model struct {
	runner       *jj.Runner
	changeID     string // currently displayed revision
	filePath     string // currently displayed file
	layout       layout // inline or side-by-side
	showFullFile bool   // true = show entire file, false = show only changes
	lines        []string
	offset       int // scroll offset in lines
	width        int
	height       int
	focused      bool
	loading      bool
	err          error
	keymap       KeyMap
}

func New(runner *jj.Runner) Model {
	return Model{
		runner: runner,
		layout: sideBySide,
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

func (m Model) ShowFullFile() bool { return m.showFullFile }

func (m *Model) SetRevisionFile(changeID, path string) tea.Cmd {
	if changeID == m.changeID && path == m.filePath {
		return nil
	}
	m.changeID = changeID
	m.filePath = path
	m.offset = 0
	m.loading = true
	m.err = nil

	runner := m.runner
	width := m.width
	sideBySide := m.layout == sideBySide
	fullFile := m.showFullFile
	return func() tea.Msg {
		content, err := fetchDiff(runner, changeID, path, width, sideBySide, fullFile)
		if err != nil {
			return DiffContentMsg{ChangeID: changeID, FilePath: path, Err: err}
		}
		lines := strings.Split(content, "\n")
		return DiffContentMsg{ChangeID: changeID, FilePath: path, Lines: lines}
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
	filePath := m.filePath
	width := m.width
	sideBySide := m.layout == sideBySide
	fullFile := m.showFullFile
	return func() tea.Msg {
		content, err := fetchDiff(runner, changeID, filePath, width, sideBySide, fullFile)
		if err != nil {
			return DiffContentMsg{ChangeID: changeID, FilePath: filePath, Err: err}
		}
		lines := strings.Split(content, "\n")
		return DiffContentMsg{ChangeID: changeID, FilePath: filePath, Lines: lines}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DiffContentMsg:
		if msg.ChangeID == m.changeID && msg.FilePath == m.filePath {
			m.loading = false
			m.err = msg.Err
			m.lines = msg.Lines
			m.offset = 0
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
	maxOffset := m.maxOffset()

	switch {
	case key.Matches(msg, m.keymap.Back):
		return m, func() tea.Msg { return DiffCloseMsg{} }

	case key.Matches(msg, m.keymap.ToggleLayout):
		if m.layout == inline {
			m.layout = sideBySide
		} else {
			m.layout = inline
		}
		return m, m.Refresh()

	case key.Matches(msg, m.keymap.ToggleContext):
		m.showFullFile = !m.showFullFile
		return m, m.Refresh()

	case key.Matches(msg, m.keymap.Up):
		if m.offset > 0 {
			m.offset--
		}
	case key.Matches(msg, m.keymap.Down):
		if m.offset < maxOffset {
			m.offset++
		}
	case key.Matches(msg, m.keymap.HalfUp):
		m.offset -= m.height / 2
		if m.offset < 0 {
			m.offset = 0
		}
	case key.Matches(msg, m.keymap.HalfDown):
		m.offset += m.height / 2
		if m.offset > maxOffset {
			m.offset = maxOffset
		}
	case key.Matches(msg, m.keymap.Top):
		m.offset = 0
	case key.Matches(msg, m.keymap.Bottom):
		m.offset = maxOffset
	}

	return m, nil
}

func (m Model) View() string {
	if m.changeID == "" {
		return common.TextMuted.Render("No revision selected")
	}
	if m.loading {
		return common.TextMuted.Render("Loading diff...")
	}
	if m.err != nil {
		return common.ConflictStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}
	if len(m.lines) == 0 {
		return common.TextMuted.Render("No changes")
	}

	visible := m.height
	if visible <= 0 {
		visible = 40
	}

	end := min(m.offset+visible, len(m.lines))

	var b strings.Builder
	for i := m.offset; i < end; i++ {
		if i > m.offset {
			b.WriteString("\n")
		}
		b.WriteString(common.Truncate(m.lines[i], m.width))
	}

	return b.String()
}

func (m Model) maxOffset() int {
	return max(len(m.lines)-m.height, 0)
}

func fetchDiff(runner *jj.Runner, changeID, path string, width int, sideBySide bool, fullFile bool) (string, error) {
	var diffOutput string
	var err error

	if fullFile {
		diffOutput, err = runner.FileDiffFull(changeID, path)
	} else {
		diffOutput, err = runner.FileDiff(changeID, path)
	}
	if err != nil {
		return "", err
	}

	return formatWithDelta(diffOutput, width, sideBySide, fullFile)
}

// formatWithDelta formats diff output through delta for syntax highlighting.
func formatWithDelta(input string, width int, sideBySide bool, fullFile bool) (string, error) {
	args := []string{
		"--no-gitconfig",
		"--paging=never",
		"--syntax-theme=tokyonight_night",
		fmt.Sprintf("--width=%d", width),
	}
	if sideBySide {
		args = append(args, "--side-by-side")
	} else {
		args = append(args, "--color-only")
	}
	if fullFile {
		args = append(args, "--hunk-header-style=omit")
	}

	cmd := exec.Command("delta", args...)
	cmd.Stdin = strings.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("delta: %w: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
