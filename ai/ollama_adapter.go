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

	"pr-agent-reviewer/logger"
)

// OllamaAdapter implements the Provider interface for Ollama
type OllamaAdapter struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// OllamaRequest represents a request to the Ollama API
type OllamaRequest struct {
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
	System   string `json:"system,omitempty"`
	Stream   bool   `json:"stream"`
}

// OllamaResponse represents a response from the Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// NewOllamaAdapter creates a new Ollama adapter
func NewOllamaAdapter() *OllamaAdapter {
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "codellama" // Default to CodeLlama
	}

	logger.LogInfo("Initializing Ollama adapter with model: %s", model)
	return &OllamaAdapter{
		baseURL:    baseURL,
		model:      model,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// ReviewCode implements the Provider interface for Ollama
func (a *OllamaAdapter) ReviewCode(changes []string) (string, error) {
	prompt := `You are an experienced code reviewer. Please review the following code changes and provide a detailed review.
Focus on:
1. Code quality and best practices
2. Potential bugs or issues
3. Security concerns
4. Performance implications
5. Maintainability

Format your response in markdown with clear sections.

Code changes to review:
` + strings.Join(changes, "\n\n")

	logger.LogInfo("Ollama review request - Model: %s, Prompt length: %d", a.model, len(prompt))

	req := OllamaRequest{
		Model:  a.model,
		Prompt: prompt,
		System: "You are an expert code reviewer with deep knowledge of software engineering best practices. Provide detailed, actionable feedback on code changes.",
		Stream: false,
	}

	start := time.Now()
	resp, err := a.sendRequest(req)
	if err != nil {
		logger.LogError("Ollama review request failed", err)
		return "", fmt.Errorf("failed to get Ollama response: %v", err)
	}

	duration := time.Since(start)
	content := resp.Response

	// Validate response
	if len(content) < 50 {
		logger.LogError("Ollama returned suspiciously short response", fmt.Errorf("response length: %d", len(content)))
		return "", fmt.Errorf("invalid response from Ollama: response too short")
	}

	logger.LogInfo("Ollama response - Model: %s, Response length: %d, Duration: %v",
		a.model, len(content), duration)

	return content, nil
}

// GenerateReviewSummary implements the Provider interface for Ollama
func (a *OllamaAdapter) GenerateReviewSummary(review string) (string, error) {
	prompt := `You are a technical writer. Please provide a concise summary (2-3 sentences) of the following code review.
Focus on the key points and main recommendations.

Code review to summarize:
` + review

	logger.LogInfo("Ollama summary request - Model: %s, Prompt length: %d", a.model, len(prompt))

	req := OllamaRequest{
		Model:  a.model,
		Prompt: prompt,
		System: "You are a technical writer who creates clear, concise summaries of code reviews.",
		Stream: false,
	}

	start := time.Now()
	resp, err := a.sendRequest(req)
	if err != nil {
		logger.LogError("Ollama summary request failed", err)
		return "", fmt.Errorf("failed to get Ollama response: %v", err)
	}

	duration := time.Since(start)
	content := resp.Response

	// Validate response
	if len(content) < 20 {
		logger.LogError("Ollama returned suspiciously short summary", fmt.Errorf("summary length: %d", len(content)))
		return "", fmt.Errorf("invalid response from Ollama: summary too short")
	}

	logger.LogInfo("Ollama response - Model: %s, Response length: %d, Duration: %v",
		a.model, len(content), duration)

	return content, nil
}

// sendRequest sends a request to the Ollama API
func (a *OllamaAdapter) sendRequest(req OllamaRequest) (*OllamaResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := a.httpClient.Post(
		a.baseURL+"/api/generate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Validate response content
	if ollamaResp.Response == "" {
		return nil, fmt.Errorf("empty response from Ollama")
	}

	return &ollamaResp, nil
} 