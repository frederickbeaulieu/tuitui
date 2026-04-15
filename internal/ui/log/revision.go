package log

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/sahilm/fuzzy"

	"github.com/frederickbeaulieu/tuitui/internal/jj"
)

type filteredEntry struct {
	entry          jj.GraphEntry
	matchedIndexes []int
}

func (m *Model) applyFilter() {
	filter := m.filter.Value()
	if filter == "" {
		m.filtered = nil
		m.cursor = 0
		m.offset = 0
		return
	}

	var results []filteredEntry
	for _, e := range m.entries {
		if fe, ok := fuzzyMatchEntry(filter, e); ok {
			results = append(results, fe)
		}
	}

	m.filtered = results
	m.cursor = 0
	m.offset = 0
}

// fuzzyMatchEntry matches the filter against each commit field independently
// and maps highlight positions to the display text.
func fuzzyMatchEntry(filter string, e jj.GraphEntry) (filteredEntry, bool) {
	plainText := displayText(e)
	if plainText == "" {
		return filteredEntry{}, false
	}

	fields := searchableFields(e.Commit)
	matches := fuzzy.Find(filter, fields)
	if len(matches) == 0 {
		return filteredEntry{}, false
	}

	best := matches[0]
	bestField := fields[best.Index]

	// Find the field's position in the display text. For IDs, the display
	// shows a short prefix so we fall back to prefix matching.
	offset := strings.Index(plainText, bestField)
	displayLen := utf8.RuneCountInString(bestField)
	if offset < 0 {
		for _, word := range strings.Fields(plainText) {
			if strings.HasPrefix(bestField, word) {
				offset = strings.Index(plainText, word)
				displayLen = utf8.RuneCountInString(word)
				break
			}
		}
	}
	if offset < 0 {
		return filteredEntry{entry: e}, true
	}

	// Keep only positions visible in the displayed prefix.
	var mapped []int
	for _, idx := range best.MatchedIndexes {
		if idx < displayLen {
			mapped = append(mapped, idx+offset)
		}
	}

	return filteredEntry{
		entry:          e,
		matchedIndexes: mapped,
	}, true
}

func searchableFields(c jj.Commit) []string {
	fields := []string{
		c.ChangeID,
		c.CommitID,
		c.Description,
		c.Author,
		c.Timestamp.Format("2006-01-02 15:04:05"),
	}
	for _, b := range c.Bookmarks {
		fields = append(fields, b)
	}
	for _, t := range c.Tags {
		fields = append(fields, t)
	}
	return fields
}

// displayText returns the ANSI-stripped text of an entry's no-graph lines
// joined with spaces.
func displayText(e jj.GraphEntry) string {
	if len(e.NoGraphLines) == 0 {
		return ""
	}
	stripped := make([]string, len(e.NoGraphLines))
	for i, line := range e.NoGraphLines {
		stripped[i] = ansi.Strip(line)
	}
	return strings.Join(stripped, " ")
}
