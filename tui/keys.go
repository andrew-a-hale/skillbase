package tui

import (
	"slices"

	tea "charm.land/bubbletea/v2"
)

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
	Select:  []string{"space"},
	Confirm: []string{"enter"},
	Back:    []string{"left", "h"},
	Quit:    []string{"esc", "q", "ctrl+c"},
}

func IsKey(msg tea.KeyPressMsg, keys []string) bool {
	return slices.Contains(keys, msg.String())
}
