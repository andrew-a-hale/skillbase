package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetDefaultRepo(t *testing.T) {
	t.Run("env var set", func(t *testing.T) {
		t.Setenv("SKILLS_DEFAULT_REPO", "https://example.com/repo")
		repo, err := getDefaultRepo()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo != "https://example.com/repo" {
			t.Fatalf("unexpected repo: %q", repo)
		}
	})

	t.Run("env var missing", func(t *testing.T) {
		os.Unsetenv("SKILLS_DEFAULT_REPO")
		_, err := getDefaultRepo()
		if err == nil {
			t.Fatal("expected error when env var is missing")
		}
		if !strings.Contains(err.Error(), "SKILLS_DEFAULT_REPO") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		input     string
		wantRepo  string
		wantPath  string
	}{
		{"https://github.com/user/repo", "https://github.com/user/repo", ""},
		{"https://github.com/user/repo/skill/path", "https://github.com/user/repo", "skill/path"},
		{"https://github.com/user/repo.git", "https://github.com/user/repo", ""},
		{"https://github.com/user/repo.git/skill", "https://github.com/user/repo.git", "skill"},
		{"short", "short", ""},
		{"a/b", "a/b", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			repo, path := parseRepoURL(tt.input)
			if repo != tt.wantRepo {
				t.Fatalf("repo: got %q, want %q", repo, tt.wantRepo)
			}
			if path != tt.wantPath {
				t.Fatalf("path: got %q, want %q", path, tt.wantPath)
			}
		})
	}
}

func TestDispatchHelp(t *testing.T) {
	if err := Dispatch([]string{"skills", "help"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatchUnknown(t *testing.T) {
	if err := Dispatch([]string{"skills", "foobar"}); err == nil {
		t.Fatal("expected error for unknown command")
	}
}

func TestDispatchNoArgs(t *testing.T) {
	if err := Dispatch([]string{"skills"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListSkillsGlobal(t *testing.T) {
	t.Run("global empty", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldPath := SKILLS_PATH
		SKILLS_PATH = filepath.Join(tmpDir, ".skills")
		defer func() { SKILLS_PATH = oldPath }()

		if err := os.MkdirAll(SKILLS_PATH, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}

		if err := listSkills(true); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("global with skills", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldPath := SKILLS_PATH
		SKILLS_PATH = filepath.Join(tmpDir, ".skills")
		defer func() { SKILLS_PATH = oldPath }()

		skillDir := filepath.Join(SKILLS_PATH, "my-skill")
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: my-skill\n---\n"), 0o644); err != nil {
			t.Fatalf("write skill: %v", err)
		}

		if err := listSkills(true); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
