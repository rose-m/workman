package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/michael-rose/workman/internal/config"
	"github.com/michael-rose/workman/internal/git"
	"github.com/michael-rose/workman/internal/state"
)

type Model struct {
	state                   *state.AppState
	width                   int
	height                  int
	dialogType              DialogType
	addRepoDialog           AddRepoDialog
	addWorktreeDialog       AddWorktreeDialog
	confirmDeleteDialog     ConfirmDeleteDialog
	confirmDeleteRepoDialog ConfirmDeleteRepositoryDialog
	errorMsg                string
	successMsg              string
}

type editTarget int

const (
	editNotesTarget editTarget = iota
	editScriptTarget
)

type editorFinishedMsg struct {
	err          error
	target       editTarget
	repoName     string
	worktreeName string
	tempPath     string
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

	case successMsg:
		m.successMsg = msg.msg
		return m, nil

	case editorFinishedMsg:
		if msg.tempPath != "" {
			defer func() {
				_ = os.Remove(msg.tempPath)
			}()
		}

		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Editor failed: %v", msg.err)
			m.successMsg = ""
			return m, nil
		}

		content, err := os.ReadFile(msg.tempPath)
		if err != nil {
			m.errorMsg = fmt.Sprintf("Failed to read editor file: %v", err)
			m.successMsg = ""
			return m, nil
		}

		value := strings.TrimSpace(string(content))
		switch msg.target {
		case editNotesTarget:
			if err := config.SaveWorktreeNotes(msg.repoName, msg.worktreeName, value); err != nil {
				m.errorMsg = fmt.Sprintf("Failed to save notes: %v", err)
				m.successMsg = ""
				return m, nil
			}
			m.errorMsg = ""
			return m, showSuccess("Notes saved")
		case editScriptTarget:
			if err := config.SaveRepoScript(msg.repoName, value); err != nil {
				m.errorMsg = fmt.Sprintf("Failed to save script: %v", err)
				m.successMsg = ""
				return m, nil
			}
			m.errorMsg = ""
			return m, showSuccess("Post-create script saved")
		}

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

		case "h":
			m.state.ActivePane = state.ReposPane
			return m, nil

		case "l":
			m.state.ActivePane = state.WorktreesPane
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

		case "y":
			if m.state.ActivePane == state.WorktreesPane {
				if len(m.state.Worktrees) > 0 && m.state.GetSelectedRepo() != nil {
					return m.yankWorktreeCommand()
				}
			}
			return m, nil

		case "n":
			if m.state.ActivePane == state.WorktreesPane {
				if len(m.state.Worktrees) > 0 && m.state.GetSelectedRepo() != nil {
					selectedWT := m.state.Worktrees[m.state.SelectedWTIndex]
					repo := m.state.GetSelectedRepo()
					currentNotes, err := config.GetWorktreeNotes(repo.Name, selectedWT.Name)
					if err != nil {
						return m, showError(fmt.Sprintf("Failed to load notes: %v", err))
					}
					return m.openEditor(editNotesTarget, repo.Name, selectedWT.Name, currentNotes)
				}
			}
			return m, nil

		case "s":
			if m.state.ActivePane == state.ReposPane {
				if repo := m.state.GetSelectedRepo(); repo != nil {
					currentScript, err := config.GetRepoScript(repo.Name)
					if err != nil {
						return m, showError(fmt.Sprintf("Failed to load script: %v", err))
					}
					return m.openEditor(editScriptTarget, repo.Name, "", currentScript)
				}
			}
			return m, nil

		case "+":
			switch m.state.ActivePane {
			case state.ReposPane:
				// Show add repo dialog
				m.dialogType = DialogAddRepo
				m.addRepoDialog = NewAddRepoDialog()
				m.errorMsg = ""
				m.successMsg = ""
			case state.WorktreesPane:
				// Show add worktree dialog (only if a repo is selected)
				if repo := m.state.GetSelectedRepo(); repo != nil {
					// Fetch branches for autocomplete
					branches, err := git.ListBranches(repo.Path)
					if err != nil {
						branches = []string{} // If fetch fails, continue with empty list
					}

					m.dialogType = DialogAddWorktree
					m.addWorktreeDialog = NewAddWorktreeDialog(branches)
					m.errorMsg = ""
					m.successMsg = ""
				}
			}
			return m, nil

		case "-":
			// Handle deletion based on active pane
			switch m.state.ActivePane {
			case state.ReposPane:
				// Delete repository
				if len(m.state.Config.Repositories) > 0 {
					selectedRepo := m.state.GetSelectedRepo()
					if selectedRepo != nil {
						m.dialogType = DialogConfirmDeleteRepo
						m.confirmDeleteRepoDialog = NewConfirmDeleteRepositoryDialog(selectedRepo.Name)
						m.errorMsg = ""
						m.successMsg = ""
					}
				}
			case state.WorktreesPane:
				// Delete worktree
				if len(m.state.Worktrees) > 0 && m.state.GetSelectedRepo() != nil {
					selectedWT := m.state.Worktrees[m.state.SelectedWTIndex]
					// Don't allow deleting the main worktree (first one)
					if m.state.SelectedWTIndex > 0 {
						m.dialogType = DialogConfirmDelete
						m.confirmDeleteDialog = NewConfirmDeleteDialog(selectedWT.Name, selectedWT.Branch)
						m.errorMsg = ""
						m.successMsg = ""
					}
				}
			}
			return m, nil

		case "enter":
			if m.state.ActivePane == state.WorktreesPane &&
				len(m.state.Worktrees) > 0 &&
				m.state.GetSelectedRepo() != nil {
				// Execute enter_script
				if err := m.executeScript(m.state.Config.EnterScript); err != nil {
					m.errorMsg = fmt.Sprintf("Enter key: %v. Set 'enter_script' to a script file path in config.toml", err)
					m.successMsg = ""
				} else {
					m.successMsg = "Script executed"
					m.errorMsg = ""
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

	case "esc":
		// Cancel dialog for all dialog types
		m.dialogType = DialogNone
		m.errorMsg = ""
		m.successMsg = ""
		return m, nil

	case "n":
		// "n" only cancels confirmation dialogs (means "no")
		// For other dialogs, it's just a regular character
		switch m.dialogType {
		case DialogConfirmDelete, DialogConfirmDeleteRepo:
			m.dialogType = DialogNone
			m.errorMsg = ""
			m.successMsg = ""
			return m, nil
		}

	case "y":
		// Confirm action (only for confirmation dialogs)
		// For other dialogs, fall through to pass "y" to the input handler
		switch m.dialogType {
		case DialogConfirmDelete:
			return m.deleteWorktree()
		case DialogConfirmDeleteRepo:
			return m.deleteRepository()
		}

	case "ctrl+s":
		// Save based on dialog type
		switch m.dialogType {
		case DialogAddRepo:
			return m.saveRepository()
		case DialogAddWorktree:
			return m.saveWorktree()
		}
		return m, nil
	}

	// Update the dialog with the key press (for input dialogs)
	switch m.dialogType {
	case DialogAddRepo:
		cmd := m.addRepoDialog.Update(msg)
		m.errorMsg = ""
		m.successMsg = ""
		return m, cmd
	case DialogAddWorktree:
		cmd := m.addWorktreeDialog.Update(msg)
		m.errorMsg = ""
		m.successMsg = ""
		return m, cmd
	}

	return m, nil
}

