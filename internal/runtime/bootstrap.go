package runtime

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"chaosguard/internal/infra/docker"
	"chaosguard/internal/infra/sqlite"
	"chaosguard/internal/usecase/attack"
	"chaosguard/internal/usecase/recovery"
	"chaosguard/internal/usecase/scheduler"
	"chaosguard/pkg/config"
	"chaosguard/pkg/logger"
	"chaosguard/pkg/metrics"
	"chaosguard/pkg/server"

	dockerclient "github.com/docker/docker/client"
)

// Options configures application startup.
type Options struct {
	ConfigPath string
	Verbose    bool
	SafeMode   *bool
	Hooks      BootstrapHooks
}

// BootstrapHooks allows tests to inject dependencies and skip external checks.
type BootstrapHooks struct {
	DockerClient   docker.DockerClientInterface
	SkipDockerPing bool
	SkipPortCheck  bool
}

// Dependencies holds wired application components constructed during bootstrap.
type Dependencies struct {
	Config             *config.Config
	Store              *sqlite.Store
	DockerController   *docker.DockerController
	DockerCloser       io.Closer
	AttackManager      *attack.Manager
	RecoveryManager    *recovery.Manager
	Scheduler          *scheduler.Scheduler
	MetricsRegistry    *metrics.Registry
	ChaosCollector     *metrics.ChaosCollector
	ContainerCollector *metrics.Collector
	MetricsServer      *server.MetricsServer
}

// ValidatePrerequisites verifies that required external services are reachable before startup.
func ValidatePrerequisites(cfg *config.Config, hooks BootstrapHooks) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration invalid: %w", err)
	}

	if !hooks.SkipDockerPing {
		logger.Info("Checking Docker daemon connectivity")
		cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
		if err != nil {
			return fmt.Errorf("docker client initialization failed: %w", err)
		}
		defer cli.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := cli.Ping(ctx); err != nil {
			return fmt.Errorf("docker daemon unreachable: %w", err)
		}
	}

	logger.Info("Checking SQLite write permissions at %s", cfg.Database.Path)
	if err := checkDatabaseWritable(cfg.Database.Path); err != nil {
		return fmt.Errorf("sqlite not writable: %w", err)
	}

	if !hooks.SkipPortCheck {
		logger.Info("Checking metrics server port %d availability", cfg.Metrics.Port)
		if err := checkPortAvailable(cfg.Metrics.Port); err != nil {
			return fmt.Errorf("metrics port unavailable: %w", err)
		}
	}

	return nil
}

// Bootstrap constructs the application dependency graph.
func Bootstrap(opts Options) (*App, error) {
	lifecycle := NewLifecycle()
	lifecycle.SetState(StateStarting)

	logger.Info("Loading configuration")
	cfg, err := config.Load(opts.ConfigPath, nil)
	if err != nil {
		return nil, fmt.Errorf("load configuration: %w", err)
	}
	if opts.SafeMode != nil {
		cfg.SafeMode = *opts.SafeMode
	}

	if err := ValidatePrerequisites(cfg, opts.Hooks); err != nil {
		return nil, err
	}

	deps, err := buildDependencies(cfg, opts.Hooks)
	if err != nil {
		return nil, err
	}

	return &App{
		deps:      deps,
		lifecycle: lifecycle,
	}, nil
}

func buildDependencies(cfg *config.Config, hooks BootstrapHooks) (*Dependencies, error) {
	logger.Info("Opening SQLite database at %s", cfg.Database.Path)
	store, err := sqlite.NewStore(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	logger.Info("Connecting to Docker")
	var dockerCli docker.DockerClientInterface
	var dockerCloser io.Closer

	if hooks.DockerClient != nil {
		dockerCli = hooks.DockerClient
	} else {
		cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
		if err != nil {
			store.Close()
			return nil, fmt.Errorf("create docker client: %w", err)
		}
		dockerCli = cli
		dockerCloser = cli
	}

	dockerController := docker.NewDockerController(dockerCli, cfg)

	logger.Info("Registering attacks")
	metricsRegistry := metrics.NewRegistry()
	chaosCollector := metrics.NewChaosCollector(metricsRegistry)

	attackManager := attack.NewManagerWithMetrics(store.ExperimentRepo, chaosCollector)
	attackManager.Register(attack.NewPauseAttack(dockerController))
	attackManager.Register(attack.NewStopAttack(dockerController))
	attackManager.Register(attack.NewRestartAttack(dockerController))
	attackManager.Register(attack.NewKillAttack(dockerController))

	logger.Info("Creating recovery manager")
	recoveryManager := recovery.NewManagerWithMetrics(attackManager, store.ExperimentRepo, chaosCollector)

	collectInterval, err := time.ParseDuration(cfg.Scheduler.AttackInterval)
	if err != nil {
		collectInterval = 30 * time.Second
	}
	containerCollector := metrics.NewCollector(metricsRegistry, dockerController, store.ContainerRepo, collectInterval)

	logger.Info("Creating scheduler")
	sched := scheduler.NewSchedulerWithMetrics(
		cfg,
		dockerController,
		store.ContainerRepo,
		store.ExperimentRepo,
		attackManager,
		recoveryManager,
		chaosCollector,
		containerCollector,
	)

	metricsServer := server.NewMetricsServer(cfg.Metrics.Port)

	return &Dependencies{
		Config:             cfg,
		Store:              store,
		DockerController:   dockerController,
		DockerCloser:       dockerCloser,
		AttackManager:      attackManager,
		RecoveryManager:    recoveryManager,
		Scheduler:          sched,
		MetricsRegistry:    metricsRegistry,
		ChaosCollector:     chaosCollector,
		ContainerCollector: containerCollector,
		MetricsServer:      metricsServer,
	}, nil
}

func checkDatabaseWritable(dbPath string) error {
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}

	testFile := filepath.Join(dbDir, ".chaosguard_write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("write test failed: %w", err)
	}
	return os.Remove(testFile)
}

func checkPortAvailable(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	return ln.Close()
}
