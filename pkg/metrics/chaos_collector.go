package metrics

import (
	"sync"
	"time"

	"chaosguard/pkg/logger"
)

// ChaosCollector tracks chaos engineering specific metrics
type ChaosCollector struct {
	mu       sync.RWMutex
	registry *Registry

	// Internal state for tracking
	runningExperiments map[string]time.Time // experimentID -> startTime
}

// NewChaosCollector creates a new chaos metrics collector
func NewChaosCollector(registry *Registry) *ChaosCollector {
	return &ChaosCollector{
		registry:           registry,
		runningExperiments: make(map[string]time.Time),
	}
}

// RecordExperimentStarted records the start of an experiment
func (cc *ChaosCollector) RecordExperimentStarted(experimentID string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.runningExperiments[experimentID] = time.Now()
	if cc.registry.ExperimentsTotal != nil {
		cc.registry.ExperimentsTotal.Inc()
	}
	if cc.registry.ExperimentsRunning != nil {
		cc.registry.ExperimentsRunning.Inc()
	}

	logger.Debug("Recorded experiment started: %s", experimentID)
}

// RecordExperimentCompleted records the completion of an experiment
func (cc *ChaosCollector) RecordExperimentCompleted(experimentID string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	startTime, exists := cc.runningExperiments[experimentID]
	if !exists {
		logger.Warn("Completed experiment not found in running experiments: %s", experimentID)
		return
	}

	duration := time.Since(startTime).Milliseconds()
	delete(cc.runningExperiments, experimentID)

	if cc.registry.ExperimentsRunning != nil {
		cc.registry.ExperimentsRunning.Dec()
	}
	if cc.registry.ExperimentsCompleted != nil {
		cc.registry.ExperimentsCompleted.Inc()
	}
	if cc.registry.ExperimentDurationMs != nil {
		cc.registry.ExperimentDurationMs.Observe(float64(duration))
	}

	logger.Debug("Recorded experiment completed: %s (duration: %dms)", experimentID, duration)
}

// RecordExperimentFailed records the failure of an experiment
func (cc *ChaosCollector) RecordExperimentFailed(experimentID string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	startTime, exists := cc.runningExperiments[experimentID]
	if exists {
		duration := time.Since(startTime).Milliseconds()
		delete(cc.runningExperiments, experimentID)

		if cc.registry.ExperimentsRunning != nil {
			cc.registry.ExperimentsRunning.Dec()
		}
		if cc.registry.ExperimentDurationMs != nil {
			cc.registry.ExperimentDurationMs.Observe(float64(duration))
		}
	} else {
		if cc.registry.ExperimentsRunning != nil {
			cc.registry.ExperimentsRunning.Dec() // Still decrement if running but not tracked
		}
	}

	if cc.registry.ExperimentsFailed != nil {
		cc.registry.ExperimentsFailed.Inc()
	}

	logger.Debug("Recorded experiment failed: %s", experimentID)
}

// RecordExperimentRecovered records the recovery of an experiment
func (cc *ChaosCollector) RecordExperimentRecovered(experimentID string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// Remove from tracking if still there
	if _, exists := cc.runningExperiments[experimentID]; exists {
		delete(cc.runningExperiments, experimentID)
	}

	if cc.registry.ExperimentsRecovered != nil {
		cc.registry.ExperimentsRecovered.Inc()
	}

	logger.Debug("Recorded experiment recovered: %s", experimentID)
}

// RecordAttackExecuted records the execution of an attack
func (cc *ChaosCollector) RecordAttackExecuted() {
	if cc.registry.AttacksExecuted != nil {
		cc.registry.AttacksExecuted.Inc()
	}
	logger.Debug("Recorded attack executed")
}

// RecordAttackFailed records the failure of an attack
func (cc *ChaosCollector) RecordAttackFailed() {
	if cc.registry.AttacksFailed != nil {
		cc.registry.AttacksFailed.Inc()
	}
	logger.Debug("Recorded attack failed")
}

// RecordRecoveryExecuted records the execution of a recovery
func (cc *ChaosCollector) RecordRecoveryExecuted() {
	if cc.registry.RecoveriesExecuted != nil {
		cc.registry.RecoveriesExecuted.Inc()
	}
	logger.Debug("Recorded recovery executed")
}

// RecordRecoveryFailed records the failure of a recovery
func (cc *ChaosCollector) RecordRecoveryFailed() {
	if cc.registry.RecoveriesFailed != nil {
		cc.registry.RecoveriesFailed.Inc()
	}
	logger.Debug("Recorded recovery failed")
}

// RecordSchedulerStatusChange records the scheduler status change
func (cc *ChaosCollector) RecordSchedulerStatusChange(running bool) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.registry.SchedulerRunning != nil {
		if running {
			cc.registry.SchedulerRunning.Set(1)
		} else {
			cc.registry.SchedulerRunning.Set(0)
		}
	}

	logger.Debug("Recorded scheduler status change: running=%v", running)
}

// RecordLastExperimentTime records the timestamp of the last experiment
func (cc *ChaosCollector) RecordLastExperimentTime() {
	if cc.registry.LastExperimentAt != nil {
		cc.registry.LastExperimentAt.Set(float64(time.Now().Unix()))
	}
	logger.Debug("Recorded last experiment timestamp")
}

// GetRunningExperimentsCount returns the number of currently running experiments
func (cc *ChaosCollector) GetRunningExperimentsCount() int {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return len(cc.runningExperiments)
}
