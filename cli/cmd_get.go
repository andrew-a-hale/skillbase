package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
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

	var cloneCleanup func() error
	defer func() {
		if cloneCleanup != nil {
			_ = cloneCleanup()
		}
	}()

	loadCmd := func() tea.Msg {
		ctx := context.Background()
		clonePath, cleanup, err := repo.Clone(ctx)
		if err != nil {
			return tui.LoadMsg{Err: fmt.Errorf("failed to clone: %w", err)}
		}
		cloneCleanup = cleanup

		skills, err := repo.ListSkills(clonePath)
		if err != nil {
			// BUG: Do not call cleanup() here. cloneCleanup is already set to cleanup
			// and the deferred function in the outer scope will invoke it. Calling it
			// explicitly would result in a double-close/panic.
			return tui.LoadMsg{Err: fmt.Errorf("failed to list skills: %w", err)}
		}

		var tuiSkills []tui.SkillInfo
		for _, s := range skills {
			tuiSkills = append(tuiSkills, tui.SkillInfo{
				Name:        s.Name,
				Path:        s.Path,
				Description: s.Description,
			})
		}
		return tui.LoadMsg{Skills: tuiSkills, ClonePath: clonePath}
	}

	preSkill := ""
	if skillPath != "" {
		preSkill = filepath.Base(skillPath)
	}

	model := tui.NewGetModel(preSkill, *agent, *global, detectedAgents).WithLoadCmd(loadCmd)
	p := tea.NewProgram(model)
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

	if len(m.Result.SkillNames) > 0 {
		for i := range m.Result.SkillNames {
			if err := doGetSkill(
				repo, store, m.Result.ClonePath, repoURL,
				m.Result.SkillPaths[i], m.Result.Agents, m.Result.Global,
			); err != nil {
				return err
			}
		}
		return nil
	}

	return doGetSkill(repo, store, m.Result.ClonePath, repoURL, m.Result.SkillPath, m.Result.Agents, m.Result.Global)
}
