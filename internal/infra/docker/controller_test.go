package docker

import (
	"context"
	"errors"
	"testing"
	"time"

	"chaosguard/pkg/config"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

type mockDockerClient struct {
	containers  []container.Summary
	inspectJSON types.ContainerJSON
	err         error
	called      map[string]bool
}

func newMockDockerClient() *mockDockerClient {
	return &mockDockerClient{
		called: make(map[string]bool),
	}
}

func (m *mockDockerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	m.called["ContainerList"] = true
	if m.err != nil {
		return nil, m.err
	}
	return m.containers, nil
}

func (m *mockDockerClient) ContainerPause(ctx context.Context, containerID string) error {
	m.called["ContainerPause"] = true
	return m.err
}

func (m *mockDockerClient) ContainerUnpause(ctx context.Context, containerID string) error {
	m.called["ContainerUnpause"] = true
	return m.err
}

func (m *mockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	m.called["ContainerStop"] = true
	return m.err
}

func (m *mockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	m.called["ContainerStart"] = true
	return m.err
}

func (m *mockDockerClient) ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error {
	m.called["ContainerRestart"] = true
	return m.err
}

func (m *mockDockerClient) ContainerKill(ctx context.Context, containerID string, signal string) error {
	m.called["ContainerKill"] = true
	return m.err
}

func (m *mockDockerClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	m.called["ContainerInspect"] = true
	if m.err != nil {
		return types.ContainerJSON{}, m.err
	}
	return m.inspectJSON, nil
}

func TestDockerController_Discover(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SafeMode = true
	cfg.Containers.Exclude = []string{"postgres", "ignored-service"}

	mockCli := newMockDockerClient()
	mockCli.containers = []container.Summary{
		{
			ID:     "c1",
			Names:  []string{"/web-service"},
			Image:  "nginx:alpine",
			State:  "running",
			Created: 1625097600,
		},
		{
			ID:     "c2",
			Names:  []string{"/my-postgres-db"},
			Image:  "postgres:13",
			State:  "running",
			Created: 1625097600,
		},
		{
			ID:     "c3",
			Names:  []string{"/ignored-service"},
			Image:  "alpine",
			State:  "stopped",
			Created: 1625097600,
		},
	}

	controller := NewDockerController(mockCli, cfg)
	list, err := controller.Discover()
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("expected 3 containers, got %d", len(list))
	}

	for _, c := range list {
		if c.ID == "c1" {
			if !c.IsMonitored {
				t.Error("expected web-service to be monitored")
			}
		} else if c.ID == "c2" {
			if c.IsMonitored {
				t.Error("expected database to be filtered in SafeMode")
			}
		} else if c.ID == "c3" {
			if c.IsMonitored {
				t.Error("expected ignored-service to be filtered by Exclude list")
			}
		}
	}
}

func TestDockerController_Actions(t *testing.T) {
	cfg := config.DefaultConfig()
	mockCli := newMockDockerClient()
	controller := NewDockerController(mockCli, cfg)

	tests := []struct {
		name       string
		action     func() error
		expectedOp string
	}{
		{"Pause", func() error { return controller.Pause("c1") }, "ContainerPause"},
		{"Unpause", func() error { return controller.Unpause("c1") }, "ContainerUnpause"},
		{"Stop", func() error { return controller.Stop("c1") }, "ContainerStop"},
		{"Start", func() error { return controller.Start("c1") }, "ContainerStart"},
		{"Restart", func() error { return controller.Restart("c1") }, "ContainerRestart"},
		{"Kill", func() error { return controller.Kill("c1") }, "ContainerKill"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCli.err = nil
			mockCli.called[tt.expectedOp] = false

			err := tt.action()
			if err != nil {
				t.Errorf("%s failed: %v", tt.name, err)
			}
			if !mockCli.called[tt.expectedOp] {
				t.Errorf("%s did not invoke mock client operation %s", tt.name, tt.expectedOp)
			}
		})
	}
}

func TestDockerController_Inspect(t *testing.T) {
	cfg := config.DefaultConfig()
	mockCli := newMockDockerClient()
	mockCli.inspectJSON = types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:      "c1",
			Name:    "/my-inspect-container",
			Created: "2026-07-10T12:00:00Z",
			State: &types.ContainerState{
				Status: "running",
			},
		},
		Config: &container.Config{
			Image: "ubuntu:latest",
		},
	}

	controller := NewDockerController(mockCli, cfg)
	c, err := controller.Inspect("c1")
	if err != nil {
		t.Fatalf("Inspect failed: %v", err)
	}

	if c.ID != "c1" {
		t.Errorf("expected ID 'c1', got %s", c.ID)
	}
	if c.Name != "my-inspect-container" {
		t.Errorf("expected name 'my-inspect-container', got %s", c.Name)
	}
	if c.Status != "running" {
		t.Errorf("expected status 'running', got %s", c.Status)
	}
	if c.Image != "ubuntu:latest" {
		t.Errorf("expected image 'ubuntu:latest', got %s", c.Image)
	}
}

func TestDockerController_InspectFallbackTime(t *testing.T) {
	cfg := config.DefaultConfig()
	mockCli := newMockDockerClient()
	mockCli.inspectJSON = types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:      "c1",
			Name:    "/my-inspect-container",
			Created: "invalid-time-string",
			State: &types.ContainerState{
				Status: "running",
			},
		},
		Config: &container.Config{
			Image: "ubuntu:latest",
		},
	}

	controller := NewDockerController(mockCli, cfg)
	c, err := controller.Inspect("c1")
	if err != nil {
		t.Fatalf("Inspect failed: %v", err)
	}

	// Should not crash, and should set CreatedAt to time.Now() (which is close to time.Now())
	if time.Since(c.CreatedAt) > 5*time.Second {
		t.Errorf("expected fallback time close to now, got %v", c.CreatedAt)
	}
}

func TestDockerController_DiscoverError(t *testing.T) {
	cfg := config.DefaultConfig()
	mockCli := newMockDockerClient()
	mockCli.err = errors.New("docker daemon is offline")
	controller := NewDockerController(mockCli, cfg)

	_, err := controller.Discover()
	if err == nil {
		t.Error("expected error from Discover when Docker client returns error")
	}
}
