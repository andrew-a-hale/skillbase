package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSkillStoreInstallAndExists(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewFileSystemSkillStore(&fakeResolver{}, tmpDir)

	srcDir := filepath.Join(tmpDir, "src", "my-skill")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("---\nname: my-skill\n---\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if store.Exists("my-skill") {
		t.Fatal("expected skill to not exist before install")
	}

	if err := store.Install("my-skill", srcDir); err != nil {
		t.Fatalf("install: %v", err)
	}

	if !store.Exists("my-skill") {
		t.Fatal("expected skill to exist after install")
	}

	skillFile := filepath.Join(tmpDir, "my-skill", "SKILL.md")
	if _, err := os.Stat(skillFile); err != nil {
		t.Fatalf("skill file missing: %v", err)
	}
}

func TestSkillStoreProvenance(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewFileSystemSkillStore(&fakeResolver{}, tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, "test-skill"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	prov := Provenance{
		Source:      "https://example.com/repo",
		SkillPath:   "skills/test-skill",
		InstalledAt: "2024-01-01T00:00:00Z",
	}

	if err := store.WriteProvenance("test-skill", prov); err != nil {
		t.Fatalf("write: %v", err)
	}

	read, err := store.ReadProvenance("test-skill")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if read.Source != prov.Source {
		t.Fatalf("source: got %q, want %q", read.Source, prov.Source)
	}
	if read.SkillPath != prov.SkillPath {
		t.Fatalf("skillPath: got %q, want %q", read.SkillPath, prov.SkillPath)
	}
}

func TestSkillStoreLink(t *testing.T) {
	tmpDir := t.TempDir()

	skillbase := filepath.Join(tmpDir, ".skillbase")
	agentDir := filepath.Join(tmpDir, ".claude", "skills")

	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("mkdir agent: %v", err)
	}

	resolver, err := NewFileSystemScopeResolver(tmpDir, tmpDir)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	store := NewFileSystemSkillStore(resolver, skillbase)

	srcDir := filepath.Join(tmpDir, "src", "my-skill")
	os.MkdirAll(srcDir, 0o755)
	os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte(""), 0o644)
	store.Install("my-skill", srcDir)

	if err := store.Link("my-skill", "", false); err != nil {
		t.Fatalf("link: %v", err)
	}

	linkPath := filepath.Join(agentDir, "my-skill")
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("link missing: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("expected symlink")
	}
}

func TestSkillStoreRemove(t *testing.T) {
	tmpDir := t.TempDir()

	skillbase := filepath.Join(tmpDir, ".skillbase")
	agentDir := filepath.Join(tmpDir, ".claude", "skills")

	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("mkdir agent: %v", err)
	}

	resolver, err := NewFileSystemScopeResolver(tmpDir, tmpDir)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	store := NewFileSystemSkillStore(resolver, skillbase)

	srcDir := filepath.Join(tmpDir, "src", "my-skill")
	os.MkdirAll(srcDir, 0o755)
	os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte(""), 0o644)
	store.Install("my-skill", srcDir)
	store.Link("my-skill", "", false)

	if err := store.Remove("my-skill", "", false); err != nil {
		t.Fatalf("remove: %v", err)
	}

	linkPath := filepath.Join(agentDir, "my-skill")
	if _, err := os.Lstat(linkPath); err == nil {
		t.Fatal("expected symlink to be removed")
	}
}

func TestSkillStoreListInstalled_ProjectScopeDescription(t *testing.T) {
	tmpDir := t.TempDir()

	skillbase := filepath.Join(tmpDir, ".skillbase")
	agentDir := filepath.Join(tmpDir, ".claude", "skills")

	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("mkdir agent: %v", err)
	}

	resolver, err := NewFileSystemScopeResolver(tmpDir, tmpDir)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	store := NewFileSystemSkillStore(resolver, skillbase)

	srcDir := filepath.Join(tmpDir, "src", "my-skill")
	os.MkdirAll(srcDir, 0o755)
	os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("---\ndescription: hello world\n---\n"), 0o644)
	store.Install("my-skill", srcDir)
	store.Link("my-skill", "", false)

	project, _, err := store.ListInstalled()
	if err != nil {
		t.Fatalf("list installed: %v", err)
	}

	if len(project) != 1 {
		t.Fatalf("expected 1 project skill, got %d", len(project))
	}
	if project[0].Description != "hello world" {
		t.Fatalf("expected description %q, got %q", "hello world", project[0].Description)
	}
}

type fakeResolver struct{}

func (f *fakeResolver) DetectAgents(global bool) []string { return nil }
func (f *fakeResolver) Resolve(skillName string, global bool, agentFilter string) ([]InstallationTarget, error) {
	return nil, nil
}
