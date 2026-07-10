package attack

import (
	"context"
	"fmt"

	"chaosguard/internal/domain"
)

type RestartAttack struct {
	controller domain.ContainerController
}

func NewRestartAttack(controller domain.ContainerController) *RestartAttack {
	return &RestartAttack{controller: controller}
}

func (a *RestartAttack) Name() string {
	return "restart"
}

func (a *RestartAttack) Run(ctx context.Context, containerID string, parameters string) error {
	return a.controller.Restart(containerID)
}

func (a *RestartAttack) Recover(ctx context.Context, containerID string) error {
	// Restarting is transient. Recovery is generally not needed as the container
	// self-heals back to running.
	return nil
}

func (a *RestartAttack) Validate(ctx context.Context, containerID string) (bool, error) {
	c, err := a.controller.Inspect(containerID)
	if err != nil {
		return false, fmt.Errorf("failed to inspect container during validation: %w", err)
	}

	return c.Status == "running" || c.Status == "restarting", nil
}
