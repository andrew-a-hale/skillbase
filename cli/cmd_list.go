package cli

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/andrew-a-hale/skillbase/tui"
)

func cmdList(args []string) error {
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
	project, global, err := store.ListInstalled()
	if err != nil {
		return err
	}

	model := tui.NewListModel(toTuiSkills(project, "project"), toTuiSkills(global, "global"))
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = p.Run()
	return err
}
