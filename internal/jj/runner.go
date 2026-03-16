package jj

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Runner executes jj CLI commands.
type Runner struct {
	RepoPath string
}

func NewRunner(repoPath string) *Runner {
	return &Runner{RepoPath: repoPath}
}

func (r *Runner) Run(args ...string) (string, error) {
	return r.run("never", args...)
}

func (r *Runner) RunWithColor(args ...string) (string, error) {
	return r.run("always", args...)
}

func (r *Runner) run(color string, args ...string) (string, error) {
	baseArgs := []string{"--no-pager", "--color", color}
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

	return stdout.String(), nil
}
