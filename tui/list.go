package tui

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"golang.org/x/term"
)

type ListModel struct {
	list

	projectSkills         []SkillInfo
	globalSkills          []SkillInfo
	filter                string
	filteredProjectSkills []SkillInfo
	filteredGlobalSkills  []SkillInfo

	width, height int

	Err       error
	Cancelled bool
}

func NewListModel(projectSkills, globalSkills []SkillInfo) *ListModel {
	w, h, _ := term.GetSize(int(os.Stdout.Fd()))
	m := &ListModel{
		projectSkills: projectSkills,
		globalSkills:  globalSkills,
		width:         w - 4,
		height:        h,
	}
	m.buildFiltered()
	return m
}

func (m *ListModel) buildFiltered() {
	m.filteredProjectSkills = nil
	m.filteredGlobalSkills = nil
	if m.filter == "" {
		m.filteredProjectSkills = append([]SkillInfo(nil), m.projectSkills...)
		m.filteredGlobalSkills = append([]SkillInfo(nil), m.globalSkills...)
		return
	}
	f := strings.ToLower(m.filter)
	for _, s := range m.projectSkills {
		if strings.Contains(strings.ToLower(s.Name), f) ||
			strings.Contains(strings.ToLower(s.Path), f) ||
			strings.Contains(strings.ToLower(s.Description), f) {
			m.filteredProjectSkills = append(m.filteredProjectSkills, s)
		}
	}
	for _, s := range m.globalSkills {
		if strings.Contains(strings.ToLower(s.Name), f) ||
			strings.Contains(strings.ToLower(s.Path), f) ||
			strings.Contains(strings.ToLower(s.Description), f) {
			m.filteredGlobalSkills = append(m.filteredGlobalSkills, s)
		}
	}
}

func (m *ListModel) allSkills() []SkillInfo {
	return append(m.projectSkills, m.globalSkills...)
}

func (m *ListModel) Init() tea.Cmd { return nil }

func (m *ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.Err != nil {
		if _, ok := msg.(tea.KeyMsg); ok {
			return m, tea.Quit
		}
		return m, nil
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
		allCount := len(m.filteredProjectSkills) + len(m.filteredGlobalSkills)
		switch {
		case IsKey(msg, DefaultKeyMap.Down):
			m.down(allCount)
		case IsKey(msg, DefaultKeyMap.Up):
			m.up(allCount)
		case msg.Code == tea.KeyBackspace:
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.buildFiltered()
				m.reset()
			}
		case len(msg.Text) > 0:
			m.filter += strings.ToLower(msg.Text)
			m.buildFiltered()
			m.reset()
		}
	case tea.MouseMsg:
		m.handleMouse(msg, len(m.filteredProjectSkills)+len(m.filteredGlobalSkills))
	}
	return m, nil
}

func (m *ListModel) View() tea.View {
	v := tea.NewView("")
	if m.Err != nil {
		v.SetContent(ErrorStyle.Render(fmt.Sprintf("Error: %v\n\nPress any key to quit.", m.Err)))
		return v
	}

	var b strings.Builder
	b.WriteString(TitleStyle.Render("skillbase list"))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Filter: %s_", m.filter)))
	b.WriteString("\n\n")

	cursor := 0
	if len(m.filteredProjectSkills) > 0 {
		b.WriteString(SubtitleStyle.Render("Project Scope"))
		b.WriteString("\n")
		for _, skill := range m.filteredProjectSkills {
			if cursor-m.cursor > 5 || cursor-m.cursor < -5 {
				cursor++
				continue
			}
			selected := m.cursor == cursor
			agents := ""
			if len(skill.Agents) > 0 {
				agents = fmt.Sprintf(" [%s]", strings.Join(skill.Agents, ", "))
			}
			desc := skill.Description
			if desc == "" {
				desc = "(no description)"
			}
			b.WriteString(renderListItem(selected, m.width, skill.Name+MutedStyle.Render(agents), desc))
			cursor++
		}
		b.WriteString("\n")
	}

	if len(m.filteredGlobalSkills) > 0 {
		b.WriteString(SubtitleStyle.Render("Global Scope"))
		b.WriteString("\n")
		for _, skill := range m.filteredGlobalSkills {
			if cursor-m.cursor > 5 || cursor-m.cursor < -5 {
				cursor++
				continue
			}

			selected := m.cursor == cursor
			desc := skill.Description
			if desc == "" {
				desc = "(no description)"
			}
			b.WriteString(renderListItem(selected, m.width, skill.Name, desc))
			cursor++
		}
	}

	if len(m.projectSkills) == 0 && len(m.globalSkills) == 0 {
		b.WriteString(MutedStyle.Render("No skills installed"))
		b.WriteString("\n")
	} else if len(m.filteredProjectSkills) == 0 && len(m.filteredGlobalSkills) == 0 {
		b.WriteString(MutedStyle.Render("No skills match the filter"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("j/\u2193 k/\u2191 navigate \u2022 q/esc quit"))

	v.SetContent(viewMargin(b.String()))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}
