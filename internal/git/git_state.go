package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GitStateType represents the current state of the git repository
type GitStateType int

const (
	// StateNormal indicates normal git operations
	StateNormal GitStateType = iota
	// StateMerge indicates a merge is in progress
	StateMerge
	// StateRebase indicates a rebase is in progress
	StateRebase
	// StateCherryPick indicates a cherry-pick is in progress
	StateCherryPick
)

// String returns the string representation of GitStateType
func (s GitStateType) String() string {
	switch s {
	case StateNormal:
		return "normal"
	case StateMerge:
		return "merge"
	case StateRebase:
		return "rebase"
	case StateCherryPick:
		return "cherry-pick"
	default:
		return "unknown"
	}
}

// GitState represents the current state of the git repository
type GitState struct {
	// Type is the type of operation in progress
	Type GitStateType
	// OriginalMessage is the original commit/merge message (if available)
	OriginalMessage string
	// ConflictMode indicates if there are conflicts to resolve
	ConflictMode bool
}

// DetectGitState detects the current git state by inspecting the .git directory
func DetectGitState(repoRoot string) (*GitState, error) {
	gitDir := filepath.Join(repoRoot, ".git")

	// Check if .git exists
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("not a git repository: %s", repoRoot)
	}

	state := &GitState{
		Type:         StateNormal,
		ConflictMode: false,
	}

	// Check for merge state
	mergeHeadPath := filepath.Join(gitDir, "MERGE_HEAD")
	if _, err := os.Stat(mergeHeadPath); err == nil {
		state.Type = StateMerge
		state.ConflictMode = true

		// Read merge message
		mergeMsgPath := filepath.Join(gitDir, "MERGE_MSG")
		if content, err := os.ReadFile(mergeMsgPath); err == nil {
			state.OriginalMessage = filterCommentLines(strings.TrimSpace(string(content)))
		}
		return state, nil
	}

	// Check for cherry-pick state
	cherryPickHeadPath := filepath.Join(gitDir, "CHERRY_PICK_HEAD")
	if _, err := os.Stat(cherryPickHeadPath); err == nil {
		state.Type = StateCherryPick
		state.ConflictMode = true

		// Read cherry-pick message from COMMIT_EDITMSG
		commitEditMsgPath := filepath.Join(gitDir, "COMMIT_EDITMSG")
		if content, err := os.ReadFile(commitEditMsgPath); err == nil {
			state.OriginalMessage = filterCommentLines(strings.TrimSpace(string(content)))
		}
		return state, nil
	}

	// Check for rebase state
	rebaseMergePath := filepath.Join(gitDir, "rebase-merge")
	rebaseApplyPath := filepath.Join(gitDir, "rebase-apply")

	if info, err := os.Stat(rebaseMergePath); err == nil && info.IsDir() {
		state.Type = StateRebase
		state.ConflictMode = true

		// Try to read the original commit message
		headNamePath := filepath.Join(rebaseMergePath, "head-name")
		if content, err := os.ReadFile(headNamePath); err == nil {
			state.OriginalMessage = fmt.Sprintf("Rebase branch: %s", strings.TrimSpace(string(content)))
		}
		return state, nil
	}

	if info, err := os.Stat(rebaseApplyPath); err == nil && info.IsDir() {
		state.Type = StateRebase
		state.ConflictMode = true

		// Try to read the original commit message
		headNamePath := filepath.Join(rebaseApplyPath, "head-name")
		if content, err := os.ReadFile(headNamePath); err == nil {
			state.OriginalMessage = fmt.Sprintf("Rebase branch: %s", strings.TrimSpace(string(content)))
		}
		return state, nil
	}

	// Normal state
	return state, nil
}

// filterCommentLines removes git comment lines (starting with #) from a message
func filterCommentLines(message string) string {
	lines := strings.Split(message, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") && trimmed != "" {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}
