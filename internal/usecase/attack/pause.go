package attack

import (
	"context"
	"fmt"

	"chaosguard/internal/domain"
)

type PauseAttack struct {
	controller domain.ContainerController
}

func NewPauseAttack(controller domain.ContainerController) *PauseAttack {
	return &PauseAttack{controller: controller}
}

func (a *PauseAttack) Name() string {
	return "pause"
}

func (a *PauseAttack) Run(ctx context.Context, containerID string, parameters string) error {
	return a.controller.Pause(containerID)
}

func (a *PauseAttack) Recover(ctx context.Context, containerID string) error {
	return a.controller.Unpause(containerID)
}

func (a *PauseAttack) Validate(ctx context.Context, containerID string) (bool, error) {
	c, err := a.controller.Inspect(containerID)
	if err != nil {
		return false, fmt.Errorf("failed to inspect container during validation: %w", err)
	}

	return c.Status == "paused", nil
}
