package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupConfigBench(t *testing.B, withRules bool) (string, func()) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)

	// Create .git directory to make it a repo
	os.Mkdir(".git", 0755)

	if withRules {
		// Create rules file
		rulesContent := `- Always start with a verb (Add, Fix, Update).
- If the change affects the UI, mention it.
- Max 50 characters for the subject line.
- Always mention Jira ID if applicable.`
		os.WriteFile(".git-commit-rules-for-ai", []byte(rulesContent), 0644)
	}

	cleanup := func() {
		os.Chdir(originalWd)
	}

	return tempDir, cleanup
}

func BenchmarkLoadRules_WithFile(b *testing.B) {
	_, cleanup := setupConfigBench(b, true)
	defer cleanup()

	loader := NewLoader()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = loader.LoadRules()
	}
}

func BenchmarkLoadRules_WithoutFile(b *testing.B) {
	_, cleanup := setupConfigBench(b, false)
	defer cleanup()

	loader := NewLoader()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = loader.LoadRules()
	}
}

func BenchmarkFindRepoRoot_Shallow(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, _ := os.Getwd()
	defer func() { os.Chdir(originalWd) }()

	os.Chdir(tempDir)
	os.Mkdir(".git", 0755)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = findRepoRoot()
	}
}

func BenchmarkFindRepoRoot_Deep(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, _ := os.Getwd()
	defer func() { os.Chdir(originalWd) }()

	// Create nested directory structure
	deepDir := filepath.Join(tempDir, "level1", "level2", "level3", "level4", "level5")
	os.MkdirAll(deepDir, 0755)
	os.Chdir(deepDir)

	// Create .git at root
	os.Mkdir(filepath.Join(tempDir, ".git"), 0755)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = findRepoRoot()
	}
}

func BenchmarkFindRepoRoot_VeryDeep(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, _ := os.Getwd()
	defer func() { os.Chdir(originalWd) }()

	// Create very nested directory structure
	deepDir := filepath.Join(tempDir, "a", "b", "c", "d", "e", "f", "g", "h", "i", "j")
	os.MkdirAll(deepDir, 0755)
	os.Chdir(deepDir)

	// Create .git at root
	os.Mkdir(filepath.Join(tempDir, ".git"), 0755)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = findRepoRoot()
	}
}
