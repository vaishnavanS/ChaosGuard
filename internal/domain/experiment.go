package domain

import (
	"errors"
	"time"
)

// Experiment status constants
const (
	ExperimentStatusPending   = "pending"
	ExperimentStatusRunning   = "running"
	ExperimentStatusCompleted = "completed"
	ExperimentStatusFailed    = "failed"
	ExperimentStatusRecovered = "recovered"
)

// ErrExperimentNotFound is returned when an experiment does not exist.
var ErrExperimentNotFound = errors.New("experiment not found")

// Experiment represents a single run of a chaos experiment.
type Experiment struct {
	ID                string     `json:"id"`                  // Unique experiment UUID
	TargetContainerID string     `json:"target_container_id"` // Target Docker Container ID
	ContainerName     string     `json:"container_name,omitempty"`
	AttackType        string     `json:"attack_type"` // Type of attack (e.g. pause, stop, latency)
	Duration          int64      `json:"duration,omitempty"`
	Status            string     `json:"status"` // Current status of experiment
	RecoveryStatus    string     `json:"recovery_status,omitempty"`
	Parameters        string     `json:"parameters"` // JSON parameter configuration (e.g. duration, latency limits)
	StartedAt         time.Time  `json:"started_at"`
	EndedAt           *time.Time `json:"ended_at,omitempty"`
	ErrorMessage      string     `json:"error_message,omitempty"`
}

// ExperimentRepository defines storage operations for Chaos Experiments.
type ExperimentRepository interface {
	Get(id string) (*Experiment, error)
	List() ([]*Experiment, error)
	Save(experiment *Experiment) error
	Create(experiment *Experiment) error
	Update(experiment *Experiment) error
	Delete(id string) error
	UpdateStatus(id string, status string, errStr string, endedAt *time.Time) error
}
