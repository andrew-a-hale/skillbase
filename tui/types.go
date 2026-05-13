package tui

type SkillInfo struct {
	Name        string
	Path        string
	Description string
	Scope       string
	Agents      []string
}

type GetResult struct {
	SkillName  string
	SkillPath  string
	SkillNames []string
	SkillPaths []string
	Agents     []string
	Global     bool
	ClonePath  string
}

type RemoveResult struct {
	SkillNames []string
	Agent      string
	Global     bool
}

type UpdateResult struct {
	SkillName string
}

type LoadMsg struct {
	Skills    []SkillInfo
	ClonePath string
	Err       error
}
