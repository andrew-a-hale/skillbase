package tui

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"golang.org/x/term"
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
	selectedSkills map[int]bool
	singleMode     bool

	global         bool
	selectedAgents map[string]bool

	preSkill       string
	preAgent       string
	preGlobal      bool
	detectedAgents []string

	loading   bool
	spinner   spinner.Model
	loadCmd   tea.Cmd
	clonePath string

	filter          string
	filteredIndices []int

	width, height int

	Result    *GetResult
	Err       error
	Cancelled bool
}

func NewGetModel(preSkill, preAgent string, preGlobal bool, detectedAgents []string) *GetModel {
	s := spinner.New(spinner.WithSpinner(spinner.Ellipsis))
	s.Style = TitleStyle
	w, h, _ := term.GetSize(int(os.Stdout.Fd()))
	m := &GetModel{
		selectedSkills: make(map[int]bool),
		selectedAgents: make(map[string]bool),
		preSkill:       preSkill,
		preAgent:       preAgent,
		preGlobal:      preGlobal,
		detectedAgents: detectedAgents,
		loading:        true,
		spinner:        s,
		width:          w - 4,
		height:         h,
	}

	m.step = getStepSkill

	return m
}

func (m *GetModel) WithLoadCmd(cmd tea.Cmd) *GetModel {
	m.loadCmd = cmd
	return m
}

func (m *GetModel) buildFilteredIndices() {
	m.filteredIndices = filteredIndices(m.skills, m.filter)
}

func (m *GetModel) applyPreselections() {
	if m.preSkill != "" {
		for i, s := range m.skills {
			if s.Name == m.preSkill {
				m.selectedSkill = i
				m.selectedSkills[i] = true
				m.singleMode = true
				m.step = getStepScope
				break
			}
		}
	}

	if m.step >= getStepScope && m.preGlobal {
		m.global = true
		m.step = getStepAgent
	}

	if m.step >= getStepAgent {
		if m.preAgent != "" {
			m.selectedAgents[m.preAgent] = true
			m.step = getStepConfirm
		} else if m.global {
			m.selectedAgents["claude"] = true
			m.selectedAgents["agents"] = true
			m.step = getStepConfirm
		} else if !m.global && len(m.detectedAgents) == 1 {
			m.selectedAgents[m.detectedAgents[0]] = true
			m.step = getStepConfirm
		}
	}

	m.buildFilteredIndices()
}

func (m *GetModel) Init() tea.Cmd {
	if m.loading && m.loadCmd != nil {
		return tea.Batch(func() tea.Msg { return m.spinner.Tick() }, m.loadCmd)
	}
	return nil
}

