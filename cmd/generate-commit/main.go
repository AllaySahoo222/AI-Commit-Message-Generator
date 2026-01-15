package main

import (
	"fmt"
	"os"

	"ai-commit-message-generator/internal/ai"
	"ai-commit-message-generator/internal/app"
	"ai-commit-message-generator/internal/config"
	"ai-commit-message-generator/internal/git"
)

func main() {
	if len(os.Args) < 2 {
		// Default behavior: generate commit message
		runGenerate()
		return
	}

	command := os.Args[1]
	switch command {
	case "init":
		runInit()
	case "generate", "gen":
		runGenerate()
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Fprintf(os.Stderr, "Run 'generate-commit help' for usage information.\n")
		os.Exit(1)
	}
}

func runInit() {
	gitClient := git.NewClient()
	rulesLoader := config.NewLoader()
	configLoader := config.NewConfigLoader()

	application := app.NewApp(gitClient, rulesLoader, configLoader, nil)

	if err := application.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runGenerate() {
	gitClient := git.NewClient()
	rulesLoader := config.NewLoader()
	configLoader := config.NewConfigLoader()

	// Load configuration
	cfg, err := configLoader.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Check for API key
	if cfg.APIKey == "" {
		fmt.Fprintf(os.Stderr, "Error: OLLAMA_API_KEY environment variable is not set and not found in config.\n")
		fmt.Fprintf(os.Stderr, "Please set your Ollama API key:\n")
		fmt.Fprintf(os.Stderr, "  export OLLAMA_API_KEY=your_api_key\n")
		fmt.Fprintf(os.Stderr, "  or add it to .commit-generator-config\n")
		os.Exit(1)
	}

	aiClient := ai.NewClient(cfg.APIKey, cfg.BaseURL, cfg.Model, cfg.GetTimeout())
	application := app.NewApp(gitClient, rulesLoader, configLoader, aiClient)

	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("AI Commit Message Generator")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  generate-commit [command]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  init       Initialize repository with config, rules, and pre-commit hook")
	fmt.Println("  generate   Generate commit message from staged changes (default)")
	fmt.Println("  help       Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  generate-commit init              # Initialize the repository")
	fmt.Println("  generate-commit generate          # Generate commit message")
	fmt.Println("  generate-commit                   # Same as 'generate'")
}
