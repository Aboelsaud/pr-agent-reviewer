package github

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"pr-agent-reviewer/logger"

	gh "github.com/google/go-github/v57/github"
)

// Client represents a GitHub client
type Client struct {
	client *gh.Client
}

// NewClient creates a new GitHub client
func NewClient() *Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		logger.LogError("GITHUB_TOKEN environment variable is not set", nil)
		return nil
	}

	client := gh.NewClient(nil).WithAuthToken(token)
	return &Client{client: client}
}

// GetChanges implements the vcs.Provider interface
func (c *Client) GetChanges(repo string, prNumber int) ([]string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s", repo)
	}
	owner, repoName := parts[0], parts[1]

	logger.LogInfo("Getting changes for PR #%d in %s", prNumber, repo)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	files, _, err := c.client.PullRequests.ListFiles(
		ctx,
		owner,
		repoName,
		prNumber,
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get PR changes", err)
		return nil, fmt.Errorf("failed to get PR changes: %v", err)
	}

	var changes []string
	for _, file := range files {
		changes = append(changes, fmt.Sprintf("File: %s\nPatch:\n%s", file.GetFilename(), file.GetPatch()))
	}

	return changes, nil
}

// CreateReview implements the vcs.Provider interface
func (c *Client) CreateReview(repo string, prNumber int, review string) error {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format: %s", repo)
	}
	owner, repoName := parts[0], parts[1]

	logger.LogInfo("Creating review for PR #%d in %s", prNumber, repo)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get PR details to check the author
	pr, _, err := c.client.PullRequests.Get(ctx, owner, repoName, prNumber)
	if err != nil {
		return fmt.Errorf("failed to get PR details: %v", err)
	}

	// Determine the review event based on the PR author
	event := "REQUEST_CHANGES"
	botUsername := os.Getenv("GITHUB_BOT_USERNAME")
	if botUsername != "" && pr.GetUser().GetLogin() == botUsername {
		event = "COMMENT"
		logger.LogInfo("Using COMMENT event for self-authored PR (author: %s, bot: %s)", pr.GetUser().GetLogin(), botUsername)
	} else {
		logger.LogInfo("Using REQUEST_CHANGES event (author: %s, bot: %s)", pr.GetUser().GetLogin(), botUsername)
	}

	// Create a review with a comment indicating requested changes
	reviewRequest := &gh.PullRequestReviewRequest{
		Body:  gh.String(review),
		Event: gh.String(event),
	}

	_, _, err = c.client.PullRequests.CreateReview(
		ctx,
		owner,
		repoName,
		prNumber,
		reviewRequest,
	)
	if err != nil {
		logger.LogError("Failed to create PR review", err)
		return fmt.Errorf("failed to create PR review: %v", err)
	}

	return nil
}

func (c *Client) GetPRDetails(owner, repo string, prNumber int) (*gh.PullRequest, error) {
	logger.LogInfo("Fetching details for PR #%d in %s/%s", prNumber, owner, repo)
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pr, _, err := c.client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		logger.LogError("Failed to get PR details", err)
		return nil, fmt.Errorf("failed to get PR details: %v", err)
	}

	logger.LogInfo("Successfully retrieved details for PR #%d", prNumber)
	return pr, nil
} 