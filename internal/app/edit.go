package app

import (
	"fmt"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/frederickbeaulieu/tuitui/internal/ui/cmdbar"
	"github.com/frederickbeaulieu/tuitui/internal/ui/files"
)

// promptTokenEdit identifies the "edit this revision?" prompt in
// cmdbar.PromptResultMsg. The cmdbar is agnostic to the prompt's
// semantics; the token is how the app correlates a result back to
// its originating request.
const promptTokenEdit = "edit-file"

// editImmutableCheckMsg is the result of an async `self.immutable()`
// probe for the revision holding the file we want to edit.
type editImmutableCheckMsg struct {
	changeID  string
	path      string
	immutable bool
	err       error
}

// editFinishedMsg is delivered by tea.ExecProcess when $EDITOR exits.
type editFinishedMsg struct {
	err error
}

// logEditCheckMsg is the result of the async immutable probe for a
// `jj edit <rev>` request coming from the log panel.
type logEditCheckMsg struct {
	changeID  string
	immutable bool
	err       error
}

// logEditDoneMsg is emitted after `jj edit <rev>` completes.
type logEditDoneMsg struct {
	err error
}

func (m *Model) handleFileEditRequest(msg files.FileEditRequestMsg) (Model, tea.Cmd) {
	// Deletions have nothing to edit. Refuse early.
	if msg.Status == "D" {
		return *m, cmdErr(fmt.Errorf("cannot edit %q: file was deleted in this revision", msg.Path))
	}

	runner := m.runner
	changeID := msg.ChangeID
	path := msg.Path
	return *m, func() tea.Msg {
		immut, err := runner.IsImmutable(changeID)
		return editImmutableCheckMsg{
			changeID:  changeID,
			path:      path,
			immutable: immut,
			err:       err,
		}
	}
}

func (m *Model) handleEditImmutableCheck(msg editImmutableCheckMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		return *m, cmdErr(fmt.Errorf("checking revision: %w", msg.err))
	}
	if msg.immutable {
		return *m, cmdErr(fmt.Errorf("cannot edit: revision %s is immutable", shortID(msg.changeID)))
	}

	// Need the current working-copy id to decide whether a `jj edit`
	// is required at all.
	current, err := m.runner.CurrentChangeID()
	if err != nil {
		return *m, cmdErr(fmt.Errorf("resolving @: %w", err))
	}

	if current == msg.changeID {
		// Already on this revision — just open the file.
		return *m, m.launchEditor(msg.path)
	}

	// Need confirmation to move the working copy.
	m.pendEdit = &pendingEdit{changeID: msg.changeID, path: msg.path}
	promptMsg := fmt.Sprintf("jj edit %s and open %s?", shortID(msg.changeID), msg.path)
	cmd := m.cmdbar.StartPrompt(promptTokenEdit, promptMsg)
	m.layoutPanels()
	return *m, cmd
}

// shortID returns the conventional jj short form (first 8 hex chars)
// of a change id for display. Falls back to the full id if it is
// already short.
func shortID(id string) string {
	const n = 8
	if len(id) <= n {
		return id
	}
	return id[:n]
}

func (m *Model) handlePromptResult(msg cmdbar.PromptResultMsg) (Model, tea.Cmd) {
	if msg.Token != promptTokenEdit {
		return *m, nil
	}
	pe := m.pendEdit
	m.pendEdit = nil
	m.layoutPanels()
	if pe == nil || !msg.Accepted {
		return *m, nil
	}

	runner := m.runner
	changeID := pe.changeID
	path := pe.path

	// Run `jj edit` then launch the editor. Chain so the editor only
	// launches if the edit succeeds.
	editRev := func() tea.Msg {
		if err := runner.Edit(changeID); err != nil {
			return cmdbar.CmdResultMsg{Err: fmt.Errorf("jj edit %s: %w", changeID, err)}
		}
		return editProceedMsg{path: path}
	}
	return *m, editRev
}

// editProceedMsg is an internal signal emitted after a successful
// `jj edit` to trigger the editor launch.
type editProceedMsg struct {
	path string
}

func (m *Model) handleEditProceed(msg editProceedMsg) (Model, tea.Cmd) {
	return *m, m.launchEditor(msg.path)
}

func (m *Model) handleEditFinished(msg editFinishedMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		return *m, cmdErr(fmt.Errorf("editor: %w", msg.err))
	}
	// jj auto-snapshots on its next invocation and the repo watcher
	// polls the op-log every second, so no explicit refresh is needed.
	return *m, nil
}

// launchEditor resolves path relative to the repo root and hands the
// terminal to $EDITOR via tea.ExecProcess.
func (m *Model) launchEditor(path string) tea.Cmd {
	root, err := m.runner.RepoRoot()
	if err != nil {
		return cmdErr(fmt.Errorf("resolving repo root: %w", err))
	}
	full := filepath.Join(root, path)
	cmd := m.editor.Command(full)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editFinishedMsg{err: err}
	})
}

// cmdErr surfaces an error through the cmdbar's error viewer.
func cmdErr(err error) tea.Cmd {
	return func() tea.Msg { return cmdbar.CmdResultMsg{Err: err} }
}

// handleLogEditRequest runs `jj edit <id>` to make the selected log
// revision the working copy, after an immutability check. No prompt:
// moving @ is the entire point of the action.
func (m *Model) handleLogEditRequest(changeID string) (Model, tea.Cmd) {
	if changeID == "" {
		return *m, nil
	}
	runner := m.runner
	return *m, func() tea.Msg {
		immut, err := runner.IsImmutable(changeID)
		return logEditCheckMsg{changeID: changeID, immutable: immut, err: err}
	}
}

func (m *Model) handleLogEditCheck(msg logEditCheckMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		return *m, cmdErr(fmt.Errorf("checking revision: %w", msg.err))
	}
	if msg.immutable {
		return *m, cmdErr(fmt.Errorf("cannot edit: revision %s is immutable", shortID(msg.changeID)))
	}

	// Skip the jj call when @ already points here.
	if current, err := m.runner.CurrentChangeID(); err == nil && current == msg.changeID {
		return *m, nil
	}

	runner := m.runner
	changeID := msg.changeID
	return *m, func() tea.Msg {
		return logEditDoneMsg{err: runner.Edit(changeID)}
	}
}

func (m *Model) handleLogEditDone(msg logEditDoneMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		return *m, cmdErr(fmt.Errorf("jj edit: %w", msg.err))
	}
	// Repo watcher will pick up the @ move on its next poll and the log
	// panel will re-render with the new working-copy marker.
	return *m, nil
}
