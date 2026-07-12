package requests

// CreateExperimentRequest represents the request body to create and start a chaos experiment
type CreateExperimentRequest struct {
	TargetContainerID string `json:"target_container_id" binding:"required" example:"c1"`
	AttackType        string `json:"attack_type" binding:"required" example:"pause"`
	DurationSec       int    `json:"duration" binding:"required,min=1" example:"10"`
}
