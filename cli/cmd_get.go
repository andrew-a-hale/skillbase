package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/andrew-a-hale/skillbase/tui"
)

func cmdGet(args []string) error {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	agent := fs.String("a", "", "target agent (claude|agents)")
	global := fs.Bool("g", false, "install to global scope")
	flagSkillPath := fs.String("p", "", "skill path within repository")
	if err := fs.Parse(args); err != nil {
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
	home := os.Getenv("HOME")
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	resolver, err := NewFileSystemScopeResolver(home, cwd)
	if err != nil {
		return err
	}
	store := NewFileSystemSkillStore(resolver, skillbasePath)

	detectedAgents := resolver.DetectAgents(false)
	isComplete := target != "" && (*global || *agent != "" || len(detectedAgents) > 0)

	if isComplete {
		return getSkill(repo, store, repoURL, skillPath, []string{*agent}, *global)
	}

	ctx := context.Background()
	clonePath, cleanup, err := repo.Clone(ctx)
	if err != nil {
		return fmt.Errorf("failed to clone: %w", err)
	}
	defer func() { _ = cleanup() }()

	skills, err := repo.ListSkills(clonePath)
	if err != nil {
		return fmt.Errorf("failed to list skills: %w", err)
	}

	var tuiSkills []tui.SkillInfo
	for _, s := range skills {
		tuiSkills = append(tuiSkills, tui.SkillInfo{
			Name:        s.Name,
			Path:        s.Path,
			Description: s.Description,
		})
	}

	preSkill := ""
	if skillPath != "" {
		preSkill = filepath.Base(skillPath)
	}

	model := tui.NewGetModel(tuiSkills, preSkill, *agent, *global, detectedAgents)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	m, ok := finalModel.(*tui.GetModel)
	if !ok || m.Cancelled {
		return nil
	}
	if m.Err != nil {
		return m.Err
	}
	if m.Result == nil {
		return nil
	}

	return doGetSkill(repo, store, clonePath, repoURL, m.Result.SkillPath, m.Result.Agents, m.Result.Global)
}
