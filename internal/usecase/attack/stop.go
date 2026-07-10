package attack

import (
	"context"
	"fmt"

	"chaosguard/internal/domain"
)

type StopAttack struct {
	controller domain.ContainerController
}

func NewStopAttack(controller domain.ContainerController) *StopAttack {
	return &StopAttack{controller: controller}
}

func (a *StopAttack) Name() string {
	return "stop"
}

func (a *StopAttack) Run(ctx context.Context, containerID string, parameters string) error {
	return a.controller.Stop(containerID)
}

func (a *StopAttack) Recover(ctx context.Context, containerID string) error {
	return a.controller.Start(containerID)
}

func (a *StopAttack) Validate(ctx context.Context, containerID string) (bool, error) {
	c, err := a.controller.Inspect(containerID)
	if err != nil {
		return false, fmt.Errorf("failed to inspect container during validation: %w", err)
	}

	return c.Status == "exited" || c.Status == "stopped", nil
}
