package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DialogType int

const (
	DialogNone DialogType = iota
	DialogAddRepo
	DialogAddWorktree
	DialogConfirmDelete
)

type AddRepoDialog struct {
	focusIndex int
	inputs     []textinput.Model
}

func NewAddRepoDialog() AddRepoDialog {
	inputs := make([]textinput.Model, 2)

	// Name input
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "my-project"
	inputs[0].Focus()
	inputs[0].CharLimit = 50
	inputs[0].Width = 50

	// Path/URL input
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "/path/to/repo or https://github.com/user/repo.git"
	inputs[1].CharLimit = 200
	inputs[1].Width = 50

	return AddRepoDialog{
		focusIndex: 0,
		inputs:     inputs,
	}
}

// inferRepoType determines if the path is a remote URL or local path
func inferRepoType(path string) string {
	path = strings.TrimSpace(path)
	if strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://") ||
		strings.HasPrefix(path, "git@") ||
		strings.HasPrefix(path, "ssh://") {
		return "remote"
	}
	return "local"
}

func (d *AddRepoDialog) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Handle key navigation
			switch s {
			case "enter", "down", "tab":
				d.focusIndex++
			case "up", "shift+tab":
				d.focusIndex--
			}

			// Wrap around
			if d.focusIndex > len(d.inputs)-1 {
				d.focusIndex = 0
			} else if d.focusIndex < 0 {
				d.focusIndex = len(d.inputs) - 1
			}

			// Update focus
			for i := 0; i < len(d.inputs); i++ {
				if i == d.focusIndex {
					cmds = append(cmds, d.inputs[i].Focus())
				} else {
					d.inputs[i].Blur()
				}
			}

			return tea.Batch(cmds...)
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	d.inputs[d.focusIndex], cmd = d.inputs[d.focusIndex].Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (d *AddRepoDialog) View() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("Add Repository"))
	b.WriteString("\n\n")

	// Name
	b.WriteString(itemStyle.Render("Name:"))
	b.WriteString("\n")
	b.WriteString(d.inputs[0].View())
	b.WriteString("\n\n")

	// Path/URL
	b.WriteString(itemStyle.Render("Path or URL:"))
	b.WriteString("\n")
	b.WriteString(d.inputs[1].View())
	b.WriteString("\n")

	// Show hint about auto-detection
	path := strings.TrimSpace(d.inputs[1].Value())
	if path != "" {
		repoType := inferRepoType(path)
		hint := infoStyle.Render(fmt.Sprintf("  → will be detected as: %s", repoType))
		b.WriteString("\n")
		b.WriteString(hint)
	}
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("Enter: next field  •  Ctrl+S: save  •  Esc: cancel"))

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Width(60)

	return dialogStyle.Render(b.String())
}

func (d *AddRepoDialog) GetValues() (name, repoType, path string) {
	name = strings.TrimSpace(d.inputs[0].Value())
	path = strings.TrimSpace(d.inputs[1].Value())
	repoType = inferRepoType(path)
	return
}

func (d *AddRepoDialog) IsValid() (bool, string) {
	name, _, path := d.GetValues()

	if name == "" {
		return false, "Name is required"
	}

	if path == "" {
		return false, "Path/URL is required"
	}

	return true, ""
}

func (d *AddRepoDialog) Reset() {
	for i := range d.inputs {
		d.inputs[i].SetValue("")
	}
	d.focusIndex = 0
	d.inputs[0].Focus()
	for i := 1; i < len(d.inputs); i++ {
		d.inputs[i].Blur()
	}
}

// AddWorktreeDialog handles the worktree creation dialog
type AddWorktreeDialog struct {
	focusIndex int
	input      textinput.Model
}

func NewAddWorktreeDialog() AddWorktreeDialog {
	input := textinput.New()
	input.Placeholder = "feature/my-feature or bugfix/issue-123"
	input.Focus()
	input.CharLimit = 100
	input.Width = 50

	return AddWorktreeDialog{
		focusIndex: 0,
		input:      input,
	}
}

func (d *AddWorktreeDialog) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)
	return cmd
}

func (d *AddWorktreeDialog) View() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("Add Worktree"))
	b.WriteString("\n\n")

	// Branch name
	b.WriteString(itemStyle.Render("Branch name:"))
	b.WriteString("\n")
	b.WriteString(d.input.View())
	b.WriteString("\n\n")

	// Show hint
	hint := infoStyle.Render("If branch doesn't exist, it will be created automatically")
	b.WriteString(hint)
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("Ctrl+S: create  •  Esc: cancel"))

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Width(60)

	return dialogStyle.Render(b.String())
}

func (d *AddWorktreeDialog) GetBranchName() string {
	return strings.TrimSpace(d.input.Value())
}

func (d *AddWorktreeDialog) IsValid() (bool, string) {
	branch := d.GetBranchName()

	if branch == "" {
		return false, "Branch name is required"
	}

	// Basic validation for branch name
	if strings.Contains(branch, " ") {
		return false, "Branch name cannot contain spaces"
	}

	return true, ""
}

func (d *AddWorktreeDialog) Reset() {
	d.input.SetValue("")
	d.input.Focus()
}

// ConfirmDeleteDialog handles the confirmation for deleting a worktree
type ConfirmDeleteDialog struct {
	worktreeName string
	branchName   string
}

func NewConfirmDeleteDialog(worktreeName, branchName string) ConfirmDeleteDialog {
	return ConfirmDeleteDialog{
		worktreeName: worktreeName,
		branchName:   branchName,
	}
}

func (d *ConfirmDeleteDialog) View() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("⚠ Confirm Delete"))
	b.WriteString("\n\n")

	warning := fmt.Sprintf("Delete worktree '%s' and branch '%s'?", d.worktreeName, d.branchName)
	b.WriteString(itemStyle.Render(warning))
	b.WriteString("\n\n")

	hint := infoStyle.Render("⚠ This will delete the branch even if unmerged!")
	b.WriteString(hint)
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("y: confirm  •  n/Esc: cancel"))

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#EF4444")).
		Padding(1, 2).
		Width(55)

	return dialogStyle.Render(b.String())
}

type errorMsg struct {
	err string
}

func (e errorMsg) Error() string {
	return e.err
}

func showError(msg string) tea.Cmd {
	return func() tea.Msg {
		return errorMsg{err: msg}
	}
}
