package tui

import tea "github.com/charmbracelet/bubbletea"

type KeyMap struct {
	Up      []string
	Down    []string
	Left    []string
	Right   []string
	Select  []string
	Confirm []string
	Back    []string
	Quit    []string
}

var DefaultKeyMap = KeyMap{
	Up:      []string{"up", "k"},
	Down:    []string{"down", "j"},
	Left:    []string{"left", "h"},
	Right:   []string{"right", "l", "enter"},
	Select:  []string{" "},
	Confirm: []string{"enter"},
	Back:    []string{"left", "h"},
	Quit:    []string{"esc", "q", "ctrl+c"},
}

func IsKey(msg tea.KeyMsg, keys []string) bool {
	for _, k := range keys {
		if msg.String() == k {
			return true
		}
	}
	return false
}
