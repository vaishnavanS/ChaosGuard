package scheduler

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"chaosguard/internal/domain"
	"chaosguard/pkg/config"
	"chaosguard/pkg/logger"
	"chaosguard/pkg/metrics"

	"github.com/google/uuid"
)

// Scheduler coordinates chaos experiments based on scheduling rules and safe mode.
type Scheduler struct {
	cfg              *config.Config
	containerCli     domain.ContainerController
	containerRepo    domain.ContainerRepository
	experimentRepo   domain.ExperimentRepository
	attackMgr        domain.AttackManager
	recoveryMgr      domain.RecoveryManager
	chaosCollector   *metrics.ChaosCollector
	containerMetrics *metrics.Collector

	// State management
	running  bool
	stopChan chan struct{}
	mu       sync.Mutex
	rrIndex  int
	randGen  *rand.Rand
}

// NewScheduler creates a new Scheduler instance.
func NewScheduler(
	cfg *config.Config,
	containerCli domain.ContainerController,
	containerRepo domain.ContainerRepository,
	experimentRepo domain.ExperimentRepository,
	attackMgr domain.AttackManager,
	recoveryMgr domain.RecoveryManager,
) *Scheduler {
	return &Scheduler{
		cfg:            cfg,
		containerCli:   containerCli,
		containerRepo:  containerRepo,
		experimentRepo: experimentRepo,
		attackMgr:      attackMgr,
		recoveryMgr:    recoveryMgr,
		stopChan:       make(chan struct{}),
		randGen:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewSchedulerWithMetrics creates a Scheduler with metrics collector integration.
func NewSchedulerWithMetrics(
	cfg *config.Config,
	containerCli domain.ContainerController,
	containerRepo domain.ContainerRepository,
	experimentRepo domain.ExperimentRepository,
	attackMgr domain.AttackManager,
	recoveryMgr domain.RecoveryManager,
	chaosCollector *metrics.ChaosCollector,
	containerMetrics *metrics.Collector,
) *Scheduler {
	s := NewScheduler(cfg, containerCli, containerRepo, experimentRepo, attackMgr, recoveryMgr)
	s.chaosCollector = chaosCollector
	s.containerMetrics = containerMetrics
	return s
}

// Start runs the scheduler polling loop in the background.
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.running = true
	s.stopChan = make(chan struct{})
	s.mu.Unlock()

	interval, err := time.ParseDuration(s.cfg.Scheduler.AttackInterval)
	if err != nil {
		return fmt.Errorf("invalid scheduler interval: %w", err)
	}

	if s.containerMetrics != nil {
		if err := s.containerMetrics.Start(ctx); err != nil {
			logger.Error(err, "Failed to start container metrics collector")
		}
	}

	if s.chaosCollector != nil {
		s.chaosCollector.RecordSchedulerStatusChange(true)
	}

	logger.Info("Starting Scheduler in mode '%s' with interval %s", s.cfg.Scheduler.Mode, interval)

	go s.loop(ctx, interval)
	return nil
}

// Stop halts the scheduler.
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("scheduler is not running")
	}

	s.running = false
	close(s.stopChan)

	if s.containerMetrics != nil {
		if err := s.containerMetrics.Stop(); err != nil {
			logger.Error(err, "Failed to stop container metrics collector")
		}
	}

	if s.chaosCollector != nil {
		s.chaosCollector.RecordSchedulerStatusChange(false)
	}

	logger.Info("Scheduler stopped successfully")
	return nil
}

// IsRunning returns the current status of the scheduler.
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Scheduler) loop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Perform initial sync of containers
	s.syncContainers()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Scheduler context canceled. Shutting down loop...")
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			if s.cfg.Scheduler.Mode == "manual" {
				logger.Debug("Scheduler is in manual mode. Skipping auto attack injection.")
				continue
			}

			// Scan container states and inject next failure
			s.syncContainers()
			s.injectChaos()
		}
	}
}

