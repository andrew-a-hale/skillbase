package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNewGetModelNoPreselections(t *testing.T) {
	skills := []SkillInfo{{Name: "foo"}, {Name: "bar"}}
	m := NewGetModel("", "", false, []string{"claude"})
	m.skills = skills
	m.loading = false
	m.applyPreselections()
	if m.step != getStepSkill {
		t.Fatalf("expected step skill, got %d", m.step)
	}
}

func TestNewGetModelPreSkill(t *testing.T) {
	skills := []SkillInfo{{Name: "foo"}, {Name: "bar"}}
	m := NewGetModel("bar", "", false, []string{"claude"})
	m.skills = skills
	m.loading = false
	m.applyPreselections()
	if m.step != getStepScope {
		t.Fatalf("expected step scope, got %d", m.step)
	}
	if m.selectedSkill != 1 {
		t.Fatalf("expected selectedSkill 1, got %d", m.selectedSkill)
	}
	if !m.singleMode {
		t.Fatal("expected singleMode true")
	}
}

func TestNewGetModelPreGlobal(t *testing.T) {
	skills := []SkillInfo{{Name: "foo"}}
	m := NewGetModel("foo", "", true, []string{"claude", "agents"})
	m.skills = skills
	m.loading = false
	m.applyPreselections()
	if m.step != getStepConfirm {
		t.Fatalf("expected step confirm, got %d", m.step)
	}
	if !m.global {
		t.Fatal("expected global true")
	}
}

func TestNewGetModelPreAgent(t *testing.T) {
	skills := []SkillInfo{{Name: "foo"}}
	m := NewGetModel("foo", "claude", false, []string{"claude", "agents"})
	m.skills = skills
	m.loading = false
	m.applyPreselections()

	// Select project scope
	model, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
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
	m := NewGetModel("", "", false, []string{"claude"})
	m.skills = skills
	m.loading = false
	m.applyPreselections()

	model, _ := m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	m = model.(*GetModel)
	model, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = model.(*GetModel)
	if m.step != getStepScope {
		t.Fatalf("expected step scope, got %d", m.step)
	}
	if m.selectedSkill != 1 {
		t.Fatalf("expected selectedSkill 1, got %d", m.selectedSkill)
	}
	if !m.selectedSkills[1] {
		t.Fatal("expected selectedSkills[1] true")
	}
}

