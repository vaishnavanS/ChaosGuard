package scheduler

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"chaosguard/internal/domain"
	"chaosguard/pkg/config"
)

// Mock ContainerRepository
type mockContainerRepo struct {
	mu   sync.Mutex
	data map[string]*domain.Container
}

func newMockContainerRepo() *mockContainerRepo {
	return &mockContainerRepo{data: make(map[string]*domain.Container)}
}

func (r *mockContainerRepo) Get(id string) (*domain.Container, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.data[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return c, nil
}

func (r *mockContainerRepo) List() ([]*domain.Container, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var list []*domain.Container
	for _, c := range r.data {
		list = append(list, c)
	}
	return list, nil
}

func (r *mockContainerRepo) Save(c *domain.Container) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[c.ID] = c
	return nil
}

func (r *mockContainerRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}

// Mock ExperimentRepository
type mockExperimentRepo struct {
	mu   sync.Mutex
	data map[string]*domain.Experiment
}

func newMockExperimentRepo() *mockExperimentRepo {
	return &mockExperimentRepo{data: make(map[string]*domain.Experiment)}
}

func (r *mockExperimentRepo) Get(id string) (*domain.Experiment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.data[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return e, nil
}

func (r *mockExperimentRepo) List() ([]*domain.Experiment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var list []*domain.Experiment
	for _, e := range r.data {
		list = append(list, e)
	}
	return list, nil
}

func (r *mockExperimentRepo) Save(e *domain.Experiment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[e.ID] = e
	return nil
}

func (r *mockExperimentRepo) UpdateStatus(id string, status string, errStr string, endedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.data[id]
	if !ok {
		return errors.New("not found")
	}
	e.Status = status
	e.ErrorMessage = errStr
	e.EndedAt = endedAt
	return nil
}

// Mock ContainerController
type mockContainerController struct {
	containers []*domain.Container
}

func (c *mockContainerController) Discover() ([]*domain.Container, error) {
	return c.containers, nil
}
func (c *mockContainerController) Pause(id string) error   { return nil }
func (c *mockContainerController) Unpause(id string) error { return nil }
func (c *mockContainerController) Stop(id string) error    { return nil }
func (c *mockContainerController) Start(id string) error   { return nil }
func (c *mockContainerController) Restart(id string) error { return nil }
func (c *mockContainerController) Kill(id string) error    { return nil }
func (c *mockContainerController) Inspect(id string) (*domain.Container, error) {
	return nil, nil
}

// Mock AttackManager
type mockAttackManager struct {
	mu       sync.Mutex
	attacks  []string
	executed []*domain.Experiment
}

func (m *mockAttackManager) Register(attack domain.Attack) {}
func (m *mockAttackManager) Get(name string) (domain.Attack, error) {
	return nil, nil
}
func (m *mockAttackManager) List() []string {
	return m.attacks
}
func (m *mockAttackManager) Execute(ctx context.Context, experiment *domain.Experiment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executed = append(m.executed, experiment)
	experiment.Status = domain.ExperimentStatusRunning
	return nil
}

// Mock RecoveryManager
type mockRecoveryManager struct {
	mu        sync.Mutex
	recovered []*domain.Experiment
}

func (m *mockRecoveryManager) Recover(ctx context.Context, experiment *domain.Experiment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recovered = append(m.recovered, experiment)
	experiment.Status = domain.ExperimentStatusRecovered
	return nil
}
func (m *mockRecoveryManager) RecoverAllActive(ctx context.Context) error {
	return nil
}

func TestScheduler_StartStop(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Scheduler.AttackInterval = "100ms"
	cfg.Scheduler.AttackDuration = "50ms"

	cRepo := newMockContainerRepo()
	eRepo := newMockExperimentRepo()
	cController := &mockContainerController{
		containers: []*domain.Container{
			{ID: "c1", Name: "web", Status: "running", IsMonitored: true},
		},
	}
	aMgr := &mockAttackManager{attacks: []string{"pause"}}
	rMgr := &mockRecoveryManager{}

	s := NewScheduler(cfg, cController, cRepo, eRepo, aMgr, rMgr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := s.Start(ctx); err != nil {
		t.Fatalf("failed to start scheduler: %v", err)
	}

	if !s.IsRunning() {
		t.Error("expected scheduler to be running")
	}

	// Try starting again (should fail)
	if err := s.Start(ctx); err == nil {
		t.Error("expected second Start to return error")
	}

	// Wait for an attack to occur
	time.Sleep(150 * time.Millisecond)

	if err := s.Stop(); err != nil {
		t.Fatalf("failed to stop scheduler: %v", err)
	}

	if s.IsRunning() {
		t.Error("expected scheduler to not be running")
	}
}

func TestScheduler_SelectionModes(t *testing.T) {
	modes := []string{"round-robin", "sequential", "random"}

	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Scheduler.Mode = mode
			cfg.Scheduler.AttackInterval = "10s" // long to prevent automatic triggers during select test
			cfg.Scheduler.AttackDuration = "5s"

			cRepo := newMockContainerRepo()
			eRepo := newMockExperimentRepo()
			cController := &mockContainerController{}
			aMgr := &mockAttackManager{attacks: []string{"pause"}}
			rMgr := &mockRecoveryManager{}

			s := NewScheduler(cfg, cController, cRepo, eRepo, aMgr, rMgr)

			targets := []*domain.Container{
				{ID: "c1", Name: "web-b", Status: "running", IsMonitored: true},
				{ID: "c2", Name: "web-c", Status: "running", IsMonitored: true},
				{ID: "c3", Name: "web-a", Status: "running", IsMonitored: true},
			}

			// Perform selection checks
			t1 := s.selectTarget(targets)
			t2 := s.selectTarget(targets)
			t3 := s.selectTarget(targets)

			if t1 == nil || t2 == nil || t3 == nil {
				t.Fatalf("selection returned nil target")
			}

			if mode == "round-robin" {
				// Round-robin sorts by ID: c1, c2, c3
				if t1.ID != "c1" || t2.ID != "c2" || t3.ID != "c3" {
					t.Errorf("round-robin sequence failed: got %s, %s, %s", t1.ID, t2.ID, t3.ID)
				}
			}

			if mode == "sequential" {
				// Sequential sorts by Name: web-a (c3), web-b (c1), web-c (c2)
				if t1.Name != "web-a" || t2.Name != "web-b" || t3.Name != "web-c" {
					t.Errorf("sequential sequence failed: got %s, %s, %s", t1.Name, t2.Name, t3.Name)
				}
			}
		})
	}
}

