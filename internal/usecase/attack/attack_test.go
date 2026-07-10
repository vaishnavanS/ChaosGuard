package attack

import (
	"context"
	"sync"
	"testing"
	"time"

	"chaosguard/internal/domain"
)

// Mock ContainerController
type mockContainerController struct {
	mu              sync.Mutex
	status          string
	called          map[string]bool
	inspectError    error
	actionError     error
	noStatusUpdate  bool // When true, actions don't update status (simulates injection failure)
}

func newMockContainerController(initialStatus string) *mockContainerController {
	return &mockContainerController{
		status: initialStatus,
		called: make(map[string]bool),
	}
}

func (m *mockContainerController) Discover() ([]*domain.Container, error) {
	return nil, nil
}

func (m *mockContainerController) Pause(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called["Pause"] = true
	if m.actionError != nil {
		return m.actionError
	}
	if !m.noStatusUpdate {
		m.status = "paused"
	}
	return nil
}

func (m *mockContainerController) Unpause(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called["Unpause"] = true
	if m.actionError != nil {
		return m.actionError
	}
	m.status = "running"
	return nil
}

func (m *mockContainerController) Stop(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called["Stop"] = true
	if m.actionError != nil {
		return m.actionError
	}
	m.status = "exited"
	return nil
}

func (m *mockContainerController) Start(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called["Start"] = true
	if m.actionError != nil {
		return m.actionError
	}
	m.status = "running"
	return nil
}

func (m *mockContainerController) Restart(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called["Restart"] = true
	if m.actionError != nil {
		return m.actionError
	}
	m.status = "running"
	return nil
}

func (m *mockContainerController) Kill(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called["Kill"] = true
	if m.actionError != nil {
		return m.actionError
	}
	m.status = "exited"
	return nil
}

func (m *mockContainerController) Inspect(id string) (*domain.Container, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called["Inspect"] = true
	if m.inspectError != nil {
		return nil, m.inspectError
	}
	return &domain.Container{
		ID:     id,
		Status: m.status,
	}, nil
}

// Mock ExperimentRepository
type mockExperimentRepo struct {
	mu     sync.Mutex
	saved  []*domain.Experiment
	status map[string]string
}

func newMockExperimentRepo() *mockExperimentRepo {
	return &mockExperimentRepo{
		status: make(map[string]string),
	}
}

func (r *mockExperimentRepo) Get(id string) (*domain.Experiment, error) {
	return nil, nil
}

func (r *mockExperimentRepo) List() ([]*domain.Experiment, error) {
	return nil, nil
}

func (r *mockExperimentRepo) Save(e *domain.Experiment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.saved = append(r.saved, e)
	r.status[e.ID] = e.Status
	return nil
}

func (r *mockExperimentRepo) UpdateStatus(id string, status string, errStr string, endedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.status[id] = status
	return nil
}

func TestAttackManager_RegisterAndGet(t *testing.T) {
	repo := newMockExperimentRepo()
	mgr := NewManager(repo)

	controller := newMockContainerController("running")
	pauseAttack := NewPauseAttack(controller)

	mgr.Register(pauseAttack)

	list := mgr.List()
	if len(list) != 1 || list[0] != "pause" {
		t.Errorf("expected list to be ['pause'], got %v", list)
	}

	atk, err := mgr.Get("pause")
	if err != nil {
		t.Fatalf("failed to get pause attack: %v", err)
	}
	if atk.Name() != "pause" {
		t.Errorf("expected attack name 'pause', got %s", atk.Name())
	}

	_, err = mgr.Get("nonexistent")
	if err == nil {
		t.Error("expected error retrieving nonexistent attack")
	}
}

func TestAttackManager_Execute_Success(t *testing.T) {
	repo := newMockExperimentRepo()
	mgr := NewManager(repo)

	controller := newMockContainerController("running")
	pauseAttack := NewPauseAttack(controller)
	mgr.Register(pauseAttack)

	exp := &domain.Experiment{
		ID:                "exp-1",
		TargetContainerID: "c1",
		AttackType:        "pause",
		Status:            domain.ExperimentStatusPending,
	}

	err := mgr.Execute(context.Background(), exp)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	repo.mu.Lock()
	finalStatus := repo.status["exp-1"]
	repo.mu.Unlock()

	if finalStatus != domain.ExperimentStatusRunning {
		t.Errorf("expected experiment status to be 'running', got %s", finalStatus)
	}

	controller.mu.Lock()
	paused := controller.called["Pause"]
	inspected := controller.called["Inspect"]
	controller.mu.Unlock()

	if !paused {
		t.Error("expected controller Pause to be called")
	}
	if !inspected {
		t.Error("expected controller Inspect to be called")
	}
}

func TestAttackManager_Execute_ValidationFailure(t *testing.T) {
	repo := newMockExperimentRepo()
	mgr := NewManager(repo)

	// Create a controller that simulates injection failure (Pause succeeds but doesn't change status)
	controller := newMockContainerController("running")
	controller.noStatusUpdate = true // Simulate failed injection
	
	pauseAttack := NewPauseAttack(controller)
	mgr.Register(pauseAttack)

	exp := &domain.Experiment{
		ID:                "exp-1",
		TargetContainerID: "c1",
		AttackType:        "pause",
		Status:            domain.ExperimentStatusPending,
	}

	err := mgr.Execute(context.Background(), exp)
	if err == nil {
		t.Error("expected execute to fail validation")
	}

	repo.mu.Lock()
	finalStatus := repo.status["exp-1"]
	repo.mu.Unlock()

	if finalStatus != domain.ExperimentStatusFailed {
		t.Errorf("expected experiment status to be 'failed', got %s", finalStatus)
	}
}

func TestConcreteAttacks(t *testing.T) {
	controller := newMockContainerController("running")

	attacks := []struct {
		attack       domain.Attack
		runCall      string
		recoverCall  string
		expectedRun  string
		expectedRec  string
	}{
		{NewPauseAttack(controller), "Pause", "Unpause", "paused", "running"},
		{NewStopAttack(controller), "Stop", "Start", "exited", "running"},
		{NewRestartAttack(controller), "Restart", "", "running", "running"}, // restart recover does nothing
		{NewKillAttack(controller), "Kill", "Start", "exited", "running"},
	}

	for _, tt := range attacks {
		t.Run(tt.attack.Name(), func(t *testing.T) {
			controller.mu.Lock()
			controller.status = "running"
			controller.called = make(map[string]bool)
			controller.mu.Unlock()

			// Run
			err := tt.attack.Run(context.Background(), "c1", "")
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}
			if tt.runCall != "" && !controller.called[tt.runCall] {
				t.Errorf("expected controller %s to be called during Run", tt.runCall)
			}

			// Validate
			ok, err := tt.attack.Validate(context.Background(), "c1")
			if err != nil {
				t.Fatalf("Validate failed: %v", err)
			}
			if !ok {
				t.Error("expected validation to succeed")
			}
			if controller.status != tt.expectedRun {
				t.Errorf("expected status %s after run, got %s", tt.expectedRun, controller.status)
			}

			// Recover
			err = tt.attack.Recover(context.Background(), "c1")
			if err != nil {
				t.Fatalf("Recover failed: %v", err)
			}
			if tt.recoverCall != "" && !controller.called[tt.recoverCall] {
				t.Errorf("expected controller %s to be called during Recover", tt.recoverCall)
			}
			if controller.status != tt.expectedRec {
				t.Errorf("expected status %s after recovery, got %s", tt.expectedRec, controller.status)
			}
		})
	}
}
