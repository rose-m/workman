package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/michael-rose/workman/internal/state"
)

// sanitizeName converts a name to only contain alphanumeric characters and dashes
func sanitizeName(name string) string {
	// Replace any non-alphanumeric characters (except dash) with dash
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-]+`)
	sanitized := reg.ReplaceAllString(name, "-")

	// Remove leading/trailing dashes and collapse multiple dashes
	sanitized = strings.Trim(sanitized, "-")
	reg = regexp.MustCompile(`-+`)
	sanitized = reg.ReplaceAllString(sanitized, "-")

	// Convert to lowercase for consistency
	return strings.ToLower(sanitized)
}

// ListWorktrees lists all worktrees for a given repository path
func ListWorktrees(repoPath string) ([]state.Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	return parseWorktreeList(string(output))
}

// parseWorktreeList parses the output of git worktree list --porcelain
func parseWorktreeList(output string) ([]state.Worktree, error) {
	var worktrees []state.Worktree
	lines := strings.Split(strings.TrimSpace(output), "\n")

	var currentPath, currentBranch string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			// Empty line separates worktrees
			if currentPath != "" {
				name := filepath.Base(currentPath)
				worktrees = append(worktrees, state.Worktree{
					Name:   name,
					Branch: currentBranch,
					Path:   currentPath,
				})
				currentPath = ""
				currentBranch = ""
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			currentPath = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch ") {
			currentBranch = strings.TrimPrefix(line, "branch refs/heads/")
		} else if strings.HasPrefix(line, "detached") {
			currentBranch = "detached HEAD"
		}
	}

	// Handle last worktree if no trailing newline
	if currentPath != "" {
		name := filepath.Base(currentPath)
		worktrees = append(worktrees, state.Worktree{
			Name:   name,
			Branch: currentBranch,
			Path:   currentPath,
		})
	}

	return worktrees, nil
}

// GetCurrentBranch returns the current branch of the repository
func GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// BranchExists checks if a branch exists locally
func BranchExists(repoPath, branch string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+branch)
	cmd.Dir = repoPath
	err := cmd.Run()
	if err != nil {
		// Branch doesn't exist
		return false, nil
	}
	return true, nil
}

// AddWorktree creates a new worktree for the repository in the root directory
// If the branch doesn't exist, it creates it based on main/master (remote) or current branch (local)
// Worktree will be created at: <rootDir>/<sanitizedRepoName>-<sanitizedBranchName>
func AddWorktree(repoPath, rootDir, repoName, branch string, isRemote bool) error {
	// Check if branch exists
	exists, err := BranchExists(repoPath, branch)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	// Sanitize names
	sanitizedRepo := sanitizeName(repoName)
	sanitizedBranch := sanitizeName(branch)

	// Determine worktree path: <rootDir>/<reponame>-<branchname>
	worktreeName := fmt.Sprintf("%s-%s", sanitizedRepo, sanitizedBranch)
	worktreePath := filepath.Join(rootDir, worktreeName)

	// Check if path already exists
	if _, err := os.Stat(worktreePath); err == nil {
		return fmt.Errorf("path already exists: %s", worktreePath)
	}

	var cmd *exec.Cmd
	if exists {
		// Branch exists, just create worktree
		cmd = exec.Command("git", "worktree", "add", worktreePath, branch)
	} else {
		// Branch doesn't exist, create it
		if isRemote {
			// For remote repos, try to base on origin/main or origin/master
			baseBranch := "origin/main"
			if !remoteBranchExists(repoPath, "origin/main") {
				baseBranch = "origin/master"
			}
			cmd = exec.Command("git", "worktree", "add", "-b", branch, worktreePath, baseBranch)
		} else {
			// For local repos, base on current branch
			currentBranch, err := GetCurrentBranch(repoPath)
			if err != nil {
				return fmt.Errorf("failed to get current branch: %w", err)
			}
			cmd = exec.Command("git", "worktree", "add", "-b", branch, worktreePath, currentBranch)
		}
	}

	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add worktree: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// remoteBranchExists checks if a remote branch exists
func remoteBranchExists(repoPath, branch string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", branch)
	cmd.Dir = repoPath
	err := cmd.Run()
	return err == nil
}

// RemoveWorktree removes a worktree forcefully
func RemoveWorktree(repoPath, worktreePath string) error {
	// Use --force to remove even if there are uncommitted changes
	cmd := exec.Command("git", "worktree", "remove", "--force", worktreePath)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// DeleteBranch deletes a branch forcefully (even if unmerged)
func DeleteBranch(repoPath, branch string) error {
	// Use -D (force delete) to remove even if unmerged
	cmd := exec.Command("git", "branch", "-D", branch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// DeleteRepository removes the entire repository directory from disk
func DeleteRepository(repoPath string) error {
	// Check if directory exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository directory does not exist: %s", repoPath)
	}

	// Remove the entire directory tree
	if err := os.RemoveAll(repoPath); err != nil {
		return fmt.Errorf("failed to delete repository directory: %w", err)
	}

	return nil
}

// ListBranches lists all local and remote branches for a repository
func ListBranches(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "branch", "-a", "--format=%(refname:short)")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var branches []string
	seen := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove "origin/" prefix for remote branches
		branch := strings.TrimPrefix(line, "origin/")

		// Skip HEAD reference
		if strings.Contains(branch, "HEAD") {
			continue
		}

		// Deduplicate branches
		if !seen[branch] {
			branches = append(branches, branch)
			seen[branch] = true
		}
	}

	return branches, nil
}

// ExecutePostCreateScript executes a bash script after worktree creation
// The script receives two arguments: repo path and worktree path
func ExecutePostCreateScript(script, repoPath, worktreePath string) error {
	if script == "" {
		return nil
	}

	cmd := exec.Command("bash", "-c", script, "--", repoPath, worktreePath)
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("script execution failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// CloneRepository clones a remote repository to the specified path
func CloneRepository(url, targetPath string) error {
	// Check if target path already exists
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("target path already exists: %s", targetPath)
	}

	// Create parent directory if it doesn't exist
	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Clone the repository
	cmd := exec.Command("git", "clone", url, targetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w\nOutput: %s", err, string(output))
	}

	return nil
}
