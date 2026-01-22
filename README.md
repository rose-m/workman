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

