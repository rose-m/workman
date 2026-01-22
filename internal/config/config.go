package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Repository struct {
	Name string `mapstructure:"name"`
	Path string `mapstructure:"path"`
	Type string `mapstructure:"type"` // "remote" or "local"
	URL  string `mapstructure:"url"`  // For remote repos
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
	viper.Set("repositories", cfg.Repositories)
	viper.Set("yank_template", cfg.YankTemplate)
	viper.Set("worktree_notes", cfg.WorktreeNotes)
	return viper.WriteConfig()
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
