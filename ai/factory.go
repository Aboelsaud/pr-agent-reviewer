package ai

import (
	"fmt"
	"os"
)

// ProviderType represents the type of AI provider
type ProviderType string

const (
	// ProviderOpenAI represents the OpenAI provider
	ProviderOpenAI ProviderType = "openai"
	// ProviderOllama represents the Ollama provider
	ProviderOllama ProviderType = "ollama"
)

// NewProvider creates a new AI provider based on the configuration
func NewProvider() (Provider, error) {
	providerType := ProviderType(os.Getenv("AI_PROVIDER"))
	if providerType == "" {
		providerType = ProviderOpenAI // Default to OpenAI
	}

	switch providerType {
	case ProviderOpenAI:
		return NewOpenAIAdapter(), nil
	case ProviderOllama:
		return NewOllamaAdapter(), nil
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", providerType)
	}
} 