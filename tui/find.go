package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type FindModel struct {
	list

	skills []SkillInfo
	filter string

	width, height int

	Err       error
	Cancelled bool
}

func NewFindModel(skills []SkillInfo, filter string) *FindModel {
	return &FindModel{
		skills: skills,
		filter: strings.ToLower(filter),
	}
}

func (m *FindModel) filteredSkills() []SkillInfo {
	if m.filter == "" {
		return m.skills
	}
	var filtered []SkillInfo
	for _, s := range m.skills {
		if strings.Contains(strings.ToLower(s.Name), m.filter) ||
			strings.Contains(strings.ToLower(s.Description), m.filter) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func (m *FindModel) Init() tea.Cmd { return nil }

func (m *FindModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	case tea.KeyMsg:
		if IsKey(msg, DefaultKeyMap.Quit) {
			m.Cancelled = true
			return m, tea.Quit
		}
		filtered := m.filteredSkills()
		switch {
		case IsKey(msg, DefaultKeyMap.Down):
			m.down(len(filtered))
		case IsKey(msg, DefaultKeyMap.Up):
			m.up(len(filtered))
		case msg.Type == tea.KeyBackspace:
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.reset()
			}
		case msg.Type == tea.KeyRunes:
			m.filter += strings.ToLower(string(msg.Runes))
			m.reset()
		}
	case tea.MouseMsg:
		m.handleMouse(msg, len(m.filteredSkills()))
	}
	return m, nil
}

func (m *FindModel) View() string {
	if m.Err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v\n\nPress any key to quit.", m.Err))
	}

	filtered := m.filteredSkills()
	var b strings.Builder
	b.WriteString(TitleStyle.Render("skillbase find"))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Filter: %s_", m.filter)))
	b.WriteString("\n\n")

	for i, skill := range filtered {
		desc := skill.Description
		if desc == "" {
			desc = "(no description)"
		}
		b.WriteString(itemLine(m.cursor == i, fmt.Sprintf("%-20s %s", skill.Name, MutedStyle.Render(desc))))
		b.WriteString("\n")
	}

	if len(filtered) == 0 {
		b.WriteString(MutedStyle.Render("No skills match the filter"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Type to filter \u2022 j/\u2193 k/\u2191 navigate \u2022 q/esc quit"))

	return b.String()
}
