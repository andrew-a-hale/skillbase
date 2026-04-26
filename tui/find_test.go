package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFindModelFilter(t *testing.T) {
	skills := []SkillInfo{
		{Name: "alpha", Description: "first"},
		{Name: "beta", Description: "second"},
		{Name: "gamma", Description: "third"},
	}
	m := NewFindModel(skills, "")

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
	m := NewFindModel(skills, "")

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = model.(*FindModel)
	if m.filter != "a" {
		t.Fatalf("expected filter 'a', got %q", m.filter)
	}

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = model.(*FindModel)
	if m.filter != "" {
		t.Fatalf("expected empty filter, got %q", m.filter)
	}
}

func TestFindModelNavigationWithFilter(t *testing.T) {
	skills := []SkillInfo{
		{Name: "alpha"},
		{Name: "beta"},
	}
	m := NewFindModel(skills, "")
	m.filter = "b"
	m.cursor = 0

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = model.(*FindModel)
	if m.cursor != 0 {
		t.Fatalf("expected cursor 0 (only 1 match), got %d", m.cursor)
	}
}

func TestFindModelQuit(t *testing.T) {
	m := NewFindModel([]SkillInfo{{Name: "a"}}, "")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = model.(*FindModel)
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	if !m.Cancelled {
		t.Fatal("expected cancelled")
	}
}
