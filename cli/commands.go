package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andrew-a-hale/skillbase/tui"
)

var (
	homeDir       = os.Getenv("HOME")
	skillbasePath = filepath.Join(homeDir, ".skillbase")
)

func getDefaultRepo() (string, error) {
	repo := os.Getenv("SKILLBASE_DEFAULT_REPO")
	if repo == "" {
		return "", fmt.Errorf("SKILLBASE_DEFAULT_REPO environment variable is required")
	}
	return repo, nil
}

func toTuiSkills(skills []InstalledSkill, scope string) []tui.SkillInfo {
	var result []tui.SkillInfo
	for _, s := range skills {
		result = append(result, tui.SkillInfo{
			Name:        s.Name,
			Description: s.Description,
			Scope:       scope,
			Agents:      s.Agents,
		})
	}
	return result
}

// Command is a subcommand handler.
type Command func([]string) error

var commands = map[string]Command{
	"help":   cmdHelp,
	"list":   cmdList,
	"ls":     cmdList,
	"find":   cmdFind,
	"get":    cmdGet,
	"remove": cmdRemove,
	"rm":     cmdRemove,
	"update": cmdUpdate,
}

func Dispatch(args []string) error {
	if len(args) < 2 {
		return cmdHelp(nil)
	}

	name := args[1]
	cmdArgs := args[2:]

	cmd, ok := commands[name]
	if !ok {
		return fmt.Errorf("unknown command: %s", name)
	}

	return cmd(cmdArgs)
}

func parseRepoURL(input string) (repoURL, skillPath string) {
	input = strings.TrimSpace(input)
	input = strings.TrimSuffix(input, ".git")
	parts := strings.Split(input, "/")

	if len(parts) < 5 {
		return input, ""
	}

	if len(parts) == 5 {
		return input, ""
	}

	repoParts := parts[:5]
	skillParts := parts[5:]
	return strings.Join(repoParts, "/"), strings.Join(skillParts, "/")
}

func doGetSkill(repo Repository, store SkillStore, clonePath, source, skillPath string, agents []string, global bool) error {
	skills, err := fetchSkillsFromRepo(repo, clonePath, skillPath)
	if err != nil {
		return err
	}

	skillNames, installedSkills := installFetchedSkills(skills, clonePath, store)

	for _, skill := range installedSkills {
		prov := Provenance{
			Source:      source,
			SkillPath:   skill.Path,
			InstalledAt: time.Now().UTC().Format(time.RFC3339),
		}
		if err := store.WriteProvenance(skill.Name, prov); err != nil {
			log.Printf("Warning: failed to write meta for %s: %v", skill.Name, err)
		}
	}

	for _, skillName := range skillNames {
		for _, agent := range agents {
			if err := store.Link(skillName, agent, global); err != nil {
				log.Printf("Warning: failed to link %s: %v", skillName, err)
			}
		}
	}
	return nil
}

func getSkill(repo Repository, store SkillStore, source, skillPath string, agents []string, global bool) error {
	ctx := context.Background()

	clonePath, cleanup, err := repo.Clone(ctx)
	if err != nil {
		return fmt.Errorf("failed to clone: %w", err)
	}
	defer func() { _ = cleanup() }()

	return doGetSkill(repo, store, clonePath, source, skillPath, agents, global)
}

func fetchSkillsFromRepo(repo Repository, clonePath, skillPath string) ([]Skill, error) {
	if skillPath == "" {
		skills, err := repo.ListSkills(clonePath)
		if err != nil {
			return nil, fmt.Errorf("failed to list skills: %w", err)
		}
		return skills, nil
	}

	skill, err := repo.GetSkill(clonePath, skillPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get skill: %w", err)
	}
	return []Skill{skill}, nil
}

func installFetchedSkills(skills []Skill, clonePath string, store SkillStore) ([]string, []Skill) {
	skillNames := make([]string, 0, len(skills))
	installed := make([]Skill, 0, len(skills))
	for _, skill := range skills {
		skillDir := filepath.Join(clonePath, skill.Path)
		if err := store.Install(skill.Name, skillDir); err != nil {
			log.Printf("Failed to install %s: %v", skill.Name, err)
			continue
		}
		skillNames = append(skillNames, skill.Name)
		installed = append(installed, skill)
	}
	if len(skillNames) == 1 {
		fmt.Printf("Installed skill: %s\n", skillNames[0])
	} else {
		fmt.Printf("Installed %d skills\n", len(skillNames))
	}
	return skillNames, installed
}

func removeSkills(store SkillStore, skillNames []string, agent string, global bool) error {
	for _, name := range skillNames {
		if err := store.Remove(name, agent, global); err != nil {
			log.Printf("Warning: failed to remove %s: %v", name, err)
		}
	}
	return nil
}

func updateSkill(repo Repository, store SkillStore, source, skillPath string) error {
	skillName := filepath.Base(skillPath)
	if !store.Exists(skillName) {
		return fmt.Errorf("skill %q not installed", skillName)
	}
	return getSkill(repo, store, source, skillPath, []string{""}, false)
}
