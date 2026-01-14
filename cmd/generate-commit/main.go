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
	apiKey := os.Getenv("OLLAMA_API_KEY")
	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "Error: OLLAMA_API_KEY environment variable is not set.\n")
		fmt.Fprintf(os.Stderr, "Please set your Ollama API key:\n")
		fmt.Fprintf(os.Stderr, "export OLLAMA_API_KEY=your_api_key\n")
		os.Exit(1)
	}

	gitClient := git.NewClient()
	configLoader := config.NewLoader()
	aiClient := ai.NewClient(apiKey)

	application := app.NewApp(gitClient, configLoader, aiClient)

	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
