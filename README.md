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
make install    # Build and install to ~/.local/bin
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

# Template for the 'y' (yank) command
# Variables: ${repo_name}, ${branch_name}, ${worktree_path}, ${worktree_name}
yank_template = 'wt "${repo_name} - ${branch_name}"; cd "${worktree_path}"'

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

## Terminal Integration

Workman can execute a custom script when you press `Enter` on a worktree. This allows you to open terminals, create splits, or run any command with the worktree path.

Configure this in your `~/.config/workman/config.toml`:

### Simple Examples

**tmux - Open split:**
```toml
enter_script = "tmux split-window -h -c '${worktree_path}'"
```

**Ghostty (macOS) - Open new window:**
```toml
enter_script = "open -na ghostty --args --working-directory='${worktree_path}'"
```

**Ghostty (macOS) - Create split using Command+D keybinding:**
```toml
enter_script = "osascript -e 'tell application \"System Events\" to keystroke \"d\" using command down' && sleep 0.2 && osascript -e 'tell application \"System Events\" to keystroke \"cd ${worktree_path}\" & return'"
```

**Ghostty (Linux) - Open new window:**
```toml
enter_script = "ghostty --working-directory='${worktree_path}' &"
```

### Advanced Examples: Running Commands After Opening

**Open Ghostty window and start Claude Code:**
```toml
# macOS
enter_script = "open -na ghostty --args --working-directory='${worktree_path}' -e 'claude' &"

# Linux
enter_script = "ghostty --working-directory='${worktree_path}' -e 'claude' &"
```

**tmux split and start Claude Code:**
```toml
enter_script = "tmux split-window -h -c '${worktree_path}' claude"
```

**iTerm2 - Open new window, cd, and start editor:**
```toml
enter_script = "osascript -e 'tell application \"iTerm\" to create window with default profile command \"cd ${worktree_path} && nvim\"'"
```

**Alacritty - Open window and start shell with command:**
```toml
enter_script = "alacritty --working-directory '${worktree_path}' -e zsh -c 'clear && pwd && zsh' &"
```

**Terminal.app (macOS) - Open tab and run command:**
```toml
enter_script = "osascript -e 'tell application \"Terminal\" to do script \"cd ${worktree_path} && claude\"'"
```

**Open split in current terminal and run git status:**
```toml
# In tmux
enter_script = "tmux split-window -h -c '${worktree_path}' 'git status; exec $SHELL'"

# In kitty
enter_script = "kitty @ launch --type=tab --cwd='${worktree_path}' --title '${repo_name}/${branch}' sh -c 'git status; exec $SHELL'"
```

### Available Variables

- `${worktree_path}` or `${path}` - Full path to the worktree
- `${branch_name}` or `${branch}` - Git branch name
- `${repo_name}` or `${repo}` - Repository name

### Tips

- Leave scripts empty to disable the keybinding (pressing the key will show an error prompting configuration)
- Scripts are executed with `sh -c`, so you can chain commands with `&&` or `;`
- Add `&` at the end to run commands in the background (non-blocking)
- Use `exec $SHELL` at the end to keep terminal open after command finishes
- Wrap commands in quotes when they contain spaces or special characters
- Use `-e` flag in terminals to execute commands directly

## Keyboard Shortcuts

### Main View
- `↑/↓` or `j/k` - Navigate items in the active pane
- `Tab` or `h/l` - Switch between repositories and worktrees panes (h=left, l=right)
- `+` - Add repository (when in repos pane) or add worktree (when in worktrees pane)
- `-` - Delete worktree (when in worktrees pane, with confirmation)
- `n` - Edit notes for selected worktree
- `s` - Edit post-create script for selected repository
- `y` - Yank (copy) command to clipboard (when worktree is selected)
- `Enter` - Execute configured script for worktree (see Terminal Integration below)
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
6. ~~Add ability to open worktree in editor/terminal~~ ✅ (implemented via configurable scripts)
7. Add Git stash management
8. Add search/filter for repositories and worktrees
