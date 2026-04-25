package cli

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrew-a-hale/skills/internal/fsutil"
)

const APP = "skills"

var (
	HOME        = os.Getenv("HOME")
	SKILLS_PATH = filepath.Join(HOME, ".skills")
)

func getDefaultRepo() (string, error) {
	repo := os.Getenv("SKILLS_DEFAULT_REPO")
	if repo == "" {
		return "", fmt.Errorf("SKILLS_DEFAULT_REPO environment variable is required")
	}
	return repo, nil
}

func Dispatch(args []string) error {
	if len(args) < 2 {
		help("")
		return nil
	}

	name := args[1]
	cmdArgs := args[2:]

	switch name {
	case "help":
		message := ""
		if len(cmdArgs) > 0 {
			message = cmdArgs[0]
		}
		help(message)
		return nil
	case "list", "ls":
		fs := flag.NewFlagSet("list", flag.ContinueOnError)
		global := fs.Bool("g", false, "list global skills")
		if err := fs.Parse(cmdArgs); err != nil {
			return err
		}
		return listSkills(*global)
	case "find":
		filter := ""
		if len(cmdArgs) > 0 {
			filter = cmdArgs[0]
		}
		defaultRepo, err := getDefaultRepo()
		if err != nil {
			return err
		}
		repo, err := NewGitRepository(defaultRepo)
		if err != nil {
			return err
		}
		return findSkills(repo, defaultRepo, filter)
	case "get":
		fs := flag.NewFlagSet("get", flag.ContinueOnError)
		agent := fs.String("a", "", "target agent (claude|agents)")
		global := fs.Bool("g", false, "install to global scope")
		flagSkillPath := fs.String("p", "", "skill path within repository")
		if err := fs.Parse(cmdArgs); err != nil {
			return err
		}
		target := ""
		if len(fs.Args()) > 0 {
			target = fs.Args()[0]
		}

		repoURL, err := getDefaultRepo()
		if err != nil {
			return err
		}
		skillPath := ""
		if target != "" {
			if strings.Contains(target, "/") || strings.Contains(target, ":") {
				repoURL, skillPath = parseRepoURL(target)
				if *flagSkillPath != "" {
					skillPath = *flagSkillPath
				}
			} else {
				skillPath = target
			}
		}

		repo, err := NewGitRepository(repoURL)
		if err != nil {
			return err
		}
		resolver, err := NewFileSystemScopeResolver()
		if err != nil {
			return err
		}
		return getSkill(repo, resolver, skillPath, *agent, *global)
	case "remove", "rm":
		fs := flag.NewFlagSet("remove", flag.ContinueOnError)
		agent := fs.String("a", "", "target agent")
		global := fs.Bool("g", false, "remove from global storage")
		if err := fs.Parse(cmdArgs); err != nil {
			return err
		}
		if len(fs.Args()) < 1 {
			return fmt.Errorf("expected skill name")
		}
		resolver, err := NewFileSystemScopeResolver()
		if err != nil {
			return err
		}
		return removeSkill(resolver, fs.Args()[0], *agent, *global)
	case "update":
		if len(cmdArgs) < 1 {
			return fmt.Errorf("expected skill name")
		}
		defaultRepo, err := getDefaultRepo()
		if err != nil {
			return err
		}
		repo, err := NewGitRepository(defaultRepo)
		if err != nil {
			return err
		}
		resolver, err := NewFileSystemScopeResolver()
		if err != nil {
			return err
		}
		return updateSkill(repo, resolver, cmdArgs[0])
	default:
		return fmt.Errorf("unknown command: %s", name)
	}
}

func help(message string) {
	if message != "" {
		fmt.Printf("%s\n\n", message)
	}
	fmt.Printf(`%s - Manage agent skills

Commands:
  help              Print this help message
  list, ls          List installed skills
    -g              List global skills
  find [filter]     Find available skills in repository
  get [skill|url]   Download skill(s) and link to agent scope
    -p path         Skill path within repository
    -a agent        Target agent (claude|agents)
    -g              Install to global agent scope
  rm, remove <name> Remove skill
    -a agent        Remove from specific agent only
    -g              Remove from global storage
  update <name>     Update existing skill

Default repository: set via SKILLS_DEFAULT_REPO environment variable
`, APP)
}

