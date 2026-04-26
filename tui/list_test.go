package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestListModelAllSkills(t *testing.T) {
	project := []SkillInfo{{Name: "p1"}, {Name: "p2"}}
	global := []SkillInfo{{Name: "g1"}}
	m := NewListModel(project, global)

	all := m.allSkills()
	if len(all) != 3 {
		t.Fatalf("expected 3 skills, got %d", len(all))
	}
}

func TestListModelQuit(t *testing.T) {
	m := NewListModel(nil, nil)
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = model.(*ListModel)
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	if !m.Cancelled {
		t.Fatal("expected cancelled")
	}
}
