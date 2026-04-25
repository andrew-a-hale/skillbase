package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.txt")
	dst := filepath.Join(tmpDir, "dst.txt")

	content := []byte("hello, world")
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("copy file: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("content mismatch: got %q, want %q", got, content)
	}
}

func TestCopyFileMissingSource(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "missing.txt")
	dst := filepath.Join(tmpDir, "dst.txt")

	if err := CopyFile(src, dst); err == nil {
		t.Fatal("expected error for missing source file")
	}
}

func TestCopyDir(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")

	if err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0o755); err != nil {
		t.Fatalf("create src subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatalf("write a.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "subdir", "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatalf("write b.txt: %v", err)
	}

	if err := CopyDir(srcDir, dstDir); err != nil {
		t.Fatalf("copy dir: %v", err)
	}

	for _, want := range []struct {
		path    string
		content string
	}{
		{filepath.Join(dstDir, "a.txt"), "a"},
		{filepath.Join(dstDir, "subdir", "b.txt"), "b"},
	} {
		got, err := os.ReadFile(want.path)
		if err != nil {
			t.Fatalf("read %s: %v", want.path, err)
		}
		if string(got) != want.content {
			t.Fatalf("content mismatch in %s: got %q, want %q", want.path, got, want.content)
		}
	}
}
