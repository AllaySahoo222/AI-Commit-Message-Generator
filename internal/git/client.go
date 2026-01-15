package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Client defines the interface for git operations
type Client interface {
	IsInsideRepo() (bool, error)
	HasStagedChanges() (bool, error)
	GetStagedDiff() (string, error)
	CommitWithMessage(message string) error
	GetRepoRoot() (string, error)
}

// ClientImpl implements the Client interface using go-git
type ClientImpl struct {
	repo     *git.Repository
	repoPath string
	mu       sync.Mutex
}

// NewClient creates a new Git client
func NewClient() Client {
	return &ClientImpl{}
}

// openRepo opens a git repository from the current working directory
// Uses caching to avoid repeated opens
func (c *ClientImpl) openRepo() (*git.Repository, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	// Return cached repo if it exists and we're in the same directory
	if c.repo != nil && c.repoPath == wd {
		return c.repo, nil
	}

	repo, err := git.PlainOpenWithOptions(wd, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, err
	}

	// Cache the repo
	c.repo = repo
	c.repoPath = wd

	return repo, nil
}

// IsInsideRepo checks if the current directory is inside a git repository
func (c *ClientImpl) IsInsideRepo() (bool, error) {
	_, err := c.openRepo()
	if err == git.ErrRepositoryNotExists {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// HasStagedChanges checks if there are staged changes
func (c *ClientImpl) HasStagedChanges() (bool, error) {
	repo, err := c.openRepo()
	if err != nil {
		return false, fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}

	// Check if there are any staged changes
	// Short-circuit: return immediately after finding first staged file
	for _, fileStatus := range status {
		// Staged changes are files that have been added to the index
		// but not yet committed. This includes:
		// - Added files (Staging == Added)
		// - Modified files (Staging == Modified)
		// - Deleted files (Staging == Deleted)
		// - Renamed files (Staging == Renamed)
		// - Copied files (Staging == Copied)
		if fileStatus.Staging != git.Unmodified && fileStatus.Staging != git.Untracked {
			return true, nil
		}
	}

	return false, nil
}

// GetStagedDiff returns the diff of staged changes
func (c *ClientImpl) GetStagedDiff() (string, error) {
	repo, err := c.openRepo()
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	// Pre-allocate builder capacity based on estimated diff size
	// Estimate: ~100 bytes per file header + ~50 bytes per line
	estimatedSize := len(status) * 500
	var diffBuilder strings.Builder
	diffBuilder.Grow(estimatedSize)

	// Cache working directory
	wd, _ := os.Getwd()

	// Get HEAD commit for comparison
	head, err := repo.Head()
	if err != nil && err != plumbing.ErrReferenceNotFound {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	var headTree *object.Tree
	if err == nil {
		headCommit, err := repo.CommitObject(head.Hash())
		if err == nil {
			headTree, err = headCommit.Tree()
			if err != nil {
				return "", fmt.Errorf("failed to get HEAD tree: %w", err)
			}
		}
	}

	// Process each staged file
	for filePath, fileStatus := range status {
		// Only process staged changes
		if fileStatus.Staging == git.Unmodified || fileStatus.Staging == git.Untracked {
			continue
		}

		switch fileStatus.Staging {
		case git.Added:
			// New file - show all lines as additions
			diffBuilder.WriteString("diff --git a/")
			diffBuilder.WriteString(filePath)
			diffBuilder.WriteString(" b/")
			diffBuilder.WriteString(filePath)
			diffBuilder.WriteString("\nnew file mode 100644\nindex 0000000..")
			diffBuilder.WriteString(fileStatus.Extra)
			diffBuilder.WriteString("\n--- /dev/null\n+++ b/")
			diffBuilder.WriteString(filePath)
			diffBuilder.WriteString("\n")

			// Read file content
			fullPath := filepath.Join(wd, filePath)
			content, err := os.ReadFile(fullPath)
			if err == nil {
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					diffBuilder.WriteString("+")
					diffBuilder.WriteString(line)
					diffBuilder.WriteString("\n")
				}
			}

		case git.Deleted:
			// Deleted file
			diffBuilder.WriteString("diff --git a/")
			diffBuilder.WriteString(filePath)
			diffBuilder.WriteString(" b/")
			diffBuilder.WriteString(filePath)
			diffBuilder.WriteString("\ndeleted file mode 100644\nindex ")
			diffBuilder.WriteString(fileStatus.Extra)
			diffBuilder.WriteString("..0000000\n--- a/")
			diffBuilder.WriteString(filePath)
			diffBuilder.WriteString("\n+++ /dev/null\n")

			// Try to get content from HEAD
			if headTree != nil {
				entry, err := headTree.FindEntry(filePath)
				if err == nil {
					blob, err := repo.BlobObject(entry.Hash)
					if err == nil {
						reader, err := blob.Reader()
						if err == nil {
							content := make([]byte, blob.Size)
							reader.Read(content)
							reader.Close()
							lines := strings.Split(string(content), "\n")
							for _, line := range lines {
								diffBuilder.WriteString("-")
								diffBuilder.WriteString(line)
								diffBuilder.WriteString("\n")
							}
						}
					}
				}
			}

		case git.Modified:
			// Modified file - get diff between HEAD and staged version
			diffBuilder.WriteString("diff --git a/")
			diffBuilder.WriteString(filePath)
			diffBuilder.WriteString(" b/")
			diffBuilder.WriteString(filePath)
			diffBuilder.WriteString("\nindex ")
			diffBuilder.WriteString(fileStatus.Extra)
			diffBuilder.WriteString("..")
			diffBuilder.WriteString(fileStatus.Extra)
			diffBuilder.WriteString(" 100644\n--- a/")
			diffBuilder.WriteString(filePath)
			diffBuilder.WriteString("\n+++ b/")
			diffBuilder.WriteString(filePath)
			diffBuilder.WriteString("\n")

			// Get old content from HEAD
			var oldContent []byte
			if headTree != nil {
				entry, err := headTree.FindEntry(filePath)
				if err == nil {
					blob, err := repo.BlobObject(entry.Hash)
					if err == nil {
						reader, err := blob.Reader()
						if err == nil {
							oldContent = make([]byte, blob.Size)
							reader.Read(oldContent)
							reader.Close()
						}
					}
				}
			}

			// Get new content from working directory
			fullPath := filepath.Join(wd, filePath)
			newContent, err := os.ReadFile(fullPath)
			if err != nil {
				newContent = []byte{}
			}

			// Simple line-by-line diff
			oldLines := strings.Split(string(oldContent), "\n")
			newLines := strings.Split(string(newContent), "\n")

			// For simplicity, show old lines as removed and new lines as added
			// A more sophisticated diff algorithm could be used here
			for _, line := range oldLines {
				diffBuilder.WriteString("-")
				diffBuilder.WriteString(line)
				diffBuilder.WriteString("\n")
			}
			for _, line := range newLines {
				diffBuilder.WriteString("+")
				diffBuilder.WriteString(line)
				diffBuilder.WriteString("\n")
			}

		case git.Renamed:
			// Renamed file
			diffBuilder.WriteString("diff --git a/")
			diffBuilder.WriteString(fileStatus.Extra)
			diffBuilder.WriteString(" b/")
			diffBuilder.WriteString(filePath)
			diffBuilder.WriteString("\nrename from ")
			diffBuilder.WriteString(fileStatus.Extra)
			diffBuilder.WriteString("\nrename to ")
			diffBuilder.WriteString(filePath)
			diffBuilder.WriteString("\n")
		}
	}

	diff := diffBuilder.String()
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
