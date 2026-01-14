package config

import (
	"os"
	"path/filepath"
)

// Loader defines the interface for loading configuration
type Loader interface {
	LoadRules() (string, error)
}

// FileLoader implements the Loader interface
type FileLoader struct{}

// NewLoader creates a new Config loader
func NewLoader() Loader {
	return &FileLoader{}
}

// LoadRules reads the .git-commit-rules-for-ai file from the repo root.
// It assumes the current working directory is the repo root (or we could look it up).
// For simplicity and per requirements ("in the root of the git repository"),
// we will try to find it in the current directory or traverse up?
// The requirement says: "Look for a file named .git-commit-rules-for-ai in the root of the git repository."
// Since the tool is likely run from the root, we check current dir.
// We can also double check by finding the .git dir if needed, but 'internal/git' handles repo check.
// We'll trust the user invokes it from within the repo.
func (c *FileLoader) LoadRules() (string, error) {
	// 1. Try to find the root of the git repo.
	// We can cheat a bit and just look in current dir and maybe parent dirs?
	// But strictly, we should invoke git to find root, OR just look in CWD.
	// "Look for a file ... in the root of the git repository"
	// Let's assume the tool is run from the root for now, or we can try to walk up until we find .git
	
	repoRoot, err := findRepoRoot()
	if err != nil {
		// If we can't find repo root, we can't find the rules file there.
		// Return empty, but maybe this isn't an error for the rules loader itself?
		// The App will verify we are in a repo first.
		// If we are, findRepoRoot should succeed.
		return "", nil 
	}

	rulesPath := filepath.Join(repoRoot, ".git-commit-rules-for-ai")
	
	content, err := os.ReadFile(rulesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // Optional file
		}
		return "", err
	}

	return string(content), nil
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
