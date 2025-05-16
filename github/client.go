package github

import (
	"context"
	"fmt"
	"os"

	"pr-agent-reviewer/logger"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	ctx    context.Context
}

func NewClient() *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	logger.LogInfo("GitHub client initialized")
	return &Client{
		client: client,
		ctx:    ctx,
	}
}

func (c *Client) GetPRChanges(owner, repo string, prNumber int) ([]*github.CommitFile, error) {
	logger.LogInfo("Fetching changes for PR #%d in %s/%s", prNumber, owner, repo)
	
	files, _, err := c.client.PullRequests.ListFiles(c.ctx, owner, repo, prNumber, nil)
	if err != nil {
		logger.LogError("Failed to get PR files", err)
		return nil, fmt.Errorf("failed to get PR files: %v", err)
	}

	logger.LogInfo("Successfully retrieved %d files from PR #%d", len(files), prNumber)
	return files, nil
}

func (c *Client) CreateReview(owner, repo string, prNumber int, review *github.PullRequestReviewRequest) error {
	logger.LogInfo("Creating review for PR #%d in %s/%s", prNumber, owner, repo)
	
	_, _, err := c.client.PullRequests.CreateReview(c.ctx, owner, repo, prNumber, review)
	if err != nil {
		logger.LogError("Failed to create review", err)
		return fmt.Errorf("failed to create review: %v", err)
	}

	logger.LogInfo("Successfully created review for PR #%d", prNumber)
	return nil
}

func (c *Client) GetPRDetails(owner, repo string, prNumber int) (*github.PullRequest, error) {
	logger.LogInfo("Fetching details for PR #%d in %s/%s", prNumber, owner, repo)
	
	pr, _, err := c.client.PullRequests.Get(c.ctx, owner, repo, prNumber)
	if err != nil {
		logger.LogError("Failed to get PR details", err)
		return nil, fmt.Errorf("failed to get PR details: %v", err)
	}

	logger.LogInfo("Successfully retrieved details for PR #%d", prNumber)
	return pr, nil
} 