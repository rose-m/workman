---
date: 2026-01-22T08:09:59Z
git_commit: 4cf11bb26a9e9c9ac7acb8bc9b2a4fa0b7fee8b4
branch: main
topic: "Terminal Spawning for Worktrees with Ghostty Support"
tags: [research, codebase, terminal, ghostty, worktree, external-process]
status: complete
last_updated: 2026-01-22
---

# Research: Terminal Spawning for Worktrees with Ghostty Support

**Date**: 2026-01-22T08:09:59Z
**Git Commit**: 4cf11bb26a9e9c9ac7acb8bc9b2a4fa0b7fee8b4
**Branch**: main

## Research Question

How to spawn a terminal for a worktree that automatically opens the worktree folder in a new window/tab (depending on terminal application support). Initial focus on Ghostty terminal emulator with potential future support for other terminals. Ideally track opened tabs/windows to return to existing ones instead of spawning new ones.

## Summary

The workman codebase currently has no terminal spawning functionality - this would be a new feature. The README explicitly lists "Add ability to open worktree in editor/terminal" as a planned feature. Ghostty provides platform-specific IPC mechanisms: D-Bus on Linux (`ghostty +new-window`) and AppleScript/System Events workarounds on macOS. Window/tab tracking is limited - Ghostty does not expose external APIs for enumerating or identifying windows/tabs, though window titles can be set for potential identification.

## Detailed Findings

### Current Codebase Architecture

#### Language and Framework
- **Language**: Go 1.25
- **TUI Framework**: Bubble Tea (Elm Architecture pattern)
- **Styling**: Lipgloss
- **Config**: Viper with TOML files

#### Directory Structure
```
/Users/michael.rose/Software/private/workman/
├── main.go                          # Application entry point
├── go.mod                           # Module definition
├── internal/
│   ├── config/config.go             # Configuration loading/saving
│   ├── state/state.go               # Application state management
│   ├── git/repository.go            # Git worktree operations
│   └── ui/
│       ├── model.go                 # Main Bubble Tea model
│       ├── styles.go                # Lipgloss styling
│       └── dialog.go                # Dialog components
```

#### Worktree Data Model

**`internal/state/state.go:5-9`**
```go
type Worktree struct {
    Name   string  // Directory name (derived from path)
    Branch string  // Branch name associated with the worktree
    Path   string  // Full filesystem path
}
```

The worktree `Path` field contains the full filesystem path needed for opening a terminal in that directory.

#### Current External Process Execution Pattern

The codebase uses `exec.Command` for git operations in `internal/git/repository.go`:

- Line 31: `exec.Command("git", "worktree", "list", "--porcelain")`
- Line 135, 144, 151: `exec.Command("git", "worktree", "add", ...)`
- Line 175: `exec.Command("git", "worktree", "remove", ...)`
- Line 187: `exec.Command("git", "branch", "-D", ...)`

**Pattern used**:
```go
cmd := exec.Command("git", "worktree", "list", "--porcelain")
cmd.Dir = repoPath
output, err := cmd.Output()
```

#### Missing Functionality

Per README.md line 162: **"Add ability to open worktree in editor/terminal"** is explicitly listed as a future feature.

Currently there is:
- No terminal spawning code
- No editor integration
- No configuration for terminal/editor preferences
- No keybinding for opening worktrees externally

### Ghostty Terminal Capabilities

#### Platform-Specific IPC

Ghostty uses platform-native IPC mechanisms rather than a unified cross-platform API.

##### Linux (D-Bus)

**Fast window creation via D-Bus**:
```bash
# Creates new window via D-Bus (~20ms latency vs ~300ms for new process)
ghostty +new-window
```

- D-Bus bus name: `com.mitchellh.ghostty`
- If Ghostty isn't running, D-Bus activation starts it automatically
- **Limitation**: `--working-directory` flag is ignored when using `+new-window` in daemon mode

**Configuration for daemon mode**:
```
quit-after-last-window-closed = false
# or
quit-after-last-window-closed-delay = 5m
```

##### macOS

**No direct CLI command for new window in existing instance**. Workarounds:

**AppleScript via System Events**:
```applescript
#!/usr/bin/osascript
set termName to "Ghostty"
tell application termName
    if it is running
        tell application "System Events" to tell process termName
            click menu item "New Window" of menu "File" of menu bar 1
        end tell
    else
        activate
    end if
end tell
```

**Opening new instance with specific directory**:
```bash
open -na ghostty --args --title=my-terminal --working-directory="$(pwd)"
```

Note: This creates a new Ghostty process/instance, not a window in an existing instance.

#### Opening Terminal in Specific Directory

**CLI flag**:
```bash
ghostty --working-directory=/path/to/directory
```

