package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"chaosguard/internal/api"
	"chaosguard/internal/api/handlers"
	"chaosguard/internal/api/responses"
	"chaosguard/internal/domain"
	"chaosguard/internal/usecase/scheduler"
	"chaosguard/pkg/config"
)

// Mock ContainerRepository
type mockContainerRepo struct {
	mu   sync.Mutex
	data map[string]*domain.Container
}

func (r *mockContainerRepo) Get(id string) (*domain.Container, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.data[id]
	if !ok {
		return nil, errors.New("container not found")
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

func (r *mockContainerRepo) Create(c *domain.Container) error { return r.Save(c) }
func (r *mockContainerRepo) Update(c *domain.Container) error { return r.Save(c) }
func (r *mockContainerRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}
func (r *mockContainerRepo) UpdateState(id string, state string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.data[id]; ok {
		c.Status = state
	}
	return nil
}

// Mock ExperimentRepository
type mockExperimentRepo struct {
	mu   sync.Mutex
	data map[string]*domain.Experiment
}

func (r *mockExperimentRepo) Get(id string) (*domain.Experiment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.data[id]
	if !ok {
		return nil, errors.New("experiment not found")
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

func (r *mockExperimentRepo) Create(e *domain.Experiment) error { return r.Save(e) }
func (r *mockExperimentRepo) Update(e *domain.Experiment) error { return r.Save(e) }
func (r *mockExperimentRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}
func (r *mockExperimentRepo) UpdateStatus(id string, status string, errStr string, endedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.data[id]; ok {
		e.Status = status
		e.ErrorMessage = errStr
		e.EndedAt = endedAt
	}
	return nil
}

// Mock ContainerController
type mockContainerController struct{}

func (c *mockContainerController) Discover() ([]*domain.Container, error) { return nil, nil }
func (c *mockContainerController) Pause(id string) error                  { return nil }
func (c *mockContainerController) Unpause(id string) error                { return nil }
func (c *mockContainerController) Stop(id string) error                   { return nil }
func (c *mockContainerController) Start(id string) error                  { return nil }
func (c *mockContainerController) Restart(id string) error                { return nil }
func (c *mockContainerController) Kill(id string) error                   { return nil }
func (c *mockContainerController) Inspect(id string) (*domain.Container, error) {
	return &domain.Container{ID: id, Name: "web", Status: "running"}, nil
}

// Mock AttackManager
type mockAttackManager struct {
	attacks map[string]domain.Attack
}

func (m *mockAttackManager) Register(attack domain.Attack) {}
func (m *mockAttackManager) Get(name string) (domain.Attack, error) {
	if name == "invalid" {
		return nil, errors.New("invalid attack")
	}
	return nil, nil
}
func (m *mockAttackManager) List() []string { return []string{"pause"} }
func (m *mockAttackManager) Execute(ctx context.Context, experiment *domain.Experiment) error {
	experiment.Status = domain.ExperimentStatusRunning
	return nil
}

// Mock RecoveryManager
type mockRecoveryManager struct{}

func (m *mockRecoveryManager) Recover(ctx context.Context, experiment *domain.Experiment) error {
	experiment.Status = domain.ExperimentStatusRecovered
	return nil
}
func (m *mockRecoveryManager) RecoverAllActive(ctx context.Context) error    { return nil }
func (m *mockRecoveryManager) TrackExperiment(experiment *domain.Experiment) {}

// Mock StateProvider
type mockStateProvider struct {
	state string
}

func (p *mockStateProvider) GetState() string {
	return p.state
}

func setupTestRouter() (*handlers.Handler, http.Handler) {
	cfg := config.DefaultConfig()
	cRepo := &mockContainerRepo{data: make(map[string]*domain.Container)}
	eRepo := &mockExperimentRepo{data: make(map[string]*domain.Experiment)}
	cController := &mockContainerController{}
	aMgr := &mockAttackManager{}
	rMgr := &mockRecoveryManager{}
	sched := scheduler.NewScheduler(cfg, cController, cRepo, eRepo, aMgr, rMgr)
	stateProv := &mockStateProvider{state: "running"}

	// Seed repositories
	_ = cRepo.Save(&domain.Container{
		ID:          "c1",
		Name:        "web-service",
		Image:       "nginx:alpine",
		Status:      "running",
		IsMonitored: true,
		CreatedAt:   time.Now().Add(-1 * time.Hour),
		UpdatedAt:   time.Now(),
	})

	_ = eRepo.Save(&domain.Experiment{
		ID:                "exp1",
		TargetContainerID: "c1",
		ContainerName:     "web-service",
		AttackType:        "pause",
		Duration:          10,
		Status:            "completed",
		StartedAt:         time.Now().Add(-1 * time.Minute),
	})

	h := handlers.NewHandler(cfg, cRepo, cController, eRepo, aMgr, rMgr, sched, stateProv, "v0.1.0-test")
	r := api.SetupRouter(h)

	return h, r
}

func TestGetHealth(t *testing.T) {
	_, router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp responses.SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success = true")
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected health response data format")
	}

	if dataMap["status"] != "healthy" || dataMap["state"] != "running" {
		t.Errorf("unexpected health data: %v", resp.Data)
	}
}

func TestGetContainers(t *testing.T) {
	_, router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/containers", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp responses.SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success = true")
	}
}

func TestGetContainerByID(t *testing.T) {
	_, router := setupTestRouter()

	// 1. Success case
	req, _ := http.NewRequest("GET", "/containers/c1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// 2. Not found case
	reqNotFound, _ := http.NewRequest("GET", "/containers/invalid", nil)
	wNotFound := httptest.NewRecorder()
	router.ServeHTTP(wNotFound, reqNotFound)

	if wNotFound.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", wNotFound.Code)
	}
}

func TestGetExperiments(t *testing.T) {
	_, router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/experiments", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetExperimentByID(t *testing.T) {
	_, router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/experiments/exp1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestCreateExperiment(t *testing.T) {
	_, router := setupTestRouter()

	// 1. Valid request
	payload := []byte(`{"target_container_id":"c1","attack_type":"pause","duration":10}`)
	req, _ := http.NewRequest("POST", "/experiments", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 2. Missing fields validation
	invalidPayload := []byte(`{"target_container_id":"","attack_type":"pause"}`)
	reqInvalid, _ := http.NewRequest("POST", "/experiments", bytes.NewBuffer(invalidPayload))
	reqInvalid.Header.Set("Content-Type", "application/json")
	wInvalid := httptest.NewRecorder()
	router.ServeHTTP(wInvalid, reqInvalid)

	if wInvalid.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", wInvalid.Code)
	}
}

func TestDeleteExperiment(t *testing.T) {
	_, router := setupTestRouter()

	req, _ := http.NewRequest("DELETE", "/experiments/exp1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetSchedulerStatus(t *testing.T) {
	_, router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/scheduler/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestSchedulerStartStop(t *testing.T) {
	_, router := setupTestRouter()

	// Stop scheduler
	reqStop, _ := http.NewRequest("POST", "/scheduler/stop", nil)
	wStop := httptest.NewRecorder()
	router.ServeHTTP(wStop, reqStop)

	// Since scheduler is not running initially in test setup, Stop should fail/warn or return 400 if it is already stopped.
	// Actually, initially scheduler is NOT running, so stop should return 400.
	if wStop.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 when stopping an already stopped scheduler, got %d", wStop.Code)
	}

	// Start scheduler
	reqStart, _ := http.NewRequest("POST", "/scheduler/start", nil)
	wStart := httptest.NewRecorder()
	router.ServeHTTP(wStart, reqStart)

	if wStart.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", wStart.Code)
	}
}

func TestGetRuntime(t *testing.T) {
	_, router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/runtime", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_CORS(t *testing.T) {
	_, router := setupTestRouter()

	req, _ := http.NewRequest("OPTIONS", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204 for OPTIONS, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin header to be '*'")
	}
}

func TestMiddleware_RequestID(t *testing.T) {
	_, router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	reqID := w.Header().Get("X-Request-ID")
	if reqID == "" {
		t.Error("expected X-Request-ID header to be set in response")
	}
}

func TestStopRuntime(t *testing.T) {
	h, router := setupTestRouter()

	// 1. Without stop callback registered
	req, _ := http.NewRequest("POST", "/runtime/stop", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 when stopFunc is nil, got %d", w.Code)
	}

	// 2. With stop callback registered
	stopCalled := false
	h.SetStopFunc(func() {
		stopCalled = true
	})

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200 when stopFunc is set, got %d", w2.Code)
	}

	// Wait to let async goroutine execute stopFunc
	time.Sleep(150 * time.Millisecond)
	if !stopCalled {
		t.Error("expected stopFunc callback to be called")
	}
}
