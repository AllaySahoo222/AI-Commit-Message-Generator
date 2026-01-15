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
)

// Client defines the interface for AI operations
type Client interface {
	GenerateCommitMessage(diff string, rules string) (string, error)
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
func (c *OllamaClient) GenerateCommitMessage(diff string, rules string) (string, error) {
	prompt := c.buildPrompt(diff, rules)

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

func (c *OllamaClient) buildPrompt(diff string, rules string) string {
	var sb strings.Builder
	sb.WriteString("You are an expert DevOps engineer specialized in writing git commit messages.\n\n")
	sb.WriteString("Analyze the following code diff.\n\n")
	sb.WriteString("First, determine whether the diff represents a single logical change or multiple independent changes that should be split into smaller commits to follow clean code and best practices.\n\n")
	sb.WriteString("If the diff should be split, briefly state that it can be broken down and list the suggested commit scopes or purposes (do not generate the commits yet).\n\n")
	sb.WriteString("If the diff represents a single logical change, generate a single-line git commit message following the Conventional Commits specification.\n\n")
	sb.WriteString("Format for commit message:\n<type>(<scope>): <description>\n\n")
	sb.WriteString("Allowed types: feat, fix, docs, style, refactor, test, chore.\n\n")
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
