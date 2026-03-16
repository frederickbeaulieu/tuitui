package jj

import (
	"strings"
)

// FileDiff returns the git-format diff for a single file in the given revision.
func (r *Runner) FileDiff(revision, path string) (string, error) {
	return r.Run("diff", "--revision", revision, "--git", path)
}

// FileDiffFull returns the diff with full file context (all unchanged lines included).
func (r *Runner) FileDiffFull(revision, path string) (string, error) {
	return r.Run("diff", "--revision", revision, "--git", "--context", "999999", path)
}

// ChangedFiles returns the list of files changed in the given revision.
func (r *Runner) ChangedFiles(revision string) ([]FileChange, error) {
	output, err := r.Run("diff", "--revision", revision, "--summary")
	if err != nil {
		return nil, err
	}
	return parseFileChanges(output), nil
}

func parseFileChanges(output string) []FileChange {
	var changes []FileChange
	for _, line := range nonEmptyLines(output) {
		fc := parseFileChange(line)
		if fc != nil {
			changes = append(changes, *fc)
		}
	}
	return changes
}

func nonEmptyLines(s string) []string {
	var lines []string
	for line := range strings.SplitSeq(s, "\n") {
		line = strings.TrimRight(line, "\r")
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func parseFileChange(line string) *FileChange {
	line = strings.TrimSpace(line)
	if len(line) < 2 {
		return nil
	}

	status := string(line[0])
	path := strings.TrimSpace(line[1:])

	if path == "" {
		return nil
	}

	return &FileChange{
		Status: status,
		Path:   path,
	}
}
