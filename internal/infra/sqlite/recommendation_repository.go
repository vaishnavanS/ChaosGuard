package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"chaosguard/internal/domain"

	"github.com/google/uuid"
)

// RecommendationRepository stores recommendations in SQLite.
type RecommendationRepository struct {
	db *sql.DB
}

func (r *RecommendationRepository) Get(id string) (*domain.Recommendation, error) {
	row := r.db.QueryRow(`SELECT id, issue_id, recommendation, priority, created_at FROM recommendations WHERE id = ?`, id)
	var recommendation domain.Recommendation
	var createdAt string
	if err := row.Scan(&recommendation.ID, &recommendation.IssueID, &recommendation.Recommendation, &recommendation.Priority, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("recommendation %s not found", id)
		}
		return nil, fmt.Errorf("get recommendation: %w", err)
	}
	recommendation.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return &recommendation, nil
}

func (r *RecommendationRepository) GetByIssue(issueID string) (*domain.Recommendation, error) {
	row := r.db.QueryRow(`SELECT id, issue_id, recommendation, priority, created_at FROM recommendations WHERE issue_id = ? ORDER BY created_at DESC LIMIT 1`, issueID)
	var recommendation domain.Recommendation
	var createdAt string
	if err := row.Scan(&recommendation.ID, &recommendation.IssueID, &recommendation.Recommendation, &recommendation.Priority, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("recommendation for issue %s not found", issueID)
		}
		return nil, fmt.Errorf("get recommendation by issue: %w", err)
	}
	recommendation.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return &recommendation, nil
}

func (r *RecommendationRepository) ListByIssue(issueID string) ([]*domain.Recommendation, error) {
	rows, err := r.db.Query(`SELECT id, issue_id, recommendation, priority, created_at FROM recommendations WHERE issue_id = ? ORDER BY created_at DESC`, issueID)
	if err != nil {
		return nil, fmt.Errorf("list recommendations by issue: %w", err)
	}
	defer rows.Close()

	var recommendations []*domain.Recommendation
	for rows.Next() {
		var recommendation domain.Recommendation
		var createdAt string
		if err := rows.Scan(&recommendation.ID, &recommendation.IssueID, &recommendation.Recommendation, &recommendation.Priority, &createdAt); err != nil {
			return nil, fmt.Errorf("scan recommendation: %w", err)
		}
		recommendation.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		recommendations = append(recommendations, &recommendation)
	}
	return recommendations, nil
}

func (r *RecommendationRepository) Save(recommendation *domain.Recommendation) error {
	if recommendation == nil {
		return fmt.Errorf("recommendation cannot be nil")
	}
	if recommendation.ID == "" {
		recommendation.ID = uuid.NewString()
	}
	if recommendation.CreatedAt.IsZero() {
		recommendation.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(`
		INSERT INTO recommendations (id, issue_id, recommendation, priority, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
		issue_id = excluded.issue_id,
		recommendation = excluded.recommendation,
		priority = excluded.priority,
		created_at = excluded.created_at
	`, recommendation.ID, recommendation.IssueID, recommendation.Recommendation, recommendation.Priority, recommendation.CreatedAt.Format(time.RFC3339Nano))
	return err
}

func (r *RecommendationRepository) Create(recommendation *domain.Recommendation) error {
	if recommendation == nil {
		return fmt.Errorf("recommendation cannot be nil")
	}
	if recommendation.ID == "" {
		recommendation.ID = uuid.NewString()
	}
	if recommendation.CreatedAt.IsZero() {
		recommendation.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(`
		INSERT INTO recommendations (id, issue_id, recommendation, priority, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, recommendation.ID, recommendation.IssueID, recommendation.Recommendation, recommendation.Priority, recommendation.CreatedAt.Format(time.RFC3339Nano))
	return err
}
