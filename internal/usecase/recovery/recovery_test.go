package recovery

import (
	"context"
	"sync"
	"testing"
	"time"

	"chaosguard/internal/domain"
)

// Mock AttackManager
type mockAttackManager struct {
	mu      sync.Mutex
	attacks map[string]domain.Attack
}

func newMockAttackManager() *mockAttackManager {
	return &mockAttackManager{
		attacks: make(map[string]domain.Attack),
	}
}

func (m *mockAttackManager) Register(attack domain.Attack) {}

func (m *mockAttackManager) Get(name string) (domain.Attack, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	atk, ok := m.attacks[name]
	if !ok {
		return nil, domain.ErrExperimentNotFound // Using as a generic error
	}
	return atk, nil
}

func (m *mockAttackManager) List() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var list []string
	for name := range m.attacks {
		list = append(list, name)
	}
	return list
}

func (m *mockAttackManager) Execute(ctx context.Context, experiment *domain.Experiment) error {
	return nil
}

func (m *mockAttackManager) addAttack(name string, attack domain.Attack) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.attacks[name] = attack
}

// Mock Attack
type mockAttack struct {
	name         string
	recoverCalls int
	recoverError error
}

func newMockAttack(name string) *mockAttack {
	return &mockAttack{name: name}
}

func (a *mockAttack) Name() string {
	return a.name
}

func (a *mockAttack) Run(ctx context.Context, containerID string, parameters string) error {
	return nil
}

func (a *mockAttack) Recover(ctx context.Context, containerID string) error {
	a.recoverCalls++
	return a.recoverError
}

func (a *mockAttack) Validate(ctx context.Context, containerID string) (bool, error) {
	return true, nil
}

// Mock ExperimentRepository
type mockExperimentRepo struct {
	mu       sync.Mutex
	status   map[string]string
	endTimes map[string]time.Time
}

func newMockExperimentRepo() *mockExperimentRepo {
	return &mockExperimentRepo{
		status:   make(map[string]string),
		endTimes: make(map[string]time.Time),
	}
}

func (r *mockExperimentRepo) Get(id string) (*domain.Experiment, error) {
	return nil, nil
}

func (r *mockExperimentRepo) List() ([]*domain.Experiment, error) {
	return nil, nil
}

func (r *mockExperimentRepo) Save(e *domain.Experiment) error {
	return nil
}

func (r *mockExperimentRepo) UpdateStatus(id string, status string, errStr string, endedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.status[id] = status
	if endedAt != nil {
		r.endTimes[id] = *endedAt
	}
	return nil
}

func TestRecoveryManager_Recover_Success(t *testing.T) {
	aMgr := newMockAttackManager()
	eRepo := newMockExperimentRepo()
	rMgr := NewManager(aMgr, eRepo)

	// Register a mock attack
	mockAtk := newMockAttack("pause")
	aMgr.addAttack("pause", mockAtk)

	exp := &domain.Experiment{
		ID:                "exp-1",
		TargetContainerID: "c1",
		AttackType:        "pause",
		Status:            domain.ExperimentStatusRunning,
	}

	err := rMgr.Recover(context.Background(), exp)
	if err != nil {
		t.Fatalf("Recover failed: %v", err)
	}

	eRepo.mu.Lock()
	finalStatus := eRepo.status["exp-1"]
	eRepo.mu.Unlock()

	if finalStatus != domain.ExperimentStatusRecovered {
		t.Errorf("expected status to be 'recovered', got %s", finalStatus)
	}

	if mockAtk.recoverCalls != 1 {
		t.Errorf("expected Recover to be called once on attack, got %d calls", mockAtk.recoverCalls)
	}
}

func TestRecoveryManager_Recover_AttackNotFound(t *testing.T) {
	aMgr := newMockAttackManager()
	eRepo := newMockExperimentRepo()
	rMgr := NewManager(aMgr, eRepo)

	exp := &domain.Experiment{
		ID:                "exp-1",
		TargetContainerID: "c1",
		AttackType:        "nonexistent",
		Status:            domain.ExperimentStatusRunning,
	}

	err := rMgr.Recover(context.Background(), exp)
	if err == nil {
		t.Error("expected Recover to fail when attack not found")
	}

	eRepo.mu.Lock()
	finalStatus := eRepo.status["exp-1"]
	eRepo.mu.Unlock()

	if finalStatus != domain.ExperimentStatusFailed {
		t.Errorf("expected status to be 'failed', got %s", finalStatus)
	}
}

