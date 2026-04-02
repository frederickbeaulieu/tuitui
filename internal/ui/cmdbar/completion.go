package cmdbar

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/frederickbeaulieu/tuitui/internal/jj"
	"github.com/frederickbeaulieu/tuitui/internal/ui/common"
	"github.com/google/shlex"
)

const (
	maxVisibleCompletions = 10
	minPanelHeight        = 5
	debounceDelay         = 150 * time.Millisecond
)

type completionState struct {
	items        []jj.Completion
	selected     int
	visible      bool
	showDropdown bool
	tag          int
}

type CompletionMsg struct {
	Items []jj.Completion
	Tag   int
}

type completionTickMsg struct {
	Tag   int
	Input string
}

func requestCompletionTick(tag int, input string) tea.Cmd {
	return tea.Tick(debounceDelay, func(time.Time) tea.Msg {
		return completionTickMsg{Tag: tag, Input: input}
	})
}

func requestCompletion(runner *jj.Runner, input string, tag int) tea.Cmd {
	return func() tea.Msg {
		if cursorInQuote(input) {
			return CompletionMsg{Tag: tag}
		}
		words, index := parseInputForCompletion(input)
		items, err := runner.Complete(words, index)
		if err != nil {
			return CompletionMsg{Tag: tag}
		}
		return CompletionMsg{Items: items, Tag: tag}
	}
}

func cursorInQuote(input string) bool {
	inSingle := false
	inDouble := false
	for _, r := range input {
		switch {
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '"' && !inSingle:
			inDouble = !inDouble
		}
	}
	return inSingle || inDouble
}

func parseInputForCompletion(input string) (words []string, index int) {
	parts, err := shlex.Split(input)
	if err != nil || len(parts) == 0 {
		return []string{""}, 0
	}
	if strings.HasSuffix(input, " ") {
		parts = append(parts, "")
	}
	return parts, len(parts) - 1
}

func completionView(cs completionState, width int, maxItems int) string {
	if !cs.visible || !cs.showDropdown || len(cs.items) == 0 || maxItems <= 0 {
		return ""
	}

	items, offset := visibleItems(cs, maxItems)

	var lines []string
	for i, item := range items {
		line := formatCompletionLine(item, width)
		line = styleCompletionLine(line, i+offset, cs.selected, width)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func visibleItems(cs completionState, maxItems int) ([]jj.Completion, int) {
	items := cs.items
	visible := min(maxItems, len(items), maxVisibleCompletions)
	offset := 0

	if len(items) > visible {
		if cs.selected >= visible {
			offset = cs.selected - visible + 1
		}
		end := offset + visible
		if end > len(items) {
			end = len(items)
			offset = end - visible
		}
		items = items[offset:end]
	}

	return items, offset
}

func formatCompletionLine(item jj.Completion, width int) string {
	valueStyle := lipgloss.NewStyle().Foreground(common.ColorText)
	descStyle := lipgloss.NewStyle().Foreground(common.ColorOverlay)
	innerWidth := width - 2

	value := item.Value
	if len(value) > innerWidth/2 {
		value = value[:innerWidth/2]
	}

	if item.Description == "" {
		gap := max(innerWidth-len(value), 0)
		return " " + valueStyle.Render(value) + strings.Repeat(" ", gap) + " "
	}

	desc := item.Description
	maxDescLen := innerWidth - len(value) - 2
	if maxDescLen > 0 && len(desc) > maxDescLen {
		desc = desc[:maxDescLen]
	}
	gap := max(innerWidth-len(value)-len(desc), 1)
	return " " + valueStyle.Render(value) + strings.Repeat(" ", gap) + descStyle.Render(desc) + " "
}

func styleCompletionLine(line string, index, selected, width int) string {
	if index == selected {
		return common.HighlightLine(line, width)
	}
	return lipgloss.NewStyle().Background(common.ColorSurface).Width(width).Render(line)
}

func acceptCompletion(input string, completion string) string {
	parts, err := shlex.Split(input)
	if err != nil || len(parts) == 0 {
		return completion
	}

	if strings.HasSuffix(input, " ") {
		return input + completion + " "
	}

	parts[len(parts)-1] = completion
	var result []string
	for _, p := range parts {
		if strings.ContainsAny(p, " \t") {
			result = append(result, `"`+p+`"`)
		} else {
			result = append(result, p)
		}
	}
	return strings.Join(result, " ") + " "
}

func currentPartialWord(input string) string {
	if input == "" || strings.HasSuffix(input, " ") {
		return ""
	}
	parts, err := shlex.Split(input)
	if err != nil || len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func ghostText(input string, cs completionState) string {
	if !cs.visible || len(cs.items) == 0 || cursorInQuote(input) {
		return ""
	}
	idx := cs.selected
	if idx < 0 {
		idx = 0
	}
	item := cs.items[idx]
	partial := currentPartialWord(input)
	if strings.HasPrefix(item.Value, partial) {
		return item.Value[len(partial):]
	}
	return ""
}
