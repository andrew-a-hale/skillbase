package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type UpdateModel struct {
	list

	skills []SkillInfo

	width, height int

	Result    *UpdateResult
	Err       error
	Cancelled bool
}

func NewUpdateModel(skills []SkillInfo) *UpdateModel {
	return &UpdateModel{skills: skills}
}

func (m *UpdateModel) Init() tea.Cmd { return nil }

func (m *UpdateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		switch {
		case IsKey(msg, DefaultKeyMap.Down):
			m.down(len(m.skills))
		case IsKey(msg, DefaultKeyMap.Up):
			m.up(len(m.skills))
		case IsKey(msg, DefaultKeyMap.Confirm):
			if len(m.skills) > 0 {
				m.Result = &UpdateResult{SkillName: m.skills[m.cursor].Name}
				return m, tea.Quit
			}
		}
	case tea.MouseMsg:
		m.handleMouse(msg, len(m.skills))
	}
	return m, nil
}

func (m *UpdateModel) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion

	if m.Err != nil {
		v.SetContent(ErrorStyle.Render(fmt.Sprintf("Error: %v\n\nPress any key to quit.", m.Err)))
		return v
	}

	var b strings.Builder
	b.WriteString(TitleStyle.Render("skillbase update"))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render("Select a skill to update"))
	b.WriteString("\n\n")

	for i, skill := range m.skills {
		desc := skill.Description
		if desc == "" {
			desc = "(no description)"
		}
		agents := ""
		if len(skill.Agents) > 0 {
			agents = fmt.Sprintf(" [%s]", strings.Join(skill.Agents, ", "))
		}
		b.WriteString(renderListItem(m.cursor == i, m.width, skill.Name+MutedStyle.Render(agents), desc))
	}

	if len(m.skills) == 0 {
		b.WriteString(MutedStyle.Render("No skills installed"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("j/\u2193 k/\u2191 navigate \u2022 enter/l/\u2192 select \u2022 q/esc quit"))

	v.SetContent(viewMargin(b.String()))
	return v
}
