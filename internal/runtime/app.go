package runtime

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"chaosguard/pkg/logger"
)

// App is the ChaosGuard application runtime and composition root.
type App struct {
	deps      *Dependencies
	lifecycle *Lifecycle
	runCtx    context.Context
	runCancel context.CancelFunc
}

// Start bootstraps dependencies and runs the application until shutdown.
func Start(opts Options) error {
	logger.Setup(opts.Verbose, false)

	app, err := Bootstrap(opts)
	if err != nil {
		return err
	}
	return app.Run()
}

// Run starts background services, waits for shutdown, and performs graceful teardown.
func (a *App) Run() error {
	a.runCtx, a.runCancel = context.WithCancel(context.Background())
	defer a.runCancel()

	if err := a.deps.MetricsServer.Start(a.runCtx); err != nil {
		return err
	}

	if err := a.deps.APIServer.Start(a.runCtx); err != nil {
		_ = a.Shutdown(context.Background())
		return err
	}

	if err := a.deps.Scheduler.Start(a.runCtx); err != nil {
		_ = a.Shutdown(context.Background())
		return err
	}

	a.lifecycle.SetState(StateRunning)
	logger.Info("Application ready")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		logger.Info("Shutdown signal received")
	case <-a.runCtx.Done():
		logger.Info("Runtime context canceled")
	}
	signal.Stop(sigCh)

	return a.Shutdown(context.Background())
}

// LifecycleState exposes the current lifecycle state for operators and future APIs.
func (a *App) LifecycleState() State {
	return a.lifecycle.State()
}

// Lifecycle returns the lifecycle manager instance.
func (a *App) Lifecycle() *Lifecycle {
	return a.lifecycle
}

// Dependencies returns the wired dependency graph.
func (a *App) Dependencies() *Dependencies {
	return a.deps
}

// Stop triggers an in-process shutdown without waiting for OS signals.
func (a *App) Stop() {
	if a.runCancel != nil {
		a.runCancel()
	}
}
