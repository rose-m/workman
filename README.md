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
root_directory = "/path/to/your/workspace"

[[repositories]]
name = "my-repo"
type = "local"
path = "/path/to/existing/repo"
url = ""
```

## Keyboard Shortcuts

- `↑/↓` or `j/k` - Navigate items in the active pane
- `Tab` - Switch between repositories and worktrees panes
- `+` - Add repository or worktree (TODO)
- `q` or `Ctrl+C` - Quit

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

This is a basic scaffolding with:
- ✅ Split pane UI (repositories left, worktrees right)
- ✅ Keyboard navigation (arrows, j/k, tab)
- ✅ Configuration file management
- ✅ Basic state management
- ⏳ Git operations (placeholder implementations)
- ⏳ Add/remove repositories
- ⏳ Create/manage worktrees

## Next Steps

1. Implement actual git operations using `go-git` or git CLI
2. Add dialogs for adding repositories and worktrees
3. Add repository cloning functionality
4. Add worktree creation/deletion
5. Display more repository and worktree details
6. Add status indicators (clean, dirty, ahead/behind)
