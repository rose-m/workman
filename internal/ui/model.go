package ui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/michael-rose/workman/internal/config"
	"github.com/michael-rose/workman/internal/git"
	"github.com/michael-rose/workman/internal/state"
)

type Model struct {
	state               *state.AppState
	width               int
	height              int
	dialogType          DialogType
	addRepoDialog       AddRepoDialog
	addWorktreeDialog   AddWorktreeDialog
	confirmDeleteDialog ConfirmDeleteDialog
	errorMsg            string
}

func NewModel(appState *state.AppState) Model {
	m := Model{
		state:      appState,
		width:      80,
		height:     24,
		dialogType: DialogNone,
	}
	// Load initial worktrees
	m = m.loadWorktrees()
	return m
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

	case errorMsg:
		m.errorMsg = msg.err
		return m, nil

	case tea.KeyMsg:
		// Handle dialog mode
		if m.dialogType != DialogNone {
			return m.handleDialogKeys(msg)
		}

		// Normal mode key handling
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			m.state.TogglePane()
			return m, nil

		case "up", "k":
			if m.state.ActivePane == state.ReposPane {
				m.state.PrevRepo()
				m = m.loadWorktrees()
			} else {
				m.state.PrevWorktree()
			}
			return m, nil

		case "down", "j":
			if m.state.ActivePane == state.ReposPane {
				m.state.NextRepo()
				m = m.loadWorktrees()
			} else {
				m.state.NextWorktree()
			}
			return m, nil

		case "+":
			switch m.state.ActivePane {
			case state.ReposPane:
				// Show add repo dialog
				m.dialogType = DialogAddRepo
				m.addRepoDialog = NewAddRepoDialog()
				m.errorMsg = ""
			case state.WorktreesPane:
				// Show add worktree dialog (only if a repo is selected)
				if m.state.GetSelectedRepo() != nil {
					m.dialogType = DialogAddWorktree
					m.addWorktreeDialog = NewAddWorktreeDialog()
					m.errorMsg = ""
				}
			}
			return m, nil

		case "-":
			// Only allow deletion in worktrees pane
			if m.state.ActivePane == state.WorktreesPane {
				if len(m.state.Worktrees) > 0 && m.state.GetSelectedRepo() != nil {
					selectedWT := m.state.Worktrees[m.state.SelectedWTIndex]
					// Don't allow deleting the main worktree (first one)
					if m.state.SelectedWTIndex > 0 {
						m.dialogType = DialogConfirmDelete
						m.confirmDeleteDialog = NewConfirmDeleteDialog(selectedWT.Name, selectedWT.Branch)
						m.errorMsg = ""
					}
				}
			}
			return m, nil
		}
	}

	return m, nil
}

func (m Model) handleDialogKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc", "n":
		// Cancel dialog (or "n" for confirmation dialog)
		m.dialogType = DialogNone
		m.errorMsg = ""
		return m, nil

	case "y":
		// Confirm action (only for confirmation dialog)
		if m.dialogType == DialogConfirmDelete {
			return m.deleteWorktree()
		}
		return m, nil

	case "ctrl+s":
		// Save based on dialog type
		switch m.dialogType {
		case DialogAddRepo:
			return m.saveRepository()
		case DialogAddWorktree:
			return m.saveWorktree()
		}
		return m, nil

	default:
		// Update the dialog with the key press
		switch m.dialogType {
		case DialogAddRepo:
			cmd := m.addRepoDialog.Update(msg)
			m.errorMsg = ""
			return m, cmd
		case DialogAddWorktree:
			cmd := m.addWorktreeDialog.Update(msg)
			m.errorMsg = ""
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) saveRepository() (tea.Model, tea.Cmd) {
	// Validate inputs
	valid, errMsg := m.addRepoDialog.IsValid()
	if !valid {
		return m, showError(errMsg)
	}

	// Get values
	name, repoType, path := m.addRepoDialog.GetValues()

	// For local repos, verify path exists
	if repoType == "local" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return m, showError("Path does not exist")
		}
	}

	// Check for duplicate names
	for _, repo := range m.state.Config.Repositories {
		if repo.Name == name {
			return m, showError("Repository with this name already exists")
		}
	}

	// Create new repository
	newRepo := config.Repository{
		Name: name,
		Type: repoType,
		Path: path,
	}

	if repoType == "remote" {
		newRepo.URL = path
		// For remote repos, we'll need to clone them later
		// For now, just save the URL
	}

	// Add to config
	m.state.Config.Repositories = append(m.state.Config.Repositories, newRepo)

	// Save config
	if err := config.Save(m.state.Config); err != nil {
		return m, showError(fmt.Sprintf("Failed to save config: %v", err))
	}

	// Close dialog
	m.dialogType = DialogNone
	m.errorMsg = ""

	return m, nil
}

