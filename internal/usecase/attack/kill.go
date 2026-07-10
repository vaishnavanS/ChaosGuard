package attack

import (
	"context"
	"fmt"

	"chaosguard/internal/domain"
)

type KillAttack struct {
	controller domain.ContainerController
}

func NewKillAttack(controller domain.ContainerController) *KillAttack {
	return &KillAttack{controller: controller}
}

func (a *KillAttack) Name() string {
	return "kill"
}

func (a *KillAttack) Run(ctx context.Context, containerID string, parameters string) error {
	return a.controller.Kill(containerID)
}

func (a *KillAttack) Recover(ctx context.Context, containerID string) error {
	return a.controller.Start(containerID)
}

func (a *KillAttack) Validate(ctx context.Context, containerID string) (bool, error) {
	c, err := a.controller.Inspect(containerID)
	if err != nil {
		return false, fmt.Errorf("failed to inspect container during validation: %w", err)
	}

	return c.Status == "exited" || c.Status == "stopped", nil
}
