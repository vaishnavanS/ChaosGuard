package runtime

import (
	"context"
	"os"
	"time"

	"chaosguard/internal/domain"
	"chaosguard/pkg/logger"
)

const (
	defaultShutdownTimeout   = 2 * time.Minute
	activeExperimentWaitTime = 90 * time.Second
)

// Shutdown gracefully stops all services and releases resources.
func (a *App) Shutdown(ctx context.Context) error {
	if a.lifecycle.State() == StateStopping || a.lifecycle.State() == StateStopped {
		return nil
	}

	a.lifecycle.SetState(StateStopping)
	logger.Info("Beginning graceful shutdown")

	if ctx == nil {
		ctx = context.Background()
	}
	shutdownCtx, cancel := context.WithTimeout(ctx, defaultShutdownTimeout)
	defer cancel()

	if a.deps.Scheduler.IsRunning() {
		logger.Info("Stopping scheduler")
		if err := a.deps.Scheduler.Stop(); err != nil {
			logger.Error(err, "Failed to stop scheduler")
		}
	}

	logger.Info("Waiting for active experiments to finish")
	waitCtx, waitCancel := context.WithTimeout(shutdownCtx, activeExperimentWaitTime)
	if err := a.deps.Scheduler.WaitForActive(waitCtx); err != nil {
		logger.Warn("Timed out waiting for active experiments: %v", err)
	}
	waitCancel()

	logger.Info("Recovering active experiments")
	if err := a.deps.RecoveryManager.RecoverAllActive(shutdownCtx); err != nil {
		logger.Error(err, "Failed to recover all active experiments")
	}

	logger.Info("Recovering paused containers")
	if err := recoverPausedContainers(shutdownCtx, a.deps.DockerController); err != nil {
		logger.Error(err, "Failed to recover paused containers")
	}

	if a.deps.APIServer.IsRunning() {
		logger.Info("Stopping REST API server")
		if err := a.deps.APIServer.Stop(shutdownCtx); err != nil {
			logger.Error(err, "Failed to stop REST API server")
		}
	}

	if a.deps.MetricsServer.IsRunning() {
		logger.Info("Stopping metrics server")
		if err := a.deps.MetricsServer.Stop(shutdownCtx); err != nil {
			logger.Error(err, "Failed to stop metrics server")
		}
	}

	if a.deps.DockerCloser != nil {
		logger.Info("Closing Docker client")
		if err := a.deps.DockerCloser.Close(); err != nil {
			logger.Error(err, "Failed to close Docker client")
		}
	}

	if a.deps.Store != nil {
		logger.Info("Closing SQLite database")
		if err := a.deps.Store.Close(); err != nil {
			logger.Error(err, "Failed to close SQLite database")
		}
	}

	flushLogger()
	a.lifecycle.SetState(StateStopped)
	logger.Info("Shutdown complete")
	return nil
}

func recoverPausedContainers(ctx context.Context, controller domain.ContainerController) error {
	containers, err := controller.Discover()
	if err != nil {
		return err
	}

	for _, container := range containers {
		if container.Status != "paused" {
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		logger.Info("Unpausing container %s during shutdown recovery", container.Name)
		if err := controller.Unpause(container.ID); err != nil {
			logger.Error(err, "Failed to unpause container %s", container.Name)
		}
	}

	return nil
}

func flushLogger() {
	_ = os.Stdout.Sync()
}
