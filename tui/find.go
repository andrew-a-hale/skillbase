package tui

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"golang.org/x/term"
)

type FindModel struct {
	list

	skills  []SkillInfo
	filter  string
	loading bool

	spinner spinner.Model
	loadCmd tea.Cmd

	width, height int

	Err       error
	Cancelled bool
}

func NewFindModel(filter string) *FindModel {
	s := spinner.New(spinner.WithSpinner(spinner.Ellipsis))
	s.Style = TitleStyle
	w, h, _ := term.GetSize(int(os.Stdout.Fd()))
	return &FindModel{
		filter:  strings.ToLower(filter),
		loading: true,
		spinner: s,
		width:   w - 4,
		height:  h,
	}
}

func (m *FindModel) WithLoadCmd(cmd tea.Cmd) *FindModel {
	m.loadCmd = cmd
	return m
}

func (m *FindModel) filteredSkills() []SkillInfo {
	indices := filteredIndices(m.skills, m.filter)
	result := make([]SkillInfo, len(indices))
	for i, idx := range indices {
		result[i] = m.skills[idx]
	}
	return result
}

func (m *FindModel) Init() tea.Cmd {
	if m.loading && m.loadCmd != nil {
		return tea.Batch(func() tea.Msg { return m.spinner.Tick() }, m.loadCmd)
	}
	return nil
}

func (m *FindModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.loading {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		case LoadMsg:
			if msg.Err != nil {
				m.Err = msg.Err
			} else {
				m.skills = msg.Skills
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

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
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
		case msg.Code == tea.KeyBackspace:
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.reset()
			}
		case len(msg.Text) > 0:
			m.filter += strings.ToLower(msg.Text)
			m.reset()
		}
	case tea.MouseMsg:
		m.handleMouse(msg, len(m.filteredSkills()))
	}
	return m, nil
}

func (m *FindModel) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion

	if m.loading {
		var b strings.Builder
		b.WriteString(fmt.Sprintf("Pulling Repository%s", m.spinner.View()))
		v.SetContent(b.String())
		return v
	}

	if m.Err != nil {
		v.SetContent(ErrorStyle.Render(fmt.Sprintf("Error: %v\n\nPress any key to quit.", m.Err)))
		return v
	}

	filtered := m.filteredSkills()
	var b strings.Builder
	b.WriteString(TitleStyle.Render("skillbase find"))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Filter: %s_", m.filter)))
	b.WriteString("\n\n")

	for i, skill := range filtered {
		if i-m.cursor > 5 || i-m.cursor < -5 {
			continue // skip
		}

		desc := skill.Description
		if desc == "" {
			desc = "(no description)"
		}
		b.WriteString(renderListItem(m.cursor == i, m.width, skill.Name, desc))
	}

	if len(filtered) == 0 {
		b.WriteString(MutedStyle.Render("No skills match the filter"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Type to filter \u2022 j/\u2193 k/\u2191 navigate \u2022 q/esc quit"))

	v.SetContent(viewMargin(b.String()))
	return v
}
