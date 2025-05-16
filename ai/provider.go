package ai

// Provider defines the interface for AI review providers
type Provider interface {
	// ReviewCode reviews the provided code changes and returns a detailed review
	ReviewCode(changes []string) (string, error)
	
	// GenerateReviewSummary generates a brief summary of a review
	GenerateReviewSummary(review string) (string, error)
}

// ReviewRequest represents a request for code review
type ReviewRequest struct {
	Changes []string
	Model   string
}

// ReviewResponse represents a response from the AI provider
type ReviewResponse struct {
	Review  string
	Summary string
	Error   error
} 