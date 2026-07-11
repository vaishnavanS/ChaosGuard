package recovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"chaosguard/internal/domain"
	"chaosguard/pkg/logger"
	"chaosguard/pkg/metrics"
)

// Manager implements domain.RecoveryManager to handle container restoration.
type Manager struct {
	mu             sync.RWMutex
	attackMgr      domain.AttackManager
	experimentRepo domain.ExperimentRepository
	chaosCollector *metrics.ChaosCollector
	active         map[string]*domain.Experiment // Track active experiments for recovery
}

// NewManager creates a new Recovery Manager instance.
func NewManager(
	attackMgr domain.AttackManager,
	experimentRepo domain.ExperimentRepository,
) *Manager {
	return &Manager{
		attackMgr:      attackMgr,
		experimentRepo: experimentRepo,
		active:         make(map[string]*domain.Experiment),
	}
}

// NewManagerWithMetrics creates a new Recovery Manager instance with metrics support.
func NewManagerWithMetrics(
	attackMgr domain.AttackManager,
	experimentRepo domain.ExperimentRepository,
	chaosCollector *metrics.ChaosCollector,
) *Manager {
	m := NewManager(attackMgr, experimentRepo)
	m.chaosCollector = chaosCollector
	return m
}

// Recover restores a container affected by a chaos experiment.
func (m *Manager) Recover(ctx context.Context, experiment *domain.Experiment) error {
	if experiment == nil || experiment.ID == "" {
		return fmt.Errorf("invalid experiment: experiment must not be nil and must have an ID")
	}

	logger.Info("Starting recovery for experiment %s on container %s", experiment.ID, experiment.TargetContainerID)

	// Get the attack that was used
	atk, err := m.attackMgr.Get(experiment.AttackType)
	if err != nil {
		msg := fmt.Sprintf("failed to retrieve attack '%s' for recovery: %v", experiment.AttackType, err)
		logger.Error(nil, "%s", msg)
		m.experimentRepo.UpdateStatus(experiment.ID, domain.ExperimentStatusFailed, msg, nil)
		return fmt.Errorf("%s", msg)
	}

	// Execute recovery
	if err := atk.Recover(ctx, experiment.TargetContainerID); err != nil {
		msg := fmt.Sprintf("recovery failed for experiment %s: %v", experiment.ID, err)
		logger.Error(nil, "%s", msg)
		m.experimentRepo.UpdateStatus(experiment.ID, domain.ExperimentStatusFailed, msg, nil)
		if m.chaosCollector != nil {
			m.chaosCollector.RecordRecoveryFailed()
		}
		return fmt.Errorf("%s", msg)
	}

	// Update experiment status to recovered
	endedAt := time.Now()
	if err := m.experimentRepo.UpdateStatus(experiment.ID, domain.ExperimentStatusRecovered, "", &endedAt); err != nil {
		logger.Error(err, "Failed to update experiment %s status to recovered", experiment.ID)
		return err
	}

	if m.chaosCollector != nil {
		m.chaosCollector.RecordRecoveryExecuted()
		m.chaosCollector.RecordExperimentRecovered(experiment.ID)
	}

	logger.Info("Recovery completed successfully for experiment %s", experiment.ID)

	// Remove from active tracking
	m.mu.Lock()
	delete(m.active, experiment.ID)
	m.mu.Unlock()

	return nil
}

// RecoverAllActive recovers all active experiments in the system.
func (m *Manager) RecoverAllActive(ctx context.Context) error {
	m.mu.RLock()
	// Make a copy of the active experiments
	activeExp := make([]*domain.Experiment, 0, len(m.active))
	for _, exp := range m.active {
		activeExp = append(activeExp, exp)
	}
	m.mu.RUnlock()

	if len(activeExp) == 0 {
		logger.Info("No active experiments to recover")
		return nil
	}

	logger.Info("Starting recovery for %d active experiments", len(activeExp))

	var lastErr error
	for _, exp := range activeExp {
		if err := m.Recover(ctx, exp); err != nil {
			logger.Error(err, "Failed to recover experiment %s", exp.ID)
			lastErr = err
		}
	}

	if lastErr != nil {
		return lastErr
	}

	logger.Info("All active experiments recovered successfully")
	return nil
}

// TrackExperiment marks an experiment as active for recovery tracking.
func (m *Manager) TrackExperiment(experiment *domain.Experiment) {
	if experiment == nil || experiment.ID == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.active[experiment.ID] = experiment
}
