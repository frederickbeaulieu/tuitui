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
	return m.fetchContent(!changed)
}

func (m *Model) Refresh() tea.Cmd {
	if m.changeID == "" {
		return nil
	}
	m.loading = true
	m.err = nil
	return m.fetchContent(m.plainFile)
}

func (m *Model) fetchContent(plain bool) tea.Cmd {
	runner := m.runner
	changeID := m.changeID
	filePath := m.filePath
	width := m.width
	sideBySide := m.layout == sideBySide
	fullFile := m.showFullFile
	return func() tea.Msg {
		var content string
		var err error
		if plain {
			content, err = fetchPlain(runner, changeID, filePath)
		} else {
			content, err = fetchDiff(runner, changeID, filePath, width, sideBySide, fullFile)
		}
		if err != nil {
			return DiffContentMsg{ChangeID: changeID, FilePath: filePath, Err: err}
		}
		var lines []string
		if content != "" {
			lines = strings.Split(content, "\n")
		}
		return DiffContentMsg{ChangeID: changeID, FilePath: filePath, Lines: lines, PlainFile: plain}
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

	return m.renderLines()
}

func (m Model) renderLines() string {
	visible := common.ViewportHeight(m.height)

	end := min(m.offset+visible, len(m.lines))
	visibleLines := m.lines[m.offset:end]

	var b strings.Builder
	for i, line := range visibleLines {
		if i > 0 {
			b.WriteString("\n")
		}
		if m.plainFile {
			b.WriteString(lineGutter(m.offset+i+1, len(m.lines)))
			b.WriteString(line)
		} else {
			b.WriteString(ansi.Truncate(line, m.width, ""))
		}
	}

	return b.String()
}

func lineGutter(lineNum, totalLines int) string {
	w := max(3, len(fmt.Sprintf("%d", totalLines)))
	numStyle := ansi.NewStyle().ForegroundColor(ansi.BrightBlack).String()
	return fmt.Sprintf("%s%*d%s  ", numStyle, w, lineNum, ansi.ResetStyle)
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

	style := plainStyle()
	formatter := formatters.Get("terminal16")

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

// Tokyo Night style adapted for terminal output, with some adjustments for better visibility in a terminal.
func plainStyle() *chroma.Style {
	style, _ := chroma.NewStyle("plain", chroma.StyleEntries{
		chroma.Keyword:            "#800080", // magenta
		chroma.KeywordConstant:    "#800000", // red — nil, true, false, iota
		chroma.KeywordDeclaration: "#800080",
		chroma.KeywordType:        "#008080", // cyan — types
		chroma.KeywordNamespace:   "#000080", // blue — package/import

		chroma.Operator:     "bold #008000", // green bold
		chroma.OperatorWord: "bold #008000",

		chroma.NameFunction: "#000080", // blue
		chroma.NameClass:    "#bbaa00", // bright yellow
		chroma.NameBuiltin:  "#008000", // green

		chroma.LiteralString:         "#008000", // green
		chroma.LiteralStringEscape:   "#000080", // blue
		chroma.LiteralStringInterpol: "#bbaa00", // bright yellow
		chroma.LiteralNumber:         "#bbaa00", // bright yellow

		chroma.Comment:          "italic #666666", // bright black italic
		chroma.CommentSingle:    "italic #666666",
		chroma.CommentMultiline: "italic #666666",
		chroma.CommentSpecial:   "italic #666666",

		chroma.GenericDeleted:  "#800000",
		chroma.GenericInserted: "#008000",
		chroma.GenericEmph:     "italic",
		chroma.GenericStrong:   "bold",
	})
	return style
}

// formatDiff formats diff output through delta for syntax highlighting.
func formatDiff(input string, width int, sideBySide bool, fullFile bool) (string, error) {
	args := []string{
		"--no-gitconfig",
		"--paging=never",
		"--syntax-theme=ansi",
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
