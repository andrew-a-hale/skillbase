package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type getStep int

const (
	getStepSkill getStep = iota
	getStepScope
	getStepAgent
	getStepConfirm
)

type GetModel struct {
	list

	step getStep

	skills         []SkillInfo
	selectedSkill  int
	global         bool
	selectedAgents map[string]bool

	preSkill       string
	preAgent       string
	preGlobal      bool
	detectedAgents []string

	width, height int

	Result    *GetResult
	Err       error
	Cancelled bool
}

func NewGetModel(skills []SkillInfo, preSkill, preAgent string, preGlobal bool, detectedAgents []string) *GetModel {
	m := &GetModel{
		skills:         skills,
		selectedAgents: make(map[string]bool),
		preSkill:       preSkill,
		preAgent:       preAgent,
		preGlobal:      preGlobal,
		detectedAgents: detectedAgents,
	}

	m.step = getStepSkill

	if preSkill != "" {
		for i, s := range skills {
			if s.Name == preSkill {
				m.selectedSkill = i
				m.step = getStepScope
				break
			}
		}
	}

	if m.step >= getStepScope && preGlobal {
		m.global = true
		m.step = getStepAgent
	}

	if m.step >= getStepAgent {
		if preAgent != "" {
			m.selectedAgents[preAgent] = true
			m.step = getStepConfirm
		} else if m.global {
			m.selectedAgents["claude"] = true
			m.selectedAgents["agents"] = true
			m.step = getStepConfirm
		} else if !m.global && len(detectedAgents) == 1 {
			m.selectedAgents[detectedAgents[0]] = true
			m.step = getStepConfirm
		}
	}

	return m
}

func (m *GetModel) Init() tea.Cmd { return nil }

func (m *GetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.Err != nil {
		if _, ok := msg.(tea.KeyMsg); ok {
			return m, tea.Quit
		}
		return m, nil
	}

	if m.Result != nil {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if IsKey(msg, DefaultKeyMap.Quit) {
			m.Cancelled = true
			return m, tea.Quit
		}
		if IsKey(msg, DefaultKeyMap.Back) && m.step > getStepSkill {
			m.step--
			return m, nil
		}
		switch m.step {
		case getStepSkill:
			m.updateSkillStep(msg)
		case getStepScope:
			m.updateScopeStep(msg)
		case getStepAgent:
			m.updateAgentStep(msg)
		case getStepConfirm:
			if cmd := m.updateConfirmStep(msg); cmd != nil {
				return m, cmd
			}
		}
	case tea.MouseMsg:
		m.handleMouse(msg)
	}
	return m, nil
}

func (m *GetModel) updateSkillStep(msg tea.KeyMsg) {
	switch {
	case IsKey(msg, DefaultKeyMap.Down):
		m.down(len(m.skills))
	case IsKey(msg, DefaultKeyMap.Up):
		m.up(len(m.skills))
	case IsKey(msg, DefaultKeyMap.Confirm):
		if len(m.skills) > 0 {
			m.selectedSkill = m.cursor
			m.step = getStepScope
			m.reset()
		}
	}
}

func (m *GetModel) updateScopeStep(msg tea.KeyMsg) {
	switch {
	case IsKey(msg, DefaultKeyMap.Up), IsKey(msg, DefaultKeyMap.Down):
		m.global = !m.global
	case IsKey(msg, DefaultKeyMap.Confirm):
		if m.global {
			m.selectedAgents = map[string]bool{"claude": true, "agents": true}
			m.step = getStepConfirm
		} else {
			if len(m.detectedAgents) == 0 {
				m.Err = fmt.Errorf("no agent scopes detected in project; use -g or -a")
				return
			}
			if len(m.detectedAgents) == 1 {
				m.selectedAgents[m.detectedAgents[0]] = true
				m.step = getStepConfirm
			} else {
				if m.preAgent != "" {
					m.selectedAgents[m.preAgent] = true
				}
				m.step = getStepAgent
				m.reset()
			}
		}
	}
}

func (m *GetModel) updateAgentStep(msg tea.KeyMsg) {
	switch {
	case IsKey(msg, DefaultKeyMap.Down):
		m.down(len(m.detectedAgents))
	case IsKey(msg, DefaultKeyMap.Up):
		m.up(len(m.detectedAgents))
	case IsKey(msg, DefaultKeyMap.Select):
		agent := m.detectedAgents[m.cursor]
		m.selectedAgents[agent] = !m.selectedAgents[agent]
	case IsKey(msg, DefaultKeyMap.Confirm):
		hasSelected := false
		for _, v := range m.selectedAgents {
			if v {
				hasSelected = true
				break
			}
		}
		if hasSelected {
			m.step = getStepConfirm
		}
	}
}

