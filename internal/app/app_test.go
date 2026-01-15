package app

import (
	"errors"
	"strings"
	"testing"
)

// Manual Mocks

type MockGit struct {
	IsInsideRepoFunc      func() (bool, error)
	HasStagedChangesFunc  func() (bool, error)
	GetStagedDiffFunc     func() (string, error)
	CommitWithMessageFunc func(message string) error
	GetRepoRootFunc       func() (string, error)
}

func (m *MockGit) IsInsideRepo() (bool, error) {
	return m.IsInsideRepoFunc()
}

func (m *MockGit) HasStagedChanges() (bool, error) {
	return m.HasStagedChangesFunc()
}

func (m *MockGit) GetStagedDiff() (string, error) {
	return m.GetStagedDiffFunc()
}

func (m *MockGit) CommitWithMessage(message string) error {
	if m.CommitWithMessageFunc != nil {
		return m.CommitWithMessageFunc(message)
	}
	return nil
}

func (m *MockGit) GetRepoRoot() (string, error) {
	if m.GetRepoRootFunc != nil {
		return m.GetRepoRootFunc()
	}
	return "/tmp/test-repo", nil
}

type MockConfig struct {
	LoadRulesFunc func() (string, error)
}

func (m *MockConfig) LoadRules() (string, error) {
	return m.LoadRulesFunc()
}

type MockAI struct {
	GenerateCommitMessageFunc func(diff string, rules string) (string, error)
}

func (m *MockAI) GenerateCommitMessage(diff string, rules string) (string, error) {
	return m.GenerateCommitMessageFunc(diff, rules)
}

func TestApp_Run(t *testing.T) {
	tests := []struct {
		name          string
		mockGit       *MockGit
		mockConfig    *MockConfig
		mockAI        *MockAI
		expectedError string
	}{
		{
			name: "Success with rules",
			mockGit: &MockGit{
				IsInsideRepoFunc:     func() (bool, error) { return true, nil },
				HasStagedChangesFunc: func() (bool, error) { return true, nil },
				GetStagedDiffFunc:    func() (string, error) { return "diff content", nil },
			},
			mockConfig: &MockConfig{
				LoadRulesFunc: func() (string, error) { return "some rules", nil },
			},
			mockAI: &MockAI{
				GenerateCommitMessageFunc: func(diff, rules string) (string, error) {
					if diff != "diff content" {
						return "", errors.New("unexpected diff")
					}
					if rules != "some rules" {
						return "", errors.New("unexpected rules")
					}
					return "feat: something", nil
				},
			},
			expectedError: "",
		},
		{
			name: "Success without rules",
			mockGit: &MockGit{
				IsInsideRepoFunc:     func() (bool, error) { return true, nil },
				HasStagedChangesFunc: func() (bool, error) { return true, nil },
				GetStagedDiffFunc:    func() (string, error) { return "diff content", nil },
			},
			mockConfig: &MockConfig{
				LoadRulesFunc: func() (string, error) { return "", nil },
			},
			mockAI: &MockAI{
				GenerateCommitMessageFunc: func(diff, rules string) (string, error) {
					if rules != "" {
						return "", errors.New("expected empty rules")
					}
					return "fix: something", nil
				},
			},
			expectedError: "",
		},
		{
			name: "Not a git repo",
			mockGit: &MockGit{
				IsInsideRepoFunc: func() (bool, error) { return false, nil },
			},
			mockConfig:    &MockConfig{}, // Should not be called
			mockAI:        &MockAI{},     // Should not be called
			expectedError: "not a git repository",
		},
		{
			name: "No staged changes",
			mockGit: &MockGit{
				IsInsideRepoFunc:     func() (bool, error) { return true, nil },
				HasStagedChangesFunc: func() (bool, error) { return false, nil },
			},
			mockConfig:    &MockConfig{}, // Should not be called
			mockAI:        &MockAI{},     // Should not be called
			expectedError: "no staged changes found",
		},
		{
			name: "Git Diff Error",
			mockGit: &MockGit{
				IsInsideRepoFunc:     func() (bool, error) { return true, nil },
				HasStagedChangesFunc: func() (bool, error) { return true, nil },
				GetStagedDiffFunc:    func() (string, error) { return "", errors.New("git error") },
			},
			mockConfig: &MockConfig{
				LoadRulesFunc: func() (string, error) { return "", nil },
			},
			mockAI:        &MockAI{}, // Should not be called
			expectedError: "failed to get diff: git error",
		},
		{
			name: "AI Error",
			mockGit: &MockGit{
				IsInsideRepoFunc:     func() (bool, error) { return true, nil },
				HasStagedChangesFunc: func() (bool, error) { return true, nil },
				GetStagedDiffFunc:    func() (string, error) { return "diff", nil },
			},
			mockConfig: &MockConfig{
				LoadRulesFunc: func() (string, error) { return "", nil },
			},
			mockAI: &MockAI{
				GenerateCommitMessageFunc: func(diff, rules string) (string, error) {
					return "", errors.New("ai service down")
				},
			},
			expectedError: "failed to generate commit message: ai service down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(tt.mockGit, tt.mockConfig, nil, tt.mockAI)
			err := app.Run()

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}
