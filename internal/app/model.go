// Package app implements the root Bubble Tea model and panel orchestration.
package app

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/frederickbeaulieu/tuitui/internal/jj"
	"github.com/frederickbeaulieu/tuitui/internal/ui/cmdbar"
	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
	"github.com/frederickbeaulieu/tuitui/internal/ui/diff"
	"github.com/frederickbeaulieu/tuitui/internal/ui/files"
	logpanel "github.com/frederickbeaulieu/tuitui/internal/ui/log"
)

const statusBarHeight = 1
const cmdbarHeight = 2

type mode int

const (
	modeLog mode = iota
	modeFiles
	modeDiff
)

type Model struct {
	log    logpanel.Model
	files  files.Model
	diff   diff.Model
	cmdbar cmdbar.Model
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
		cmdbar: cmdbar.New(runner),
		mode:   modeLog,
		keymap: common.DefaultKeyMap(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.log.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	cmd, handled := m.updateCmdbar(msg)
	cmds = appendCmd(cmds, cmd)
	if handled {
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleResize(msg)
	case tea.KeyPressMsg:
		if model, cmd, ok := m.handleKey(msg); ok {
			return model, cmd
		}
	case logpanel.LogSelectMsg:
		return m.handleLogSelect(msg)
	case files.FileSelectedMsg:
		return m.handleFileSelected(msg)
	case files.FilesCloseMsg:
		return m.handleFilesClose()
	case diff.DiffCloseMsg:
		return m.handleDiffClose()
	case logpanel.CursorChangedMsg:
		if m.mode == modeFiles && msg.ChangeID != "" {
			cmds = appendCmd(cmds, m.files.SetRevision(msg.ChangeID))
		}
	case logpanel.RepoChangedMsg:
		if m.mode == modeFiles || m.mode == modeDiff {
			cmds = appendCmd(cmds, m.files.Refresh())
		}
		if m.mode == modeDiff {
			cmds = appendCmd(cmds, m.diff.Refresh())
		}
	}

	cmds = append(cmds, m.updatePanels(msg)...)
	return m, tea.Batch(cmds...)
}

func (m *Model) updateCmdbar(msg tea.Msg) (tea.Cmd, bool) {
	switch msg.(type) {
	case cmdbar.CmdResultMsg, cmdbar.CmdCloseMsg, cmdbar.CompletionMsg:
		var cmd tea.Cmd
		m.cmdbar, cmd = m.cmdbar.Update(msg)
		m.layoutPanels()
		return cmd, false
	case tea.KeyPressMsg:
		if m.cmdbar.Active() || m.cmdbar.ShowingError() {
			var cmd tea.Cmd
			m.cmdbar, cmd = m.cmdbar.Update(msg)
			m.layoutPanels()
			return cmd, true
		}
	default:
		if m.cmdbar.Active() {
			var cmd tea.Cmd
			m.cmdbar, cmd = m.cmdbar.Update(msg)
			return cmd, false
		}
	}
	return nil, false
}

func (m *Model) handleResize(msg tea.WindowSizeMsg) (Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.cmdbar.SetSize(m.width, m.panelHeight()-2)
	m.layoutPanels()
	if m.mode == modeDiff {
		return *m, m.diff.Refresh()
	}
	return *m, nil
}

func (m *Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keymap.Quit):
		return *m, tea.Quit, true
	case key.Matches(msg, m.keymap.Command):
		cmd := m.cmdbar.Activate()
		m.layoutPanels()
		return *m, cmd, true
	}
	return *m, nil, false
}

func (m *Model) handleLogSelect(msg logpanel.LogSelectMsg) (Model, tea.Cmd) {
	m.setMode(modeFiles)
	return *m, m.files.SetRevision(msg.ChangeID)
}

func (m *Model) handleFileSelected(msg files.FileSelectedMsg) (Model, tea.Cmd) {
	m.setMode(modeDiff)
	return *m, m.diff.SetRevisionFile(msg.ChangeID, msg.Path)
}

func (m *Model) handleFilesClose() (Model, tea.Cmd) {
	m.setMode(modeLog)
	return *m, nil
}

func (m *Model) handleDiffClose() (Model, tea.Cmd) {
	m.setMode(modeFiles)
	return *m, nil
}

func (m *Model) setMode(newMode mode) {
	m.mode = newMode
	m.log.Blur()
	m.files.Blur()
	m.diff.Blur()
	switch newMode {
	case modeLog:
		m.log.Focus()
	case modeFiles:
		m.files.Focus()
	case modeDiff:
		m.diff.Focus()
	}
	m.layoutPanels()
}

func (m *Model) updatePanels(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.log, cmd = m.log.Update(msg)
	cmds = appendCmd(cmds, cmd)

	if m.mode == modeFiles || m.mode == modeDiff {
		m.files, cmd = m.files.Update(msg)
		cmds = appendCmd(cmds, cmd)
	}

	if m.mode == modeDiff {
		m.diff, cmd = m.diff.Update(msg)
		cmds = appendCmd(cmds, cmd)
	}

	return cmds
}

func (m Model) splitWidths() (int, int) {
	logWidth := max(m.width*3/5, 30)
	filesWidth := max(m.width-logWidth, 20)
	return logWidth, filesWidth
}

func (m Model) panelHeight() int {
	bottomHeight := statusBarHeight
	if m.cmdbar.Active() {
		bottomHeight = cmdbarHeight
	}
	availableHeight := m.height - bottomHeight
	return max(availableHeight-m.cmdbar.CompletionHeight(availableHeight), 3)
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

func appendCmd(cmds []tea.Cmd, cmd tea.Cmd) []tea.Cmd {
	if cmd != nil {
		return append(cmds, cmd)
	}
	return cmds
}