func (m Model) openEditor(target editTarget, repoName, worktreeName, content string) (tea.Model, tea.Cmd) {
	var prefix string
	switch target {
	case editNotesTarget:
		prefix = "workman-notes-"
	case editScriptTarget:
		prefix = "workman-script-"
	default:
		prefix = "workman-edit-"
	}

	tmpFile, err := os.CreateTemp("", prefix)
	if err != nil {
		return m, showError(fmt.Sprintf("Failed to create temp file: %v", err))
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return m, showError(fmt.Sprintf("Failed to write temp file: %v", err))
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpFile.Name())
		return m, showError(fmt.Sprintf("Failed to close temp file: %v", err))
	}

	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if editor == "" {
		editor = "vi"
	}

	cmd := exec.Command(editor, tmpFile.Name())
	m.errorMsg = ""
	m.successMsg = ""
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorFinishedMsg{
			err:          err,
			target:       target,
			repoName:     repoName,
			worktreeName: worktreeName,
			tempPath:     tmpFile.Name(),
		}
	})
}

func (m Model) saveRepository() (tea.Model, tea.Cmd) {
	// Validate inputs
	valid, errMsg := m.addRepoDialog.IsValid()
	if !valid {
		return m, showError(errMsg)
	}

	// Get values
	name, repoType, pathOrURL := m.addRepoDialog.GetValues()

	// Check for duplicate names
	for _, repo := range m.state.Config.Repositories {
		if repo.Name == name {
			return m, showError("Repository with this name already exists")
		}
	}

	var repoPath string
	var repoURL string

	if repoType == "local" {
		// For local repos, verify path exists
		if _, err := os.Stat(pathOrURL); os.IsNotExist(err) {
			return m, showError("Path does not exist")
		}
		repoPath = pathOrURL
	} else {
		// For remote repos, clone immediately
		repoURL = pathOrURL

		// Sanitize the repo name for the directory
		sanitizedName := sanitizeRepoName(name)

		// Construct target path: <rootDir>/<sanitized-name>
		rootDir := strings.TrimSpace(m.state.Config.RootDirectory)
		if rootDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return m, showError(fmt.Sprintf("Failed to get home directory: %v", err))
			}
			rootDir = homeDir
		}
		repoPath = filepath.Join(rootDir, sanitizedName)

		// Clone the repository
		if err := git.CloneRepository(repoURL, repoPath); err != nil {
			return m, showError(fmt.Sprintf("Failed to clone repository: %v", err))
		}
	}

	// Create new repository
	newRepo := config.Repository{
		Name: name,
		Type: repoType,
		Path: repoPath,
		URL:  repoURL,
	}

	// Add to config
	m.state.Config.Repositories = append(m.state.Config.Repositories, newRepo)

	// Save config
	if err := config.Save(m.state.Config); err != nil {
		return m, showError(fmt.Sprintf("Failed to save config: %v", err))
	}

	// Select the newly added repository and load its worktrees
	m.state.SelectedRepoIndex = len(m.state.Config.Repositories) - 1
	m = m.loadWorktrees()

	// Close dialog
	m.dialogType = DialogNone
	m.errorMsg = ""

	return m, showSuccess(fmt.Sprintf("Repository '%s' added successfully", name))
}

