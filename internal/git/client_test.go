package git

import (
	"os"
	"os/exec"
	"testing"
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

	// Initialize Git Repo
	if output, err := exec.Command("git", "init").CombinedOutput(); err != nil {
		t.Fatalf("failed to git init: %v\nOutput: %s", err, output)
	}
	// Configure git user/email for commits (required in some envs)
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

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

	// Stage the file
	if output, err := exec.Command("git", "add", "test.txt").CombinedOutput(); err != nil {
		t.Fatalf("failed to git add: %v\nOutput: %s", err, output)
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
}
