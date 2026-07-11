package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"chaosguard/internal/domain"

	"github.com/google/uuid"
)

// ContainerRepository stores containers in SQLite.
type ContainerRepository struct {
	db *sql.DB
}

func (r *ContainerRepository) Get(id string) (*domain.Container, error) {
	row := r.db.QueryRow(`SELECT id, name, image, state, created_at, updated_at, is_monitored FROM containers WHERE id = ?`, id)
	var container domain.Container
	var createdAt, updatedAt string
	var monitored int
	if err := row.Scan(&container.ID, &container.Name, &container.Image, &container.Status, &createdAt, &updatedAt, &monitored); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("container %s not found", id)
		}
		return nil, fmt.Errorf("get container: %w", err)
	}
	container.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	container.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	container.IsMonitored = monitored == 1
	return &container, nil
}

func (r *ContainerRepository) List() ([]*domain.Container, error) {
	rows, err := r.db.Query(`SELECT id, name, image, state, created_at, updated_at, is_monitored FROM containers ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}
	defer rows.Close()

	var containers []*domain.Container
	for rows.Next() {
		var container domain.Container
		var createdAt, updatedAt string
		var monitored int
		if err := rows.Scan(&container.ID, &container.Name, &container.Image, &container.Status, &createdAt, &updatedAt, &monitored); err != nil {
			return nil, fmt.Errorf("scan container: %w", err)
		}
		container.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		container.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
		container.IsMonitored = monitored == 1
		containers = append(containers, &container)
	}
	return containers, nil
}

func (r *ContainerRepository) Save(container *domain.Container) error {
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}
	if container.ID == "" {
		container.ID = uuid.NewString()
	}
	if container.CreatedAt.IsZero() {
		container.CreatedAt = time.Now().UTC()
	}
	container.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(`
		INSERT INTO containers (id, name, image, state, created_at, updated_at, is_monitored)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		image = excluded.image,
		state = excluded.state,
		created_at = excluded.created_at,
		updated_at = excluded.updated_at,
		is_monitored = excluded.is_monitored
	`, container.ID, container.Name, container.Image, container.Status, container.CreatedAt.Format(time.RFC3339Nano), container.UpdatedAt.Format(time.RFC3339Nano), boolToInt(container.IsMonitored))
	return err
}

func (r *ContainerRepository) Create(container *domain.Container) error {
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}
	if container.ID == "" {
		container.ID = uuid.NewString()
	}
	if container.CreatedAt.IsZero() {
		container.CreatedAt = time.Now().UTC()
	}
	container.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(`
		INSERT INTO containers (id, name, image, state, created_at, updated_at, is_monitored)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, container.ID, container.Name, container.Image, container.Status, container.CreatedAt.Format(time.RFC3339Nano), container.UpdatedAt.Format(time.RFC3339Nano), boolToInt(container.IsMonitored))
	return err
}

func (r *ContainerRepository) Update(container *domain.Container) error {
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}
	if container.ID == "" {
		return fmt.Errorf("container ID cannot be empty")
	}
	container.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(`
		UPDATE containers SET name = ?, image = ?, state = ?, updated_at = ?, is_monitored = ? WHERE id = ?
	`, container.Name, container.Image, container.Status, container.UpdatedAt.Format(time.RFC3339Nano), boolToInt(container.IsMonitored), container.ID)
	return err
}

func (r *ContainerRepository) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("container ID cannot be empty")
	}
	_, err := r.db.Exec(`DELETE FROM containers WHERE id = ?`, id)
	return err
}

func (r *ContainerRepository) UpdateState(id string, state string) error {
	if id == "" {
		return fmt.Errorf("container ID cannot be empty")
	}
	_, err := r.db.Exec(`UPDATE containers SET state = ?, updated_at = ? WHERE id = ?`, state, time.Now().UTC().Format(time.RFC3339Nano), id)
	return err
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
