package domain

import "time"

// Issue severity levels
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// Issue represents an anomaly or vulnerability detected by ChaosGuard.
type Issue struct {
	ID           string    `json:"id"`            // Unique issue UUID
	ExperimentID string    `json:"experiment_id"` // Associated experiment UUID (optional)
	RuleName     string    `json:"rule_name"`     // Triggered rule name (e.g. crash_loop, backend_instability)
	Severity     string    `json:"severity"`      // Severity level
	Description  string    `json:"description"`   // Human-readable details
	DetectedAt   time.Time `json:"detected_at"`
}

// IssueRepository defines storage operations for Issues.
type IssueRepository interface {
	Get(id string) (*Issue, error)
	List() ([]*Issue, error)
	ListByExperiment(experimentID string) ([]*Issue, error)
	Save(issue *Issue) error
	Delete(id string) error
}
