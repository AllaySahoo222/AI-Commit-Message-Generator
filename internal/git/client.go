package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Client defines the interface for git operations
type Client interface {
	IsInsideRepo() (bool, error)
	HasStagedChanges() (bool, error)
	GetStagedDiff() (string, error)
}

// ClientImpl implements the Client interface using go-git
type ClientImpl struct{}

// NewClient creates a new Git client
func NewClient() Client {
	return &ClientImpl{}
}

// openRepo opens a git repository from the current working directory
func (c *ClientImpl) openRepo() (*git.Repository, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	repo, err := git.PlainOpenWithOptions(wd, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, err
	}

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

	var diffBuilder strings.Builder

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
			diffBuilder.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", filePath, filePath))
			diffBuilder.WriteString(fmt.Sprintf("new file mode 100644\n"))
			diffBuilder.WriteString(fmt.Sprintf("index 0000000..%s\n", fileStatus.Extra))
			diffBuilder.WriteString(fmt.Sprintf("--- /dev/null\n"))
			diffBuilder.WriteString(fmt.Sprintf("+++ b/%s\n", filePath))

			// Read file content
			wd, _ := os.Getwd()
			fullPath := filepath.Join(wd, filePath)
			content, err := os.ReadFile(fullPath)
			if err == nil {
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					diffBuilder.WriteString(fmt.Sprintf("+%s\n", line))
				}
			}

		case git.Deleted:
			// Deleted file
			diffBuilder.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", filePath, filePath))
			diffBuilder.WriteString(fmt.Sprintf("deleted file mode 100644\n"))
			diffBuilder.WriteString(fmt.Sprintf("index %s..0000000\n", fileStatus.Extra))
			diffBuilder.WriteString(fmt.Sprintf("--- a/%s\n", filePath))
			diffBuilder.WriteString(fmt.Sprintf("+++ /dev/null\n"))

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
								diffBuilder.WriteString(fmt.Sprintf("-%s\n", line))
							}
						}
					}
				}
			}

		case git.Modified:
			// Modified file - get diff between HEAD and staged version
			diffBuilder.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", filePath, filePath))
			diffBuilder.WriteString(fmt.Sprintf("index %s..%s 100644\n", fileStatus.Extra, fileStatus.Extra))
			diffBuilder.WriteString(fmt.Sprintf("--- a/%s\n", filePath))
			diffBuilder.WriteString(fmt.Sprintf("+++ b/%s\n", filePath))

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
			wd, _ := os.Getwd()
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
				diffBuilder.WriteString(fmt.Sprintf("-%s\n", line))
			}
			for _, line := range newLines {
				diffBuilder.WriteString(fmt.Sprintf("+%s\n", line))
			}

		case git.Renamed:
			// Renamed file
			diffBuilder.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", fileStatus.Extra, filePath))
			diffBuilder.WriteString(fmt.Sprintf("rename from %s\n", fileStatus.Extra))
			diffBuilder.WriteString(fmt.Sprintf("rename to %s\n", filePath))
		}
	}

	diff := diffBuilder.String()
	if len(diff) > 10000 {
		return diff[:10000] + "\n...[TRUNCATED]", nil
	}
	return diff, nil
}
