package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the application configuration
type Config struct {
	APIKey         string `json:"api_key"`
	Model          string `json:"model"`
	BaseURL        string `json:"base_url"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

// ConfigLoader handles loading configuration from file, env, or defaults
type ConfigLoader struct{}

// NewConfigLoader creates a new config loader
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{}
}

// LoadConfig loads configuration with priority: file > env > defaults
func (c *ConfigLoader) LoadConfig() (*Config, error) {
	config := &Config{
		Model:          "gpt-oss:120b",
		BaseURL:        "http://localhost:11434/api/generate",
		TimeoutSeconds: 60,
	}

	// Try to load from config file
	repoRoot, err := findRepoRoot()
	if err == nil {
		configPath := filepath.Join(repoRoot, ".commit-generator-config")
		if fileData, err := os.ReadFile(configPath); err == nil {
			if err := json.Unmarshal(fileData, config); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Override with environment variable if config file doesn't have it
	if config.APIKey == "" {
		config.APIKey = os.Getenv("OLLAMA_API_KEY")
	}

	return config, nil
}

// GetTimeout returns the timeout as a time.Duration
func (c *Config) GetTimeout() time.Duration {
	return time.Duration(c.TimeoutSeconds) * time.Second
}

// SaveDefaultConfig saves a default config file to the repo root
func (c *ConfigLoader) SaveDefaultConfig(repoRoot string) error {
	config := &Config{
		APIKey:         os.Getenv("OLLAMA_API_KEY"), // Pre-fill from env if available
		Model:          "gpt-oss:120b",
		BaseURL:        "http://localhost:11434/api/generate",
		TimeoutSeconds: 60,
	}

	configPath := filepath.Join(repoRoot, ".commit-generator-config")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ConfigExists checks if a config file already exists
func (c *ConfigLoader) ConfigExists() (bool, error) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return false, err
	}
	configPath := filepath.Join(repoRoot, ".commit-generator-config")
	_, err = os.Stat(configPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
