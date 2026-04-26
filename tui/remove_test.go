package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewRemoveModelNoPreselections(t *testing.T) {
	m := NewRemoveModel([]string{"g1"}, []SkillInfo{{Name: "p1"}}, false, "")
	if m.step != removeStepScope {
		t.Fatalf("expected step scope, got %d", m.step)
	}
}

func TestNewRemoveModelPreGlobal(t *testing.T) {
	m := NewRemoveModel([]string{"g1"}, []SkillInfo{{Name: "p1"}}, true, "")
	if m.step != removeStepSkills {
		t.Fatalf("expected step skills, got %d", m.step)
	}
	if !m.global {
		t.Fatal("expected global true")
	}
}

func TestRemoveModelScopeToggle(t *testing.T) {
	m := NewRemoveModel([]string{"g1"}, []SkillInfo{{Name: "p1"}}, false, "")
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = model.(*RemoveModel)
	if !m.global {
		t.Fatal("expected global true")
	}
}

func TestRemoveModelSkillSelect(t *testing.T) {
	m := NewRemoveModel([]string{"g1", "g2"}, []SkillInfo{{Name: "p1"}}, true, "")
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = model.(*RemoveModel)
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = model.(*RemoveModel)
	if !m.selected[1] {
		t.Fatal("expected index 1 selected")
	}
}

func TestRemoveModelConfirm(t *testing.T) {
	m := NewRemoveModel([]string{"g1", "g2"}, []SkillInfo{{Name: "p1"}}, true, "")
	m.selected[0] = true
	m.selected[1] = true
	m.step = removeStepConfirm

	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*RemoveModel)
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	if m.Result == nil {
		t.Fatal("expected result")
	}
	if len(m.Result.SkillNames) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(m.Result.SkillNames))
	}
}

func TestRemoveModelBack(t *testing.T) {
	m := NewRemoveModel([]string{"g1"}, []SkillInfo{{Name: "p1"}}, true, "")
	m.step = removeStepSkills
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = model.(*RemoveModel)
	if m.step != removeStepScope {
		t.Fatalf("expected step scope, got %d", m.step)
	}
}

func TestRemoveModelQuit(t *testing.T) {
	m := NewRemoveModel(nil, nil, false, "")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = model.(*RemoveModel)
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	if !m.Cancelled {
		t.Fatal("expected cancelled")
	}
}
