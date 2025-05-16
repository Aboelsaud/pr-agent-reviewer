package slack

import (
	"fmt"
	"os"

	"pr-agent-reviewer/logger"

	"github.com/slack-go/slack"
)

type Client struct {
	client *slack.Client
}

func NewClient() *Client {
	logger.LogInfo("Initializing Slack client")
	return &Client{
		client: slack.New(os.Getenv("SLACK_BOT_TOKEN")),
	}
}

func (c *Client) SendPRReviewNotification(prTitle, prURL, reviewSummary string) error {
	channelID := os.Getenv("SLACK_CHANNEL_ID")
	
	logger.LogInfo("Preparing Slack notification for PR: %s", prTitle)
	
	// Format the message
	message := fmt.Sprintf("*New PR Review*\nTitle: %s\nURL: %s\n\nReview Summary:\n%s",
		prTitle, prURL, reviewSummary)

	// Send the message
	_, _, err := c.client.PostMessage(
		channelID,
		slack.MsgOptionText(message, false),
	)
	if err != nil {
		logger.LogError("Failed to send Slack message", err)
		return fmt.Errorf("failed to send Slack message: %v", err)
	}

	logger.LogInfo("Successfully sent Slack notification to channel %s", channelID)
	return nil
} 