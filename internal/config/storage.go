package config

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	notesDirName   = "notes"
	scriptsDirName = "scripts"
	nameSeparator  = "__"
)

var storageNameCleaner = regexp.MustCompile(`[^a-zA-Z0-9-_]+`)

func ConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "workman"), nil
}

func sanitizeStorageName(name string) string {
	name = strings.TrimSpace(name)
	name = storageNameCleaner.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	if name == "" {
		return "unnamed"
	}
	return strings.ToLower(name)
}

func repoScriptPath(repoName string) (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	fileName := sanitizeStorageName(repoName)
	return filepath.Join(configDir, scriptsDirName, fileName), nil
}

func worktreeNotesPath(repoName, worktreeName string) (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	fileName := sanitizeStorageName(repoName) + nameSeparator + sanitizeStorageName(worktreeName)
	return filepath.Join(configDir, notesDirName, fileName), nil
}

func HasRepoScript(repoName string) (bool, error) {
	path, err := repoScriptPath(repoName)
	if err != nil {
		return false, err
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.Size() > 0, nil
}

func GetRepoScript(repoName string) (string, error) {
	path, err := repoScriptPath(repoName)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func SaveRepoScript(repoName, script string) error {
	script = strings.TrimSpace(script)
	if script == "" {
		return DeleteRepoScript(repoName)
	}
	path, err := repoScriptPath(repoName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(script), 0644)
}

func DeleteRepoScript(repoName string) error {
	path, err := repoScriptPath(repoName)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func GetWorktreeNotes(repoName, worktreeName string) (string, error) {
	path, err := worktreeNotesPath(repoName, worktreeName)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func SaveWorktreeNotes(repoName, worktreeName, notes string) error {
	notes = strings.TrimSpace(notes)
	if notes == "" {
		return DeleteWorktreeNotes(repoName, worktreeName)
	}
	path, err := worktreeNotesPath(repoName, worktreeName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(notes), 0644)
}

func DeleteWorktreeNotes(repoName, worktreeName string) error {
	path, err := worktreeNotesPath(repoName, worktreeName)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
