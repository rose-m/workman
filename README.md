# Workman

The goal of this terminal UI app is to have an easy way of managing Git repositories and especially its worktrees.

Repositories can be added as:
* Remote URLs - they will be cloned to a configurable root directory
* Existing local repositories - they will just be referenced

The TUI is split:
* List of repositories to the left
* List of worktrees for the selected repository on the right

There will be extensive keyboard support:
* arrow up / down OR j/k will select previous / next item in a list
* "+" will allow to add a repo or new worktree

## Tech Stack

- **Language:** Go
- **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) - A powerful TUI framework based on The Elm Architecture
- **UI Components:** [Bubbles](https://github.com/charmbracelet/bubbles) - Common UI components
- **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling and layout
- **Config Management:** [Viper](https://github.com/spf13/viper) - TOML-based configuration

## Building

Using Make (recommended):
```bash
make build      # Build the binary
make fmt        # Format code
make lint       # Run linter
make all        # Format, lint, and build
```

Or directly with Go:
```bash
go build -o workman .
```

## Running

```bash
make run        # Build and run
# or
./workman
```

## Development

```bash
make help           # Show all available targets
make install-tools  # Install golangci-lint
make tidy          # Tidy Go modules
make test          # Run tests
make clean         # Clean build artifacts
```

## Configuration

Configuration is stored in `~/.config/workman/config.toml` and will be created automatically on first run with default values.

See `config.example.toml` for a complete example.

### Config Structure

```toml
# Root directory where worktrees will be created
# Defaults to ~/workspace
root_directory = "/path/to/your/workspace"

[[repositories]]
name = "my-repo"
type = "local"
path = "/path/to/existing/repo"
url = ""
```

**Important:** The `root_directory` is where all worktrees will be created with the naming pattern `<reponame>-<branchname>`.

You can also add repositories directly through the UI by pressing `+` when in the repositories pane (left side). The type will be automatically detected:
- URLs starting with `http://`, `https://`, `git@`, or `ssh://` are detected as **remote**
- All other paths are detected as **local**

## Keyboard Shortcuts

### Main View
- `↑/↓` or `j/k` - Navigate items in the active pane
- `Tab` - Switch between repositories and worktrees panes
- `+` - Add repository (when in repos pane) or add worktree (when in worktrees pane)
- `-` - Delete worktree (when in worktrees pane, with confirmation)
- `q` or `Ctrl+C` - Quit

### Add Repository Dialog
- `Enter` / `Tab` / `↓` - Move to next field
- `Shift+Tab` / `↑` - Move to previous field
- `Ctrl+S` - Save repository
- `Esc` - Cancel

**Note:** Repository type (local vs remote) is automatically detected based on the path/URL you enter.

### Add Worktree Dialog
- Type branch name
- `Ctrl+S` - Create worktree
- `Esc` - Cancel

**Worktree Creation:**
- Worktrees are created in the configured `root_directory`
- Path format: `<root_directory>/<reponame>-<branchname>`
- Both repo and branch names are sanitized (alphanumeric + dashes only, lowercase)
- Example: repo "My Repo" + branch "feature/new-thing" → `~/workspace/my-repo-feature-new-thing`

**Branch Creation:**
- If the branch doesn't exist:
  - For **remote** repos: new branch is created based on `origin/main` (or `origin/master`)
  - For **local** repos: new branch is created based on the currently checked out branch

### Delete Worktree Confirmation
- `y` - Confirm deletion
- `n` or `Esc` - Cancel

**Warning:** Deleting a worktree will:
- Remove the worktree directory and all files
- Delete the branch (even if unmerged!)
- Cannot delete the main worktree (the first one in the list)

## Project Structure

```
workman/
├── main.go                 # Entry point
├── internal/
│   ├── config/            # Configuration management
│   ├── git/               # Git operations (TODO)
│   ├── state/             # Application state
│   └── ui/                # Bubble Tea UI components
├── config.example.toml    # Example configuration
└── README.md
```

## Current Status

This is a working TUI with:
- ✅ Split pane UI (repositories left, worktrees right)
- ✅ Keyboard navigation (arrows, j/k, tab)
- ✅ Configuration file management (TOML)
- ✅ Basic state management
- ✅ Add repositories (interactive dialog with auto-type detection)
- ✅ List/display worktrees for selected repository
- ✅ Create worktrees with automatic branch creation
- ✅ Delete worktrees with confirmation (removes branch even if unmerged)
- ✅ Git operations (list, create, delete worktrees and branches)
- ⏳ Delete repositories
- ⏳ Clone remote repositories

## Next Steps

1. Add repository cloning functionality for remote repos
2. Add repository deletion from config (with confirmation)
3. Display more repository details (current branch, status)
4. Display more worktree details (commit hash, ahead/behind status)
5. Add status indicators (clean, dirty, ahead/behind)
6. Add ability to open worktree in editor/terminal
7. Add Git stash management
8. Add search/filter for repositories and worktrees
