package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileSystemScopeResolverDetectAgents(t *testing.T) {
	tmpDir := t.TempDir()
	resolver := &FileSystemScopeResolver{home: tmpDir, cwd: tmpDir}

	t.Run("project agents", func(t *testing.T) {
		if err := os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}

		agents := resolver.DetectAgents(false)
		if len(agents) != 2 {
			t.Fatalf("expected 2 agents, got %d", len(agents))
		}
	})

	t.Run("global agents", func(t *testing.T) {
		agents := resolver.DetectAgents(true)
		if len(agents) != 2 {
			t.Fatalf("expected 2 global agents, got %d", len(agents))
		}
	})

	t.Run("no agents", func(t *testing.T) {
		emptyDir := t.TempDir()
		r := &FileSystemScopeResolver{home: emptyDir, cwd: emptyDir}
		agents := r.DetectAgents(false)
		if len(agents) != 0 {
			t.Fatalf("expected 0 agents, got %d", len(agents))
		}
	})
}

func TestFileSystemScopeResolverResolve(t *testing.T) {
	tmpDir := t.TempDir()
	resolver := &FileSystemScopeResolver{home: tmpDir, cwd: tmpDir}

	if err := os.MkdirAll(filepath.Join(tmpDir, ".claude", "skillbase"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	t.Run("project scope", func(t *testing.T) {
		targets, err := resolver.Resolve("my-skill", false, "claude")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(targets) != 1 {
			t.Fatalf("expected 1 target, got %d", len(targets))
		}
		if targets[0].Agent != "claude" {
			t.Fatalf("agent mismatch: got %q", targets[0].Agent)
		}
		want := filepath.Join(tmpDir, ".claude", "skillbase")
		if targets[0].Path != want {
			t.Fatalf("path mismatch: got %q, want %q", targets[0].Path, want)
		}
	})

	t.Run("global scope", func(t *testing.T) {
		targets, err := resolver.Resolve("my-skill", true, "claude")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join(tmpDir, ".claude", "skillbase")
		if targets[0].Path != want {
			t.Fatalf("path mismatch: got %q, want %q", targets[0].Path, want)
		}
	})

	t.Run("no agents found", func(t *testing.T) {
		emptyDir := t.TempDir()
		r := &FileSystemScopeResolver{home: emptyDir, cwd: emptyDir}
		_, err := r.Resolve("my-skill", false, "")
		if err == nil {
			t.Fatal("expected error when no agents found")
		}
	})
}

func TestNewFileSystemScopeResolver(t *testing.T) {
	resolver, err := NewFileSystemScopeResolver()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolver.home == "" {
		t.Fatal("expected home to be set")
	}
	if resolver.cwd == "" {
		t.Fatal("expected cwd to be set")
	}
}
