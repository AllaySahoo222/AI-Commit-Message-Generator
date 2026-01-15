package git

import (
	"os"
	"strings"
	"testing"

	git "github.com/go-git/go-git/v5"
)

func TestClientImpl_Integration(t *testing.T) {
	// Setup temp dir
	tempDir := t.TempDir()

	// Capture WD
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get WD: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	// Change to temp dir
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}

	client := NewClient()

	// 1. Test IsInsideRepo - False (No repo yet)
	inRepo, err := client.IsInsideRepo()
	if err != nil {
		t.Errorf("expected no error checking repo status outside repo, got %v", err)
	}
	if inRepo {
		t.Error("expected IsInsideRepo to be false")
	}

	// Initialize Git Repo using go-git
	repo, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to git init: %v", err)
	}

	// Configure git user/email for commits (required in some envs)
	config, err := repo.Config()
	if err != nil {
		t.Fatalf("failed to get config: %v", err)
	}
	config.User.Name = "Test User"
	config.User.Email = "test@example.com"
	repo.SetConfig(config)

	// 2. Test IsInsideRepo - True
	inRepo, err = client.IsInsideRepo()
	if err != nil {
		t.Errorf("expected no error checking repo status inside repo, got %v", err)
	}
	if !inRepo {
		t.Error("expected IsInsideRepo to be true")
	}

	// 3. Test HasStagedChanges - False (Empty repo)
	staged, err := client.HasStagedChanges()
	if err != nil {
		t.Errorf("unexpected error checking staged changes: %v", err)
	}
	if staged {
		t.Error("expected no staged changes")
	}

	// Create a file
	if err := os.WriteFile("test.txt", []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// 4. Test HasStagedChanges - False (Unstaged file)
	staged, err = client.HasStagedChanges()
	if err != nil {
		t.Errorf("unexpected error checking staged changes: %v", err)
	}
	if staged {
		t.Error("expected no staged changes for unstaged file")
	}

	// Stage the file using go-git
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}
	if _, err := worktree.Add("test.txt"); err != nil {
		t.Fatalf("failed to git add: %v", err)
	}

	// 5. Test HasStagedChanges - True
	staged, err = client.HasStagedChanges()
	if err != nil {
		t.Errorf("unexpected error checking staged changes: %v", err)
	}
	if !staged {
		t.Error("expected staged changes")
	}

	// 6. Test GetStagedDiff
	diff, err := client.GetStagedDiff()
	if err != nil {
		t.Errorf("unexpected error getting diff: %v", err)
	}
	if diff == "" {
		t.Error("expected diff content, got empty string")
	}
	// Verify diff contains filename or content
	// git diff output formats vary, but should contain "test.txt"
	// diff --staged on a new file shows mostly +lines
	if !strings.Contains(diff, "test.txt") {
		t.Errorf("expected diff to contain 'test.txt', got: %s", diff)
	}
}
