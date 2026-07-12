package attack

import (
	"context"
	"fmt"
	"sync"
	"time"

	"chaosguard/internal/domain"
	"chaosguard/pkg/logger"
	"chaosguard/pkg/metrics"
)

// Manager implements domain.AttackManager to register and execute chaos attacks.
type Manager struct {
	mu             sync.RWMutex
	attacks        map[string]domain.Attack
	experimentRepo domain.ExperimentRepository
	chaosCollector *metrics.ChaosCollector
}

// NewManager creates a new Manager instance.
func NewManager(experimentRepo domain.ExperimentRepository) *Manager {
	return &Manager{
		attacks:        make(map[string]domain.Attack),
		experimentRepo: experimentRepo,
	}
}

// NewManagerWithMetrics creates a new Manager instance with metrics support.
func NewManagerWithMetrics(experimentRepo domain.ExperimentRepository, chaosCollector *metrics.ChaosCollector) *Manager {
	m := NewManager(experimentRepo)
	m.chaosCollector = chaosCollector
	return m
}

// Register adds a new attack plugin to the manager.
func (m *Manager) Register(attack domain.Attack) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.attacks[attack.Name()] = attack
	logger.Info("Registered attack plugin: %s", attack.Name())
}

// Get retrieves an attack plugin by name.
func (m *Manager) Get(name string) (domain.Attack, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	atk, ok := m.attacks[name]
	if !ok {
		return nil, fmt.Errorf("attack plugin '%s' not registered", name)
	}
	return atk, nil
}

// List returns a sorted list of registered attack names.
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []string
	for name := range m.attacks {
		list = append(list, name)
	}
	return list
}

// Execute runs the selected chaos attack, updates the DB, and validates the failure injection.
func (m *Manager) Execute(ctx context.Context, exp *domain.Experiment) error {
	atk, err := m.Get(exp.AttackType)
	if err != nil {
		sErr := err.Error()
		m.experimentRepo.UpdateStatus(exp.ID, domain.ExperimentStatusFailed, sErr, nil)
		return err
	}

	logger.Info("Starting execution of attack '%s' for experiment %s", exp.AttackType, exp.ID)
	exp.Status = domain.ExperimentStatusRunning
	if err := m.experimentRepo.Save(exp); err != nil {
		logger.Error(err, "Failed to update experiment %s status to running", exp.ID)
	}

	// Execute attack injection
	if err := atk.Run(ctx, exp.TargetContainerID, exp.Parameters); err != nil {
		logger.Error(err, "Attack injection failed for container %s", exp.TargetContainerID)
		endedAt := time.Now()
		m.experimentRepo.UpdateStatus(exp.ID, domain.ExperimentStatusFailed, err.Error(), &endedAt)
		if m.chaosCollector != nil {
			m.chaosCollector.RecordAttackFailed()
			m.chaosCollector.RecordExperimentFailed(exp.ID)
		}
		return err
	}

	// Validate attack injection
	ok, valErr := atk.Validate(ctx, exp.TargetContainerID)
	if valErr != nil || !ok {
		var msg string
		if valErr != nil {
			msg = fmt.Sprintf("attack validation error: %v", valErr)
			logger.Error(valErr, "%s", msg)
		} else {
			msg = "attack validation failed: container not in expected state"
			logger.Error(nil, "%s", msg)
		}
		endedAt := time.Now()
		m.experimentRepo.UpdateStatus(exp.ID, domain.ExperimentStatusFailed, msg, &endedAt)
		if m.chaosCollector != nil {
			m.chaosCollector.RecordAttackFailed()
			m.chaosCollector.RecordExperimentFailed(exp.ID)
		}
		return fmt.Errorf("%s", msg)
	}

	if m.chaosCollector != nil {
		m.chaosCollector.RecordAttackExecuted()
	}

	logger.Info("Attack '%s' successfully injected and validated for container %s", exp.AttackType, exp.TargetContainerID)
	return nil
}
