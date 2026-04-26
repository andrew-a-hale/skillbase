package tui

type SkillInfo struct {
	Name        string
	Path        string
	Description string
	Scope       string
	Agents      []string
}

type GetResult struct {
	SkillName string
	SkillPath string
	Agents    []string
	Global    bool
}

type RemoveResult struct {
	SkillNames []string
	Agent      string
	Global     bool
}

type UpdateResult struct {
	SkillName string
}