func (m *GetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.loading {
		switch msg := msg.(type) {
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		case LoadMsg:
			if msg.Err != nil {
				m.Err = msg.Err
			} else {
				m.skills = msg.Skills
				m.clonePath = msg.ClonePath
				m.applyPreselections()
			}
			m.loading = false
			return m, nil
		case tea.KeyPressMsg:
			if IsKey(msg, DefaultKeyMap.Quit) {
				m.Cancelled = true
				return m, tea.Quit
			}
		}
		return m, nil
	}

	if m.Err != nil {
		if _, ok := msg.(tea.KeyPressMsg); ok {
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
	case tea.KeyPressMsg:
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

func (m *GetModel) updateSkillStep(msg tea.KeyPressMsg) {
	filtered := m.filteredIndices
	switch {
	case IsKey(msg, DefaultKeyMap.Down):
		m.down(len(filtered))
	case IsKey(msg, DefaultKeyMap.Up):
		m.up(len(filtered))
	case IsKey(msg, DefaultKeyMap.Select):
		if !m.singleMode && len(filtered) > 0 {
			origIdx := filtered[m.cursor]
			m.selectedSkills[origIdx] = !m.selectedSkills[origIdx]
		}
	case IsKey(msg, DefaultKeyMap.Confirm):
		if m.singleMode {
			if len(filtered) > 0 {
				m.selectedSkill = filtered[m.cursor]
				m.step = getStepScope
				m.reset()
			}
		} else {
			hasSelected := false
			for _, v := range m.selectedSkills {
				if v {
					hasSelected = true
					break
				}
			}
			if !hasSelected && len(filtered) > 0 {
				origIdx := filtered[m.cursor]
				m.selectedSkills[origIdx] = true
				m.selectedSkill = origIdx
			}
			m.step = getStepScope
			m.reset()
		}
	case msg.Code == tea.KeyBackspace:
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.buildFilteredIndices()
			m.reset()
		}
	case len(msg.Text) > 0:
		m.filter += strings.ToLower(msg.Text)
		m.buildFilteredIndices()
		m.reset()
	}
}

func (m *GetModel) updateScopeStep(msg tea.KeyPressMsg) {
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

func (m *GetModel) updateAgentStep(msg tea.KeyPressMsg) {
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

func (m *GetModel) updateConfirmStep(msg tea.KeyPressMsg) tea.Cmd {
	if IsKey(msg, DefaultKeyMap.Confirm) {
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

		if m.singleMode {
			skill := m.skills[m.selectedSkill]
			m.Result = &GetResult{
				SkillName: skill.Name,
				SkillPath: skill.Path,
				Agents:    agents,
				Global:    m.global,
				ClonePath: m.clonePath,
			}
		} else {
			var names, paths []string
			for i, skill := range m.skills {
				if m.selectedSkills[i] {
					names = append(names, skill.Name)
					paths = append(paths, skill.Path)
				}
			}
			m.Result = &GetResult{
				SkillNames: names,
				SkillPaths: paths,
				Agents:     agents,
				Global:     m.global,
				ClonePath:  m.clonePath,
			}
		}
		return tea.Quit
	}
	return nil
}

func (m *GetModel) handleMouse(msg tea.MouseMsg) {
	count := 0
	switch m.step {
	case getStepSkill:
		count = len(m.filteredIndices)
	case getStepScope:
		count = 2
	case getStepAgent:
		count = len(m.detectedAgents)
	}
	m.list.handleMouse(msg, count)
}

func (m *GetModel) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion

	if m.loading {
		var b strings.Builder
		b.WriteString("Pulling Repository" + m.spinner.View())
		v.SetContent(b.String())
		return v
	}

	if m.Err != nil {
		v.SetContent(viewMargin(ErrorStyle.Render(fmt.Sprintf("Error: %v\n\nPress any key to quit.", m.Err))))
		return v
	}

	var b strings.Builder
	b.WriteString(TitleStyle.Render("skillbase get"))
	b.WriteString("\n")

	switch m.step {
	case getStepSkill:
		b.WriteString(SubtitleStyle.Render("Select skills to install"))
		b.WriteString("\n")
		b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Filter: %s_", m.filter)))
		b.WriteString("\n\n")
		for i, origIdx := range m.filteredIndices {
			if i-m.cursor > 5 || i-m.cursor < -5 {
				continue
			}

			skill := m.skills[origIdx]
			desc := skill.Description
			if desc == "" {
				desc = "(no description)"
			}
			if m.singleMode {
				b.WriteString(renderListItem(m.cursor == i, m.width, skill.Name, desc))
			} else {
				checked := "[ ] "
				if m.selectedSkills[origIdx] {
					checked = "[x] "
				}
				b.WriteString(renderListItem(m.cursor == i, m.width, checked+skill.Name, desc))
			}
		}

		if len(m.filteredIndices) == 0 {
			b.WriteString(MutedStyle.Render("No skills match the filter"))
			b.WriteString("\n")
		}

	case getStepScope:
		if m.singleMode {
			b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Selected: %s", m.skills[m.selectedSkill].Name)))
		} else {
			var selectedNames []string
			for i, skill := range m.skills {
				if m.selectedSkills[i] {
					selectedNames = append(selectedNames, skill.Name)
				}
			}
			b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Selected: %s", strings.Join(selectedNames, ", "))))
		}
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
		if m.singleMode {
			b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Selected: %s | Scope: Project", m.skills[m.selectedSkill].Name)))
		} else {
			var selectedNames []string
			for i, skill := range m.skills {
				if m.selectedSkills[i] {
					selectedNames = append(selectedNames, skill.Name)
				}
			}
			b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Selected: %s | Scope: Project", strings.Join(selectedNames, ", "))))
		}
		b.WriteString("\n")
		b.WriteString(SubtitleStyle.Render("Select agents"))
		b.WriteString("\n\n")
		for i, agent := range m.detectedAgents {
			checked := "[ ] "
			if m.selectedAgents[agent] {
				checked = "[x] "
			}
			b.WriteString(renderListItem(m.cursor == i, m.width, checked+agent, ""))
		}

	case getStepConfirm:
		var selectedNames []string
		if m.singleMode {
			selectedNames = []string{m.skills[m.selectedSkill].Name}
		} else {
			for i, skill := range m.skills {
				if m.selectedSkills[i] {
					selectedNames = append(selectedNames, skill.Name)
				}
			}
		}
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
		b.WriteString(ItemStyle.Render(fmt.Sprintf("Skills: %s", strings.Join(selectedNames, ", "))))
		b.WriteString("\n")
		b.WriteString(ItemStyle.Render(fmt.Sprintf("Scope:  %s", scope)))
		b.WriteString("\n")
		b.WriteString(ItemStyle.Render(fmt.Sprintf("Agents: %s", strings.Join(agents, ", "))))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("Press Enter to confirm, h/\u2190 to go back"))
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("\u2193/\u2191 navigate \u2022 space select \u2022 enter/l/\u2192 confirm \u2022 h/\u2190 back \u2022 q/esc quit"))

	v.SetContent(viewMargin(b.String()))
	return v
}