func listSkills(global bool) error {
	var path string
	if global {
		path = SKILLS_PATH
	} else {
		path = "."
	}

	var found bool
	err := fs.WalkDir(os.DirFS(path), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasPrefix(path, ".") && !global {
			return filepath.SkipDir
		}

		if d.Name() == "SKILL.md" {
			parts := strings.Split(path, string(os.PathSeparator))
			skillName, _ := strings.CutSuffix(parts[len(parts)-2], string(os.PathSeparator))
			fmt.Printf("  %s: %s\n", skillName, path)
			found = true
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("%v: %v", ErrSkillNotFound, err)
	}

	if !found {
		fmt.Println("No skills installed")
	}
	return nil
}

func findSkills(repo Repository, repoURL string, filter string) error {
	ctx := context.Background()

	fmt.Printf("Searching repository: %s\n\n", repoURL)

	clonePath, cleanup, err := repo.Clone(ctx)
	if err != nil {
		return fmt.Errorf("failed to clone: %v", err)
	}
	defer func() { _ = cleanup() }()

	skills, err := repo.ListSkills(clonePath)
	if err != nil {
		return fmt.Errorf("failed to list skills: %v", err)
	}

	if len(skills) == 0 {
		if filter != "" {
			fmt.Printf("No skills matching %q found\n", filter)
		} else {
			fmt.Println("No skills found")
		}
		return nil
	}

	for _, skill := range skills {
		if filter != "" && !strings.Contains(strings.ToLower(skill.Name), strings.ToLower(filter)) {
			continue
		}
		desc := skill.Description
		if desc == "" {
			desc = "(no description)"
		}
		fmt.Printf("  %-20s %s\n", skill.Name, desc)
	}
	return nil
}

func installSkill(srcDir, skillName string) error {
	if err := os.MkdirAll(SKILLS_PATH, 0o755); err != nil {
		return err
	}
	dstDir := filepath.Join(SKILLS_PATH, skillName)
	if err := os.RemoveAll(dstDir); err != nil {
		return err
	}
	return fsutil.CopyDir(srcDir, dstDir)
}

func linkSkill(resolver ScopeResolver, skillName string, agents []string, global bool) error {
	storageDir := filepath.Join(SKILLS_PATH, skillName)

	for _, agent := range agents {
		targets, err := resolver.Resolve(skillName, global, agent)
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
				return fmt.Errorf("failed to link %s: %v", target.Agent, err)
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

func getSkill(repo Repository, resolver ScopeResolver, skillPath, agent string, global bool) error {
	ctx := context.Background()

	clonePath, cleanup, err := repo.Clone(ctx)
	if err != nil {
		return fmt.Errorf("failed to clone: %v", err)
	}
	defer func() { _ = cleanup() }()

	skills, err := fetchSkillsFromRepo(repo, clonePath, skillPath)
	if err != nil {
		return err
	}

	skillNames := installFetchedSkills(skills, clonePath)

	agents, err := resolveTargetAgents(resolver, agent, global)
	if err != nil {
		return err
	}

	for _, skillName := range skillNames {
		if err := linkSkill(resolver, skillName, agents, global); err != nil {
			log.Printf("Warning: failed to link %s: %v", skillName, err)
		}
	}
	return nil
}

func fetchSkillsFromRepo(repo Repository, clonePath, skillPath string) ([]Skill, error) {
	if skillPath == "" {
		skills, err := repo.ListSkills(clonePath)
		if err != nil {
			return nil, fmt.Errorf("failed to list skills: %v", err)
		}
		return skills, nil
	}

	skill, err := repo.GetSkill(clonePath, skillPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get skill: %v", err)
	}
	return []Skill{skill}, nil
}

func installFetchedSkills(skills []Skill, clonePath string) []string {
	var skillNames []string
	for _, skill := range skills {
		skillDir := filepath.Join(clonePath, skill.Path)
		if err := installSkill(skillDir, skill.Name); err != nil {
			log.Printf("Failed to install %s: %v", skill.Name, err)
			continue
		}
		skillNames = append(skillNames, skill.Name)
	}
	if len(skillNames) == 1 {
		fmt.Printf("Installed skill: %s\n", skillNames[0])
	} else {
		fmt.Printf("Installed %d skills\n", len(skillNames))
	}
	return skillNames
}

func resolveTargetAgents(resolver ScopeResolver, agent string, global bool) ([]string, error) {
	if agent != "" {
		return []string{agent}, nil
	}
	if global {
		return []string{"claude", "agents"}, nil
	}
	agents := resolver.DetectAgents(false)
	if len(agents) == 0 {
		return nil, fmt.Errorf("no agent scopes detected; use -a or -g")
	}
	return agents, nil
}

func removeSkill(resolver ScopeResolver, skillName, agent string, global bool) error {
	if global {
		storageDir := filepath.Join(SKILLS_PATH, skillName)
		if _, err := os.Stat(storageDir); os.IsNotExist(err) {
			fmt.Printf("Skill %q not found\n", skillName)
			return nil
		}

		for _, ag := range []string{"claude", "agents"} {
			targets, _ := resolver.Resolve(skillName, true, ag)
			for _, target := range targets {
				linkPath := filepath.Join(target.Path, skillName)
				_ = os.Remove(linkPath)
			}
		}

		_ = os.RemoveAll(storageDir)
		fmt.Printf("Removed %s\n", skillName)
	} else {
		agents := resolver.DetectAgents(false)
		if agent != "" {
			agents = []string{agent}
		}

		removed := false
		for _, ag := range agents {
			targets, _ := resolver.Resolve(skillName, false, ag)
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

func updateSkill(repo Repository, resolver ScopeResolver, skillName string) error {
	storageDir := filepath.Join(SKILLS_PATH, skillName)
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		return fmt.Errorf("skill %q not installed", skillName)
	}
	return getSkill(repo, resolver, skillName, "", false)
}
