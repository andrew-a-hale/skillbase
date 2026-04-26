package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/andrew-a-hale/skillbase/tui"
)

func cmdUpdate(args []string) error {
	defaultRepo, err := getDefaultRepo()
	if err != nil {
		return err
	}
	repo, err := NewGitRepository(defaultRepo)
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

	var skillName string
	if len(args) > 0 {
		skillName = args[0]
	} else {
		_, globalSkills, err := store.ListInstalled()
		if err != nil {
			return err
		}

		model := tui.NewUpdateModel(toTuiSkills(globalSkills, "global"))
		p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
		finalModel, err := p.Run()
		if err != nil {
			return err
		}

		m, ok := finalModel.(*tui.UpdateModel)
		if !ok || m.Cancelled {
			return nil
		}
		if m.Err != nil {
			return m.Err
		}
		if m.Result == nil {
			return nil
		}
		skillName = m.Result.SkillName
	}

	if skillName == "" {
		return nil
	}

	if !store.Exists(skillName) {
		return fmt.Errorf("skill %q not installed", skillName)
	}

	prov, err := store.ReadProvenance(skillName)
	if err == nil && prov.Source != "" {
		origRepo, err := NewGitRepository(prov.Source)
		if err != nil {
			return err
		}
		return updateSkill(origRepo, store, prov.Source, prov.SkillPath)
	}

	return updateSkill(repo, store, defaultRepo, skillName)
}