func TestRecoveryManager_Recover_RecoveryError(t *testing.T) {
	aMgr := newMockAttackManager()
	eRepo := newMockExperimentRepo()
	rMgr := NewManager(aMgr, eRepo)

	// Register an attack that fails on recovery
	mockAtk := newMockAttack("pause")
	mockAtk.recoverError = domain.ErrExperimentNotFound
	aMgr.addAttack("pause", mockAtk)

	exp := &domain.Experiment{
		ID:                "exp-1",
		TargetContainerID: "c1",
		AttackType:        "pause",
		Status:            domain.ExperimentStatusRunning,
	}

	err := rMgr.Recover(context.Background(), exp)
	if err == nil {
		t.Error("expected Recover to fail when recovery error occurs")
	}

	eRepo.mu.Lock()
	finalStatus := eRepo.status["exp-1"]
	eRepo.mu.Unlock()

	if finalStatus != domain.ExperimentStatusFailed {
		t.Errorf("expected status to be 'failed', got %s", finalStatus)
	}
}

func TestRecoveryManager_RecoverAllActive(t *testing.T) {
	aMgr := newMockAttackManager()
	eRepo := newMockExperimentRepo()
	rMgr := NewManager(aMgr, eRepo)

	// Register attacks
	mockAtk1 := newMockAttack("pause")
	mockAtk2 := newMockAttack("stop")
	aMgr.addAttack("pause", mockAtk1)
	aMgr.addAttack("stop", mockAtk2)

	// Track active experiments
	exp1 := &domain.Experiment{
		ID:                "exp-1",
		TargetContainerID: "c1",
		AttackType:        "pause",
		Status:            domain.ExperimentStatusRunning,
	}
	exp2 := &domain.Experiment{
		ID:                "exp-2",
		TargetContainerID: "c2",
		AttackType:        "stop",
		Status:            domain.ExperimentStatusRunning,
	}

	rMgr.TrackExperiment(exp1)
	rMgr.TrackExperiment(exp2)

	err := rMgr.RecoverAllActive(context.Background())
	if err != nil {
		t.Fatalf("RecoverAllActive failed: %v", err)
	}

	eRepo.mu.Lock()
	status1 := eRepo.status["exp-1"]
	status2 := eRepo.status["exp-2"]
	eRepo.mu.Unlock()

	if status1 != domain.ExperimentStatusRecovered {
		t.Errorf("expected exp-1 status to be 'recovered', got %s", status1)
	}
	if status2 != domain.ExperimentStatusRecovered {
		t.Errorf("expected exp-2 status to be 'recovered', got %s", status2)
	}

	if mockAtk1.recoverCalls != 1 {
		t.Errorf("expected mockAtk1 Recover to be called once, got %d", mockAtk1.recoverCalls)
	}
	if mockAtk2.recoverCalls != 1 {
		t.Errorf("expected mockAtk2 Recover to be called once, got %d", mockAtk2.recoverCalls)
	}
}

func TestRecoveryManager_TrackExperiment(t *testing.T) {
	aMgr := newMockAttackManager()
	eRepo := newMockExperimentRepo()
	rMgr := NewManager(aMgr, eRepo)

	exp := &domain.Experiment{
		ID:                "exp-1",
		TargetContainerID: "c1",
		AttackType:        "pause",
	}

	rMgr.TrackExperiment(exp)

	rMgr.mu.RLock()
	tracked := rMgr.active["exp-1"]
	rMgr.mu.RUnlock()

	if tracked == nil || tracked.ID != "exp-1" {
		t.Error("expected experiment to be tracked")
	}

	// Test recovery removes from tracking
	mockAtk := newMockAttack("pause")
	aMgr.addAttack("pause", mockAtk)

	rMgr.Recover(context.Background(), exp)

	rMgr.mu.RLock()
	tracked = rMgr.active["exp-1"]
	rMgr.mu.RUnlock()

	if tracked != nil {
		t.Error("expected experiment to be removed from tracking after recovery")
	}
}
