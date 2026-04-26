package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewGetModelNoPreselections(t *testing.T) {
	skills := []SkillInfo{{Name: "foo"}, {Name: "bar"}}
	m := NewGetModel(skills, "", "", false, []string{"claude"})
	if m.step != getStepSkill {
		t.Fatalf("expected step skill, got %d", m.step)
	}
}

func TestNewGetModelPreSkill(t *testing.T) {
	skills := []SkillInfo{{Name: "foo"}, {Name: "bar"}}
	m := NewGetModel(skills, "bar", "", false, []string{"claude"})
	if m.step != getStepScope {
		t.Fatalf("expected step scope, got %d", m.step)
	}
	if m.selectedSkill != 1 {
		t.Fatalf("expected selectedSkill 1, got %d", m.selectedSkill)
	}
}

func TestNewGetModelPreGlobal(t *testing.T) {
	skills := []SkillInfo{{Name: "foo"}}
	m := NewGetModel(skills, "foo", "", true, []string{"claude", "agents"})
	if m.step != getStepConfirm {
		t.Fatalf("expected step confirm, got %d", m.step)
	}
	if !m.global {
		t.Fatal("expected global true")
	}
}

func TestNewGetModelPreAgent(t *testing.T) {
	skills := []SkillInfo{{Name: "foo"}}
	m := NewGetModel(skills, "foo", "claude", false, []string{"claude", "agents"})
	if m.step != getStepScope {
		t.Fatalf("expected step scope, got %d", m.step)
	}

	// Select project scope
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*GetModel)
	if m.step != getStepAgent {
		t.Fatalf("expected step agent, got %d", m.step)
	}
	if !m.selectedAgents["claude"] {
		t.Fatal("expected claude pre-selected")
	}
}

func TestGetModelSkillConfirm(t *testing.T) {
	skills := []SkillInfo{{Name: "a"}, {Name: "b"}}
	m := NewGetModel(skills, "", "", false, []string{"claude"})

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = model.(*GetModel)
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*GetModel)
	if m.step != getStepScope {
		t.Fatalf("expected step scope, got %d", m.step)
	}
	if m.selectedSkill != 1 {
		t.Fatalf("expected selectedSkill 1, got %d", m.selectedSkill)
	}
}

func TestGetModelScopeToggle(t *testing.T) {
	skills := []SkillInfo{{Name: "a"}}
	m := NewGetModel(skills, "a", "", false, []string{"claude"})

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = model.(*GetModel)
	if !m.global {
		t.Fatal("expected global true after toggle")
	}

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = model.(*GetModel)
	if m.global {
		t.Fatal("expected global false after toggle back")
	}
}

func TestGetModelScopeConfirmProject(t *testing.T) {
	skills := []SkillInfo{{Name: "a"}}
	m := NewGetModel(skills, "a", "", false, []string{"claude"})

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*GetModel)
	if m.step != getStepConfirm {
		t.Fatalf("expected step confirm (single agent), got %d", m.step)
	}
}

func TestGetModelScopeConfirmGlobal(t *testing.T) {
	skills := []SkillInfo{{Name: "a"}}
	m := NewGetModel(skills, "a", "", false, []string{"claude", "agents"})

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = model.(*GetModel)
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*GetModel)
	if m.step != getStepConfirm {
		t.Fatalf("expected step confirm (global), got %d", m.step)
	}
}

func TestGetModelAgentSelect(t *testing.T) {
	skills := []SkillInfo{{Name: "a"}}
	m := NewGetModel(skills, "a", "", false, []string{"claude", "agents"})

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*GetModel)
	if m.step != getStepAgent {
		t.Fatalf("expected step agent, got %d", m.step)
	}

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = model.(*GetModel)
	if !m.selectedAgents["claude"] {
		t.Fatal("expected claude selected")
	}

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = model.(*GetModel)
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = model.(*GetModel)
	if !m.selectedAgents["agents"] {
		t.Fatal("expected agents selected")
	}
}

func TestGetModelAgentConfirm(t *testing.T) {
	skills := []SkillInfo{{Name: "a"}}
	m := NewGetModel(skills, "a", "", false, []string{"claude"})
	m.step = getStepAgent
	m.cursor = 0
	m.selectedAgents["claude"] = true

	// Press Enter to advance to confirm
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*GetModel)
	if m.step != getStepConfirm {
		t.Fatalf("expected step confirm, got %d", m.step)
	}

	// Press Enter again to confirm and quit
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*GetModel)
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	if m.Result == nil {
		t.Fatal("expected result")
	}
	if m.Result.SkillName != "a" {
		t.Fatalf("expected skill a, got %s", m.Result.SkillName)
	}
}

func TestGetModelBack(t *testing.T) {
	skills := []SkillInfo{{Name: "a"}}
	m := NewGetModel(skills, "a", "", false, []string{"claude"})
	m.step = getStepScope

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = model.(*GetModel)
	if m.step != getStepSkill {
		t.Fatalf("expected step skill, got %d", m.step)
	}
}

func TestGetModelQuit(t *testing.T) {
	m := NewGetModel([]SkillInfo{{Name: "a"}}, "", "", false, nil)
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = model.(*GetModel)
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	if !m.Cancelled {
		t.Fatal("expected cancelled")
	}
}