func TestScheduler_ExecutionAndRecovery(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Scheduler.AttackInterval = "10s"
	cfg.Scheduler.AttackDuration = "10ms"

	cRepo := newMockContainerRepo()
	eRepo := newMockExperimentRepo()
	cController := &mockContainerController{}
	aMgr := &mockAttackManager{attacks: []string{"pause"}}
	rMgr := &mockRecoveryManager{}

	s := NewScheduler(cfg, cController, cRepo, eRepo, aMgr, rMgr)

	// Save target to repo
	target := &domain.Container{ID: "c1", Name: "web", Status: "running", IsMonitored: true}
	cRepo.Save(target)

	// Run manual trigger of chaos injection
	s.injectChaos()

	// Wait for sleep in attack execution and recovery routine
	time.Sleep(30 * time.Millisecond)

	aMgr.mu.Lock()
	executedCount := len(aMgr.executed)
	aMgr.mu.Unlock()

	rMgr.mu.Lock()
	recoveredCount := len(rMgr.recovered)
	rMgr.mu.Unlock()

	if executedCount != 1 {
		t.Errorf("expected 1 executed attack, got %d", executedCount)
	}

	if recoveredCount != 1 {
		t.Errorf("expected 1 recovered attack, got %d", recoveredCount)
	}
}
