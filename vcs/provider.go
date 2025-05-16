package vcs

// Provider defines the interface for VCS providers
type Provider interface {
	// GetChanges gets the changes in a pull/merge request
	GetChanges(repo string, prNumber int) ([]string, error)
	
	// CreateReview creates a review on a pull/merge request
	CreateReview(repo string, prNumber int, review string) error
} 