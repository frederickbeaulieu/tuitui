package jj

import (
	"strings"
	"time"
	"unicode/utf8"
)

// logTemplate is the jj template for structured log output.
// Fields are tab-separated; explicit tab concatenation is used
// instead of separate() because separate() skips empty values.
const logTemplate = `change_id ++ "\t" ++ commit_id ++ "\t" ++ if(description, description.first_line(), "") ++ "\t" ++ author.email() ++ "\t" ++ author.timestamp() ++ "\t" ++ bookmarks ++ "\t" ++ tags ++ "\t" ++ if(empty, "true", "false") ++ "\t" ++ if(conflict, "true", "false") ++ "\t" ++ parents.map(|p| p.commit_id()).join(",") ++ "\n"`

// Field indices for the tab-separated logTemplate output.
// Order must match the template above.
const (
	fieldChangeID    = iota // 0
	fieldCommitID           // 1
	fieldDescription        // 2
	fieldAuthor             // 3
	fieldTimestamp          // 4
	fieldBookmarks          // 5
	fieldTags               // 6
	fieldEmpty              // 7
	fieldConflict           // 8
	fieldParents            // 9
	fieldCount              // total number of fields
)

// jjTimestampLayout is the time format produced by jj's author.timestamp().
const jjTimestampLayout = "2006-01-02 15:04:05.000 -07:00"

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
	structuredArgs := []string{"log", "-T", logTemplate}
	graphArgs := []string{"log"}
	if revset != "" {
		structuredArgs = append(structuredArgs, "-r", revset)
		graphArgs = append(graphArgs, "-r", revset)
	}

	structuredOutput, err := r.Run(structuredArgs...)
	if err != nil {
		return nil, err
	}
	commits := parseGraphLogOutput(structuredOutput)

	graphOutput, err := r.RunWithColor(graphArgs...)
	if err != nil {
		return nil, err
	}
	blocks := splitGraphEntries(graphOutput)

	entries := make([]GraphEntry, 0, len(blocks))
	for i, block := range blocks {
		entry := GraphEntry{Lines: block}
		if i < len(commits) {
			entry.Commit = commits[i]
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

func parseGraphLogOutput(output string) []Commit {
	var commits []Commit
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		tab := strings.IndexByte(line, '\t')
		if tab == -1 {
			continue
		}
		// The graph prefix (e.g. "│ ◆  ") is before the first field.
		// Strip it by finding the change ID right before the first tab.
		// The change ID is the last space-delimited word before the tab.
		prefix := line[:tab]
		lastSpace := strings.LastIndexByte(prefix, ' ')
		changeID := prefix[lastSpace+1:]

		// Build a full field slice: [changeID, commitID, desc, ...]
		fields := append([]string{changeID}, strings.Split(line[tab+1:], "\t")...)
		if commit, ok := commitFromFields(fields); ok {
			commits = append(commits, commit)
		}
	}
	return commits
}

func parseLogOutput(output string) ([]Commit, error) {
	var commits []Commit

	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if commit, ok := commitFromFields(fields); ok {
			commits = append(commits, commit)
		}
	}

	return commits, nil
}

// commitFromFields parses a Commit from a tab-separated field slice
// matching the logTemplate field order. Returns false if fields are
// insufficient.
func commitFromFields(fields []string) (Commit, bool) {
	if len(fields) < fieldCount {
		return Commit{}, false
	}

	commit := Commit{
		ChangeID:    fields[fieldChangeID],
		CommitID:    fields[fieldCommitID],
		Description: fields[fieldDescription],
		Author:      fields[fieldAuthor],
		IsEmpty:     fields[fieldEmpty] == "true",
		IsConflict:  fields[fieldConflict] == "true",
	}

	if ts, err := time.Parse(jjTimestampLayout, fields[fieldTimestamp]); err == nil {
		commit.Timestamp = ts
	}

	if fields[fieldBookmarks] != "" {
		commit.Bookmarks = strings.Split(fields[fieldBookmarks], " ")
	}

	if fields[fieldTags] != "" {
		commit.Tags = strings.Split(fields[fieldTags], " ")
	}

	if fields[fieldParents] != "" {
		commit.Parents = strings.Split(fields[fieldParents], ",")
	}

	return commit, true
}
