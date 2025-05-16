package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body
	expected := `OK`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestWebhookHandler(t *testing.T) {
	// Set up test environment
	os.Setenv("GITHUB_WEBHOOK_SECRET", "test-secret")

	// Create a test webhook payload
	webhook := PRWebhook{
		Action: "opened",
		PullRequest: struct {
			Number int    `json:"number"`
			Title  string `json:"title"`
			Body   string `json:"body"`
			URL    string `json:"html_url"`
		}{
			Number: 1,
			Title:  "Test PR",
			Body:   "Test PR body",
			URL:    "https://github.com/test/repo/pull/1",
		},
		Repository: struct {
			FullName string `json:"full_name"`
		}{
			FullName: "test/repo",
		},
	}

	// Convert webhook to JSON
	payload, err := json.Marshal(webhook)
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to pass to our handler
	req, err := http.NewRequest("POST", "/webhook", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}

	// Add GitHub webhook signature header
	req.Header.Set("X-Hub-Signature-256", "sha256=test-signature")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleWebhook)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}
} 