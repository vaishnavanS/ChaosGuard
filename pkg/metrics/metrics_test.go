package metrics

import (
	"context"
	"sync"
	"testing"
	"time"

	"chaosguard/internal/domain"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// Mock ContainerRepository for testing
type mockContainerRepo struct {
	mu         sync.Mutex
	containers []*domain.Container
}

func newMockContainerRepo() *mockContainerRepo {
	return &mockContainerRepo{
		containers: make([]*domain.Container, 0),
	}
}

func (m *mockContainerRepo) Get(id string) (*domain.Container, error) {
	return nil, nil
}

func (m *mockContainerRepo) List() ([]*domain.Container, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]*domain.Container, len(m.containers))
	copy(result, m.containers)
	return result, nil
}

func (m *mockContainerRepo) Save(c *domain.Container) error {
	return nil
}

func (m *mockContainerRepo) Create(c *domain.Container) error {
	return nil
}

func (m *mockContainerRepo) Update(c *domain.Container) error {
	return nil
}

func (m *mockContainerRepo) Delete(id string) error {
	return nil
}

func (m *mockContainerRepo) UpdateState(id string, state string) error {
	return nil
}

func (m *mockContainerRepo) addContainer(c *domain.Container) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.containers = append(m.containers, c)
}

// Mock ContainerController for testing
type mockContainerController struct{}

func (m *mockContainerController) Discover() ([]*domain.Container, error) {
	return nil, nil
}

func (m *mockContainerController) Pause(id string) error   { return nil }
func (m *mockContainerController) Unpause(id string) error { return nil }
func (m *mockContainerController) Stop(id string) error    { return nil }
func (m *mockContainerController) Start(id string) error   { return nil }
func (m *mockContainerController) Restart(id string) error { return nil }
func (m *mockContainerController) Kill(id string) error    { return nil }
func (m *mockContainerController) Inspect(id string) (*domain.Container, error) {
	return nil, nil
}

func TestCollector_StartStop(t *testing.T) {
	registry := NewTestRegistry(prometheus.NewRegistry())
	collector := NewCollector(registry, &mockContainerController{}, newMockContainerRepo(), 10*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := collector.Start(ctx); err != nil {
		t.Fatalf("failed to start collector: %v", err)
	}
	if !collector.IsRunning() {
		t.Fatal("expected collector to be running")
	}

	if err := collector.Stop(); err != nil {
		t.Fatalf("failed to stop collector: %v", err)
	}
	if collector.IsRunning() {
		t.Fatal("expected collector to be stopped")
	}
}

func TestCollector_Collect(t *testing.T) {
	registry := NewTestRegistry(prometheus.NewRegistry())
	repo := newMockContainerRepo()
	repo.addContainer(&domain.Container{ID: "c1", Status: "running"})
	repo.addContainer(&domain.Container{ID: "c2", Status: "paused"})
	repo.addContainer(&domain.Container{ID: "c3", Status: "exited"})

	collector := NewCollector(registry, &mockContainerController{}, repo, time.Second)
	collector.collect(context.Background())

	if got := testutil.ToFloat64(registry.ContainersRunning); got != 1 {
		t.Errorf("expected 1 running container, got %v", got)
	}
	if got := testutil.ToFloat64(registry.ContainersPaused); got != 1 {
		t.Errorf("expected 1 paused container, got %v", got)
	}
	if got := testutil.ToFloat64(registry.ContainersStopped); got != 1 {
		t.Errorf("expected 1 stopped container, got %v", got)
	}
}

func TestChaosCollector_ExperimentLifecycle(t *testing.T) {
	registry := &Registry{}

	collector := NewChaosCollector(registry)

	expID := "exp-1"

	collector.RecordExperimentStarted(expID)
	if count := collector.GetRunningExperimentsCount(); count != 1 {
		t.Errorf("expected 1 running experiment, got %d", count)
	}

	time.Sleep(10 * time.Millisecond)
	collector.RecordExperimentCompleted(expID)
	if count := collector.GetRunningExperimentsCount(); count != 0 {
		t.Errorf("expected 0 running experiments after completion, got %d", count)
	}
}

func TestChaosCollector_ExperimentFailed(t *testing.T) {
	registry := &Registry{}

	collector := NewChaosCollector(registry)
	expID := "exp-2"

	collector.RecordExperimentStarted(expID)
	if count := collector.GetRunningExperimentsCount(); count != 1 {
		t.Errorf("expected 1 running experiment, got %d", count)
	}

	collector.RecordExperimentFailed(expID)
	if count := collector.GetRunningExperimentsCount(); count != 0 {
		t.Errorf("expected 0 running experiments after failure, got %d", count)
	}
}

func TestChaosCollector_MultipleExperiments(t *testing.T) {
	registry := &Registry{}

	collector := NewChaosCollector(registry)

	collector.RecordExperimentStarted("exp-1")
	collector.RecordExperimentStarted("exp-2")
	collector.RecordExperimentStarted("exp-3")

	if count := collector.GetRunningExperimentsCount(); count != 3 {
		t.Errorf("expected 3 running experiments, got %d", count)
	}

	collector.RecordExperimentCompleted("exp-1")
	if count := collector.GetRunningExperimentsCount(); count != 2 {
		t.Errorf("expected 2 running experiments after completing one, got %d", count)
	}

	collector.RecordExperimentFailed("exp-2")
	if count := collector.GetRunningExperimentsCount(); count != 1 {
		t.Errorf("expected 1 running experiment after failing one, got %d", count)
	}

	collector.RecordExperimentRecovered("exp-3")
	if count := collector.GetRunningExperimentsCount(); count != 0 {
		t.Errorf("expected 0 running experiments after recovery, got %d", count)
	}
}

func TestChaosCollector_RecordMetrics(t *testing.T) {
	registry := &Registry{}

	collector := NewChaosCollector(registry)

	collector.RecordAttackExecuted()
	collector.RecordAttackFailed()
	collector.RecordRecoveryExecuted()
	collector.RecordRecoveryFailed()
	collector.RecordSchedulerStatusChange(true)
	collector.RecordLastExperimentTime()

	// Verify no panics occurred
}

// Test helpers

// No mock Prometheus metrics needed - just use empty registry for tests