// sanitizeRepoName converts a repository name to a safe directory name
func sanitizeRepoName(name string) string {
	// Replace any non-alphanumeric characters (except dash) with dash
	result := ""
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' {
			result += string(ch)
		} else {
			result += "-"
		}
	}

	// Remove leading/trailing dashes and collapse multiple dashes
	result = strings.Trim(result, "-")
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}

	// Convert to lowercase for consistency
	return strings.ToLower(result)
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

	// Reload worktrees to get the new worktree path
	worktrees, err := git.ListWorktrees(repo.Path)
	if err != nil {
		return m, showError(fmt.Sprintf("Failed to list worktrees: %v", err))
	}
	m.state.Worktrees = worktrees

	// Select the newly created worktree
	for i, wt := range worktrees {
		if wt.Branch == branch {
			m.state.SelectedWTIndex = i
			break
		}
	}

	// Execute post-create script if configured
	script, err := config.GetRepoScript(repo.Name)
	if err != nil {
		return m, showError(fmt.Sprintf("Failed to load post-create script: %v", err))
	}
	if script != "" {
		// Find the newly created worktree
		var newWorktreePath string
		for _, wt := range worktrees {
			if wt.Branch == branch {
				newWorktreePath = wt.Path
				break
			}
		}

		if newWorktreePath != "" {
			if err := git.ExecutePostCreateScript(script, repo.Path, newWorktreePath); err != nil {
				m.dialogType = DialogNone
				m.errorMsg = ""
				return m, showError(fmt.Sprintf("Worktree created but script failed: %v", err))
			}
		}
	}

	// Close dialog
	m.dialogType = DialogNone
	m.errorMsg = ""

	return m, showSuccess("Worktree created successfully")
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

	// Remove notes for this worktree
	_ = config.DeleteWorktreeNotes(repo.Name, selectedWT.Name)

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

