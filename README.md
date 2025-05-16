# PR Agent Reviewer

A Go-based PR review agent that automatically reviews pull requests using AI and posts reviews as comments.

## Features

- Listens for GitHub webhook events for new PRs
- Analyzes PR changes using OpenAI
- Posts detailed reviews as PR comments
- Sends notifications to Slack

## Setup

1. Clone this repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Create a `.env` file based on `.env.example` and fill in your credentials:

   - GitHub webhook secret
   - GitHub access token
   - OpenAI API key
   - Slack bot token
   - Slack channel ID

4. Run the application:
   ```bash
   go run main.go
   ```

## Configuration

The application requires the following environment variables:

- `GITHUB_WEBHOOK_SECRET`: Secret for GitHub webhook verification
- `GITHUB_ACCESS_TOKEN`: GitHub personal access token
- `OPENAI_API_KEY`: OpenAI API key
- `SLACK_BOT_TOKEN`: Slack bot token
- `SLACK_CHANNEL_ID`: Slack channel ID for notifications

## Usage

1. Set up a GitHub webhook in your repository pointing to your server's `/webhook` endpoint
2. The agent will automatically review new PRs and post comments
3. Review summaries will be posted to the configured Slack channel
