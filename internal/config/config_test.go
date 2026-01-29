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

	// Create config with a repository that has a post_create_script
	testScript := "echo 'Hello from test script'\necho 'Worktree path: $2'"
	cfg := &Config{
		RootDirectory: "/test/workspace",
		YankTemplate:  "${worktree_path}",
		Repositories: []Repository{
			{
				Name:             "test-repo",
				Path:             "/test/path",
				Type:             "local",
				URL:              "",
				PostCreateScript: testScript,
			},
		},
		WorktreeNotes: []WorktreeNote{
			{
				Path:  "/test/worktree",
				Notes: "Test notes",
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

	// Verify snake_case keys are used (not PascalCase)
	contentStr := string(content)
	if !strings.Contains(contentStr, "post_create_script") {
		t.Errorf("Expected 'post_create_script' key in config, got:\n%s", contentStr)
	}
	if strings.Contains(contentStr, "PostCreateScript") {
		t.Errorf("Found 'PostCreateScript' (PascalCase) in config, should be 'post_create_script':\n%s", contentStr)
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

	// Verify the script was persisted and loaded correctly
	if len(loaded.Repositories) != 1 {
		t.Fatalf("Expected 1 repository, got %d", len(loaded.Repositories))
	}

	if loaded.Repositories[0].PostCreateScript != testScript {
		t.Errorf("PostCreateScript not persisted correctly.\nExpected: %q\nGot: %q",
			testScript, loaded.Repositories[0].PostCreateScript)
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

	// Verify worktree notes
	if len(loaded.WorktreeNotes) != 1 {
		t.Fatalf("Expected 1 worktree note, got %d", len(loaded.WorktreeNotes))
	}
	if loaded.WorktreeNotes[0].Path != "/test/worktree" {
		t.Errorf("WorktreeNote Path not persisted correctly: %s", loaded.WorktreeNotes[0].Path)
	}
	if loaded.WorktreeNotes[0].Notes != "Test notes" {
		t.Errorf("WorktreeNote Notes not persisted correctly: %s", loaded.WorktreeNotes[0].Notes)
	}
}

func TestLoad_BackwardsCompatibility_PascalCaseKeys(t *testing.T) {
	// This test documents that old configs with PascalCase keys have a known issue:
	// - Simple fields (Name, Path, Type, URL) work due to case-insensitive matching
	// - Compound fields like PostCreateScript do NOT work because mapstructure
	//   expects "post_create_script" but the old buggy format wrote "PostCreateScript"
	//
	// The fix ensures new saves use the correct snake_case format. Existing users
	// will need to either:
	// 1. Re-enter their scripts (which will be saved correctly), or
	// 2. Manually edit their config.toml to use snake_case keys
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.toml")

	// Write a config with PascalCase keys (the old buggy format)
	oldFormatConfig := `root_directory = '/test/workspace'
yank_template = '${worktree_path}'

[[repositories]]
Name = 'test-repo'
Path = '/test/path'
Type = 'local'
URL = ''
PostCreateScript = 'echo hello'

[[worktree_notes]]
Path = '/test/worktree'
Notes = 'Test notes'
`
	if err := os.WriteFile(configFile, []byte(oldFormatConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Reset viper and load the old format
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

	// Simple fields work with case-insensitive matching
	if len(loaded.Repositories) != 1 {
		t.Fatalf("Expected 1 repository, got %d", len(loaded.Repositories))
	}
	if loaded.Repositories[0].Name != "test-repo" {
		t.Errorf("Name not loaded correctly: %s", loaded.Repositories[0].Name)
	}

	// PostCreateScript does NOT load from PascalCase - this is the bug we're fixing
	// The field will be empty, but once the user saves (any change), it will be
	// written correctly as snake_case and will persist properly going forward.
	if loaded.Repositories[0].PostCreateScript != "" {
		t.Errorf("Expected PostCreateScript to be empty from old format, got: %q",
			loaded.Repositories[0].PostCreateScript)
	}
}

func TestSaveAndLoad_MigratesOldFormat(t *testing.T) {
	// This test verifies that loading an old config and re-saving it
	// will migrate to the correct snake_case format
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.toml")

	// Write a config with PascalCase keys (the old buggy format)
	// Note: PostCreateScript won't be loaded, but other fields will
	oldFormatConfig := `root_directory = '/test/workspace'
yank_template = '${worktree_path}'

[[repositories]]
Name = 'test-repo'
Path = '/test/path'
Type = 'local'
URL = ''
PostCreateScript = ''
`
	if err := os.WriteFile(configFile, []byte(oldFormatConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load the old format
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

	// Add a script and save
	loaded.Repositories[0].PostCreateScript = "echo migrated"
	if err := Save(&loaded); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Read raw file to verify format
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "post_create_script") {
		t.Errorf("Expected 'post_create_script' in migrated config, got:\n%s", contentStr)
	}
	if strings.Contains(contentStr, "PostCreateScript") {
		t.Errorf("Found 'PostCreateScript' in migrated config, should be snake_case:\n%s", contentStr)
	}

	// Reload and verify script persists
	viper.Reset()
	viper.SetConfigFile(configFile)
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to re-read config: %v", err)
	}

	var reloaded Config
	if err := viper.Unmarshal(&reloaded); err != nil {
		t.Fatalf("Failed to re-unmarshal config: %v", err)
	}

	if reloaded.Repositories[0].PostCreateScript != "echo migrated" {
		t.Errorf("PostCreateScript not persisted after migration: %q",
			reloaded.Repositories[0].PostCreateScript)
	}
}
