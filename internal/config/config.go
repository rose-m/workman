package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Repository struct {
	Name             string `mapstructure:"name"`
	Path             string `mapstructure:"path"`
	Type             string `mapstructure:"type"`               // "remote" or "local"
	URL              string `mapstructure:"url"`                // For remote repos
	PostCreateScript string `mapstructure:"post_create_script"` // Script to run after creating worktrees
}

type WorktreeNote struct {
	Path  string `mapstructure:"path"`
	Notes string `mapstructure:"notes"`
}

type Config struct {
	RootDirectory string         `mapstructure:"root_directory"`
	Repositories  []Repository   `mapstructure:"repositories"`
	YankTemplate  string         `mapstructure:"yank_template"`
	WorktreeNotes []WorktreeNote `mapstructure:"worktree_notes"`
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		RootDirectory: filepath.Join(homeDir, "workspace"),
		Repositories:  []Repository{},
		YankTemplate:  "${worktree_path}",
		WorktreeNotes: []WorktreeNote{},
	}
}

func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".config", "workman")
	configFile := filepath.Join(configDir, "config.toml")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(configDir)

	// Set defaults
	defaultCfg := DefaultConfig()
	viper.SetDefault("root_directory", defaultCfg.RootDirectory)
	viper.SetDefault("repositories", defaultCfg.Repositories)
	viper.SetDefault("yank_template", defaultCfg.YankTemplate)
	viper.SetDefault("worktree_notes", defaultCfg.WorktreeNotes)

	// If config file doesn't exist, create it with defaults
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := viper.SafeWriteConfig(); err != nil {
			return nil, err
		}
	}

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// Initialize WorktreeNotes if nil
	if config.WorktreeNotes == nil {
		config.WorktreeNotes = []WorktreeNote{}
	}

	return &config, nil
}

func Save(cfg *Config) error {
	viper.Set("root_directory", cfg.RootDirectory)
	viper.Set("repositories", repositoriesToMaps(cfg.Repositories))
	viper.Set("yank_template", cfg.YankTemplate)
	viper.Set("worktree_notes", worktreeNotesToMaps(cfg.WorktreeNotes))
	return viper.WriteConfig()
}

// repositoriesToMaps converts Repository structs to maps with snake_case keys.
// This is necessary because viper.Set() uses Go field names (capitalized) when
// serializing structs, but viper.Unmarshal() expects mapstructure tag names
// (snake_case). Without this conversion, fields like PostCreateScript would be
// written as "PostCreateScript" but read as "post_create_script", causing data loss.
func repositoriesToMaps(repos []Repository) []map[string]interface{} {
	result := make([]map[string]interface{}, len(repos))
	for i, repo := range repos {
		result[i] = map[string]interface{}{
			"name":               repo.Name,
			"path":               repo.Path,
			"type":               repo.Type,
			"url":                repo.URL,
			"post_create_script": repo.PostCreateScript,
		}
	}
	return result
}

// worktreeNotesToMaps converts WorktreeNote structs to maps with snake_case keys.
// See repositoriesToMaps for explanation of why this is necessary.
func worktreeNotesToMaps(notes []WorktreeNote) []map[string]interface{} {
	result := make([]map[string]interface{}, len(notes))
	for i, note := range notes {
		result[i] = map[string]interface{}{
			"path":  note.Path,
			"notes": note.Notes,
		}
	}
	return result
}

// GetWorktreeNotes returns the notes for a given worktree path
func (c *Config) GetWorktreeNotes(path string) string {
	for _, note := range c.WorktreeNotes {
		if note.Path == path {
			return note.Notes
		}
	}
	return ""
}

// SetWorktreeNotes sets or updates the notes for a given worktree path
func (c *Config) SetWorktreeNotes(path, notes string) {
	// Find existing note and update it
	for i := range c.WorktreeNotes {
		if c.WorktreeNotes[i].Path == path {
			if notes == "" {
				// Remove the note if empty
				c.WorktreeNotes = append(c.WorktreeNotes[:i], c.WorktreeNotes[i+1:]...)
			} else {
				c.WorktreeNotes[i].Notes = notes
			}
			return
		}
	}

	// Add new note if not empty
	if notes != "" {
		c.WorktreeNotes = append(c.WorktreeNotes, WorktreeNote{
			Path:  path,
			Notes: notes,
		})
	}
}

// DeleteWorktreeNotes removes the notes for a given worktree path
func (c *Config) DeleteWorktreeNotes(path string) {
	for i := range c.WorktreeNotes {
		if c.WorktreeNotes[i].Path == path {
			c.WorktreeNotes = append(c.WorktreeNotes[:i], c.WorktreeNotes[i+1:]...)
			return
		}
	}
}
