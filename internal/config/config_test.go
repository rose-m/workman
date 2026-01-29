package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestSave_RepositoryFieldsPersisted(t *testing.T) {
	// Setup: create a temp config directory
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.toml")

	// Reset viper for this test
	viper.Reset()
	viper.SetConfigFile(configFile)
	viper.SetConfigType("toml")

	cfg := &Config{
		RootDirectory: "/test/workspace",
		YankTemplate:  "${worktree_path}",
		Repositories: []Repository{
			{
				Name: "test-repo",
				Path: "/test/path",
				Type: "local",
				URL:  "",
			},
		},
	}

	// Save
	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Read the raw file to verify the format
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Verify script and notes are not persisted in config
	contentStr := string(content)
	if strings.Contains(contentStr, "post_create_script") {
		t.Errorf("Found 'post_create_script' in config, should be stored in files:\n%s", contentStr)
	}
	if strings.Contains(contentStr, "worktree_notes") {
		t.Errorf("Found 'worktree_notes' in config, should be stored in files:\n%s", contentStr)
	}

	// Reset viper and reload
	viper.Reset()
	viper.SetConfigFile(configFile)
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var loaded Config
	if err := viper.Unmarshal(&loaded); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify repository data persisted and loaded correctly
	if len(loaded.Repositories) != 1 {
		t.Fatalf("Expected 1 repository, got %d", len(loaded.Repositories))
	}
	if loaded.Repositories[0].PostCreateScript != "" {
		t.Errorf("Expected PostCreateScript to be empty, got: %q", loaded.Repositories[0].PostCreateScript)
	}

	// Verify other fields
	if loaded.Repositories[0].Name != "test-repo" {
		t.Errorf("Name not persisted correctly: %s", loaded.Repositories[0].Name)
	}
	if loaded.Repositories[0].Path != "/test/path" {
		t.Errorf("Path not persisted correctly: %s", loaded.Repositories[0].Path)
	}
	if loaded.Repositories[0].Type != "local" {
		t.Errorf("Type not persisted correctly: %s", loaded.Repositories[0].Type)
	}

}

func TestRepoScriptStorage(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	t.Cleanup(func() {
		_ = os.Setenv("HOME", oldHome)
	})
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	script := "echo hello"
	if err := SaveRepoScript("test-repo", script); err != nil {
		t.Fatalf("SaveRepoScript failed: %v", err)
	}

	loaded, err := GetRepoScript("test-repo")
	if err != nil {
		t.Fatalf("GetRepoScript failed: %v", err)
	}
	if loaded != script {
		t.Errorf("Script mismatch. Expected %q, got %q", script, loaded)
	}

	if err := SaveRepoScript("test-repo", ""); err != nil {
		t.Fatalf("SaveRepoScript delete failed: %v", err)
	}

	loaded, err = GetRepoScript("test-repo")
	if err != nil {
		t.Fatalf("GetRepoScript after delete failed: %v", err)
	}
	if loaded != "" {
		t.Errorf("Expected script to be deleted, got %q", loaded)
	}
}

func TestWorktreeNotesStorage(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	t.Cleanup(func() {
		_ = os.Setenv("HOME", oldHome)
	})
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	notes := "some notes"
	if err := SaveWorktreeNotes("test-repo", "feature-branch", notes); err != nil {
		t.Fatalf("SaveWorktreeNotes failed: %v", err)
	}

	loaded, err := GetWorktreeNotes("test-repo", "feature-branch")
	if err != nil {
		t.Fatalf("GetWorktreeNotes failed: %v", err)
	}
	if loaded != notes {
		t.Errorf("Notes mismatch. Expected %q, got %q", notes, loaded)
	}

	if err := SaveWorktreeNotes("test-repo", "feature-branch", ""); err != nil {
		t.Fatalf("SaveWorktreeNotes delete failed: %v", err)
	}

	loaded, err = GetWorktreeNotes("test-repo", "feature-branch")
	if err != nil {
		t.Fatalf("GetWorktreeNotes after delete failed: %v", err)
	}
	if loaded != "" {
		t.Errorf("Expected notes to be deleted, got %q", loaded)
	}
}
