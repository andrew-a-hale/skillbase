package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateModelConfirm(t *testing.T) {
	skills := []SkillInfo{{Name: "a"}, {Name: "b"}}
	m := NewUpdateModel(skills)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = model.(*UpdateModel)
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*UpdateModel)
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	if m.Result == nil {
		t.Fatal("expected result")
	}
	if m.Result.SkillName != "b" {
		t.Fatalf("expected skill b, got %s", m.Result.SkillName)
	}
}

func TestUpdateModelQuit(t *testing.T) {
	m := NewUpdateModel([]SkillInfo{{Name: "a"}})
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = model.(*UpdateModel)
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	if !m.Cancelled {
		t.Fatal("expected cancelled")
	}
}


