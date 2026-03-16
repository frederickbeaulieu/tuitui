package jj

import (
	"strings"
	"time"
	"unicode/utf8"
)

// logTemplate is the jj template for structured log output.
// Fields are tab-separated; explicit tab concatenation is used
// instead of separate() because separate() skips empty values.
const logTemplate = `change_id ++ "\t" ++ commit_id ++ "\t" ++ if(description, description.first_line(), "") ++ "\t" ++ author.email() ++ "\t" ++ author.timestamp() ++ "\t" ++ bookmarks ++ "\t" ++ if(empty, "true", "false") ++ "\t" ++ if(conflict, "true", "false") ++ "\t" ++ parents.map(|p| p.commit_id()).join(",") ++ "\n"`

func (r *Runner) Log(revset string) ([]Commit, error) {
	args := []string{"log", "--no-graph", "-T", logTemplate}
	if revset != "" {
		args = append(args, "-r", revset)
	}

	output, err := r.Run(args...)
	if err != nil {
		return nil, err
	}

	return parseLogOutput(output)
}

func (r *Runner) LogGraphEntries(revset string) ([]GraphEntry, error) {
	commits, err := r.Log(revset)
	if err != nil {
		return nil, err
	}

	args := []string{"log"}
	if revset != "" {
		args = append(args, "-r", revset)
	}
	graphOutput, err := r.RunWithColor(args...)
	if err != nil {
		return nil, err
	}

	blocks := splitGraphEntries(graphOutput)

	entries := make([]GraphEntry, 0, len(commits))
	for i, commit := range commits {
		entry := GraphEntry{Commit: commit}
		if i < len(blocks) {
			entry.Lines = blocks[i]
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func splitGraphEntries(output string) [][]string {
	lines := strings.Split(output, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	var blocks [][]string
	var current []string

	for _, line := range lines {
		if hasNodeGlyph(line) {
			if current != nil {
				blocks = append(blocks, current)
			}
			current = []string{line}
		} else {
			if current == nil {
				current = []string{line}
			} else {
				current = append(current, line)
			}
		}
	}
	if current != nil {
		blocks = append(blocks, current)
	}

	return blocks
}

func hasNodeGlyph(line string) bool {
	visCount := 0
	i := 0
	for i < len(line) && visCount < 10 {
		if line[i] == '\x1b' {
			i++
			for i < len(line) {
				b := line[i]
				i++
				if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') {
					break
				}
			}
			continue
		}

		r, size := utf8.DecodeRuneInString(line[i:])
		i += size
		visCount++

		if isNodeGlyph(r) {
			return true
		}
	}
	return false
}

func isNodeGlyph(r rune) bool {
	switch r {
	case '@', '○', '◆', '●', '×', '◉':
		return true
	}
	return false
}

func parseLogOutput(output string) ([]Commit, error) {
	var commits []Commit

	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 8 {
			continue
		}

		commit := Commit{
			ChangeID:    fields[0],
			CommitID:    fields[1],
			Description: fields[2],
			Author:      fields[3],
			IsEmpty:     fields[6] == "true",
			IsConflict:  fields[7] == "true",
		}

		if ts, err := time.Parse("2006-01-02 15:04:05.000 -07:00", fields[4]); err == nil {
			commit.Timestamp = ts
		}

		if fields[5] != "" {
			commit.Bookmarks = strings.Split(fields[5], " ")
		}

		if len(fields) > 8 && fields[8] != "" {
			commit.Parents = strings.Split(fields[8], ",")
		}

		commits = append(commits, commit)
	}

	return commits, nil
}