func (m *GetModel) updateConfirmStep(msg tea.KeyMsg) tea.Cmd {
	if IsKey(msg, DefaultKeyMap.Confirm) {
		skill := m.skills[m.selectedSkill]
		var agents []string
		if m.preAgent != "" {
			agents = []string{m.preAgent}
		} else if m.global {
			agents = []string{""}
		} else if len(m.detectedAgents) == 1 {
			agents = []string{m.detectedAgents[0]}
		} else {
			for _, agent := range m.detectedAgents {
				if m.selectedAgents[agent] {
					agents = append(agents, agent)
				}
			}
		}
		m.Result = &GetResult{
			SkillName: skill.Name,
			SkillPath: skill.Path,
			Agents:    agents,
			Global:    m.global,
		}
		return tea.Quit
	}
	return nil
}

func (m *GetModel) handleMouse(msg tea.MouseMsg) {
	count := 0
	switch m.step {
	case getStepSkill:
		count = len(m.skills)
	case getStepScope:
		count = 2
	case getStepAgent:
		count = len(m.detectedAgents)
	}
	m.list.handleMouse(msg, count)
}

func (m *GetModel) View() string {
	if m.Err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v\n\nPress any key to quit.", m.Err))
	}

	var b strings.Builder
	b.WriteString(TitleStyle.Render("skillbase get"))
	b.WriteString("\n")

	switch m.step {
	case getStepSkill:
		b.WriteString(SubtitleStyle.Render("Select a skill to install"))
		b.WriteString("\n\n")
		for i, skill := range m.skills {
			desc := skill.Description
			if desc == "" {
				desc = "(no description)"
			}
			b.WriteString(itemLine(m.cursor == i, fmt.Sprintf("%s  %s", skill.Name, MutedStyle.Render(desc))))
			b.WriteString("\n")
		}

	case getStepScope:
		b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Selected: %s", m.skills[m.selectedSkill].Name)))
		b.WriteString("\n")
		b.WriteString(SubtitleStyle.Render("Select scope"))
		b.WriteString("\n\n")
		options := []struct {
			label string
			value bool
		}{
			{"Project", false},
			{"Global", true},
		}
		for _, opt := range options {
			cursor := " "
			style := ItemStyle
			if m.global == opt.value {
				cursor = ">"
				style = SelectedItemStyle
			}
			b.WriteString(style.Render(fmt.Sprintf("%s %s", cursor, opt.label)))
			b.WriteString("\n")
		}

	case getStepAgent:
		b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Selected: %s | Scope: Project", m.skills[m.selectedSkill].Name)))
		b.WriteString("\n")
		b.WriteString(SubtitleStyle.Render("Select agents"))
		b.WriteString("\n\n")
		for i, agent := range m.detectedAgents {
			checked := "[ ] "
			if m.selectedAgents[agent] {
				checked = "[x] "
			}
			b.WriteString(itemLine(m.cursor == i, fmt.Sprintf("%s%s", checked, agent)))
			b.WriteString("\n")
		}

	case getStepConfirm:
		skill := m.skills[m.selectedSkill]
		agents := []string{}
		for a, selected := range m.selectedAgents {
			if selected {
				agents = append(agents, a)
			}
		}
		scope := "project"
		if m.global {
			scope = "global"
		}
		b.WriteString(SubtitleStyle.Render("Confirm installation"))
		b.WriteString("\n\n")
		b.WriteString(ItemStyle.Render(fmt.Sprintf("Skill:  %s", skill.Name)))
		b.WriteString("\n")
		b.WriteString(ItemStyle.Render(fmt.Sprintf("Scope:  %s", scope)))
		b.WriteString("\n")
		b.WriteString(ItemStyle.Render(fmt.Sprintf("Agents: %s", strings.Join(agents, ", "))))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("Press Enter to confirm, h/\u2190 to go back"))
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("j/\u2193 k/\u2191 navigate \u2022 space select \u2022 enter/l/\u2192 confirm \u2022 h/\u2190 back \u2022 q/esc quit"))

	return b.String()
}