**Important**: Paths must be absolute (e.g., `~/Location` won't work).

**Configuration options**:
```
working-directory = /path/to/default
working-directory = inherit
window-inherit-working-directory = true
```

#### Window/Tab Identification and Tracking

**Current capabilities**:

1. **Window Titles**: Can set via CLI or config
   ```bash
   ghostty --title="worktree-name"
   ```

2. **OSC Sequences** (from within terminal):
   - OSC 2: `printf '\033]2;My Title\007'` - Set window title
   - OSC 7: `printf '\033]7;file:///path/to/dir\007'` - Set working directory

**What's NOT available**:
- No window/tab IDs exposed externally
- No API to enumerate or identify specific windows/tabs from outside Ghostty
- No tab switching API from external processes
- macOS tabs appear as separate windows to window managers

#### Focusing Existing Windows

**No dedicated IPC mechanism for focusing specific windows/tabs**. This is being discussed in [GitHub Discussion #2353](https://github.com/ghostty-org/ghostty/discussions/2353).

**macOS workaround**:
```applescript
tell application "Ghostty" to activate
```

**Community solutions**: Hammerspoon (macOS) for querying windows via accessibility APIs.

### Available Ghostty CLI Commands

| Command | Description |
|---------|-------------|
| `ghostty` | Launch new Ghostty instance |
| `ghostty +new-window` | Create window via D-Bus (Linux only) |
| `ghostty -e <command>` | Execute a command |
| `ghostty --working-directory=/path` | Set working directory |
| `ghostty --title="Title"` | Set window title |

### Implementation Considerations

#### Terminal Abstraction Layer

To support multiple terminals in the future, an abstraction is needed:

```go
type TerminalLauncher interface {
    // OpenInDirectory opens a terminal window/tab in the specified directory
    OpenInDirectory(path string, options LaunchOptions) (WindowHandle, error)
    // FocusWindow focuses an existing window if possible
    FocusWindow(handle WindowHandle) error
    // IsWindowOpen checks if a tracked window is still open
    IsWindowOpen(handle WindowHandle) bool
}

type LaunchOptions struct {
    Title       string  // Window title for identification
    ReuseWindow bool    // Try to reuse existing window
}

type WindowHandle struct {
    // Platform-specific identifier
    // Could be PID, window title, or other identifier
}
```

#### Window Tracking Challenges

1. **No external enumeration API**: Cannot query Ghostty for open windows
2. **PID-based tracking**: Could track spawned process PIDs, but:
   - On macOS with `open -na`, the PID is for the `open` command, not Ghostty
   - D-Bus mode on Linux doesn't spawn new processes
3. **Title-based identification**: Most reliable approach
   - Set unique title per worktree: `ghostty --title="workman:<worktree-path>"`
   - Use OS-level window enumeration (platform-specific)
4. **Graceful degradation**: If window tracking fails, just open new terminal

#### Platform-Specific Implementation

**Linux with D-Bus**:
```go
// Fast window creation but can't set working directory in daemon mode
cmd := exec.Command("ghostty", "+new-window")
// Workaround: Send cd command after window opens
```

**macOS**:
```go
// Opens new instance with directory
cmd := exec.Command("open", "-na", "ghostty", "--args",
    "--title="+title,
    "--working-directory="+path)
```

#### Configuration Extension

Current config structure in `internal/config/config.go`:

```go
type Config struct {
    RootDirectory string       `mapstructure:"root_directory"`
    Repositories  []Repository `mapstructure:"repositories"`
    // Potential additions:
    // Terminal      string       `mapstructure:"terminal"`  // "ghostty", "iterm", etc.
    // TerminalOpts  map[string]string `mapstructure:"terminal_options"`
}
```

## Code References

- `internal/state/state.go:5-9` - Worktree struct with Path field
- `internal/state/state.go:14` - SelectedWTIndex for current selection
- `internal/git/repository.go:30-39` - Pattern for exec.Command usage
- `internal/ui/model.go:269` - Access to selected worktree path
- `internal/config/config.go:17-20` - Config struct to extend
- `README.md:162` - TODO item for terminal/editor integration

## Architecture Documentation

### Current Patterns

1. **External commands**: Use `exec.Command` with `cmd.Dir` for working directory
2. **Configuration**: Viper with TOML, structs with mapstructure tags
3. **UI updates**: Return `tea.Cmd` for async operations
4. **State management**: Centralized in `state.AppState`

### Suggested File Organization

New files to create:
- `internal/terminal/launcher.go` - Terminal abstraction interface
- `internal/terminal/ghostty.go` - Ghostty-specific implementation
- `internal/terminal/tracker.go` - Window tracking logic (optional)

## Open Questions

1. **Keybinding**: What key should spawn terminal? Likely `Enter` or a new binding like `t`
2. **Window tracking scope**: How important is returning to existing windows vs always spawning new?
3. **Linux D-Bus working directory**: The `--working-directory` flag is ignored in daemon mode - acceptable limitation or need workaround?
4. **macOS behavior**: Is spawning new instances (vs reusing) acceptable?
5. **Other terminals**: Which terminals should be prioritized after Ghostty? (iTerm2, Terminal.app, Alacritty, Kitty, etc.)

## External References

- [Ghostty Documentation](https://ghostty.org/docs/)
- [Ghostty Configuration Reference](https://ghostty.org/docs/config/reference)
- [Ghostty Action Reference](https://ghostty.org/docs/config/keybind/reference)
- [Ghostty Systemd/D-Bus (Linux)](https://ghostty.org/docs/linux/systemd)
- [Ghostty Shell Integration](https://ghostty.org/docs/features/shell-integration)
- [Scripting API Discussion - GitHub #2353](https://github.com/ghostty-org/ghostty/discussions/2353)
- [Open New Windows on macOS - GitHub #6053](https://github.com/ghostty-org/ghostty/discussions/6053)
