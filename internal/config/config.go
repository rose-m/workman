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

type Config struct {
	RootDirectory string       `mapstructure:"root_directory"`
	Repositories  []Repository `mapstructure:"repositories"`
	YankTemplate  string       `mapstructure:"yank_template"`
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		RootDirectory: filepath.Join(homeDir, "workspace"),
		Repositories:  []Repository{},
		YankTemplate:  "${worktree_path}",
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

	return &config, nil
}

func Save(cfg *Config) error {
	viper.Set("root_directory", cfg.RootDirectory)
	viper.Set("repositories", cfg.Repositories)
	viper.Set("yank_template", cfg.YankTemplate)
	return viper.WriteConfig()
}
