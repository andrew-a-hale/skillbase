package cli

import (
	"context"
	"fmt"

	"github.com/andrew-a-hale/skillbase/tui"
	tea "charm.land/bubbletea/v2"
)

func cmdFind(args []string) error {
	filter := ""
	if len(args) > 0 {
		filter = args[0]
	}
	defaultRepo, err := getDefaultRepo()
	if err != nil {
		return err
	}
	repo, err := NewGitRepository(defaultRepo)
	if err != nil {
		return err
	}

	loadCmd := func() tea.Msg {
		ctx := context.Background()
		clonePath, cleanup, err := repo.Clone(ctx)
		if err != nil {
			return tui.LoadMsg{Err: fmt.Errorf("failed to clone: %w", err)}
		}
		defer func() { _ = cleanup() }()

		skills, err := repo.ListSkills(clonePath)
		if err != nil {
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
		return tui.LoadMsg{Skills: tuiSkills}
	}

	model := tui.NewFindModel(filter).WithLoadCmd(loadCmd)
	p := tea.NewProgram(model)
	_, err = p.Run()
	return err
}