func (m Model) deleteRepository() (tea.Model, tea.Cmd) {
	// Get selected repository
	repo := m.state.GetSelectedRepo()
	if repo == nil {
		return m, showError("No repository selected")
	}

	// Track errors
	var errors []string

	// List all worktrees for the repository
	worktrees, err := git.ListWorktrees(repo.Path)
	if err != nil && !os.IsNotExist(err) {
		// If we can't list worktrees and it's not because the repo doesn't exist, fail
		errors = append(errors, fmt.Sprintf("Failed to list worktrees: %v", err))
	} else {
		// Delete each worktree (except the main one which will be deleted with the repo)
		for i, wt := range worktrees {
			if i == 0 {
				// Skip the main worktree - it will be deleted with the repository
				continue
			}

			// Remove worktree
			if err := git.RemoveWorktree(repo.Path, wt.Path); err != nil {
				errors = append(errors, fmt.Sprintf("Failed to remove worktree '%s': %v", wt.Name, err))
				continue
			}

			// Delete branch - failure is non-critical since worktree is already removed
			_ = git.DeleteBranch(repo.Path, wt.Branch)
		}
	}

	// If we had errors deleting worktrees, don't proceed with repo deletion
	if len(errors) > 0 {
		m.dialogType = DialogNone
		return m, showError(fmt.Sprintf("Errors during deletion:\n%s\nRepository kept in config.", strings.Join(errors, "\n")))
	}

	// Delete repository directory
	if err := git.DeleteRepository(repo.Path); err != nil {
		// Check if directory doesn't exist - that's ok, we'll still remove from config
		if !os.IsNotExist(err) {
			m.dialogType = DialogNone
			return m, showError(fmt.Sprintf("Failed to delete repository: %v\nRepository kept in config.", err))
		}
		// Directory doesn't exist - that's fine, continue with config removal
	}

	// Remove all notes and script for this repository
	for _, wt := range worktrees {
		_ = config.DeleteWorktreeNotes(repo.Name, wt.Name)
	}
	_ = config.DeleteRepoScript(repo.Name)

	// Remove repository from config
	repoIndex := m.state.SelectedRepoIndex
	m.state.Config.Repositories = append(
		m.state.Config.Repositories[:repoIndex],
		m.state.Config.Repositories[repoIndex+1:]...,
	)

	// Save config
	if err := config.Save(m.state.Config); err != nil {
		return m, showError(fmt.Sprintf("Failed to save config: %v", err))
	}

	// Adjust selected repository index
	if len(m.state.Config.Repositories) == 0 {
		m.state.SelectedRepoIndex = 0
	} else if m.state.SelectedRepoIndex >= len(m.state.Config.Repositories) {
		m.state.SelectedRepoIndex = len(m.state.Config.Repositories) - 1
	}

	// Reload worktrees for the new selected repository
	m = m.loadWorktrees()

	// Close dialog
	m.dialogType = DialogNone
	m.errorMsg = ""

	return m, showSuccess(fmt.Sprintf("Repository '%s' deleted successfully", repo.Name))
}

