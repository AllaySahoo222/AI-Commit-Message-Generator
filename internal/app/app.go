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

	// 3. Smart Diff Reading
	diff, err := a.Git.GetStagedDiff()
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}

	fmt.Println("Generating commit message...")

	// 4. AI Integration
	message, err := a.AI.GenerateCommitMessage(diff, rules)
	if err != nil {
		return fmt.Errorf("failed to generate commit message: %w", err)
	}

	// 5. Output
	// Check if the response suggests splitting (multi-line or specific keywords)
	// Heuristic: If it has multiple lines, it's likely a split suggestion or discussion.
	// Conventional commits are typically single line (subject).
	if strings.Contains(message, "\n") {
		// Output split suggestion in Yellow
		fmt.Println("\n\033[33mAI Suggestion (Split Changes):\033[0m")
		fmt.Println(message)
	} else {
		// Output commit message in Cyan
		fmt.Println("\n\033[36m" + message + "\033[0m")
	}

	return nil
}

// Init initializes the repository with config, rules file, and pre-commit hook
func (a *App) Init() error {
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
	configExists, err := a.ConfigLoader.ConfigExists()
	if err != nil {
		return fmt.Errorf("failed to check config existence: %w", err)
	}
	if configExists {
		fmt.Println("Repository already initialized. Use --force to reinitialize.")
		return nil
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
	return `#!/bin/bash
# Pre-commit hook for AI commit message generator

# Check if there are staged changes
if ! git diff --staged --quiet; then
    # Generate commit message
    COMMIT_MSG=$(generate-commit 2>&1)
    EXIT_CODE=$?
    
    if [ $EXIT_CODE -ne 0 ]; then
        echo "Error generating commit message: $COMMIT_MSG"
        exit 1
    fi
    
    # Extract just the message (skip "Generating commit message..." line)
    COMMIT_MSG=$(echo "$COMMIT_MSG" | grep -v "Generating commit message" | sed 's/^[[:space:]]*//' | sed '/^$/d')
    
    if [ -z "$COMMIT_MSG" ]; then
        echo "No commit message generated"
        exit 1
    fi
    
    # Display the generated message
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
    read -p "Your choice (A/R/E): " choice
    
    case "$choice" in
        [Aa]*)
            # Accept: commit with the generated message
            git commit -m "$COMMIT_MSG" --no-verify
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
            # Edit: allow user to modify
            echo "$COMMIT_MSG" > /tmp/commit_msg.txt
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
`
}

// generateWindowsHook generates a batch pre-commit hook for Windows
func (a *App) generateWindowsHook() string {
	return "@echo off\n" +
		"REM Pre-commit hook for AI commit message generator (Windows)\n\n" +
		"REM Check if there are staged changes\n" +
		"git diff --staged --quiet >nul 2>&1\n" +
		"if %errorlevel% equ 0 exit /b 0\n\n" +
		"REM Generate commit message\n" +
		"for /f \"delims=\" %%i in ('generate-commit 2^>^&1') do set OUTPUT=%%i\n" +
		"if errorlevel 1 (\n" +
		"    echo Error generating commit message\n" +
		"    exit /b 1\n" +
		")\n\n" +
		"REM Extract commit message (basic extraction - may need refinement)\n" +
		"set COMMIT_MSG=%OUTPUT%\n" +
		"REM Remove \"Generating commit message...\" line if present\n" +
		"set COMMIT_MSG=%COMMIT_MSG:Generating commit message...=%\n\n" +
		"if \"%COMMIT_MSG%\"==\"\" (\n" +
		"    echo No commit message generated\n" +
		"    exit /b 1\n" +
		")\n\n" +
		"REM Display the generated message\n" +
		"echo.\n" +
		"echo Generated commit message:\n" +
		"echo ==========================\n" +
		"echo %COMMIT_MSG%\n" +
		"echo ==========================\n" +
		"echo.\n" +
		"echo Options:\n" +
		"echo   [A]ccept and commit\n" +
		"echo   [R]eject (abort commit)\n" +
		"echo   [E]dit message\n" +
		"echo.\n" +
		"set /p CHOICE=Your choice (A/R/E): \n\n" +
		"if /i \"%CHOICE%\"==\"A\" goto accept\n" +
		"if /i \"%CHOICE:~0,1%\"==\"A\" goto accept\n" +
		"if /i \"%CHOICE%\"==\"R\" goto reject\n" +
		"if /i \"%CHOICE:~0,1%\"==\"R\" goto reject\n" +
		"if /i \"%CHOICE%\"==\"E\" goto edit\n" +
		"if /i \"%CHOICE:~0,1%\"==\"E\" goto edit\n" +
		"echo Invalid choice. Aborting commit.\n" +
		"exit /b 1\n\n" +
		":accept\n" +
		"git commit -m \"%COMMIT_MSG%\" --no-verify\n" +
		"exit /b 1\n\n" +
		":reject\n" +
		"echo Commit aborted by user\n" +
		"exit /b 1\n\n" +
		":edit\n" +
		"echo %COMMIT_MSG% > %TEMP%\\commit_msg.txt\n" +
		"notepad %TEMP%\\commit_msg.txt\n" +
		"set /p EDITED_MSG=<%TEMP%\\commit_msg.txt\n" +
		"git commit -m \"%EDITED_MSG%\" --no-verify\n" +
		"del %TEMP%\\commit_msg.txt\n" +
		"exit /b 1\n"
}
