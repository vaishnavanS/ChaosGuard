package sqlite

import (
	"database/sql"
	"fmt"
)

var migrationStatements = []string{
	`CREATE TABLE IF NOT EXISTS experiments (
		id TEXT PRIMARY KEY,
		container_name TEXT,
		attack_type TEXT NOT NULL,
		duration INTEGER,
		status TEXT NOT NULL,
		recovery_status TEXT,
		target_container_id TEXT,
		parameters TEXT,
		started_at TEXT NOT NULL,
		finished_at TEXT,
		error_message TEXT
	)`,
	`CREATE TABLE IF NOT EXISTS containers (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		image TEXT,
		state TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		is_monitored INTEGER NOT NULL DEFAULT 0
	)`,
	`CREATE TABLE IF NOT EXISTS issues (
		id TEXT PRIMARY KEY,
		experiment_id TEXT,
		severity TEXT NOT NULL,
		category TEXT,
		title TEXT,
		description TEXT NOT NULL,
		detected_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS recommendations (
		id TEXT PRIMARY KEY,
		issue_id TEXT NOT NULL,
		recommendation TEXT NOT NULL,
		priority TEXT,
		created_at TEXT NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_experiments_status ON experiments(status)`,
	`CREATE INDEX IF NOT EXISTS idx_issues_experiment_id ON issues(experiment_id)`,
	`CREATE INDEX IF NOT EXISTS idx_recommendations_issue_id ON recommendations(issue_id)`,
}

// AutoMigrate applies the SQLite schema migration set.
func AutoMigrate(db *sql.DB) error {
	for _, statement := range migrationStatements {
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("exec migration %q: %w", statement, err)
		}
	}
	return nil
}
