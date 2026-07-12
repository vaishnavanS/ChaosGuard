package responses

import "time"

// SuccessResponse represents a generic success response wrapper
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents a generic error response wrapper
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"error message"`
}

// HealthResponse represents the health endpoint response payload
type HealthResponse struct {
	Status  string `json:"status" example:"healthy"`
	State   string `json:"state" example:"running"`
	Version string `json:"version" example:"v0.1.0-dev"`
}

// ContainerResponse represents the container payload returned by the API
type ContainerResponse struct {
	ID          string            `json:"id" example:"c1"`
	Name        string            `json:"name" example:"web"`
	Image       string            `json:"image" example:"nginx:latest"`
	Status      string            `json:"status" example:"running"`
	IsMonitored bool              `json:"is_monitored" example:"true"`
	Uptime      string            `json:"uptime" example:"2h45m10s"`
	Labels      map[string]string `json:"labels"`
	CreatedAt   time.Time         `json:"created_at" example:"2026-07-12T16:00:00Z"`
	UpdatedAt   time.Time         `json:"updated_at" example:"2026-07-12T16:05:00Z"`
}

// ExperimentResponse represents the experiment payload returned by the API
type ExperimentResponse struct {
	ID                string     `json:"id" example:"uuid-string"`
	TargetContainerID string     `json:"target_container_id" example:"c1"`
	ContainerName     string     `json:"container_name,omitempty" example:"web"`
	AttackType        string     `json:"attack_type" example:"pause"`
	Duration          int64      `json:"duration,omitempty" example:"10"`
	Status            string     `json:"status" example:"completed"`
	RecoveryStatus    string     `json:"recovery_status,omitempty" example:"recovered"`
	Parameters        string     `json:"parameters" example:"{\"duration\":\"10s\"}"`
	StartedAt         time.Time  `json:"started_at" example:"2026-07-12T16:10:00Z"`
	EndedAt           *time.Time `json:"ended_at,omitempty" example:"2026-07-12T16:10:10Z"`
	ErrorMessage      string     `json:"error_message,omitempty" example:"attack failed"`
}

// SchedulerStatusResponse represents the scheduler status and config payload
type SchedulerStatusResponse struct {
	Running        bool   `json:"running" example:"true"`
	Mode           string `json:"mode" example:"random"`
	AttackInterval string `json:"attack_interval" example:"30s"`
	AttackDuration string `json:"attack_duration" example:"10s"`
}

// RuntimeResponse represents the application runtime state
type RuntimeResponse struct {
	State string `json:"state" example:"running"`
}
