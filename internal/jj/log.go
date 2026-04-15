package jj

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

// Tab-separated jj template. Uses explicit tabs instead of separate()
// because separate() skips empty values.
const logTemplate = `change_id ++ "\t" ++ commit_id ++ "\t" ++ if(description, description.first_line(), "") ++ "\t" ++ author.email() ++ "\t" ++ author.timestamp() ++ "\t" ++ bookmarks ++ "\t" ++ tags ++ "\t" ++ if(empty, "true", "false") ++ "\t" ++ if(conflict, "true", "false") ++ "\t" ++ parents.map(|p| p.commit_id()).join(",") ++ "\n"`

// Field indices matching the logTemplate order.
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
	fieldCount              // total
)

// jjTimestampLayout is the format produced by jj's author.timestamp().
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
	noGraphArgs := []string{"log", "--no-graph"}
	if revset != "" {
		structuredArgs = append(structuredArgs, "-r", revset)
		graphArgs = append(graphArgs, "-r", revset)
		noGraphArgs = append(noGraphArgs, "-r", revset)
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

	noGraphOutput, err := r.RunWithColor(noGraphArgs...)
	if err != nil {
		return nil, err
	}
	noGraphBlocks := splitNoGraphEntries(noGraphOutput, commits)

	entries := make([]GraphEntry, 0, len(blocks))
	for i, block := range blocks {
		entry := GraphEntry{Lines: block}
		if i < len(commits) {
			entry.Commit = commits[i]
		}
		if i < len(noGraphBlocks) {
			entry.NoGraphLines = noGraphBlocks[i]
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// splitByRevision splits output into per-revision blocks of lines.
func splitByRevision(output string, isNewRevision func(line string) bool) [][]string {
	lines := strings.Split(output, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	var blocks [][]string
	var current []string
	for _, line := range lines {
		if isNewRevision(line) && current != nil {
			blocks = append(blocks, current)
			current = nil
		}
		current = append(current, line)
	}
	if current != nil {
		blocks = append(blocks, current)
	}
	return blocks
}

func splitGraphEntries(output string) [][]string {
	return splitByRevision(output, isRevisionStart)
}

// splitNoGraphEntries splits --no-graph output into per-revision blocks,
// matching displayed short changeID prefixes against full IDs.
func splitNoGraphEntries(output string, commits []Commit) [][]string {
	if len(commits) == 0 {
		return nil
	}

	fullIDs := make([]string, len(commits))
	for i, c := range commits {
		fullIDs[i] = c.ChangeID
	}

	return splitByRevision(output, func(line string) bool {
		plain := ansi.Strip(line)
		firstWord := plain
		if sp := strings.IndexByte(plain, ' '); sp != -1 {
			firstWord = plain[:sp]
		}
		return matchesChangeID(firstWord, fullIDs)
	})
}

func matchesChangeID(prefix string, fullIDs []string) bool {
	if prefix == "" {
		return false
	}
	for _, id := range fullIDs {
		if strings.HasPrefix(id, prefix) {
			return true
		}
	}
	return false
}

func isRevisionStart(line string) bool {
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
		// Strip graph prefix by finding the changeID before the first tab.
		prefix := line[:tab]
		lastSpace := strings.LastIndexByte(prefix, ' ')
		changeID := prefix[lastSpace+1:]

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
