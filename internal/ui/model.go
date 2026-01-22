package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/michael-rose/workman/internal/state"
)

type Model struct {
	state  *state.AppState
	width  int
	height int
}

func NewModel(appState *state.AppState) Model {
	return Model{
		state:  appState,
		width:  80,
		height: 24,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			m.state.TogglePane()
			return m, nil

		case "up", "k":
			if m.state.ActivePane == state.ReposPane {
				m.state.PrevRepo()
			} else {
				m.state.PrevWorktree()
			}
			return m, nil

		case "down", "j":
			if m.state.ActivePane == state.ReposPane {
				m.state.NextRepo()
			} else {
				m.state.NextWorktree()
			}
			return m, nil

		case "+":
			// TODO: Implement add repo/worktree dialog
			return m, nil
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.width < 40 || m.height < 10 {
		return "Terminal too small. Please resize."
	}

	// Calculate panel dimensions (split view: 40% left, 60% right)
	leftWidth := m.width*40/100 - 4
	rightWidth := m.width*60/100 - 4
	panelHeight := m.height - 6

	// Render left panel (repositories)
	leftPanel := m.renderReposPanel(leftWidth, panelHeight)

	// Render right panel (worktrees)
	rightPanel := m.renderWorktreesPanel(rightWidth, panelHeight)

	// Combine panels side by side
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// Render help text
	help := m.renderHelp()

	return lipgloss.JoinVertical(lipgloss.Left, panels, help)
}

func (m Model) renderReposPanel(width, height int) string {
	isActive := m.state.ActivePane == state.ReposPane
	style := panelStyle
	if isActive {
		style = activePanelStyle
	}

	header := headerStyle.Render("Repositories")
	var items []string

	if len(m.state.Config.Repositories) == 0 {
		items = append(items, infoStyle.Render("No repositories yet"))
		items = append(items, infoStyle.Render("Press '+' to add one"))
	} else {
		for i, repo := range m.state.Config.Repositories {
			itemText := fmt.Sprintf("%s (%s)", repo.Name, repo.Type)
			if isActive && i == m.state.SelectedRepoIndex {
				items = append(items, selectedItemStyle.Render("> "+itemText))
			} else {
				items = append(items, itemStyle.Render("  "+itemText))
			}
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, header, strings.Join(items, "\n"))

	return style.
		Width(width).
		Height(height).
		Render(content)
}

func (m Model) renderWorktreesPanel(width, height int) string {
	isActive := m.state.ActivePane == state.WorktreesPane
	style := panelStyle
	if isActive {
		style = activePanelStyle
	}

	selectedRepo := m.state.GetSelectedRepo()
	var header string
	if selectedRepo != nil {
		header = headerStyle.Render(fmt.Sprintf("Worktrees - %s", selectedRepo.Name))
	} else {
		header = headerStyle.Render("Worktrees")
	}

	var items []string

	if len(m.state.Worktrees) == 0 {
		if selectedRepo == nil {
			items = append(items, infoStyle.Render("Select a repository first"))
		} else {
			items = append(items, infoStyle.Render("No worktrees yet"))
			items = append(items, infoStyle.Render("Press '+' to add one"))
		}
	} else {
		for i, wt := range m.state.Worktrees {
			itemText := fmt.Sprintf("%s [%s]", wt.Name, wt.Branch)
			if isActive && i == m.state.SelectedWTIndex {
				items = append(items, selectedItemStyle.Render("> "+itemText))
			} else {
				items = append(items, itemStyle.Render("  "+itemText))
			}
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, header, strings.Join(items, "\n"))

	return style.
		Width(width).
		Height(height).
		Render(content)
}

func (m Model) renderHelp() string {
	help := []string{
		"Navigation: ↑↓ or j/k   Switch pane: tab   Add: +   Quit: q or ctrl+c",
	}
	return helpStyle.Render(strings.Join(help, " • "))
}
