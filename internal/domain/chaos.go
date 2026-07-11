package domain

import "context"

// Attack defines the interface that all chaos attack executors must implement.
type Attack interface {
	Name() string
	Run(ctx context.Context, containerID string, parameters string) error
	Recover(ctx context.Context, containerID string) error
	Validate(ctx context.Context, containerID string) (bool, error)
}

// AttackManager manages registration and execution of chaos attacks.
type AttackManager interface {
	Register(attack Attack)
	Get(name string) (Attack, error)
	List() []string
	Execute(ctx context.Context, experiment *Experiment) error
}

// RecoveryManager handles service restoration for containers affected by chaos experiments.
type RecoveryManager interface {
	Recover(ctx context.Context, experiment *Experiment) error
	RecoverAllActive(ctx context.Context) error
	TrackExperiment(experiment *Experiment)
}