func (m Model) saveWorktree() (tea.Model, tea.Cmd) {
	// Validate inputs
	valid, errMsg := m.addWorktreeDialog.IsValid()
	if !valid {
		return m, showError(errMsg)
	}

	// Get selected repository
	repo := m.state.GetSelectedRepo()
	if repo == nil {
		return m, showError("No repository selected")
	}

	// Get branch name
	branch := m.addWorktreeDialog.GetBranchName()

	// Create worktree in configured root directory
	isRemote := repo.Type == "remote"
	rootDir := m.state.Config.RootDirectory
	if err := git.AddWorktree(repo.Path, rootDir, repo.Name, branch, isRemote); err != nil {
		return m, showError(fmt.Sprintf("Failed to create worktree: %v", err))
	}

	// Reload worktrees
	worktrees, err := git.ListWorktrees(repo.Path)
	if err != nil {
		return m, showError(fmt.Sprintf("Failed to list worktrees: %v", err))
	}
	m.state.Worktrees = worktrees

	// Close dialog
	m.dialogType = DialogNone
	m.errorMsg = ""

	return m, nil
}

func (m Model) deleteWorktree() (tea.Model, tea.Cmd) {
	// Get selected repository
	repo := m.state.GetSelectedRepo()
	if repo == nil {
		return m, showError("No repository selected")
	}

	// Get selected worktree
	if len(m.state.Worktrees) == 0 || m.state.SelectedWTIndex >= len(m.state.Worktrees) {
		return m, showError("No worktree selected")
	}

	selectedWT := m.state.Worktrees[m.state.SelectedWTIndex]

	// Remove worktree
	if err := git.RemoveWorktree(repo.Path, selectedWT.Path); err != nil {
		return m, showError(fmt.Sprintf("Failed to remove worktree: %v", err))
	}

	// Delete branch
	if err := git.DeleteBranch(repo.Path, selectedWT.Branch); err != nil {
		return m, showError(fmt.Sprintf("Failed to delete branch: %v", err))
	}

	// Reload worktrees
	worktrees, err := git.ListWorktrees(repo.Path)
	if err != nil {
		return m, showError(fmt.Sprintf("Failed to list worktrees: %v", err))
	}
	m.state.Worktrees = worktrees

	// Adjust selected index if needed
	if m.state.SelectedWTIndex >= len(m.state.Worktrees) && len(m.state.Worktrees) > 0 {
		m.state.SelectedWTIndex = len(m.state.Worktrees) - 1
	}

	// Close dialog
	m.dialogType = DialogNone
	m.errorMsg = ""

	return m, nil
}

func (m Model) loadWorktrees() Model {
	repo := m.state.GetSelectedRepo()
	if repo == nil {
		m.state.Worktrees = []state.Worktree{}
		return m
	}

	// Only try to load worktrees for local repos or if path exists
	if _, err := os.Stat(repo.Path); os.IsNotExist(err) {
		m.state.Worktrees = []state.Worktree{}
		return m
	}

	worktrees, err := git.ListWorktrees(repo.Path)
	if err != nil {
		// Failed to load worktrees, just set empty list
		m.state.Worktrees = []state.Worktree{}
		return m
	}

	m.state.Worktrees = worktrees
	m.state.SelectedWTIndex = 0
	return m
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

	mainView := lipgloss.JoinVertical(lipgloss.Left, panels, help)

	// Show dialog if active
	if m.dialogType != DialogNone {
		var dialog string
		switch m.dialogType {
		case DialogAddRepo:
			dialog = m.addRepoDialog.View()
		case DialogAddWorktree:
			dialog = m.addWorktreeDialog.View()
		case DialogConfirmDelete:
			dialog = m.confirmDeleteDialog.View()
		}

		// Add error message if present
		if m.errorMsg != "" {
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444")).
				Bold(true).
				Padding(0, 2)
			dialog = lipgloss.JoinVertical(lipgloss.Left, dialog, errorStyle.Render("Error: "+m.errorMsg))
		}

		// Center the dialog
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			dialog,
			lipgloss.WithWhitespaceChars("░"),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#1F2937")),
		)
	}

	return mainView
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
		"Navigation: ↑↓ or j/k   Switch pane: tab   Add: +   Delete: -   Quit: q or ctrl+c",
	}
	return helpStyle.Render(strings.Join(help, " • "))
}
