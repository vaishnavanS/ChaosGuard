package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"chaosguard/internal/domain"

	"github.com/google/uuid"
)

// IssueRepository stores issues in SQLite.
type IssueRepository struct {
	db *sql.DB
}

func (r *IssueRepository) Get(id string) (*domain.Issue, error) {
	row := r.db.QueryRow(`SELECT id, experiment_id, severity, category, title, description, detected_at FROM issues WHERE id = ?`, id)
	var issue domain.Issue
	var detectedAt string
	if err := row.Scan(&issue.ID, &issue.ExperimentID, &issue.Severity, &issue.Category, &issue.Title, &issue.Description, &detectedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("issue %s not found", id)
		}
		return nil, fmt.Errorf("get issue: %w", err)
	}
	issue.DetectedAt, _ = time.Parse(time.RFC3339Nano, detectedAt)
	return &issue, nil
}

func (r *IssueRepository) List() ([]*domain.Issue, error) {
	rows, err := r.db.Query(`SELECT id, experiment_id, severity, category, title, description, detected_at FROM issues ORDER BY detected_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list issues: %w", err)
	}
	defer rows.Close()

	var issues []*domain.Issue
	for rows.Next() {
		var issue domain.Issue
		var detectedAt string
		if err := rows.Scan(&issue.ID, &issue.ExperimentID, &issue.Severity, &issue.Category, &issue.Title, &issue.Description, &detectedAt); err != nil {
			return nil, fmt.Errorf("scan issue: %w", err)
		}
		issue.DetectedAt, _ = time.Parse(time.RFC3339Nano, detectedAt)
		issues = append(issues, &issue)
	}
	return issues, nil
}

func (r *IssueRepository) ListByExperiment(experimentID string) ([]*domain.Issue, error) {
	rows, err := r.db.Query(`SELECT id, experiment_id, severity, category, title, description, detected_at FROM issues WHERE experiment_id = ? ORDER BY detected_at DESC`, experimentID)
	if err != nil {
		return nil, fmt.Errorf("list issues by experiment: %w", err)
	}
	defer rows.Close()

	var issues []*domain.Issue
	for rows.Next() {
		var issue domain.Issue
		var detectedAt string
		if err := rows.Scan(&issue.ID, &issue.ExperimentID, &issue.Severity, &issue.Category, &issue.Title, &issue.Description, &detectedAt); err != nil {
			return nil, fmt.Errorf("scan issue: %w", err)
		}
		issue.DetectedAt, _ = time.Parse(time.RFC3339Nano, detectedAt)
		issues = append(issues, &issue)
	}
	return issues, nil
}

func (r *IssueRepository) Save(issue *domain.Issue) error {
	if issue == nil {
		return fmt.Errorf("issue cannot be nil")
	}
	if issue.ID == "" {
		issue.ID = uuid.NewString()
	}
	if issue.DetectedAt.IsZero() {
		issue.DetectedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(`
		INSERT INTO issues (id, experiment_id, severity, category, title, description, detected_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
		experiment_id = excluded.experiment_id,
		severity = excluded.severity,
		category = excluded.category,
		title = excluded.title,
		description = excluded.description,
		detected_at = excluded.detected_at
	`, issue.ID, issue.ExperimentID, issue.Severity, issue.Category, issue.Title, issue.Description, issue.DetectedAt.Format(time.RFC3339Nano))
	return err
}

func (r *IssueRepository) Create(issue *domain.Issue) error {
	if issue == nil {
		return fmt.Errorf("issue cannot be nil")
	}
	if issue.ID == "" {
		issue.ID = uuid.NewString()
	}
	if issue.DetectedAt.IsZero() {
		issue.DetectedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(`
		INSERT INTO issues (id, experiment_id, severity, category, title, description, detected_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, issue.ID, issue.ExperimentID, issue.Severity, issue.Category, issue.Title, issue.Description, issue.DetectedAt.Format(time.RFC3339Nano))
	return err
}

func (r *IssueRepository) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("issue ID cannot be empty")
	}
	_, err := r.db.Exec(`DELETE FROM issues WHERE id = ?`, id)
	return err
}
