package docker

import (
	"context"
	"strings"
	"time"

	"chaosguard/internal/domain"
	"chaosguard/pkg/config"
	"chaosguard/pkg/logger"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

// DockerClientInterface defines the subset of Docker API Client methods we use.
// This allows us to easily mock the client in unit tests.
type DockerClientInterface interface {
	ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	ContainerPause(ctx context.Context, containerID string) error
	ContainerUnpause(ctx context.Context, containerID string) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerKill(ctx context.Context, containerID string, signal string) error
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
}

type DockerController struct {
	cli DockerClientInterface
	cfg *config.Config
}

func NewDockerController(cli DockerClientInterface, cfg *config.Config) *DockerController {
	return &DockerController{
		cli: cli,
		cfg: cfg,
	}
}

// Discover fetches all running/stopped containers and evaluates if they should be monitored
func (d *DockerController) Discover() ([]*domain.Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	containers, err := d.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	var results []*domain.Container
	for _, c := range containers {
		// Clean name (Docker prepends slash)
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		status := c.State
		if status == "" {
			status = c.Status
		}

		dc := &domain.Container{
			ID:          c.ID,
			Name:        name,
			Image:       c.Image,
			Status:      status,
			IsMonitored: d.shouldMonitor(name, c.Image),
			CreatedAt:   time.Unix(c.Created, 0),
			UpdatedAt:   time.Now(),
		}
		results = append(results, dc)
	}

	return results, nil
}

// shouldMonitor checks config filter lists to see if container is in scope
func (d *DockerController) shouldMonitor(name string, image string) bool {
	// 1. Safe mode database / system exclusions
	if d.cfg.SafeMode {
		for _, ex := range d.cfg.Containers.Exclude {
			if strings.Contains(strings.ToLower(name), strings.ToLower(ex)) ||
				strings.Contains(strings.ToLower(image), strings.ToLower(ex)) {
				return false
			}
		}
	}

	// 2. Explicit exclusions (even if safe mode is off)
	for _, ex := range d.cfg.Containers.Exclude {
		if name == ex || image == ex {
			return false
		}
	}

	// 3. Explicit inclusions (if defined, must match)
	if len(d.cfg.Containers.Include) > 0 {
		matched := false
		for _, inc := range d.cfg.Containers.Include {
			if name == inc || image == inc || strings.Contains(name, inc) {
				matched = true
				break
			}
		}
		return matched
	}

	return true
}

func (d *DockerController) Pause(id string) error {
	ctx := context.Background()
	logger.Info("Pausing Docker container: %s", id)
	return d.cli.ContainerPause(ctx, id)
}

func (d *DockerController) Unpause(id string) error {
	ctx := context.Background()
	logger.Info("Unpausing Docker container: %s", id)
	return d.cli.ContainerUnpause(ctx, id)
}

func (d *DockerController) Stop(id string) error {
	ctx := context.Background()
	logger.Info("Stopping Docker container: %s", id)
	return d.cli.ContainerStop(ctx, id, container.StopOptions{})
}

func (d *DockerController) Start(id string) error {
	ctx := context.Background()
	logger.Info("Starting Docker container: %s", id)
	return d.cli.ContainerStart(ctx, id, container.StartOptions{})
}

func (d *DockerController) Restart(id string) error {
	ctx := context.Background()
	logger.Info("Restarting Docker container: %s", id)
	return d.cli.ContainerRestart(ctx, id, container.StopOptions{})
}

func (d *DockerController) Kill(id string) error {
	ctx := context.Background()
	logger.Info("Killing Docker container: %s", id)
	return d.cli.ContainerKill(ctx, id, "SIGKILL")
}

func (d *DockerController) Inspect(id string) (*domain.Container, error) {
	ctx := context.Background()
	json, err := d.cli.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}

	name := strings.TrimPrefix(json.Name, "/")
	status := json.State.Status

	createdTime, err := time.Parse(time.RFC3339Nano, json.Created)
	if err != nil {
		createdTime, err = time.Parse(time.RFC3339, json.Created)
		if err != nil {
			createdTime = time.Now()
		}
	}

	return &domain.Container{
		ID:          json.ID,
		Name:        name,
		Image:       json.Config.Image,
		Status:      status,
		IsMonitored: d.shouldMonitor(name, json.Config.Image),
		CreatedAt:   createdTime,
		UpdatedAt:   time.Now(),
	}, nil
}
