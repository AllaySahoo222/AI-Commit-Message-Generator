package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"ai-commit-message-generator/internal/ai"
	"ai-commit-message-generator/internal/config"
	"ai-commit-message-generator/internal/git"
)

// App is the main application struct
type App struct {
	Git          git.Client
	RulesLoader  config.Loader
	ConfigLoader *config.ConfigLoader
	AI           ai.Client
}

// NewApp creates a new App
func NewApp(gitClient git.Client, rulesLoader config.Loader, configLoader *config.ConfigLoader, aiClient ai.Client) *App {
	return &App{
		Git:          gitClient,
		RulesLoader:  rulesLoader,
		ConfigLoader: configLoader,
		AI:           aiClient,
	}
}

// Run executes the main logic
func (a *App) Run() error {
	// 1. Pre-flight Checks
	isRepo, err := a.Git.IsInsideRepo()
	if err != nil {
		return fmt.Errorf("failed to check repository status: %w", err)
	}
	if !isRepo {
		return errors.New("not a git repository")
	}

	hasChanges, err := a.Git.HasStagedChanges()
	if err != nil {
		return fmt.Errorf("failed to check for staged changes: %w", err)
	}
	if !hasChanges {
		return errors.New("no staged changes found. Please stage your changes using 'git add'")
	}

	// 2. Custom Rule Injection
	rules, err := a.RulesLoader.LoadRules()
	if err != nil {
		fmt.Printf("Warning: failed to load rules: %v. Proceeding without rules.\n", err)
	}

	// 3. Detect Git State (merge, rebase, cherry-pick)
	gitState, err := a.Git.DetectState()
	if err != nil {
		fmt.Printf("Warning: failed to detect git state: %v. Proceeding with normal state.\n", err)
		gitState = &git.GitState{Type: git.StateNormal}
	}

	// Display state information if not normal
	if gitState.Type != git.StateNormal {
		fmt.Fprintf(os.Stderr, "\n\033[33m⚠ Git State Detected: %s\033[0m\n", gitState.Type)
		if gitState.OriginalMessage != "" {
			fmt.Fprintf(os.Stderr, "\033[33mOriginal message: %s\033[0m\n", gitState.OriginalMessage)
		}
		fmt.Fprintln(os.Stderr)
	}

	// 4. Smart Diff Reading
	diff, err := a.Git.GetStagedDiff()
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}

	fmt.Println("Generating commit message...")

	// 5. AI Integration (with git state context)
	message, err := a.AI.GenerateCommitMessage(diff, rules, gitState)
	if err != nil {
		return fmt.Errorf("failed to generate commit message: %w", err)
	}

	// 6. Output
	// Check if the response suggests splitting into multiple commits
	// Look for explicit keywords that indicate the AI is suggesting a split
	lowerMessage := strings.ToLower(message)
	isSplitSuggestion := strings.Contains(lowerMessage, "split") ||
		strings.Contains(lowerMessage, "separate commit") ||
		strings.Contains(lowerMessage, "multiple commit") ||
		strings.Contains(lowerMessage, "should be committed separately")
	
	if isSplitSuggestion {
		// Output split suggestion in Yellow
		fmt.Println("\n\033[33mAI Suggestion (Split Changes):\033[0m")
		fmt.Println(message)
	} else {
		// Output commit message in Cyan (can be multi-line)
		fmt.Println("\n\033[36m" + message + "\033[0m")
	}

	return nil
}

