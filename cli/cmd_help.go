package cli

import "fmt"

func cmdHelp(args []string) error {
	message := ""
	if len(args) > 0 {
		message = args[0]
	}
	if message != "" {
		fmt.Printf("%s\n\n", message)
	}
	fmt.Printf(`skillbase - Manage agent skills

Commands:
  help              Print this help message
  list, ls          List installed skills (interactive)
  find [filter]     Find available skills in repository (interactive)
  get [skill|url]   Download skill(s) and link to agent scope
    -p path         Skill path within repository
    -a agent        Target agent (claude|agents)
    -g              Install to global agent scope
  rm, remove <name> Remove skill
    -a agent        Remove from specific agent only
    -g              Remove from global storage
  update <name>     Update existing skill

Default repository: set via SKILLBASE_DEFAULT_REPO environment variable
`)
	return nil
}
