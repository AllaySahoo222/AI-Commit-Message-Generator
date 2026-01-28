package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"ai-commit-message-generator/internal/git"
)

// Client defines the interface for AI operations
type Client interface {
	GenerateCommitMessage(diff string, rules string, gitState *git.GitState) (string, error)
}

// OllamaClient implements the Client interface for Ollama API
type OllamaClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewClient creates a new Ollama AI client from config
func NewClient(apiKey, baseURL, model string, timeout time.Duration) Client {
	if baseURL == "" {
		baseURL = "http://localhost:11434/api/generate"
	}
	if model == "" {
		model = "gpt-oss:120b"
	}
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	return &OllamaClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Request/Response structures for Ollama API
type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// GenerateCommitMessage sends the diff and rules to Ollama and returns the generated message
func (c *OllamaClient) GenerateCommitMessage(diff string, rules string, gitState *git.GitState) (string, error) {
	prompt := c.buildPrompt(diff, rules, gitState)

	reqBody := ollamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Retry loop
	maxRetries := 3
	baseDelay := 2 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Backoff logic
			delay := baseDelay * time.Duration(1<<uint(attempt-1)) // 2s, 4s, 8s
			fmt.Fprintf(os.Stderr, "\033[33mRate limit hit. Retrying in %v...\033[0m\n", delay)
			time.Sleep(delay)
		}

		req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonBody))
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, err := c.client.Do(req)
		if err != nil {
			return "", fmt.Errorf("API call failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			if attempt == maxRetries {
				body, _ := io.ReadAll(resp.Body)
				return "", fmt.Errorf("API rate limit exceeded after %d retries: %s", maxRetries, string(body))
			}
			continue // Retry
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return "", fmt.Errorf("API returned error: %s (body: %s)", resp.Status, string(body))
		}

		var ollamaResp ollamaResponse
		if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
			return "", fmt.Errorf("failed to decode response: %w", err)
		}

		if ollamaResp.Response == "" {
			return "", fmt.Errorf("empty response from model")
		}

		return strings.TrimSpace(ollamaResp.Response), nil
	}
	return "", fmt.Errorf("unreachable")
}