func (s *Scheduler) syncContainers() {
	list, err := s.containerCli.Discover()
	if err != nil {
		logger.Error(err, "Scheduler failed to discover containers")
		return
	}

	for _, c := range list {
		if err := s.containerRepo.Save(c); err != nil {
			logger.Error(err, "Failed to save container %s to repository", c.Name)
		}
	}
}

func (s *Scheduler) injectChaos() {
	// Find candidates
	containers, err := s.containerRepo.List()
	if err != nil {
		logger.Error(err, "Failed to list containers from database for chaos selection")
		return
	}

	var targets []*domain.Container
	for _, c := range containers {
		if c.IsMonitored && c.Status == "running" {
			targets = append(targets, c)
		}
	}

	if len(targets) == 0 {
		logger.Info("No running containers match monitoring filters. Chaos injection skipped.")
		return
	}

	// Determine next target container
	target := s.selectTarget(targets)
	if target == nil {
		return
	}

	// Select attack type
	attacks := s.attackMgr.List()
	if len(attacks) == 0 {
		logger.Warn("No chaos attacks registered in AttackManager. Chaos injection skipped.")
		return
	}

	s.mu.Lock()
	attackType := attacks[s.randGen.Intn(len(attacks))]
	s.mu.Unlock()

	// Parse duration
	duration, err := time.ParseDuration(s.cfg.Scheduler.AttackDuration)
	if err != nil {
		logger.Error(err, "Invalid attack duration configured")
		return
	}

	// Launch experiment
	experimentID := uuid.New().String()
	exp := &domain.Experiment{
		ID:                experimentID,
		TargetContainerID: target.ID,
		AttackType:        attackType,
		Status:            domain.ExperimentStatusPending,
		Parameters:        fmt.Sprintf(`{"duration":"%s"}`, duration.String()),
		StartedAt:         time.Now(),
	}

	if err := s.experimentRepo.Save(exp); err != nil {
		logger.Error(err, "Failed to create experiment record: %v", err)
		return
	}

	if s.recoveryMgr != nil {
		s.recoveryMgr.TrackExperiment(exp)
	}

	if s.chaosCollector != nil {
		s.chaosCollector.RecordExperimentStarted(exp.ID)
		s.chaosCollector.RecordLastExperimentTime()
	}

	logger.Info("Scheduler selected container '%s' (ID: %s) for attack '%s'", target.Name, target.ID, attackType)

	// Async run attack
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), duration+10*time.Second)
		defer cancel()

		if err := s.attackMgr.Execute(ctx, exp); err != nil {
			logger.Error(err, "Failed to execute chaos attack %s on %s", attackType, target.Name)
			return
		}

		// Schedule recovery after attack duration
		time.Sleep(duration)

		logger.Info("Scheduler initiating recovery for experiment %s", exp.ID)
		if err := s.recoveryMgr.Recover(context.Background(), exp); err != nil {
			logger.Error(err, "Failed to recover container %s after attack", target.Name)
		}
	}()
}

func (s *Scheduler) selectTarget(targets []*domain.Container) *domain.Container {
	s.mu.Lock()
	defer s.mu.Unlock()

	mode := s.cfg.Scheduler.Mode
	switch mode {
	case "random":
		return targets[s.randGen.Intn(len(targets))]

	case "round-robin":
		// Sort by ID to ensure a stable cycle
		sort.Slice(targets, func(i, j int) bool {
			return targets[i].ID < targets[j].ID
		})
		if s.rrIndex >= len(targets) {
			s.rrIndex = 0
		}
		target := targets[s.rrIndex]
		s.rrIndex = (s.rrIndex + 1) % len(targets)
		return target

	case "sequential":
		// Sort by Name to ensure sequential alphabetical execution
		sort.Slice(targets, func(i, j int) bool {
			return targets[i].Name < targets[j].Name
		})
		if s.rrIndex >= len(targets) {
			s.rrIndex = 0
		}
		target := targets[s.rrIndex]
		s.rrIndex = (s.rrIndex + 1) % len(targets)
		return target

	default:
		logger.Warn("Unknown scheduling mode: %s. Defaulting to first candidate.", mode)
		return targets[0]
	}
}