// Init initializes the repository with config, rules file, and pre-commit hook
func (a *App) Init(force bool) error {
	// Check if we're in a git repo
	isRepo, err := a.Git.IsInsideRepo()
	if err != nil {
		return fmt.Errorf("failed to check repository status: %w", err)
	}
	if !isRepo {
		return errors.New("not a git repository. Please run this command from within a git repository")
	}

	// Get repo root
	repoRoot, err := a.Git.GetRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get repository root: %w", err)
	}

	// Check if already initialized
	if !force {
		configExists, err := a.ConfigLoader.ConfigExists()
		if err != nil {
			return fmt.Errorf("failed to check config existence: %w", err)
		}
		if configExists {
			fmt.Println("Repository already initialized. Use --force to reinitialize.")
			return nil
		}
	} else {
		fmt.Println("Forcing reinitialization...")
	}

	fmt.Println("Initializing commit generator...")

	// 1. Generate config file
	if err := a.ConfigLoader.SaveDefaultConfig(repoRoot); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	fmt.Printf("✓ Created .commit-generator-config\n")

	// 2. Generate rules file
	rulesPath := filepath.Join(repoRoot, ".git-commit-rules-for-ai")
	if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
		rulesContent := `# Git Commit Rules for AI Generator
# Customize these rules to match your team's conventions

# Example rules:
# - Always start with a verb (Add, Fix, Update)
# - If the change affects the UI, mention it
# - Max 50 characters for the subject line
# - Include Jira ticket ID if applicable
`
		if err := os.WriteFile(rulesPath, []byte(rulesContent), 0644); err != nil {
			return fmt.Errorf("failed to create rules file: %w", err)
		}
		fmt.Printf("✓ Created .git-commit-rules-for-ai\n")
	} else {
		fmt.Printf("✓ Rules file already exists\n")
	}

	// 3. Generate pre-commit hook
	hookPath := filepath.Join(repoRoot, ".git", "hooks", "pre-commit")
	hookContent, err := a.generatePreCommitHook()
	if err != nil {
		return fmt.Errorf("failed to generate pre-commit hook: %w", err)
	}

	// On Windows, use .bat extension for batch files, otherwise no extension
	if runtime.GOOS == "windows" {
		// Try to detect if PowerShell is preferred, otherwise use batch
		// For now, we'll create a .bat file that can call PowerShell if needed
		hookPath = hookPath + ".bat"
	}

	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		return fmt.Errorf("failed to create pre-commit hook: %w", err)
	}
	fmt.Printf("✓ Created pre-commit hook\n")

	fmt.Println("\nInitialization complete!")
	fmt.Println("Next steps:")
	fmt.Println("1. Update .commit-generator-config with your API key if needed")
	fmt.Println("2. Customize .git-commit-rules-for-ai with your team's rules")
	fmt.Println("3. Stage your changes and commit - the hook will generate your commit message!")

	return nil
}

// generatePreCommitHook generates the pre-commit hook script for the current platform
func (a *App) generatePreCommitHook() (string, error) {
	if runtime.GOOS == "windows" {
		return a.generateWindowsHook(), nil
	}
	return a.generateUnixHook(), nil
}

