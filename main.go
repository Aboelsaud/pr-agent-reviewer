package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"pr-agent-reviewer/ai"
	"pr-agent-reviewer/logger"
	"pr-agent-reviewer/slack"
	"pr-agent-reviewer/vcs"

	"pr-agent-reviewer/types"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var (
	vcsProvider vcs.Provider
	aiProvider  ai.Provider
	slClient    *slack.Client
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.LogError("Failed to load .env file", err)
	}

	// Initialize VCS provider
	var err error
	vcsProvider, err = vcs.NewProvider()
	if err != nil {
		logger.LogError("Failed to initialize VCS provider", err)
		os.Exit(1)
	}
	
	// Initialize AI provider
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
	// Determine the provider type
	providerType := os.Getenv("VCS_PROVIDER")
	if providerType == "" {
		providerType = "github" // Default to GitHub
	}

	// Verify webhook signature based on provider
	if !verifyWebhookSignature(r, providerType) {
		logger.LogError("Webhook signature verification failed", nil)
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.LogError("Failed to read request body", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	// Restore the body for later use
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	if providerType == "gitlab" {
		handleGitLabWebhook(w, r, body)
	} else {
		handleGitHubWebhook(w, r, body)
	}
}

func handleGitHubWebhook(w http.ResponseWriter, r *http.Request, body []byte) {
	var webhook types.PRWebhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		logger.LogError("Failed to decode GitHub webhook payload", err)
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
	go processPR(webhook.PullRequest.Number, webhook.Repository.FullName, webhook.PullRequest.Title, webhook.PullRequest.URL)

	w.WriteHeader(http.StatusOK)
}

func handleGitLabWebhook(w http.ResponseWriter, r *http.Request, body []byte) {
	var webhook types.GitLabWebhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		logger.LogError("Failed to decode GitLab webhook payload", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Only process merge requests
	if webhook.ObjectKind != "merge_request" {
		logger.LogInfo("Skipping non-merge request event: %s", webhook.ObjectKind)
		w.WriteHeader(http.StatusOK)
		return
	}

	logger.LogWebhook("merge_request", webhook.ObjectAttributes.Action, webhook)

	// Only process opened MRs
	if webhook.ObjectAttributes.Action != "open" && webhook.ObjectAttributes.Action != "reopen" {
		logger.LogInfo("Skipping MR #%d: action is %s", webhook.ObjectAttributes.IID, webhook.ObjectAttributes.Action)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Process MR in a goroutine
	go processPR(webhook.ObjectAttributes.IID, webhook.Project.PathWithNamespace, webhook.ObjectAttributes.Title, webhook.ObjectAttributes.URL)

	w.WriteHeader(http.StatusOK)
}

func verifyWebhookSignature(r *http.Request, providerType string) bool {
	if providerType == "gitlab" {
		return verifyGitLabWebhook(r)
	}
	return verifyGitHubWebhook(r)
}

func verifyGitHubWebhook(r *http.Request) bool {
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if secret == "" {
		logger.LogDebug("No GitHub webhook secret set, skipping verification")
		return true
	}

	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		logger.LogError("Missing GitHub webhook signature", nil)
		return false
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.LogError("Failed to read request body", err)
		return false
	}
	// Restore the body for later use
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	isValid := hmac.Equal([]byte(signature), []byte(expectedSignature))
	if !isValid {
		logger.LogError("Invalid GitHub webhook signature", nil)
	}
	return isValid
}

func verifyGitLabWebhook(r *http.Request) bool {
	token := os.Getenv("GITLAB_WEBHOOK_TOKEN")
	if token == "" {
		logger.LogDebug("No GitLab webhook token set, skipping verification")
		return true
	}

	signature := r.Header.Get("X-GitLab-Token")
	if signature == "" {
		logger.LogError("Missing GitLab webhook token", nil)
		return false
	}

	isValid := signature == token
	if !isValid {
		logger.LogError("Invalid GitLab webhook token", nil)
	}
	return isValid
}

func processPR(prNumber int, repo string, title string, url string) {
	logger.LogPRReview(prNumber, repo, "started")

	// Get PR changes
	changes, err := vcsProvider.GetChanges(repo, prNumber)
	if err != nil {
		logger.LogError("Failed to get PR changes", err)
		return
	}
	logger.LogInfo("Retrieved %d files from PR #%d", len(changes), prNumber)

	// Get AI review
	review, err := aiProvider.ReviewCode(changes)
	if err != nil {
		logger.LogError("Failed to get AI review", err)
		return
	}
	logger.LogInfo("Generated AI review for PR #%d", prNumber)

	// Generate review summary
	summary, err := aiProvider.GenerateReviewSummary(review)
	if err != nil {
		logger.LogError("Failed to generate review summary", err)
		return
	}
	logger.LogInfo("Generated review summary for PR #%d", prNumber)

	// Create review
	if err := vcsProvider.CreateReview(repo, prNumber, review); err != nil {
		logger.LogError("Failed to create review", err)
		return
	}
	logger.LogPRReview(prNumber, repo, "review posted")

	// Send Slack notification
	if err := slClient.SendPRReviewNotification(title, url, summary); err != nil {
		logger.LogError("Failed to send Slack notification", err)
		return
	}
	logger.LogSlackNotification(os.Getenv("SLACK_CHANNEL_ID"), "PR review summary")

	logger.LogPRReview(prNumber, repo, "completed")
} 