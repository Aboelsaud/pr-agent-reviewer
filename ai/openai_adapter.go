package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"pr-agent-reviewer/logger"

	"github.com/sashabaranov/go-openai"
)

// OpenAIAdapter implements the Provider interface for OpenAI
type OpenAIAdapter struct {
	client *openai.Client
}

// NewOpenAIAdapter creates a new OpenAI adapter
func NewOpenAIAdapter() *OpenAIAdapter {
	logger.LogInfo("Initializing OpenAI adapter")
	return &OpenAIAdapter{
		client: openai.NewClient(os.Getenv("OPENAI_API_KEY")),
	}
}

// ReviewCode implements the Provider interface for OpenAI
func (a *OpenAIAdapter) ReviewCode(changes []string) (string, error) {
	prompt := "Please review the following code changes and provide a detailed review. " +
		"Focus on code quality, potential bugs, and best practices. " +
		"Format the review in markdown.\n\nChanges:\n" + strings.Join(changes, "\n\n")

	logger.LogOpenAIRequest("gpt-4", len(prompt))

	start := time.Now()
	resp, err := a.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an experienced code reviewer. Provide detailed, constructive feedback on code changes.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		logger.LogError("OpenAI review request failed", err)
		return "", fmt.Errorf("failed to get OpenAI response: %v", err)
	}

	duration := time.Since(start)
	logger.LogOpenAIResponse("gpt-4", len(resp.Choices[0].Message.Content), duration)

	return resp.Choices[0].Message.Content, nil
}

// GenerateReviewSummary implements the Provider interface for OpenAI
func (a *OpenAIAdapter) GenerateReviewSummary(review string) (string, error) {
	prompt := "Please provide a brief summary (2-3 sentences) of the following code review:\n\n" + review

	logger.LogOpenAIRequest("gpt-4", len(prompt))

	start := time.Now()
	resp, err := a.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a technical writer. Create concise summaries of code reviews.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		logger.LogError("OpenAI summary request failed", err)
		return "", fmt.Errorf("failed to get OpenAI response: %v", err)
	}

	duration := time.Since(start)
	logger.LogOpenAIResponse("gpt-4", len(resp.Choices[0].Message.Content), duration)

	return resp.Choices[0].Message.Content, nil
} 