func (m Model) yankWorktreeCommand() (tea.Model, tea.Cmd) {
	repo := m.state.GetSelectedRepo()
	if repo == nil {
		return m, showError("No repository selected")
	}

	if len(m.state.Worktrees) == 0 || m.state.SelectedWTIndex >= len(m.state.Worktrees) {
		return m, showError("No worktree selected")
	}

	selectedWT := m.state.Worktrees[m.state.SelectedWTIndex]

	// Get template from config
	template := m.state.Config.YankTemplate
	if template == "" {
		template = "${worktree_path}"
	}

	// Perform substitutions
	result := template
	result = strings.ReplaceAll(result, "${repo_name}", repo.Name)
	result = strings.ReplaceAll(result, "${branch_name}", selectedWT.Branch)
	result = strings.ReplaceAll(result, "${worktree_path}", selectedWT.Path)
	result = strings.ReplaceAll(result, "${worktree_name}", selectedWT.Name)

	// Copy to clipboard
	if err := clipboard.WriteAll(result); err != nil {
		return m, showError(fmt.Sprintf("Failed to copy: %v", err))
	}

	m.errorMsg = ""
	return m, showSuccess("Copied to clipboard")
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

	// Show success/error feedback if no dialog is active
	if m.dialogType == DialogNone {
		if m.successMsg != "" {
			successStyle := lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#047857", Dark: "#10B981"}).
				Background(lipgloss.AdaptiveColor{Light: "#D1FAE5", Dark: "#064E3B"}).
				Bold(true).
				Padding(0, 2).
				Width(m.width - 4)
			feedback := successStyle.Render("✓ " + m.successMsg)
			mainView = lipgloss.JoinVertical(lipgloss.Left, feedback, mainView)
		} else if m.errorMsg != "" {
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#B91C1C", Dark: "#EF4444"}).
				Background(lipgloss.AdaptiveColor{Light: "#FEE2E2", Dark: "#7F1D1D"}).
				Bold(true).
				Padding(0, 2).
				Width(m.width - 4)
			feedback := errorStyle.Render("✗ " + m.errorMsg)
			mainView = lipgloss.JoinVertical(lipgloss.Left, feedback, mainView)
		}
	}

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
		case DialogConfirmDeleteRepo:
			dialog = m.confirmDeleteRepoDialog.View()
		}

		// Add error message if present
		if m.errorMsg != "" {
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#B91C1C", Dark: "#EF4444"}).
				Bold(true).
				Padding(0, 2)
			dialog = lipgloss.JoinVertical(lipgloss.Left, dialog, errorStyle.Render("Error: "+m.errorMsg))
		}

		// Center the dialog
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			dialog,
			lipgloss.WithWhitespaceChars("░"),
			lipgloss.WithWhitespaceForeground(lipgloss.AdaptiveColor{Light: "#E5E7EB", Dark: "#1F2937"}),
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
			scriptIndicator := ""
			hasScript, err := config.HasRepoScript(repo.Name)
			if err == nil && hasScript {
				scriptIndicator = " 📜"
			}
			itemText := fmt.Sprintf("%s (%s)%s", repo.Name, repo.Type, scriptIndicator)
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

	// Add notes section if a worktree is selected
	var notesSection string
	if len(m.state.Worktrees) > 0 && m.state.SelectedWTIndex < len(m.state.Worktrees) {
		selectedWT := m.state.Worktrees[m.state.SelectedWTIndex]

		notesHeader := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}).
			Bold(true).
			Render("\nNotes:")

		// Show the notes in read-only mode
		notes := ""
		if selectedRepo != nil {
			storedNotes, err := config.GetWorktreeNotes(selectedRepo.Name, selectedWT.Name)
			if err == nil {
				notes = storedNotes
			}
		}

		if notes != "" {
			// Truncate notes if too long
			maxNoteLength := 200
			displayNotes := notes
			if len(notes) > maxNoteLength {
				displayNotes = notes[:maxNoteLength] + "..."
			}
			// Replace newlines with spaces for compact display
			displayNotes = strings.ReplaceAll(displayNotes, "\n", " ")
			notesContent := lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#4B5563", Dark: "#D1D5DB"}).
				Italic(true).
				Render("  " + displayNotes)
			notesSection = notesHeader + "\n" + notesContent
		} else {
			emptyNotes := lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"}).
				Italic(true).
				Render("  (no notes - press 'n' to add)")
			notesSection = notesHeader + "\n" + emptyNotes
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, header, strings.Join(items, "\n"), notesSection)

	return style.
		Width(width).
		Height(height).
		Render(content)
}

// executeScript executes a script file with variable substitution
// Returns error if script path is empty or execution fails
func (m Model) executeScript(scriptPath string) error {
	if scriptPath == "" {
		return fmt.Errorf("no script configured")
	}

	// Expand ~ to home directory
	if strings.HasPrefix(scriptPath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		scriptPath = filepath.Join(homeDir, scriptPath[1:])
	}

	// Read script file
	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read script file %s: %w", scriptPath, err)
	}

	selectedWT := m.state.Worktrees[m.state.SelectedWTIndex]
	selectedRepo := m.state.GetSelectedRepo()

	// Variable substitution
	script := string(scriptContent)
	script = strings.ReplaceAll(script, "${worktree_path}", selectedWT.Path)
	script = strings.ReplaceAll(script, "${path}", selectedWT.Path)
	script = strings.ReplaceAll(script, "${branch_name}", selectedWT.Branch)
	script = strings.ReplaceAll(script, "${branch}", selectedWT.Branch)
	if selectedRepo != nil {
		script = strings.ReplaceAll(script, "${repo_name}", selectedRepo.Name)
		script = strings.ReplaceAll(script, "${repo}", selectedRepo.Name)
	}

	// Execute the script
	cmd := exec.Command("sh", "-c", script)
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to execute script: %w", err)
	}

	// Don't wait for the command - let it run independently
	return nil
}

func (m Model) renderHelp() string {
	help := []string{
		"Navigation: ↑↓ or j/k   Switch pane: tab or h/l   Add: +   Delete: -   Notes: n   Script: s   Yank: y   Open: Enter   Quit: q or ctrl+c",
	}
	return helpStyle.Render(strings.Join(help, " • "))
}