func TestGetModelSkillMultiSelect(t *testing.T) {
	skills := []SkillInfo{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	m := NewGetModel("", "", false, []string{"claude"})
	m.skills = skills
	m.loading = false
	m.applyPreselections()

	// Select skill a (cursor starts at 0)
	model, _ := m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	m = model.(*GetModel)
	if !m.selectedSkills[0] {
		t.Fatal("expected skill 0 selected")
	}

	// Move to skill b and select it
	model, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	m = model.(*GetModel)
	model, _ = m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	m = model.(*GetModel)
	if !m.selectedSkills[1] {
		t.Fatal("expected skill 1 selected")
	}

	// Move to skill c but don't select it
	model, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	m = model.(*GetModel)

	// Confirm should advance with a and b selected
	model, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = model.(*GetModel)
	if m.step != getStepScope {
		t.Fatalf("expected step scope, got %d", m.step)
	}
}

func TestGetModelSingleModeSpaceIgnored(t *testing.T) {
	skills := []SkillInfo{{Name: "a"}, {Name: "b"}}
	m := NewGetModel("a", "", false, []string{"claude"})
	m.skills = skills
	m.loading = false
	m.applyPreselections()
	if !m.singleMode {
		t.Fatal("expected singleMode")
	}

	model, _ := m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	m = model.(*GetModel)
	if len(m.selectedSkills) != 1 {
		t.Fatalf("expected 1 selectedSkill, got %d", len(m.selectedSkills))
	}
}

func TestGetModelScopeToggle(t *testing.T) {
	skills := []SkillInfo{{Name: "a"}}
	m := NewGetModel("a", "", false, []string{"claude"})
	m.skills = skills
	m.loading = false
	m.applyPreselections()

	model, _ := m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	m = model.(*GetModel)
	if !m.global {
		t.Fatal("expected global true after toggle")
	}

	model, _ = m.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	m = model.(*GetModel)
	if m.global {
		t.Fatal("expected global false after toggle back")
	}
}

func TestGetModelScopeConfirmProject(t *testing.T) {
	skills := []SkillInfo{{Name: "a"}}
	m := NewGetModel("a", "", false, []string{"claude"})
	m.skills = skills
	m.loading = false
	m.applyPreselections()

	model, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = model.(*GetModel)
	if m.step != getStepConfirm {
		t.Fatalf("expected step confirm (single agent), got %d", m.step)
	}
}

func TestGetModelScopeConfirmGlobal(t *testing.T) {
	skills := []SkillInfo{{Name: "a"}}
	m := NewGetModel("a", "", false, []string{"claude", "agents"})
	m.skills = skills
	m.loading = false
	m.applyPreselections()

	model, _ := m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	m = model.(*GetModel)
	model, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = model.(*GetModel)
	if m.step != getStepConfirm {
		t.Fatalf("expected step confirm (global), got %d", m.step)
	}
}

func TestGetModelAgentSelect(t *testing.T) {
	skills := []SkillInfo{{Name: "a"}}
	m := NewGetModel("a", "", false, []string{"claude", "agents"})
	m.skills = skills
	m.loading = false
	m.applyPreselections()

	model, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = model.(*GetModel)
	if m.step != getStepAgent {
		t.Fatalf("expected step agent, got %d", m.step)
	}

	model, _ = m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	m = model.(*GetModel)
	if !m.selectedAgents["claude"] {
		t.Fatal("expected claude selected")
	}

	model, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	m = model.(*GetModel)
	model, _ = m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	m = model.(*GetModel)
	if !m.selectedAgents["agents"] {
		t.Fatal("expected agents selected")
	}
}

func TestGetModelAgentConfirm(t *testing.T) {
	skills := []SkillInfo{{Name: "a"}}
	m := NewGetModel("a", "", false, []string{"claude"})
	m.skills = skills
	m.loading = false
	m.applyPreselections()
	m.step = getStepAgent
	m.cursor = 0
	m.selectedAgents["claude"] = true

	// Press Enter to advance to confirm
	model, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = model.(*GetModel)
	if m.step != getStepConfirm {
		t.Fatalf("expected step confirm, got %d", m.step)
	}

	// Press Enter again to confirm and quit
	model, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
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

func TestGetModelMultiSelectConfirm(t *testing.T) {
	skills := []SkillInfo{
		{Name: "a", Path: "skills/a"},
		{Name: "b", Path: "skills/b"},
	}
	m := NewGetModel("", "", false, []string{"claude"})
	m.skills = skills
	m.loading = false
	m.applyPreselections()

	// Select both skills
	m.selectedSkills[0] = true
	m.selectedSkills[1] = true
	m.step = getStepConfirm

	model, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = model.(*GetModel)
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	if m.Result == nil {
		t.Fatal("expected result")
	}
	if len(m.Result.SkillNames) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(m.Result.SkillNames))
	}
	if m.Result.SkillNames[0] != "a" || m.Result.SkillNames[1] != "b" {
		t.Fatalf("unexpected skill names: %v", m.Result.SkillNames)
	}
	if m.Result.SkillPaths[0] != "skills/a" || m.Result.SkillPaths[1] != "skills/b" {
		t.Fatalf("unexpected skill paths: %v", m.Result.SkillPaths)
	}
}

func TestGetModelBack(t *testing.T) {
	skills := []SkillInfo{{Name: "a"}}
	m := NewGetModel("a", "", false, []string{"claude"})
	m.skills = skills
	m.loading = false
	m.applyPreselections()
	m.step = getStepScope

	model, _ := m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
	m = model.(*GetModel)
	if m.step != getStepSkill {
		t.Fatalf("expected step skill, got %d", m.step)
	}
}

func TestGetModelQuit(t *testing.T) {
	m := NewGetModel("", "", false, nil)
	m.skills = []SkillInfo{{Name: "a"}}
	m.loading = false
	m.applyPreselections()
	model, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = model.(*GetModel)
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	if !m.Cancelled {
		t.Fatal("expected cancelled")
	}
}
