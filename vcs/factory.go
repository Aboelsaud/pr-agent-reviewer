package vcs

import (
	"fmt"
	"os"

	"pr-agent-reviewer/github"
	"pr-agent-reviewer/gitlab"
)

// ProviderType represents the type of VCS provider
type ProviderType string

const (
	// ProviderGitHub represents the GitHub provider
	ProviderGitHub ProviderType = "github"
	// ProviderGitLab represents the GitLab provider
	ProviderGitLab ProviderType = "gitlab"
)

// NewProvider creates a new VCS provider based on the configuration
func NewProvider() (Provider, error) {
	providerType := ProviderType(os.Getenv("VCS_PROVIDER"))
	if providerType == "" {
		providerType = ProviderGitHub // Default to GitHub
	}

	switch providerType {
	case ProviderGitHub:
		return github.NewClient(), nil
	case ProviderGitLab:
		return gitlab.NewClient(), nil
	default:
		return nil, fmt.Errorf("unsupported VCS provider: %s", providerType)
	}
} 