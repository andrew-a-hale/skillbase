package tui

import tea "github.com/charmbracelet/bubbletea"

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
	switch msg.Button {
	case tea.MouseButtonWheelDown:
		l.down(count)
	case tea.MouseButtonWheelUp:
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
