package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Repository struct {
	Name             string `mapstructure:"name"`
	Path             string `mapstructure:"path"`
	Type             string `mapstructure:"type"`               // "remote" or "local"
	URL              string `mapstructure:"url"`                // For remote repos
	PostCreateScript string `mapstructure:"post_create_script"` // Script to run after creating worktrees
}

type Config struct {
	RootDirectory string       `mapstructure:"root_directory"`
	Repositories  []Repository `mapstructure:"repositories"`
	YankTemplate  string       `mapstructure:"yank_template"`
	EnterScript   string       `mapstructure:"enter_script"` // Path to script file to execute on Enter
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		RootDirectory: filepath.Join(homeDir, "workspace"),
		Repositories:  []Repository{},
		YankTemplate:  "${worktree_path}",
		EnterScript:   "",
	}
}

func Load() (*Config, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return nil, err
	}

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
	viper.SetDefault("enter_script", defaultCfg.EnterScript)

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

	for i := range config.Repositories {
		script := strings.TrimSpace(config.Repositories[i].PostCreateScript)
		if script == "" {
			continue
		}
		hasScript, err := HasRepoScript(config.Repositories[i].Name)
		if err == nil && !hasScript {
			_ = SaveRepoScript(config.Repositories[i].Name, script)
		}
		config.Repositories[i].PostCreateScript = ""
	}

	return &config, nil
}

func Save(cfg *Config) error {
	viper.Set("root_directory", cfg.RootDirectory)
	viper.Set("repositories", repositoriesToMaps(cfg.Repositories))
	viper.Set("yank_template", cfg.YankTemplate)
	viper.Set("enter_script", cfg.EnterScript)
	return viper.WriteConfig()
}

// repositoriesToMaps converts Repository structs to maps with snake_case keys.
// This is necessary because viper.Set() uses Go field names (capitalized) when
// serializing structs, but viper.Unmarshal() expects mapstructure tag names
// (snake_case).
func repositoriesToMaps(repos []Repository) []map[string]interface{} {
	result := make([]map[string]interface{}, len(repos))
	for i, repo := range repos {
		result[i] = map[string]interface{}{
			"name": repo.Name,
			"path": repo.Path,
			"type": repo.Type,
			"url":  repo.URL,
		}
	}
	return result
}
