package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type removeStep int

const (
	removeStepScope removeStep = iota
	removeStepSkills
	removeStepConfirm
)

type RemoveModel struct {
	list

	step removeStep

	globalSkills  []string
	projectSkills []SkillInfo

	global   bool
	selected map[int]bool

	preGlobal bool
	preAgent  string

	width, height int

	Result    *RemoveResult
	Err       error
	Cancelled bool
}

func NewRemoveModel(globalSkills []string, projectSkills []SkillInfo, preGlobal bool, preAgent string) *RemoveModel {
	m := &RemoveModel{
		globalSkills:  globalSkills,
		projectSkills: projectSkills,
		selected:      make(map[int]bool),
		preGlobal:     preGlobal,
		preAgent:      preAgent,
	}

	if preGlobal {
		m.global = true
		m.step = removeStepSkills
	}

	return m
}

func (m *RemoveModel) Init() tea.Cmd { return nil }

func (m *RemoveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if IsKey(msg, DefaultKeyMap.Back) && m.step > removeStepScope {
			m.step--
			return m, nil
		}
		switch m.step {
		case removeStepScope:
			m.updateScopeStep(msg)
		case removeStepSkills:
			m.updateSkillsStep(msg)
		case removeStepConfirm:
			if cmd := m.updateConfirmStep(msg); cmd != nil {
				return m, cmd
			}
		}
	case tea.MouseMsg:
		m.handleMouse(msg)
	}
	return m, nil
}

func (m *RemoveModel) updateScopeStep(msg tea.KeyMsg) {
	switch {
	case IsKey(msg, DefaultKeyMap.Up), IsKey(msg, DefaultKeyMap.Down):
		m.global = !m.global
	case IsKey(msg, DefaultKeyMap.Confirm):
		m.step = removeStepSkills
		m.reset()
	}
}

func (m *RemoveModel) updateSkillsStep(msg tea.KeyMsg) {
	items := m.currentItems()
	switch {
	case IsKey(msg, DefaultKeyMap.Down):
		m.down(len(items))
	case IsKey(msg, DefaultKeyMap.Up):
		m.up(len(items))
	case IsKey(msg, DefaultKeyMap.Select):
		m.selected[m.cursor] = !m.selected[m.cursor]
	case IsKey(msg, DefaultKeyMap.Confirm):
		hasSelected := false
		for _, v := range m.selected {
			if v {
				hasSelected = true
				break
			}
		}
		if hasSelected {
			m.step = removeStepConfirm
		}
	}
}

func (m *RemoveModel) updateConfirmStep(msg tea.KeyMsg) tea.Cmd {
	if IsKey(msg, DefaultKeyMap.Confirm) {
		items := m.currentItems()
		var names []string
		for i, selected := range m.selected {
			if selected {
				names = append(names, items[i].Name)
			}
		}
		m.Result = &RemoveResult{
			SkillNames: names,
			Agent:      m.preAgent,
			Global:     m.global,
		}
		return tea.Quit
	}
	return nil
}

func (m *RemoveModel) currentItems() []SkillInfo {
	if m.global {
		items := make([]SkillInfo, len(m.globalSkills))
		for i, name := range m.globalSkills {
			items[i] = SkillInfo{Name: name}
		}
		return items
	}
	return m.projectSkills
}

func (m *RemoveModel) handleMouse(msg tea.MouseMsg) {
	if m.step == removeStepSkills {
		m.list.handleMouse(msg, len(m.currentItems()))
	}
}

func (m *RemoveModel) View() string {
	if m.Err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v\n\nPress any key to quit.", m.Err))
	}

	var b strings.Builder
	b.WriteString(TitleStyle.Render("skillbase remove"))
	b.WriteString("\n")

	switch m.step {
	case removeStepScope:
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

	case removeStepSkills:
		scope := "project"
		if m.global {
			scope = "global"
		}
		b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Select skills to remove from %s scope", scope)))
		b.WriteString("\n\n")
		items := m.currentItems()
		for i, item := range items {
			checked := "[ ] "
			if m.selected[i] {
				checked = "[x] "
			}
			agents := ""
			if len(item.Agents) > 0 {
				agents = fmt.Sprintf(" (%s)", strings.Join(item.Agents, ", "))
			}
			b.WriteString(itemLine(m.cursor == i, fmt.Sprintf("%s%s%s", checked, item.Name, MutedStyle.Render(agents))))
			b.WriteString("\n")
		}

	case removeStepConfirm:
		scope := "project"
		if m.global {
			scope = "global"
		}
		b.WriteString(SubtitleStyle.Render("Confirm removal"))
		b.WriteString("\n\n")
		items := m.currentItems()
		b.WriteString(ItemStyle.Render(fmt.Sprintf("Scope: %s", scope)))
		b.WriteString("\n")
		b.WriteString(ItemStyle.Render("Skills:"))
		b.WriteString("\n")
		for i, item := range items {
			if m.selected[i] {
				b.WriteString(ItemStyle.Render(fmt.Sprintf("  \u2022 %s", item.Name)))
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("Press Enter to confirm, h/\u2190 to go back"))
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("j/\u2193 k/\u2191 navigate \u2022 space select \u2022 enter/l/\u2192 confirm \u2022 h/\u2190 back \u2022 q/esc quit"))

	return b.String()
}
