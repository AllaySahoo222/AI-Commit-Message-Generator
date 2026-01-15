package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-repo")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory to make it a repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	loader := NewConfigLoader()

	// Test loading default config (no file)
	config, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	if config.Model != "gpt-oss:120b" {
		t.Errorf("Expected default model 'gpt-oss:120b', got '%s'", config.Model)
	}

	if config.BaseURL != "http://localhost:11434/api/generate" {
		t.Errorf("Expected default base URL, got '%s'", config.BaseURL)
	}

	if config.TimeoutSeconds != 60 {
		t.Errorf("Expected default timeout 60, got %d", config.TimeoutSeconds)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-repo")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory to make it a repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	loader := NewConfigLoader()

	// Save default config
	if err := loader.SaveDefaultConfig(tmpDir); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load it back
	config, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.Model != "gpt-oss:120b" {
		t.Errorf("Expected model 'gpt-oss:120b', got '%s'", config.Model)
	}
}

func TestConfigExists(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-repo")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory to make it a repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	loader := NewConfigLoader()

	// Should not exist initially
	exists, err := loader.ConfigExists()
	if err != nil {
		t.Fatalf("ConfigExists failed: %v", err)
	}
	if exists {
		t.Error("Config should not exist initially")
	}

	// Save config
	if err := loader.SaveDefaultConfig(tmpDir); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Should exist now
	exists, err = loader.ConfigExists()
	if err != nil {
		t.Fatalf("ConfigExists failed: %v", err)
	}
	if !exists {
		t.Error("Config should exist after saving")
	}
}
