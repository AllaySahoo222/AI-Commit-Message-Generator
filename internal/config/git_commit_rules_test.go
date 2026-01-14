package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileLoader_LoadRules(t *testing.T) {
	// Setup a temporary directory
	tempDir := t.TempDir()

	// Capture current working directory to restore later
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Change to temp dir
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}

	// Helper to create a fake git repo root
	createFakeRepo := func(dir string) {
		if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
			t.Fatalf("failed to create .git dir: %v", err)
		}
	}

	// Refactoring to use subtests correctly with fresh dirs
	t.Run("Rules file exists", func(t *testing.T) {
		subDir := t.TempDir()
		createFakeRepo(subDir)
		rulesContent := "Rule 1: Be nice"
		err := os.WriteFile(filepath.Join(subDir, ".git-commit-rules-for-ai"), []byte(rulesContent), 0644)
		if err != nil {
			t.Fatalf("failed to write rules file: %v", err)
		}

		// Change WD to subDir
		if err := os.Chdir(subDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		loader := NewLoader()
		rules, err := loader.LoadRules()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if rules != rulesContent {
			t.Errorf("expected rules %q, got %q", rulesContent, rules)
		}
	})

	t.Run("Rules file missing", func(t *testing.T) {
		subDir := t.TempDir()
		createFakeRepo(subDir)
		
		if err := os.Chdir(subDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		loader := NewLoader()
		rules, err := loader.LoadRules()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if rules != "" {
			t.Errorf("expected empty rules, got %q", rules)
		}
	})
	
	t.Run("No repo root found", func(t *testing.T) {
		subDir := t.TempDir()
		// Do NOT create .git
		
		if err := os.Chdir(subDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		loader := NewLoader()
		rules, err := loader.LoadRules()
		// Expect no error, just empty rules as per implementation
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if rules != "" {
			t.Errorf("expected empty rules, got %q", rules)
		}
	})
}
