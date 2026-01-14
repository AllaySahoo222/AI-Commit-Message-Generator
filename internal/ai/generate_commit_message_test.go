package ai

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOllamaClient_GenerateCommitMessage(t *testing.T) {
	tests := []struct {
		name           string
		diff           string
		rules          string
		mockResponse   string
		mockStatusCode int
		expectedMsg    string
		expectedErr    string
	}{
		{
			name:  "Success",
			diff:  "diff content",
			rules: "some rules",
			mockResponse: `{
				"response": "feat: added login",
				"done": true
			}`,
			mockStatusCode: http.StatusOK,
			expectedMsg:    "feat: added login",
			expectedErr:    "",
		},
		{
			name:           "API Error",
			diff:           "diff",
			rules:          "",
			mockResponse:   `{"error": "bad request"}`,
			mockStatusCode: http.StatusBadRequest,
			expectedMsg:    "",
			expectedErr:    "API returned error: 400 Bad Request",
		},
		{
			name:           "Empty Response",
			diff:           "diff",
			rules:          "",
			mockResponse:   `{"response": "", "done": true}`,
			mockStatusCode: http.StatusOK,
			expectedMsg:    "",
			expectedErr:    "empty response from model",
		},
		{
			name:           "RateLimit_Retry",
			diff:           "diff",
			rules:          "",
			mockResponse:   `{"response": "retry success", "done": true}`,
			mockStatusCode: http.StatusOK,
			expectedMsg:    "retry success",
			expectedErr:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			// Create a mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				// Verify URL
				if !strings.Contains(r.URL.Path, "generate") {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				
				// Verify Method
				if r.Method != "POST" {
					t.Errorf("unexpected method: %s", r.Method)
				}

				// Verify Authorization header
				authHeader := r.Header.Get("Authorization")
				if !strings.HasPrefix(authHeader, "Bearer ") {
					t.Errorf("missing or invalid Authorization header: %s", authHeader)
				}

				// Simulate 429 for the RateLimit_Retry test case
				if tt.name == "RateLimit_Retry" && callCount <= 2 {
					w.WriteHeader(429)
					w.Write([]byte(`{"error": "rate limit"}`))
					return
				}

				// Write response
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// Create client and inject mock server URL
			client := &OllamaClient{
				apiKey:  "test-api-key",
				baseURL: server.URL + "/api/generate",
				client: &http.Client{
					Timeout: 1 * time.Second,
				},
			}

			msg, err := client.GenerateCommitMessage(tt.diff, tt.rules)

			if tt.expectedErr != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedErr)
				} else if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("expected error containing %q, got %q", tt.expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if msg != tt.expectedMsg {
					t.Errorf("expected message %q, got %q", tt.expectedMsg, msg)
				}
			}
		})
	}
}
