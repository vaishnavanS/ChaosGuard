package metrics

import (
	"context"
	"sync"
	"time"

	"chaosguard/internal/domain"
	"chaosguard/pkg/logger"
	"github.com/docker/docker/api/types"
)

// Collector gathers metrics from Docker containers and updates the Prometheus registry
type Collector struct {
	mu              sync.RWMutex
	registry        *Registry
	containerCli    domain.ContainerController
	containerRepo   domain.ContainerRepository
	ticker          *time.Ticker
	stopChan        chan struct{}
	running         bool
	collectInterval time.Duration
}

// NewCollector creates a new metrics collector
func NewCollector(
	registry *Registry,
	containerCli domain.ContainerController,
	containerRepo domain.ContainerRepository,
	collectInterval time.Duration,
) *Collector {
	return &Collector{
		registry:        registry,
		containerCli:    containerCli,
		containerRepo:   containerRepo,
		stopChan:        make(chan struct{}),
		collectInterval: collectInterval,
	}
}

// Start begins collecting metrics at regular intervals
func (c *Collector) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = true
	c.stopChan = make(chan struct{})
	c.mu.Unlock()

	c.ticker = time.NewTicker(c.collectInterval)

	logger.Info("Metrics collector started with interval %s", c.collectInterval)

	go c.loop(ctx)
	return nil
}

// Stop halts the metrics collection loop
func (c *Collector) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return nil
	}

	c.running = false
	c.ticker.Stop()
	close(c.stopChan)

	logger.Info("Metrics collector stopped")
	return nil
}

// IsRunning returns whether the collector is currently running
func (c *Collector) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

func (c *Collector) loop(ctx context.Context) {
	// Collect immediately on start
	c.collect(ctx)

	for {
		select {
		case <-ctx.Done():
			c.mu.Lock()
			c.running = false
			c.mu.Unlock()
			return
		case <-c.stopChan:
			return
		case <-c.ticker.C:
			c.collect(ctx)
		}
	}
}

func (c *Collector) collect(ctx context.Context) {
	// Get containers from repository
	containers, err := c.containerRepo.List()
	if err != nil {
		logger.Error(err, "Failed to list containers for metrics collection")
		return
	}

	if len(containers) == 0 {
		c.registry.ContainersRunning.Set(0)
		c.registry.ContainersPaused.Set(0)
		c.registry.ContainersStopped.Set(0)
		return
	}

	// Count by state
	var running, paused, stopped float64
	for _, container := range containers {
		switch container.Status {
		case "running":
			running++
		case "paused":
			paused++
		case "stopped", "exited":
			stopped++
		}
	}

	c.registry.ContainersRunning.Set(running)
	c.registry.ContainersPaused.Set(paused)
	c.registry.ContainersStopped.Set(stopped)
}

// CollectContainerStats collects detailed stats for a specific container
// This would be used for more detailed monitoring if needed
func (c *Collector) CollectContainerStats(ctx context.Context, containerID string) (*ContainerStats, error) {
	// This is a placeholder for future detailed metrics collection
	// For now, we track aggregate states and experiment metrics

	return &ContainerStats{
		ContainerID: containerID,
		Timestamp:   time.Now(),
	}, nil
}

// ContainerStats holds detailed stats for a container
type ContainerStats struct {
	ContainerID string
	CPUPercent  float64
	MemoryMB    float64
	NetworkIn   int64
	NetworkOut  int64
	Timestamp   time.Time
	State       types.ContainerState
}
