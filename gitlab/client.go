package gitlab

import (
	"fmt"
	"os"

	"pr-agent-reviewer/logger"

	"github.com/xanzy/go-gitlab"
)

// Client represents a GitLab client
type Client struct {
	client *gitlab.Client
}

// NewClient creates a new GitLab client
func NewClient() *Client {
	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		logger.LogError("GITLAB_TOKEN environment variable is not set", nil)
		return nil
	}

	client, err := gitlab.NewClient(token)
	if err != nil {
		logger.LogError("Failed to create GitLab client", err)
		return nil
	}

	return &Client{client: client}
}

// GetChanges implements the vcs.Provider interface
func (c *Client) GetChanges(repo string, mrNumber int) ([]string, error) {
	logger.LogInfo("Getting changes for MR #%d in %s", mrNumber, repo)

	diffs, _, err := c.client.MergeRequests.ListMergeRequestDiffs(repo, mrNumber, &gitlab.ListMergeRequestDiffsOptions{})
	if err != nil {
		logger.LogError("Failed to get MR changes", err)
		return nil, fmt.Errorf("failed to get MR changes: %v", err)
	}

	var changes []string
	for _, diff := range diffs {
		changes = append(changes, fmt.Sprintf("File: %s\nPatch:\n%s", diff.NewPath, diff.Diff))
	}

	return changes, nil
}

// CreateReview implements the vcs.Provider interface
func (c *Client) CreateReview(repo string, mrNumber int, review string) error {
	logger.LogInfo("Creating review for MR #%d in %s", mrNumber, repo)

	note := &gitlab.CreateMergeRequestNoteOptions{
		Body: gitlab.String(review),
	}

	_, _, err := c.client.Notes.CreateMergeRequestNote(repo, mrNumber, note)
	if err != nil {
		logger.LogError("Failed to create MR review", err)
		return fmt.Errorf("failed to create MR review: %v", err)
	}

	return nil
} 