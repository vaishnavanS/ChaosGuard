package domain

import "time"

// Recommendation represents remediation advice associated with a detected issue.
type Recommendation struct {
	ID          string    `json:"id"`           // Unique recommendation UUID
	IssueID     string    `json:"issue_id"`     // Associated issue UUID
	Title       string    `json:"title"`        // Summary title (e.g. Implement Retry Pattern)
	Remediation string    `json:"remediation"`  // Detailed recovery guidance text
	CreatedAt   time.Time `json:"created_at"`
}

// RecommendationRepository defines storage operations for Recommendations.
type RecommendationRepository interface {
	Get(id string) (*Recommendation, error)
	GetByIssue(issueID string) (*Recommendation, error)
	Save(recommendation *Recommendation) error
}
