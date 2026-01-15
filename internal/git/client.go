package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Client defines the interface for git operations
type Client interface {
	IsInsideRepo() (bool, error)
	HasStagedChanges() (bool, error)
	GetStagedDiff() (string, error)
	CommitWithMessage(message string) error
	GetRepoRoot() (string, error)
}

// ClientImpl implements the Client interface using os/exec
type ClientImpl struct{}

// NewClient creates a new Git client
func NewClient() Client {
	return &ClientImpl{}
}

// IsInsideRepo checks if the current directory is inside a git repository
func (c *ClientImpl) IsInsideRepo() (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		return false, nil // Not a repo or git not installed
	}
	return true, nil
}

// HasStagedChanges checks if there are staged changes
func (c *ClientImpl) HasStagedChanges() (bool, error) {
	// --quiet returns exit code 1 if there are changes, 0 if no changes
	// We want to know if there ARE changes.
	// Actually, git diff --staged --quiet
	// Exit code 1 implies differences exist.
	// Exit code 0 implies no differences.
	cmd := exec.Command("git", "diff", "--staged", "--quiet")
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return true, nil
			}
		}
		return false, err
	}
	return false, nil
}

// GetStagedDiff returns the diff of staged changes
func (c *ClientImpl) GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}

	diff := string(out)
	if len(diff) > 10000 {
		return diff[:10000] + "\n...[TRUNCATED]", nil
	}
	return diff, nil
}

// CommitWithMessage executes git commit with the given message
func (c *ClientImpl) CommitWithMessage(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

// GetRepoRoot returns the root directory of the git repository
func (c *ClientImpl) GetRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get repo root: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
