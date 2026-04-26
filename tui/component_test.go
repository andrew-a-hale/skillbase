package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestListUp(t *testing.T) {
	l := &list{cursor: 2}
	l.up(5)
	if l.cursor != 1 {
		t.Fatalf("expected cursor 1, got %d", l.cursor)
	}
	l.up(5)
	if l.cursor != 0 {
		t.Fatalf("expected cursor 0, got %d", l.cursor)
	}
	l.up(5)
	if l.cursor != 0 {
		t.Fatalf("expected clamped at 0, got %d", l.cursor)
	}
}

func TestListDown(t *testing.T) {
	l := &list{cursor: 0}
	l.down(3)
	if l.cursor != 1 {
		t.Fatalf("expected cursor 1, got %d", l.cursor)
	}
	l.down(3)
	if l.cursor != 2 {
		t.Fatalf("expected cursor 2, got %d", l.cursor)
	}
	l.down(3)
	if l.cursor != 2 {
		t.Fatalf("expected clamped at 2, got %d", l.cursor)
	}
}

func TestListDownZeroItems(t *testing.T) {
	l := &list{cursor: 0}
	l.down(0)
	if l.cursor != 0 {
		t.Fatalf("expected no-op on empty list, got %d", l.cursor)
	}
}

func TestListReset(t *testing.T) {
	l := &list{cursor: 5}
	l.reset()
	if l.cursor != 0 {
		t.Fatalf("expected cursor 0 after reset, got %d", l.cursor)
	}
}

func TestListHandleMouse(t *testing.T) {
	l := &list{cursor: 1}
	l.handleMouse(tea.MouseMsg{Button: tea.MouseButtonWheelDown}, 3)
	if l.cursor != 2 {
		t.Fatalf("expected cursor 2, got %d", l.cursor)
	}
	l.handleMouse(tea.MouseMsg{Button: tea.MouseButtonWheelUp}, 2)
	if l.cursor != 1 {
		t.Fatalf("expected cursor 1, got %d", l.cursor)
	}
}

func TestItemLineSelected(t *testing.T) {
	line := itemLine(true, "foo")
	if !strings.Contains(line, "> foo") {
		t.Fatalf("expected selected indicator, got %q", line)
	}
}

func TestItemLineUnselected(t *testing.T) {
	line := itemLine(false, "foo")
	if !strings.Contains(line, "  foo") {
		t.Fatalf("expected unselected indicator, got %q", line)
	}
}