func (c *OllamaClient) buildPrompt(diff string, rules string, gitState *git.GitState) string {
	var sb strings.Builder
	sb.WriteString("You are an expert DevOps engineer specialized in writing git commit messages.\n\n")
	
	// Inject git state context if not normal
	if gitState != nil && gitState.Type != git.StateNormal {
		sb.WriteString("=== SPECIAL GIT STATE CONTEXT ===\n\n")
		
		switch gitState.Type {
		case git.StateMerge:
			sb.WriteString("CONTEXT: You are completing a MERGE CONFLICT resolution.\n")
			if gitState.OriginalMessage != "" {
				sb.WriteString(fmt.Sprintf("Original merge intent: \"%s\"\n", gitState.OriginalMessage))
			}
			sb.WriteString("\nIMPORTANT INSTRUCTIONS:\n")
			sb.WriteString("1. You MUST use the following EXACT format for the first line:\n")
			sb.WriteString("   <type>(merge): Merged <Source_Branch> into <Target_Branch>\n")
			sb.WriteString("2. Analyze the diff and choose the appropriate <type> based on the changes:\n")
			sb.WriteString("   - feat: if adding new features or capabilities\n")
			sb.WriteString("   - fix: if fixing bugs or issues\n")
			sb.WriteString("   - refactor: if restructuring code without changing functionality\n")
			sb.WriteString("   - chore: if updating dependencies, configs, or maintenance tasks\n")
			sb.WriteString("   - docs: if primarily documentation changes\n")
			sb.WriteString("3. Extract <Source_Branch> from the Original merge intent (e.g. 'Merge branch feature-x' -> feature-x).\n")
			sb.WriteString("4. If <Target_Branch> is unknown, use 'main' or infer from the diff/context.\n")
			sb.WriteString("5. After the first line, leave a blank line and then provide a detailed description of what code changes were merged.\n")
			sb.WriteString("6. Explain HOW conflicts were resolved if applicable.\n")
			sb.WriteString("7. Example First Line: feat(merge): Merged feature-auth into main\n\n")
			
		case git.StateRebase:
			sb.WriteString("CONTEXT: You are completing a REBASE conflict resolution.\n")
			if gitState.OriginalMessage != "" {
				sb.WriteString(fmt.Sprintf("Rebase context: %s\n", gitState.OriginalMessage))
			}
			sb.WriteString("\nIMPORTANT INSTRUCTIONS:\n")
			sb.WriteString("1. You MUST use the following EXACT format for the first line:\n")
			sb.WriteString("   <type>(rebase): Rebased <Branch_Name> onto <Target_Branch>\n")
			sb.WriteString("2. Analyze the diff and choose the appropriate <type> based on the changes:\n")
			sb.WriteString("   - feat: if adding new features or capabilities\n")
			sb.WriteString("   - fix: if fixing bugs or issues\n")
			sb.WriteString("   - refactor: if restructuring code without changing functionality\n")
			sb.WriteString("   - chore: if updating dependencies, configs, or maintenance tasks\n")
			sb.WriteString("   - docs: if primarily documentation changes\n")
			sb.WriteString("3. Extract <Branch_Name> from the Rebase context if available, otherwise infer from diff.\n")
			sb.WriteString("4. If <Target_Branch> is unknown, use 'main' or infer from context.\n")
			sb.WriteString("5. After the first line, leave a blank line and then provide a detailed description of what code changes were rebased.\n")
			sb.WriteString("6. Explain HOW conflicts were resolved if applicable.\n")
			sb.WriteString("7. Example First Line: feat(rebase): Rebased feature-auth onto main\n\n")
			
		case git.StateCherryPick:
			sb.WriteString("CONTEXT: You are completing a CHERRY-PICK operation.\n")
			if gitState.OriginalMessage != "" {
				sb.WriteString(fmt.Sprintf("Original commit: \"%s\"\n", gitState.OriginalMessage))
			}
			sb.WriteString("\nIMPORTANT INSTRUCTIONS:\n")
			sb.WriteString("1. You MUST use the following EXACT format for the first line:\n")
			sb.WriteString("   <type>(cherry-pick): Cherry-picked <Commit_Description> into <Target_Branch>\n")
			sb.WriteString("   ⚠️  CRITICAL: The scope MUST be 'cherry-pick', NOT the original scope from the commit!\n")
			sb.WriteString("2. Analyze the diff and choose the appropriate <type> based on the changes:\n")
			sb.WriteString("   - feat: if adding new features or capabilities\n")
			sb.WriteString("   - fix: if fixing bugs or issues\n")
			sb.WriteString("   - refactor: if restructuring code without changing functionality\n")
			sb.WriteString("   - chore: if updating dependencies, configs, or maintenance tasks\n")
			sb.WriteString("   - docs: if primarily documentation changes\n")
			sb.WriteString("3. Extract <Commit_Description> from the Original commit message if available.\n")
			sb.WriteString("4. If <Target_Branch> is unknown, use 'main' or infer from context.\n")
			sb.WriteString("5. After the first line, leave a blank line and then provide a detailed description of what was cherry-picked.\n")
			sb.WriteString("6. Explain HOW conflicts were resolved and what adaptations were made if applicable.\n")
			sb.WriteString("7. CORRECT Example: docs(cherry-pick): Cherry-picked feature entries update into main\n")
			sb.WriteString("8. WRONG Example: docs(file): updated feature entries (missing cherry-pick scope!)\n\n")
		}
		
		sb.WriteString("=================================\n\n")
	}
	
	sb.WriteString("Analyze the following code diff.\n\n")
	sb.WriteString("First, determine whether the diff represents a single logical change or multiple independent changes that should be split into smaller commits to follow clean code and best practices.\n\n")
	sb.WriteString("If the diff should be split, briefly state that it can be broken down and list the suggested commit scopes or purposes (do not generate the commits yet).\n\n")
	sb.WriteString("If the diff represents a single logical change, generate a single-line git commit message following the Conventional Commits specification.\n\n")
	sb.WriteString("Format for commit message:\n<type>(<scope>): <description>\n\n")
	sb.WriteString("Allowed types: feat, fix, docs, style, refactor, test, chore.\n\n")
	sb.WriteString("IMPORTANT: Use past tense for the description (e.g., 'added feature' not 'add feature', 'fixed bug' not 'fix bug').\n\n")
	sb.WriteString("Do not output anything other than the message or the split suggestion.\n\n")

	if rules != "" {
		sb.WriteString("Team Rules:\n")
		sb.WriteString(rules)
		sb.WriteString("\n\n")
	}
	sb.WriteString("Diff:\n")
	sb.WriteString(diff)
	return sb.String()
}
