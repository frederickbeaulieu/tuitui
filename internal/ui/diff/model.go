// Package diff implements the diff panel for viewing file changes.
package diff

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/x/ansi"

	"github.com/frederickbeaulieu/tuitui/internal/jj"
	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
)

type DiffContentMsg struct {
	ChangeID  string
	FilePath  string
	Lines     []string
	PlainFile bool
	Err       error
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
	plainFile    bool   // true = showing plain file content (no diff available)
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

func (m Model) StatusBinds() []key.Help {
	return m.keymap.StatusBinds(m.showFullFile, m.plainFile)
}

func (m *Model) SetRevisionFile(changeID, path string, changed bool) tea.Cmd {
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
		var content string
		var err error
		plainFile := !changed
		if changed {
			content, err = fetchDiff(runner, changeID, path, width, sideBySide, fullFile)
		} else {
			content, err = fetchPlain(runner, changeID, path)
		}
		if err != nil {
			return DiffContentMsg{ChangeID: changeID, FilePath: path, Err: err}
		}
		var lines []string
		if content != "" {
			lines = strings.Split(content, "\n")
		}
		return DiffContentMsg{ChangeID: changeID, FilePath: path, Lines: lines, PlainFile: plainFile}
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
		var lines []string
		if content != "" {
			lines = strings.Split(content, "\n")
		}
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
			m.plainFile = msg.PlainFile
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
	switch {
	case key.Matches(msg, m.keymap.Back):
		return m, func() tea.Msg { return DiffCloseMsg{} }

	case key.Matches(msg, m.keymap.ToggleLayout):
		if m.plainFile {
			return m, nil
		}
		if m.layout == inline {
			m.layout = sideBySide
		} else {
			m.layout = inline
		}
		cmd := m.Refresh()
		m.loading = false // keep showing current content while refreshing
		return m, cmd

	case key.Matches(msg, m.keymap.ToggleContext):
		if m.plainFile {
			return m, nil
		}
		m.showFullFile = !m.showFullFile
		cmd := m.Refresh()
		m.loading = false // keep showing current content while refreshing
		return m, cmd
	}

	if newOffset, ok := m.keymap.HandleScroll(msg, m.offset, m.maxOffset(), m.height/2); ok {
		m.offset = newOffset
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
		b.WriteString(ansi.Truncate(m.lines[i], m.width, ""))
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

	return formatDiff(diffOutput, width, sideBySide, fullFile)
}

func fetchPlain(runner *jj.Runner, changeID, path string) (string, error) {
	content, err := runner.FileShow(changeID, path)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(content) == "" {
		return "", nil
	}
	return formatPlain(content, path)
}

// formatPlain formats plain file content with syntax highlighting using chroma.
func formatPlain(input string, path string) (string, error) {
	lexer := lexers.Match(filepath.Base(path))
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("tokyonight-night")
	formatter := formatters.Get("terminal16m")

	iterator, err := lexer.Tokenise(nil, input)
	if err != nil {
		return input, nil
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return input, nil
	}

	return buf.String(), nil
}

// formatDiff formats diff output through delta for syntax highlighting.
func formatDiff(input string, width int, sideBySide bool, fullFile bool) (string, error) {
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
