package logger

import (
	"log"
	"os"
	"time"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
)

func init() {
	InfoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime)
	DebugLogger = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime)
}

// LogRequest logs HTTP request details
func LogRequest(method, path, remoteAddr string, duration time.Duration) {
	InfoLogger.Printf("Request: %s %s from %s (took %v)", method, path, remoteAddr, duration)
}

// LogError logs error messages with context
func LogError(context string, err error) {
	ErrorLogger.Printf("%s: %v", context, err)
}

// LogInfo logs informational messages
func LogInfo(format string, v ...interface{}) {
	InfoLogger.Printf(format, v...)
}

// LogDebug logs debug messages
func LogDebug(format string, v ...interface{}) {
	DebugLogger.Printf(format, v...)
}

// LogWebhook logs webhook events
func LogWebhook(eventType, action string, payload interface{}) {
	InfoLogger.Printf("Webhook received - Type: %s, Action: %s, Payload: %+v", eventType, action, payload)
}

// LogPRReview logs PR review events
func LogPRReview(prNumber int, repo string, action string) {
	InfoLogger.Printf("PR Review - #%d in %s: %s", prNumber, repo, action)
}

// LogSlackNotification logs Slack notification events
func LogSlackNotification(channelID, messageType string) {
	InfoLogger.Printf("Slack notification sent to channel %s: %s", channelID, messageType)
}

// LogOpenAIRequest logs OpenAI API requests
func LogOpenAIRequest(model string, promptLength int) {
	InfoLogger.Printf("OpenAI request - Model: %s, Prompt length: %d", model, promptLength)
}

// LogOpenAIResponse logs OpenAI API responses
func LogOpenAIResponse(model string, responseLength int, duration time.Duration) {
	InfoLogger.Printf("OpenAI response - Model: %s, Response length: %d, Duration: %v", model, responseLength, duration)
} 