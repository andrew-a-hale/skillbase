package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/truncate"
	"github.com/muesli/reflow/wordwrap"
)

// filteredIndices returns the indices of skills whose Name, Path, or Description
// contains the given filter (case-insensitive). If filter is empty, all indices
// are returned in order.
func filteredIndices(skills []SkillInfo, filter string) []int {
	if filter == "" {
		indices := make([]int, len(skills))
		for i := range skills {
			indices[i] = i
		}
		return indices
	}
	f := strings.ToLower(filter)
	var indices []int
	for i, s := range skills {
		if strings.Contains(strings.ToLower(s.Name), f) ||
			strings.Contains(strings.ToLower(s.Path), f) ||
			strings.Contains(strings.ToLower(s.Description), f) {
			indices = append(indices, i)
		}
	}
	return indices
}

// list provides reusable cursor navigation.
// Embed it in a model and call its methods from Update.
type list struct {
	cursor int
}

func (l *list) up(_ int) {
	if l.cursor > 0 {
		l.cursor--
	}
}

func (l *list) down(count int) {
	if count > 0 && l.cursor < count-1 {
		l.cursor++
	}
}

func (l *list) reset() {
	l.cursor = 0
}

func (l *list) handleMouse(msg tea.MouseMsg, count int) {
	mouse := msg.Mouse()
	switch mouse.Button {
	case tea.MouseWheelDown:
		l.down(count)
	case tea.MouseWheelUp:
		l.up(count)
	}
}

// itemLine renders a single list item with a cursor indicator and lipgloss styling.
func itemLine(selected bool, content string) string {
	prefix := " "
	style := ItemStyle
	if selected {
		prefix = ">"
		style = SelectedItemStyle
	}
	return style.Render(prefix + " " + content)
}

// viewMargin wraps a rendered view with horizontal margins.
func viewMargin(view string) string {
	return ViewStyle.Render(view)
}

// renderListItem renders a list item with truncation (unselected) or word-wrapping (selected).
// It respects the terminal width, applying a right margin for readability.
func renderListItem(selected bool, width int, name, description string) string {
	contentWidth := width - 12

	if !selected {
		nameWidth := lipgloss.Width(name)
		descMax := contentWidth - 8 - nameWidth
		var line string
		if description != "" {
			truncated := truncate.StringWithTail(description, uint(descMax), "...") //nolint:gosec
			line = "  " + name + "  " + MutedStyle.Render(truncated)
		} else {
			line = "  " + name
		}
		return ItemStyle.Render(line) + "\n"
	}

	var b strings.Builder
	b.WriteString(SelectedItemStyle.Render("> " + name))
	if description != "" {
		wrapWidth := contentWidth - 4
		wrapped := wordwrap.String(description, wrapWidth)
		for line := range strings.SplitSeq(wrapped, "\n") {
			b.WriteString("\n")
			b.WriteString(SelectedItemStyle.Render("  " + MutedStyle.Render(line)))
		}
	}
	b.WriteString("\n")
	return b.String()
}
