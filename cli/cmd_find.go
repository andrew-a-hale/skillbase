package cli

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/andrew-a-hale/skillbase/tui"
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

	model := tui.NewFindModel(tuiSkills, filter)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = p.Run()
	return err
}
