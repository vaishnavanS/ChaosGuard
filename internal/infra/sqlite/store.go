package sqlite

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"chaosguard/internal/domain"

	_ "modernc.org/sqlite"
)

// Store provides SQLite-backed repositories for ChaosGuard domain entities.
type Store struct {
	db                 *sql.DB
	ExperimentRepo     domain.ExperimentRepository
	ContainerRepo      domain.ContainerRepository
	IssueRepo          domain.IssueRepository
	RecommendationRepo domain.RecommendationRepository
}

// NewStore opens a SQLite database, applies migrations, and wires repositories.
func NewStore(dbPath string) (*Store, error) {
	if dbPath == "" {
		dbPath = "./chaosguard.db"
	}

	if filepath.Ext(dbPath) == "" {
		dbPath = dbPath + ".db"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	if err := AutoMigrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate sqlite database: %w", err)
	}

	store := &Store{db: db}
	store.ExperimentRepo = &ExperimentRepository{db: db}
	store.ContainerRepo = &ContainerRepository{db: db}
	store.IssueRepo = &IssueRepository{db: db}
	store.RecommendationRepo = &RecommendationRepository{db: db}
	return store, nil
}

// Close closes the underlying SQLite connection.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}
