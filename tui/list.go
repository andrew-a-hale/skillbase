package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"github.com/muesli/reflow/wordwrap"
)

type ListModel struct {
	list

	projectSkills []SkillInfo
	globalSkills  []SkillInfo

	width, height int

	Err       error
	Cancelled bool
}

func NewListModel(projectSkills, globalSkills []SkillInfo) *ListModel {
	return &ListModel{
		projectSkills: projectSkills,
		globalSkills:  globalSkills,
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
	case tea.KeyMsg:
		if IsKey(msg, DefaultKeyMap.Quit) {
			m.Cancelled = true
			return m, tea.Quit
		}
		all := m.allSkills()
		switch {
		case IsKey(msg, DefaultKeyMap.Down):
			m.down(len(all))
		case IsKey(msg, DefaultKeyMap.Up):
			m.up(len(all))
		}
	case tea.MouseMsg:
		m.handleMouse(msg, len(m.allSkills()))
	}
	return m, nil
}

func (m *ListModel) View() string {
	if m.Err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v\n\nPress any key to quit.", m.Err))
	}

	var b strings.Builder
	b.WriteString(TitleStyle.Render("skillbase list"))
	b.WriteString("\n\n")

	cursor := 0
	if len(m.projectSkills) > 0 {
		b.WriteString(SubtitleStyle.Render("Project Scope"))
		b.WriteString("\n")
		for _, skill := range m.projectSkills {
			selected := m.cursor == cursor
			agents := ""
			if len(skill.Agents) > 0 {
				agents = fmt.Sprintf(" [%s]", strings.Join(skill.Agents, ", "))
			}
			desc := skill.Description
			if desc == "" {
				desc = "(no description)"
			}
			b.WriteString(m.renderSkillLine(selected, skill.Name, agents, desc))
			cursor++
		}
		b.WriteString("\n")
	}

	if len(m.globalSkills) > 0 {
		b.WriteString(SubtitleStyle.Render("Global Scope"))
		b.WriteString("\n")
		for _, skill := range m.globalSkills {
			selected := m.cursor == cursor
			desc := skill.Description
			if desc == "" {
				desc = "(no description)"
			}
			b.WriteString(m.renderSkillLine(selected, skill.Name, "", desc))
			cursor++
		}
	}

	if len(m.projectSkills) == 0 && len(m.globalSkills) == 0 {
		b.WriteString(MutedStyle.Render("No skills installed"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("j/\u2193 k/\u2191 navigate \u2022 q/esc quit"))

	return b.String()
}

func (m *ListModel) renderSkillLine(selected bool, name, agents, desc string) string {
	if selected {
		header := name + MutedStyle.Render(agents)
		wrapWidth := m.width - 4 // 2 padding + "> " prefix
		if wrapWidth < 10 {
			wrapWidth = 10
		}
		wrappedDesc := wordwrap.String(desc, wrapWidth)
		lines := strings.Split(wrappedDesc, "\n")

		var b strings.Builder
		b.WriteString(SelectedItemStyle.Render("> " + header))
		for _, line := range lines {
			b.WriteString("\n")
			b.WriteString(SelectedItemStyle.Render("  " + MutedStyle.Render(line)))
		}
		b.WriteString("\n")
		return b.String()
	}

	// Not selected: truncate description to fit on one line
	available := m.width - 6 // 4 padding + "  " prefix
	if available < 10 {
		available = 10
	}
	header := name + MutedStyle.Render(agents)
	prefix := header + "  "
	prefixWidth := lipgloss.Width(prefix)
	descWidth := available - prefixWidth
	if descWidth < 3 {
		descWidth = 3
	}
	truncatedDesc := truncate.StringWithTail(desc, uint(descWidth), "...")
	return ItemStyle.Render("  "+prefix+MutedStyle.Render(truncatedDesc)) + "\n"
}
