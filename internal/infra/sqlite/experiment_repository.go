package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"chaosguard/internal/domain"

	"github.com/google/uuid"
)

// ExperimentRepository stores experiments in SQLite.
type ExperimentRepository struct {
	db *sql.DB
}

func (r *ExperimentRepository) Get(id string) (*domain.Experiment, error) {
	row := r.db.QueryRow(`SELECT id, container_name, attack_type, duration, status, recovery_status, target_container_id, parameters, started_at, finished_at, error_message FROM experiments WHERE id = ?`, id)

	var exp domain.Experiment
	var finishedAt sql.NullString
	if err := row.Scan(&exp.ID, &exp.ContainerName, &exp.AttackType, &exp.Duration, &exp.Status, &exp.RecoveryStatus, &exp.TargetContainerID, &exp.Parameters, &exp.StartedAt, &finishedAt, &exp.ErrorMessage); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrExperimentNotFound
		}
		return nil, fmt.Errorf("get experiment: %w", err)
	}

	if finishedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, finishedAt.String)
		if err != nil {
			return nil, fmt.Errorf("parse finished_at: %w", err)
		}
		exp.EndedAt = &parsed
	}

	return &exp, nil
}

func (r *ExperimentRepository) List() ([]*domain.Experiment, error) {
	rows, err := r.db.Query(`SELECT id, container_name, attack_type, duration, status, recovery_status, target_container_id, parameters, started_at, finished_at, error_message FROM experiments ORDER BY started_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list experiments: %w", err)
	}
	defer rows.Close()

	var experiments []*domain.Experiment
	for rows.Next() {
		var exp domain.Experiment
		var finishedAt sql.NullString
		if err := rows.Scan(&exp.ID, &exp.ContainerName, &exp.AttackType, &exp.Duration, &exp.Status, &exp.RecoveryStatus, &exp.TargetContainerID, &exp.Parameters, &exp.StartedAt, &finishedAt, &exp.ErrorMessage); err != nil {
			return nil, fmt.Errorf("scan experiment: %w", err)
		}
		if finishedAt.Valid {
			parsed, err := time.Parse(time.RFC3339Nano, finishedAt.String)
			if err != nil {
				return nil, fmt.Errorf("parse finished_at: %w", err)
			}
			exp.EndedAt = &parsed
		}
		experiments = append(experiments, &exp)
	}

	return experiments, nil
}

func (r *ExperimentRepository) Save(experiment *domain.Experiment) error {
	if experiment == nil {
		return fmt.Errorf("experiment cannot be nil")
	}
	if experiment.ID == "" {
		experiment.ID = uuid.NewString()
	}
	if experiment.StartedAt.IsZero() {
		experiment.StartedAt = time.Now().UTC()
	}

	_, err := r.db.Exec(`
		INSERT INTO experiments (id, container_name, attack_type, duration, status, recovery_status, target_container_id, parameters, started_at, finished_at, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
		container_name = excluded.container_name,
		attack_type = excluded.attack_type,
		duration = excluded.duration,
		status = excluded.status,
		recovery_status = excluded.recovery_status,
		target_container_id = excluded.target_container_id,
		parameters = excluded.parameters,
		started_at = excluded.started_at,
		finished_at = excluded.finished_at,
		error_message = excluded.error_message
	`, experiment.ID, experiment.ContainerName, experiment.AttackType, experiment.Duration, experiment.Status, experiment.RecoveryStatus, experiment.TargetContainerID, experiment.Parameters, experiment.StartedAt.Format(time.RFC3339Nano), serializeNullableTime(experiment.EndedAt), experiment.ErrorMessage)
	return err
}

func (r *ExperimentRepository) Create(experiment *domain.Experiment) error {
	if experiment == nil {
		return fmt.Errorf("experiment cannot be nil")
	}
	if experiment.ID == "" {
		experiment.ID = uuid.NewString()
	}
	if experiment.StartedAt.IsZero() {
		experiment.StartedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(`
		INSERT INTO experiments (id, container_name, attack_type, duration, status, recovery_status, target_container_id, parameters, started_at, finished_at, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, experiment.ID, experiment.ContainerName, experiment.AttackType, experiment.Duration, experiment.Status, experiment.RecoveryStatus, experiment.TargetContainerID, experiment.Parameters, experiment.StartedAt.Format(time.RFC3339Nano), serializeNullableTime(experiment.EndedAt), experiment.ErrorMessage)
	return err
}

func (r *ExperimentRepository) Update(experiment *domain.Experiment) error {
	if experiment == nil {
		return fmt.Errorf("experiment cannot be nil")
	}
	if experiment.ID == "" {
		return fmt.Errorf("experiment ID cannot be empty")
	}
	_, err := r.db.Exec(`
		UPDATE experiments SET
		container_name = ?,
		attack_type = ?,
		duration = ?,
		status = ?,
		recovery_status = ?,
		target_container_id = ?,
		parameters = ?,
		started_at = ?,
		finished_at = ?,
		error_message = ?
		WHERE id = ?
	`, experiment.ContainerName, experiment.AttackType, experiment.Duration, experiment.Status, experiment.RecoveryStatus, experiment.TargetContainerID, experiment.Parameters, experiment.StartedAt.Format(time.RFC3339Nano), serializeNullableTime(experiment.EndedAt), experiment.ErrorMessage, experiment.ID)
	return err
}

func (r *ExperimentRepository) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("experiment ID cannot be empty")
	}
	_, err := r.db.Exec(`DELETE FROM experiments WHERE id = ?`, id)
	return err
}

func (r *ExperimentRepository) UpdateStatus(id string, status string, errStr string, endedAt *time.Time) error {
	if id == "" {
		return fmt.Errorf("experiment ID cannot be empty")
	}
	_, err := r.db.Exec(`UPDATE experiments SET status = ?, error_message = ?, finished_at = ? WHERE id = ?`, status, errStr, serializeNullableTime(endedAt), id)
	return err
}

func serializeNullableTime(value *time.Time) interface{} {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}
