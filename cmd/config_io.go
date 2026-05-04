package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	DefaultAgent string                   `yaml:"default_agent,omitempty"`
	Projects     map[string]ProjectConfig `yaml:"projects,omitempty"`
	Sessions     map[string]SessionConfig `yaml:"sessions,omitempty"`
}

type ProjectConfig struct {
	Dir                string `yaml:"dir"`
	Agent              string `yaml:"agent"`
	SessionAfterScript string `yaml:"session_after_script,omitempty"`
}

type SessionConfig struct {
	Project  string `yaml:"project,omitempty"`
	Dir      string `yaml:"dir"`
	Agent    string `yaml:"agent"`
	Worktree string `yaml:"worktree,omitempty"`
	Branch   string `yaml:"branch,omitempty"`
}

func loadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Config{
				Projects: map[string]ProjectConfig{},
				Sessions: map[string]SessionConfig{},
			}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Projects == nil {
		cfg.Projects = map[string]ProjectConfig{}
	}
	if cfg.Sessions == nil {
		cfg.Sessions = map[string]SessionConfig{}
	}
	return &cfg, nil
}

func saveConfig(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}
