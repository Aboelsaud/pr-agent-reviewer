package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"pr-agent-reviewer/ai"
	ghclient "pr-agent-reviewer/github"
	"pr-agent-reviewer/logger"
	"pr-agent-reviewer/slack"

	gh "github.com/google/go-github/v57/github"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type PRWebhook struct {
	Action      string `json:"action"`
	PullRequest struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		URL    string `json:"html_url"`
	} `json:"pull_request"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

var (
	ghClient  *ghclient.Client
	aiProvider ai.Provider
	slClient  *slack.Client
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.LogError("Failed to load .env file", err)
	}

	// Initialize clients
	ghClient = ghclient.NewClient()
	
	// Initialize AI provider
	var err error
	aiProvider, err = ai.NewProvider()
	if err != nil {
		logger.LogError("Failed to initialize AI provider", err)
		os.Exit(1)
	}
	
	slClient = slack.NewClient()

	// Initialize router
	r := mux.NewRouter()

	// Middleware for logging
	r.Use(loggingMiddleware)

	// Webhook endpoint
	r.HandleFunc("/webhook", handleWebhook).Methods("POST")

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.LogInfo("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		logger.LogError("Server failed to start", err)
		os.Exit(1)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		logger.LogRequest(r.Method, r.URL.Path, r.RemoteAddr, duration)
	})
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Verify GitHub webhook signature
	if !verifyGitHubWebhook(r) {
		logger.LogError("Webhook signature verification failed", nil)
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse webhook payload
	var webhook PRWebhook
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		logger.LogError("Failed to decode webhook payload", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	logger.LogWebhook("pull_request", webhook.Action, webhook)

	// Only process opened PRs
	if webhook.Action != "opened" && webhook.Action != "reopened" {
		logger.LogInfo("Skipping PR #%d: action is %s", webhook.PullRequest.Number, webhook.Action)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Process PR in a goroutine
	go processPR(webhook)

	w.WriteHeader(http.StatusOK)
}

func verifyGitHubWebhook(r *http.Request) bool {
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if secret == "" {
		logger.LogDebug("No webhook secret set, skipping verification")
		return true
	}

	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		logger.LogError("Missing webhook signature", nil)
		return false
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.LogError("Failed to read request body", err)
		return false
	}
	// Restore the body for later use
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	isValid := hmac.Equal([]byte(signature), []byte(expectedSignature))
	if !isValid {
		logger.LogError("Invalid webhook signature", nil)
	}
	return isValid
}

func processPR(webhook PRWebhook) {
	// Split repository full name into owner and repo
	parts := strings.Split(webhook.Repository.FullName, "/")
	if len(parts) != 2 {
		logger.LogError("Invalid repository name", fmt.Errorf("invalid format: %s", webhook.Repository.FullName))
		return
	}
	owner, repo := parts[0], parts[1]

	logger.LogPRReview(webhook.PullRequest.Number, webhook.Repository.FullName, "started")

	// Get PR changes
	files, err := ghClient.GetPRChanges(owner, repo, webhook.PullRequest.Number)
	if err != nil {
		logger.LogError("Failed to get PR changes", err)
		return
	}
	logger.LogInfo("Retrieved %d files from PR #%d", len(files), webhook.PullRequest.Number)

	// Prepare changes for review
	var changes []string
	for _, file := range files {
		changes = append(changes, fmt.Sprintf("File: %s\nPatch:\n%s", file.GetFilename(), file.GetPatch()))
	}

	// Get AI review
	review, err := aiProvider.ReviewCode(changes)
	if err != nil {
		logger.LogError("Failed to get AI review", err)
		return
	}
	logger.LogInfo("Generated AI review for PR #%d: %s", webhook.PullRequest.Number, review)

	// Generate review summary
	summary, err := aiProvider.GenerateReviewSummary(review)
	if err != nil {
		logger.LogError("Failed to generate review summary", err)
		return
	}
	logger.LogInfo("Generated review summary for PR #%d: %s", webhook.PullRequest.Number, summary)

	// Create GitHub review
	reviewRequest := &gh.PullRequestReviewRequest{
		Body:  gh.String(review),
		Event: gh.String("COMMENT"),
	}
	if err := ghClient.CreateReview(owner, repo, webhook.PullRequest.Number, reviewRequest); err != nil {
		logger.LogError("Failed to create GitHub review", err)
		return
	}
	logger.LogPRReview(webhook.PullRequest.Number, webhook.Repository.FullName, "review posted")

	// Send Slack notification
	if err := slClient.SendPRReviewNotification(
		webhook.PullRequest.Title,
		webhook.PullRequest.URL,
		summary,
	); err != nil {
		logger.LogError("Failed to send Slack notification", err)
		return
	}
	logger.LogSlackNotification(os.Getenv("SLACK_CHANNEL_ID"), "PR review summary")

	logger.LogPRReview(webhook.PullRequest.Number, webhook.Repository.FullName, "completed")
} 