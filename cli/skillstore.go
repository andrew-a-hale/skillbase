package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrew-a-hale/skillbase/internal/fsutil"
	"github.com/andrew-a-hale/skillbase/internal/skill"
)

// InstalledSkill represents a skill that has been installed locally.
type InstalledSkill struct {
	Name        string
	Description string
	Agents      []string // only populated for project scope
}

// Provenance records where an installed skill originated.
type Provenance struct {
	Source      string `json:"source"`
	SkillPath   string `json:"skill_path"`
	InstalledAt string `json:"installed_at"`
}

// SkillStore manages the installation, linking, and removal of skills.
type SkillStore interface {
	Install(skillName, srcDir string) error
	Link(skillName, agent string, global bool) error
	Remove(skillName, agent string, global bool) error
	ListInstalled() (project []InstalledSkill, global []InstalledSkill, err error)
	ReadProvenance(skillName string) (Provenance, error)
	WriteProvenance(skillName string, p Provenance) error
	Exists(skillName string) bool
}

// FileSystemSkillStore implements SkillStore using the local filesystem.
type FileSystemSkillStore struct {
	resolver      ScopeResolver
	skillbasePath string
}

// NewFileSystemSkillStore creates a FileSystemSkillStore.
func NewFileSystemSkillStore(resolver ScopeResolver, skillbasePath string) *FileSystemSkillStore {
	return &FileSystemSkillStore{
		resolver:      resolver,
		skillbasePath: skillbasePath,
	}
}

// Install copies a skill directory into the skillbase storage.
func (s *FileSystemSkillStore) Install(skillName, srcDir string) error {
	if err := os.MkdirAll(s.skillbasePath, 0o755); err != nil {
		return err
	}
	dstDir := filepath.Join(s.skillbasePath, skillName)
	if err := os.RemoveAll(dstDir); err != nil {
		return err
	}
	return fsutil.CopyDir(srcDir, dstDir)
}

// Link creates symlinks from agent scope directories to the skillbase storage.
func (s *FileSystemSkillStore) Link(skillName, agent string, global bool) error {
	storageDir := filepath.Join(s.skillbasePath, skillName)
	agents, err := s.resolveTargetAgents(agent, global)
	if err != nil {
		return err
	}

	for _, ag := range agents {
		targets, err := s.resolver.Resolve(skillName, global, ag)
		if err != nil {
			return err
		}
		for _, target := range targets {
			if err := os.MkdirAll(target.Path, 0o755); err != nil {
				return err
			}
			linkPath := filepath.Join(target.Path, skillName)
			_ = os.Remove(linkPath)
			relStorage, _ := filepath.Rel(target.Path, storageDir)
			if relStorage == "" {
				relStorage = storageDir
			}
			if err := os.Symlink(relStorage, linkPath); err != nil {
				return fmt.Errorf("failed to link %s: %w", target.Agent, err)
			}
			scopeType := "project"
			if global {
				scopeType = "global"
			}
			fmt.Printf("Linked %s to %s scope (%s)\n", skillName, target.Agent, scopeType)
		}
	}
	return nil
}

// Remove deletes a skill from storage and/or removes its symlinks.
func (s *FileSystemSkillStore) Remove(skillName, agent string, global bool) error {
	if global {
		storageDir := filepath.Join(s.skillbasePath, skillName)
		if _, err := os.Stat(storageDir); os.IsNotExist(err) {
			fmt.Printf("Skill %q not found\n", skillName)
			return nil
		}
		for _, ag := range []string{"claude", "agents"} {
			targets, _ := s.resolver.Resolve(skillName, true, ag)
			for _, target := range targets {
				linkPath := filepath.Join(target.Path, skillName)
				_ = os.Remove(linkPath)
			}
		}
		_ = os.RemoveAll(storageDir)
		fmt.Printf("Removed %s\n", skillName)
	} else {
		agents := s.resolver.DetectAgents(false)
		if agent != "" {
			agents = []string{agent}
		}
		removed := false
		for _, ag := range agents {
			targets, _ := s.resolver.Resolve(skillName, false, ag)
			for _, target := range targets {
				linkPath := filepath.Join(target.Path, skillName)
				if _, err := os.Lstat(linkPath); err == nil {
					_ = os.RemoveAll(linkPath)
					fmt.Printf("Removed %s from %s\n", skillName, ag)
					removed = true
				}
			}
		}
		if !removed {
			fmt.Printf("Skill %q not found\n", skillName)
		}
	}
	return nil
}

