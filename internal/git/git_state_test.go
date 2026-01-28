package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectGitState(t *testing.T) {
	tests := []struct {
		name                string
		setupFunc           func(t *testing.T) string // Returns temp repo path
		expectedType        GitStateType
		expectedConflict    bool
		expectedMsgContains string
		wantErr             bool
	}{
		{
			name: "Normal state - no special files",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				gitDir := filepath.Join(tmpDir, ".git")
				if err := os.Mkdir(gitDir, 0755); err != nil {
					t.Fatalf("failed to create .git dir: %v", err)
				}
				return tmpDir
			},
			expectedType:     StateNormal,
			expectedConflict: false,
			wantErr:          false,
		},
		{
			name: "Merge state - MERGE_HEAD exists",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				gitDir := filepath.Join(tmpDir, ".git")
				if err := os.Mkdir(gitDir, 0755); err != nil {
					t.Fatalf("failed to create .git dir: %v", err)
				}

				// Create MERGE_HEAD
				mergeHeadPath := filepath.Join(gitDir, "MERGE_HEAD")
				if err := os.WriteFile(mergeHeadPath, []byte("abc123\n"), 0644); err != nil {
					t.Fatalf("failed to create MERGE_HEAD: %v", err)
				}

				// Create MERGE_MSG
				mergeMsgPath := filepath.Join(gitDir, "MERGE_MSG")
				mergeMsg := "Merge branch 'feature-x' into main"
				if err := os.WriteFile(mergeMsgPath, []byte(mergeMsg), 0644); err != nil {
					t.Fatalf("failed to create MERGE_MSG: %v", err)
				}

				return tmpDir
			},
			expectedType:        StateMerge,
			expectedConflict:    true,
			expectedMsgContains: "feature-x",
			wantErr:             false,
		},
		{
			name: "Cherry-pick state - CHERRY_PICK_HEAD exists",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				gitDir := filepath.Join(tmpDir, ".git")
				if err := os.Mkdir(gitDir, 0755); err != nil {
					t.Fatalf("failed to create .git dir: %v", err)
				}

				// Create CHERRY_PICK_HEAD
				cherryPickHeadPath := filepath.Join(gitDir, "CHERRY_PICK_HEAD")
				if err := os.WriteFile(cherryPickHeadPath, []byte("def456\n"), 0644); err != nil {
					t.Fatalf("failed to create CHERRY_PICK_HEAD: %v", err)
				}

				// Create COMMIT_EDITMSG
				commitEditMsgPath := filepath.Join(gitDir, "COMMIT_EDITMSG")
				commitMsg := "feat(api): added new endpoint"
				if err := os.WriteFile(commitEditMsgPath, []byte(commitMsg), 0644); err != nil {
					t.Fatalf("failed to create COMMIT_EDITMSG: %v", err)
				}

				return tmpDir
			},
			expectedType:        StateCherryPick,
			expectedConflict:    true,
			expectedMsgContains: "added new endpoint",
			wantErr:             false,
		},
		{
			name: "Rebase state - rebase-merge exists",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				gitDir := filepath.Join(tmpDir, ".git")
				if err := os.Mkdir(gitDir, 0755); err != nil {
					t.Fatalf("failed to create .git dir: %v", err)
				}

				// Create rebase-merge directory
				rebaseMergeDir := filepath.Join(gitDir, "rebase-merge")
				if err := os.Mkdir(rebaseMergeDir, 0755); err != nil {
					t.Fatalf("failed to create rebase-merge dir: %v", err)
				}

				// Create head-name file
				headNamePath := filepath.Join(rebaseMergeDir, "head-name")
				if err := os.WriteFile(headNamePath, []byte("refs/heads/feature-branch\n"), 0644); err != nil {
					t.Fatalf("failed to create head-name: %v", err)
				}

				return tmpDir
			},
			expectedType:        StateRebase,
			expectedConflict:    true,
			expectedMsgContains: "feature-branch",
			wantErr:             false,
		},
		{
			name: "Rebase state - rebase-apply exists",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				gitDir := filepath.Join(tmpDir, ".git")
				if err := os.Mkdir(gitDir, 0755); err != nil {
					t.Fatalf("failed to create .git dir: %v", err)
				}

				// Create rebase-apply directory
				rebaseApplyDir := filepath.Join(gitDir, "rebase-apply")
				if err := os.Mkdir(rebaseApplyDir, 0755); err != nil {
					t.Fatalf("failed to create rebase-apply dir: %v", err)
				}

				// Create head-name file
				headNamePath := filepath.Join(rebaseApplyDir, "head-name")
				if err := os.WriteFile(headNamePath, []byte("refs/heads/develop\n"), 0644); err != nil {
					t.Fatalf("failed to create head-name: %v", err)
				}

				return tmpDir
			},
			expectedType:        StateRebase,
			expectedConflict:    true,
			expectedMsgContains: "develop",
			wantErr:             false,
		},
		{
			name: "Error - .git does not exist",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := tt.setupFunc(t)

			state, err := DetectGitState(repoPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if state.Type != tt.expectedType {
				t.Errorf("expected state type %v, got %v", tt.expectedType, state.Type)
			}

			if state.ConflictMode != tt.expectedConflict {
				t.Errorf("expected conflict mode %v, got %v", tt.expectedConflict, state.ConflictMode)
			}

			if tt.expectedMsgContains != "" {
				if state.OriginalMessage == "" {
					t.Errorf("expected original message to contain '%s', but message was empty", tt.expectedMsgContains)
				} else if !containsSubstring(state.OriginalMessage, tt.expectedMsgContains) {
					t.Errorf("expected original message to contain '%s', got '%s'", tt.expectedMsgContains, state.OriginalMessage)
				}
			}
		})
	}
}

func TestGitStateType_String(t *testing.T) {
	tests := []struct {
		state    GitStateType
		expected string
	}{
		{StateNormal, "normal"},
		{StateMerge, "merge"},
		{StateRebase, "rebase"},
		{StateCherryPick, "cherry-pick"},
		{GitStateType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.state.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(str) > len(substr) && findSubstring(str, substr))
}

func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