// generateUnixHook generates a bash pre-commit hook for Unix systems
func (a *App) generateUnixHook() string {
	exePath, err := os.Executable()
	if err != nil {
		exePath = "generate-commit" // Fallback
	} else {
		// Ensure we have an absolute path
		if absPath, err := filepath.Abs(exePath); err == nil {
			exePath = absPath
		}
	}
	
	return fmt.Sprintf(`#!/bin/bash
# Pre-commit hook for AI commit message generator

# Skip if commit message was provided via -m flag
# Git doesn't set any env var for this, but we can detect it by checking
# if we're being called with a message file argument
# For pre-commit hooks, if a message is provided, git will call prepare-commit-msg instead
# So we only run if no message was provided (i.e., bare 'git commit')

# Check if there are staged changes
if ! git diff --staged --quiet; then
    # Generate commit message
    COMMIT_MSG=$("%s")
    EXIT_CODE=$?
    
    if [ $EXIT_CODE -ne 0 ]; then
        echo "Error generating commit message: $COMMIT_MSG"
        exit 1
    fi
    
    # Extract just the message (skip "Generating commit message..." line)
    COMMIT_MSG=$(echo "$COMMIT_MSG" | grep -v "Generating commit message" | perl -pe 's/\e\[[0-9;]*m//g' | sed 's/^[[:space:]]*//' | sed '/^$/d')
    
    if [ -z "$COMMIT_MSG" ]; then
        echo "No commit message generated"
        exit 1
    fi
    
    # Display the generated message (with colors intact)
    echo ""
    echo "Generated commit message:"
    echo "=========================="
    echo "$COMMIT_MSG"
    echo "=========================="
    echo ""
    echo "Options:"
    echo "  [A]ccept and commit"
    echo "  [R]eject (abort commit)"
    echo "  [E]dit message"
    echo ""
    
    # Read from terminal explicitly to handle git hook context
    exec < /dev/tty
    read -p "Your choice (A/R/E): " choice
    
    # Strip ANSI color codes for the actual commit
    CLEAN_MSG=$(echo "$COMMIT_MSG" | sed 's/\x1b\[[0-9;]*m//g')
    
    case "$choice" in
        [Aa]*)
            # Accept: commit with the generated message (colors stripped)
            git commit -m "$CLEAN_MSG" --no-verify
            # Exit with error to prevent original commit from proceeding
            # (since we already committed)
            exit 1
            ;;
        [Rr]*)
            # Reject: abort the commit
            echo "Commit aborted by user"
            exit 1
            ;;
        [Ee]*)
            # Edit: allow user to modify (colors stripped)
            echo "$CLEAN_MSG" > /tmp/commit_msg.txt
            ${EDITOR:-nano} /tmp/commit_msg.txt
            EDITED_MSG=$(cat /tmp/commit_msg.txt)
            git commit -m "$EDITED_MSG" --no-verify
            rm -f /tmp/commit_msg.txt
            # Exit with error to prevent original commit from proceeding
            exit 1
            ;;
        *)
            echo "Invalid choice. Aborting commit."
            exit 1
            ;;
    esac
fi
`, exePath)
}

// generateWindowsHook generates a batch pre-commit hook for Windows
func (a *App) generateWindowsHook() string {
	exePath, err := os.Executable()
	if err != nil {
		exePath = "generate-commit"
	} else {
		// Ensure we have an absolute path
		if absPath, err := filepath.Abs(exePath); err == nil {
			exePath = absPath
		}
	}

	return fmt.Sprintf(`@echo off
REM Pre-commit hook for AI commit message generator (Windows)

REM Check if there are staged changes
git diff --staged --quiet >nul 2>&1
if %%errorlevel%% equ 0 exit /b 0

REM Generate commit message
for /f "delims=" %%%%i in ('"%s" 2^>^&1') do set OUTPUT=%%%%i
if errorlevel 1 (
    echo Error generating commit message
    exit /b 1
)

REM Extract commit message (basic extraction - may need refinement)
set COMMIT_MSG=%%OUTPUT%%
REM Remove "Generating commit message..." line if present
set COMMIT_MSG=%%COMMIT_MSG:Generating commit message...=%%

if "%%COMMIT_MSG%%"=="" (
    echo No commit message generated
    exit /b 1
)

REM Display the generated message
echo.
echo Generated commit message:
echo ==========================
echo %%COMMIT_MSG%%
echo ==========================
echo.
echo Options:
echo   [A]ccept and commit
echo   [R]eject (abort commit)
echo   [E]dit message
echo.
set /p CHOICE=Your choice (A/R/E): 

if /i "%%CHOICE%%"=="A" goto accept
if /i "%%CHOICE:~0,1%%"=="A" goto accept
if /i "%%CHOICE%%"=="R" goto reject
if /i "%%CHOICE:~0,1%%"=="R" goto reject
if /i "%%CHOICE%%"=="E" goto edit
if /i "%%CHOICE:~0,1%%"=="E" goto edit
echo Invalid choice. Aborting commit.
exit /b 1

:accept
git commit -m "%%COMMIT_MSG%%" --no-verify
exit /b 1

:reject
echo Commit aborted by user
exit /b 1

:edit
echo %%COMMIT_MSG%% > %%TEMP%%\commit_msg.txt
notepad %%TEMP%%\commit_msg.txt
set /p EDITED_MSG=<%%TEMP%%\commit_msg.txt
git commit -m "%%EDITED_MSG%%" --no-verify
del %%TEMP%%\commit_msg.txt
exit /b 1
`, exePath)
}