// ListInstalled scans the filesystem and returns all installed skills.
func (s *FileSystemSkillStore) ListInstalled() (project []InstalledSkill, global []InstalledSkill, err error) {
	entries, err := os.ReadDir(s.skillbasePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		skillFile := filepath.Join(s.skillbasePath, entry.Name(), "SKILL.md")
		desc := ""
		if data, err := os.ReadFile(skillFile); err == nil {
			fm, _ := skill.ParseFrontmatter(bytes.NewReader(data))
			desc = fm["description"]
		}
		global = append(global, InstalledSkill{
			Name:        entry.Name(),
			Description: desc,
		})
	}

	projectAgents := s.resolver.DetectAgents(false)
	projectMap := make(map[string]*InstalledSkill)
	for _, agent := range projectAgents {
		targets, err := s.resolver.Resolve("", false, agent)
		if err != nil || len(targets) == 0 {
			continue
		}
		skillDir := targets[0].Path
		entries, err := os.ReadDir(skillDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			info, err := os.Lstat(filepath.Join(skillDir, entry.Name()))
			if err != nil || info.Mode()&os.ModeSymlink == 0 {
				continue
			}
			name := entry.Name()
			if existing, ok := projectMap[name]; ok {
				existing.Agents = append(existing.Agents, agent)
			} else {
				target, _ := os.Readlink(filepath.Join(skillDir, name))
				desc := ""
				if target != "" {
					skillFile := filepath.Join(target, "SKILL.md")
					if data, err := os.ReadFile(skillFile); err == nil {
						fm, _ := skill.ParseFrontmatter(bytes.NewReader(data))
						desc = fm["description"]
					}
				}
				projectMap[name] = &InstalledSkill{
					Name:        name,
					Description: desc,
					Agents:      []string{agent},
				}
			}
		}
	}

	for _, info := range projectMap {
		project = append(project, *info)
	}
	return project, global, nil
}

// ReadProvenance reads the provenance metadata for an installed skill.
func (s *FileSystemSkillStore) ReadProvenance(skillName string) (Provenance, error) {
	path := filepath.Join(s.skillbasePath, skillName, ".skillbase-meta")
	data, err := os.ReadFile(path)
	if err != nil {
		return Provenance{}, err
	}
	var p Provenance
	if err := json.Unmarshal(data, &p); err != nil {
		return Provenance{}, err
	}
	return p, nil
}

// WriteProvenance writes provenance metadata for an installed skill.
func (s *FileSystemSkillStore) WriteProvenance(skillName string, p Provenance) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	path := filepath.Join(s.skillbasePath, skillName, ".skillbase-meta")
	return os.WriteFile(path, data, 0o644)
}

// Exists reports whether a skill is installed in the skillbase storage.
func (s *FileSystemSkillStore) Exists(skillName string) bool {
	_, err := os.Stat(filepath.Join(s.skillbasePath, skillName))
	return err == nil
}

func (s *FileSystemSkillStore) resolveTargetAgents(agent string, global bool) ([]string, error) {
	if agent != "" {
		return []string{agent}, nil
	}
	if global {
		return []string{"claude", "agents"}, nil
	}
	agents := s.resolver.DetectAgents(false)
	if len(agents) == 0 {
		return nil, fmt.Errorf("no agent scopes detected; use -a or -g")
	}
	return agents, nil
}


