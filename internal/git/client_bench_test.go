package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	git "github.com/go-git/go-git/v5"
)

// setupBenchRepo creates a temporary git repository with staged files for benchmarking
func setupBenchRepo(b *testing.B, numFiles int, linesPerFile int) func() {
	tempDir := b.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)

	// Initialize git repo
	repo, err := git.PlainInit(tempDir, false)
	if err != nil {
		b.Fatalf("failed to init repo: %v", err)
	}

	// Configure git user
	config, _ := repo.Config()
	config.User.Name = "Test User"
	config.User.Email = "test@example.com"
	repo.SetConfig(config)

	// Create initial commit
	worktree, _ := repo.Worktree()
	initialFile := "initial.txt"
	os.WriteFile(initialFile, []byte("initial content\n"), 0644)
	worktree.Add(initialFile)
	worktree.Commit("initial commit", &git.CommitOptions{})

	// Create and stage files for benchmarking
	for i := 0; i < numFiles; i++ {
		fileName := filepath.Join("bench", fmt.Sprintf("file%d.txt", i))
		os.MkdirAll(filepath.Dir(fileName), 0755)

		// Create file with specified number of lines
		var content strings.Builder
		for j := 0; j < linesPerFile; j++ {
			content.WriteString(fmt.Sprintf("Line %d in file %d\n", j, i))
		}

		os.WriteFile(fileName, []byte(content.String()), 0644)
		worktree.Add(fileName)
	}

	return func() {
		os.Chdir(originalWd)
	}
}

func BenchmarkIsInsideRepo(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, _ := os.Getwd()
	defer func() { os.Chdir(originalWd) }()

	os.Chdir(tempDir)
	git.PlainInit(tempDir, false)

	client := NewClient()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.IsInsideRepo()
	}
}

func BenchmarkHasStagedChanges_NoChanges(b *testing.B) {
	cleanup := setupBenchRepo(b, 0, 0)
	defer cleanup()

	client := NewClient()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.HasStagedChanges()
	}
}

func BenchmarkHasStagedChanges_OneFile(b *testing.B) {
	cleanup := setupBenchRepo(b, 1, 10)
	defer cleanup()

	client := NewClient()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.HasStagedChanges()
	}
}

func BenchmarkHasStagedChanges_ManyFiles(b *testing.B) {
	cleanup := setupBenchRepo(b, 50, 10)
	defer cleanup()

	client := NewClient()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.HasStagedChanges()
	}
}

func BenchmarkGetStagedDiff_Small(b *testing.B) {
	cleanup := setupBenchRepo(b, 1, 10)
	defer cleanup()

	client := NewClient()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetStagedDiff()
	}
}

func BenchmarkGetStagedDiff_Medium(b *testing.B) {
	cleanup := setupBenchRepo(b, 5, 100)
	defer cleanup()

	client := NewClient()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetStagedDiff()
	}
}

func BenchmarkGetStagedDiff_Large(b *testing.B) {
	cleanup := setupBenchRepo(b, 20, 500)
	defer cleanup()

	client := NewClient()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetStagedDiff()
	}
}

func BenchmarkGetStagedDiff_VeryLarge(b *testing.B) {
	cleanup := setupBenchRepo(b, 50, 1000)
	defer cleanup()

	client := NewClient()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetStagedDiff()
	}
}

func BenchmarkGetStagedDiff_ModifiedFiles(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, _ := os.Getwd()
	defer func() { os.Chdir(originalWd) }()

	os.Chdir(tempDir)
	repo, _ := git.PlainInit(tempDir, false)

	config, _ := repo.Config()
	config.User.Name = "Test User"
	config.User.Email = "test@example.com"
	repo.SetConfig(config)

	worktree, _ := repo.Worktree()

	// Create and commit initial files
	for i := 0; i < 10; i++ {
		fileName := fmt.Sprintf("file%d.txt", i)
		os.WriteFile(fileName, []byte(strings.Repeat("original line\n", 100)), 0644)
		worktree.Add(fileName)
	}
	worktree.Commit("initial", &git.CommitOptions{})

	// Modify files
	for i := 0; i < 10; i++ {
		fileName := fmt.Sprintf("file%d.txt", i)
		os.WriteFile(fileName, []byte(strings.Repeat("modified line\n", 100)), 0644)
		worktree.Add(fileName)
	}

	client := NewClient()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetStagedDiff()
	}
}
