package cli

import (
	"flag"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/andrew-a-hale/skillbase/tui"
)

func cmdRemove(args []string) error {
	fs := flag.NewFlagSet("remove", flag.ContinueOnError)
	agent := fs.String("a", "", "target agent")
	global := fs.Bool("g", false, "remove from global storage")
	if err := fs.Parse(args); err != nil {
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

	if len(fs.Args()) > 0 {
		return store.Remove(fs.Args()[0], *agent, *global)
	}

	project, globalSkills, err := store.ListInstalled()
	if err != nil {
		return err
	}

	var globalNames []string
	for _, g := range globalSkills {
		globalNames = append(globalNames, g.Name)
	}

	model := tui.NewRemoveModel(globalNames, toTuiSkills(project, "project"), *global, *agent)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	m, ok := finalModel.(*tui.RemoveModel)
	if !ok || m.Cancelled {
		return nil
	}
	if m.Err != nil {
		return m.Err
	}
	if m.Result == nil {
		return nil
	}

	return removeSkills(store, m.Result.SkillNames, m.Result.Agent, m.Result.Global)
}
