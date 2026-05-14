package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestFindModelFilter(t *testing.T) {
	skills := []SkillInfo{
		{Name: "alpha", Description: "first"},
		{Name: "beta", Description: "second"},
		{Name: "gamma", Description: "third"},
	}
	m := NewFindModel("")
	m.skills = skills
	m.loading = false

	filtered := m.filteredSkills()
	if len(filtered) != 3 {
		t.Fatalf("expected 3 skills, got %d", len(filtered))
	}

	m.filter = "al"
	filtered = m.filteredSkills()
	if len(filtered) != 1 || filtered[0].Name != "alpha" {
		t.Fatalf("expected 1 skill (alpha), got %v", filtered)
	}

	m.filter = "second"
	filtered = m.filteredSkills()
	if len(filtered) != 1 || filtered[0].Name != "beta" {
		t.Fatalf("expected 1 skill (beta), got %v", filtered)
	}
}

func TestFindModelTypingFilter(t *testing.T) {
	skills := []SkillInfo{
		{Name: "alpha"},
		{Name: "beta"},
	}
	m := NewFindModel("")
	m.skills = skills
	m.loading = false

	model, _ := m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	m = model.(*FindModel)
	if m.filter != "a" {
		t.Fatalf("expected filter 'a', got %q", m.filter)
	}

	model, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	m = model.(*FindModel)
	if m.filter != "" {
		t.Fatalf("expected empty filter, got %q", m.filter)
	}
}

func TestFindModelFilterByPath(t *testing.T) {
	skills := []SkillInfo{
		{Name: "write-skill", Path: "root/generic/write-skill", Description: "writes things"},
		{Name: "read-skill", Path: "root/specific/read-skill", Description: "reads things"},
	}
	m := NewFindModel("")
	m.skills = skills
	m.loading = false

	m.filter = "generic"
	filtered := m.filteredSkills()
	if len(filtered) != 1 || filtered[0].Name != "write-skill" {
		t.Fatalf("expected 1 skill (write-skill), got %v", filtered)
	}

	m.filter = "root/specific"
	filtered = m.filteredSkills()
	if len(filtered) != 1 || filtered[0].Name != "read-skill" {
		t.Fatalf("expected 1 skill (read-skill), got %v", filtered)
	}
}

func TestFindModelNavigationWithFilter(t *testing.T) {
	skills := []SkillInfo{
		{Name: "alpha"},
		{Name: "beta"},
	}
	m := NewFindModel("")
	m.skills = skills
	m.loading = false
	m.filter = "b"
	m.cursor = 0

	model, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = model.(*FindModel)
	if m.cursor != 0 {
		t.Fatalf("expected cursor 0 (only 1 match), got %d", m.cursor)
	}
}

func TestFindModelQuit(t *testing.T) {
	m := NewFindModel("")
	m.skills = []SkillInfo{{Name: "a"}}
	m.loading = false
	model, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	m = model.(*FindModel)
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	if !m.Cancelled {
		t.Fatal("expected cancelled")
	}
}
