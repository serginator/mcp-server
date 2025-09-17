package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the configuration for the application
// It contains the API tokens for the different services
// that the MCP server integrates with.
type Config struct {
	NotionToken  string `yaml:"notion_token"`
	GithubToken  string `yaml:"github_token"`
	JiraToken    string `yaml:"jira_token"`
	JiraURL      string `yaml:"jira_url"`
	JiraUsername string `yaml:"jira_username"`
}

// LoadConfig loads the configuration with the following priority:
// 1. Environment variables (highest priority)
// 2. local.yml file (if exists)
// 3. config.yml file (fallback)
func LoadConfig(configPath string) (*Config, error) {
	var cfg Config

	// First, load from the main config file
	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	}

	// Second, try to load from local.yml (overrides config.yml)
	localPath := "local.yml"
	if data, err := os.ReadFile(localPath); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	}

	// Third, override with environment variables (highest priority)
	if token := os.Getenv("NOTION_TOKEN"); token != "" {
		cfg.NotionToken = token
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		cfg.GithubToken = token
	}
	if token := os.Getenv("JIRA_TOKEN"); token != "" {
		cfg.JiraToken = token
	}
	if url := os.Getenv("JIRA_URL"); url != "" {
		cfg.JiraURL = url
	}
	if username := os.Getenv("JIRA_USERNAME"); username != "" {
		cfg.JiraUsername = username
	}

	return &cfg, nil
}
