package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeExecutor struct {
	output []byte
	err    error
	calls  [][]string
}

func (f *fakeExecutor) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, append([]string{name}, args...))
	return f.output, f.err
}

func TestNewGitRepository(t *testing.T) {
	if _, err := NewGitRepository(""); err == nil {
		t.Fatal("expected error for empty URL")
	}
	repo, err := NewGitRepository("https://example.com/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.url != "https://example.com/repo" {
		t.Fatalf("url mismatch: got %q", repo.url)
	}
}

func TestGitRepositoryClone(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		exec := &fakeExecutor{output: []byte("ok")}
		repo := &GitRepository{url: "https://example.com/repo", executor: exec}

		path, cleanup, err := repo.Clone(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer cleanup()

		if len(exec.calls) != 1 {
			t.Fatalf("expected 1 call, got %d", len(exec.calls))
		}
		if exec.calls[0][0] != "git" || exec.calls[0][1] != "clone" {
			t.Fatalf("unexpected call: %v", exec.calls[0])
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("clone path should exist: %v", err)
		}
	})

	t.Run("clone failure", func(t *testing.T) {
		exec := &fakeExecutor{err: errors.New("clone failed")}
		repo := &GitRepository{url: "https://example.com/repo", executor: exec}

		_, _, err := repo.Clone(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrCloneFailed) {
			t.Fatalf("expected ErrCloneFailed, got: %v", err)
		}
	})
}

func TestListSkills(t *testing.T) {
	tmpDir := t.TempDir()

	makeSkill := func(path, name, desc string) {
		dir := filepath.Join(tmpDir, path, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		content := fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n", name, desc)
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
			t.Fatalf("write skill: %v", err)
		}
	}

	makeSkill("skills", "test-skill", "a test skill")
	makeSkill("skills", "another-skill", "another one")

	repo := &GitRepository{url: "https://example.com/repo"}
	skills, err := repo.ListSkills(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}

	names := make(map[string]string)
	for _, s := range skills {
		names[s.Name] = s.Description
	}
	if names["test-skill"] != "a test skill" {
		t.Fatalf("description mismatch for test-skill: %q", names["test-skill"])
	}
}

func TestListSkillsCollision(t *testing.T) {
	tmpDir := t.TempDir()

	for _, path := range []string{"a/skill", "b/skill"} {
		dir := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: skill\n---\n"), 0o644); err != nil {
			t.Fatalf("write skill: %v", err)
		}
	}

	repo := &GitRepository{url: "https://example.com/repo"}
	_, err := repo.ListSkills(tmpDir)
	if err == nil {
		t.Fatal("expected error for name collision")
	}
	if !strings.Contains(err.Error(), "collision") {
		t.Fatalf("expected collision error, got: %v", err)
	}
}

func TestGetSkill(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\ndescription: my desc\n---\n"), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}

	repo := &GitRepository{url: "https://example.com/repo"}
	skill, err := repo.GetSkill(tmpDir, "my-skill")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skill.Name != "my-skill" {
		t.Fatalf("name mismatch: got %q", skill.Name)
	}
	if skill.Description != "my desc" {
		t.Fatalf("description mismatch: got %q", skill.Description)
	}
}

func TestGetSkillNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	repo := &GitRepository{url: "https://example.com/repo"}
	_, err := repo.GetSkill(tmpDir, "missing")
	if err == nil {
		t.Fatal("expected error for missing skill")
	}
	if !errors.Is(err, ErrSkillNotFound) {
		t.Fatalf("expected ErrSkillNotFound, got: %v", err)
	}
}

func TestExtractSkillDescription(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "frontmatter description",
			content:  "---\ndescription: hello world\n---\n",
			expected: "hello world",
		},
		{
			name:     "no description",
			content:  "---\nname: foo\n---\n",
			expected: "",
		},
		{
			name:     "no frontmatter",
			content:  "# Hello\n",
			expected: "",
		},
		{
			name:     "case insensitive",
			content:  "---\nDESCRIPTION: UPPER\n---\n",
			expected: "UPPER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "SKILL.md")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("write file: %v", err)
			}
			got := extractSkillDescription(path)
			if got != tt.expected {
				t.Fatalf("got %q, want %q", got, tt.expected)
			}
		})
	}
}
