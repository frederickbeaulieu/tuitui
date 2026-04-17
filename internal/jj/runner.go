// Package jj wraps the jj CLI for log, diff, and repository operations.
package jj

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

type Completion struct {
	Value       string
	Description string
}

// Runner executes jj CLI commands.
type Runner struct {
	RepoPath string
}

func NewRunner(repoPath string) *Runner {
	return &Runner{RepoPath: repoPath}
}

func (r *Runner) Run(args ...string) (string, error) {
	out, err := r.run(args...)
	if err != nil {
		return "", err
	}
	return ansi.Strip(out), nil
}

func (r *Runner) RunWithColor(args ...string) (string, error) {
	return r.run(args...)
}

// InteractiveCmd builds an *exec.Cmd suitable for interactive use
// (e.g. `jj commit`, `jj describe`, `jj split`, `jj diffedit`).
//
// Unlike run(), this does NOT wire stdio to buffers and does NOT force
// `--color always`; the caller (typically tea.ExecProcess) is expected
// to attach the real terminal so jj can spawn $EDITOR or its built-in
// diff editor.
func (r *Runner) InteractiveCmd(args ...string) *exec.Cmd {
	baseArgs := []string{"--no-pager"}
	if r.RepoPath != "" {
		baseArgs = append(baseArgs, "-R", r.RepoPath)
	}
	baseArgs = append(baseArgs, args...)
	return exec.Command("jj", baseArgs...)
}

// Complete queries jj's shell completion engine for suggestions.
func (r *Runner) Complete(words []string, index int) ([]Completion, error) {
	args := []string{"--", "jj"}
	if r.RepoPath != "" {
		args = append(args, "-R", r.RepoPath)
		index += 2
	}
	args = append(args, words...)

	completeIndex := fmt.Sprintf("%d", index+1)

	cmd := exec.Command("jj", args...)
	cmd.Env = append(cmd.Environ(),
		"COMPLETE=zsh",
		"_CLAP_IFS=\n",
		"_CLAP_COMPLETE_INDEX="+completeIndex,
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return parseCompletions(stdout.String()), nil
}

func parseCompletions(output string) []Completion {
	output = strings.TrimRight(output, "\n")
	if output == "" {
		return nil
	}
	lines := strings.Split(output, "\n")
	completions := make([]Completion, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		value, desc, _ := strings.Cut(line, ":")
		completions = append(completions, Completion{
			Value:       value,
			Description: desc,
		})
	}
	return completions
}

func (r *Runner) run(args ...string) (string, error) {
	baseArgs := []string{"--no-pager", "--color", "always"}
	if r.RepoPath != "" {
		baseArgs = append(baseArgs, "-R", r.RepoPath)
	}
	baseArgs = append(baseArgs, args...)

	cmd := exec.Command("jj", baseArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("jj %s: %w\n%s", strings.Join(args, " "), err, stderr.String())
	}

	out := stdout.String()
	if out == "" {
		out = stderr.String()
	}
	return out, nil
}
