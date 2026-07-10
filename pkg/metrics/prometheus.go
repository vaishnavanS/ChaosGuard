package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Namespace for all ChaosGuard metrics
const namespace = "chaosguard"

// Registry holds all Prometheus metrics for ChaosGuard
type Registry struct {
	// Docker metrics
	ContainerCPUUsage    prometheus.Gauge
	ContainerMemoryUsage prometheus.Gauge
	ContainerNetworkIn   prometheus.Gauge
	ContainerNetworkOut  prometheus.Gauge

	// Container state metrics
	ContainersRunning prometheus.Gauge
	ContainersPaused  prometheus.Gauge
	ContainersStopped prometheus.Gauge

	// Chaos experiment metrics
	ExperimentsTotal       prometheus.Counter
	ExperimentsRunning     prometheus.Gauge
	ExperimentsCompleted   prometheus.Counter
	ExperimentsFailed      prometheus.Counter
	ExperimentsRecovered   prometheus.Counter
	ExperimentDurationMs   prometheus.Histogram

	// Attack metrics
	AttacksExecuted prometheus.Counter
	AttacksFailed   prometheus.Counter

	// Recovery metrics
	RecoveriesExecuted prometheus.Counter
	RecoveriesFailed   prometheus.Counter

	// Health metrics
	SchedulerRunning prometheus.Gauge
	LastExperimentAt prometheus.Gauge
}

// Global default registry (uses default Prometheus registry)
var globalRegistry *Registry

// NewRegistry creates a new metrics registry with all Prometheus metrics
// Uses the default global Prometheus registry
func NewRegistry() *Registry {
	if globalRegistry != nil {
		return globalRegistry
	}

	globalRegistry = &Registry{
		// Docker metrics (per container)
		ContainerCPUUsage: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "container",
			Name:      "cpu_usage_percent",
			Help:      "CPU usage percentage for containers",
		}),

		ContainerMemoryUsage: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "container",
			Name:      "memory_usage_bytes",
			Help:      "Memory usage in bytes for containers",
		}),

		ContainerNetworkIn: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "container",
			Name:      "network_in_bytes",
			Help:      "Network bytes received by containers",
		}),

		ContainerNetworkOut: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "container",
			Name:      "network_out_bytes",
			Help:      "Network bytes transmitted by containers",
		}),

		// Container state metrics
		ContainersRunning: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "container",
			Name:      "running_total",
			Help:      "Total number of running containers",
		}),

		ContainersPaused: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "container",
			Name:      "paused_total",
			Help:      "Total number of paused containers",
		}),

		ContainersStopped: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "container",
			Name:      "stopped_total",
			Help:      "Total number of stopped containers",
		}),

		// Chaos experiment metrics
		ExperimentsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "experiment",
			Name:      "total",
			Help:      "Total number of experiments executed",
		}),

		ExperimentsRunning: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "experiment",
			Name:      "running",
			Help:      "Number of currently running experiments",
		}),

		ExperimentsCompleted: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "experiment",
			Name:      "completed_total",
			Help:      "Total number of completed experiments",
		}),

		ExperimentsFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "experiment",
			Name:      "failed_total",
			Help:      "Total number of failed experiments",
		}),

		ExperimentsRecovered: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "experiment",
			Name:      "recovered_total",
			Help:      "Total number of recovered experiments",
		}),

		ExperimentDurationMs: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "experiment",
			Name:      "duration_ms",
			Help:      "Experiment duration in milliseconds",
			Buckets:   prometheus.ExponentialBuckets(100, 2, 10), // 100ms to ~100s
		}),

		// Attack metrics
		AttacksExecuted: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "attack",
			Name:      "executed_total",
			Help:      "Total number of attacks executed",
		}),

		AttacksFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "attack",
			Name:      "failed_total",
			Help:      "Total number of failed attacks",
		}),

		// Recovery metrics
		RecoveriesExecuted: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "recovery",
			Name:      "executed_total",
			Help:      "Total number of recoveries executed",
		}),

		RecoveriesFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "recovery",
			Name:      "failed_total",
			Help:      "Total number of failed recoveries",
		}),

		// Health metrics
		SchedulerRunning: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "scheduler",
			Name:      "running",
			Help:      "Scheduler running status (1 = running, 0 = stopped)",
		}),

		LastExperimentAt: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "scheduler",
			Name:      "last_experiment_timestamp",
			Help:      "Unix timestamp of the last experiment execution",
		}),
	}

	// Register all metrics
	prometheus.MustRegister(
		globalRegistry.ContainerCPUUsage,
		globalRegistry.ContainerMemoryUsage,
		globalRegistry.ContainerNetworkIn,
		globalRegistry.ContainerNetworkOut,
		globalRegistry.ContainersRunning,
		globalRegistry.ContainersPaused,
		globalRegistry.ContainersStopped,
		globalRegistry.ExperimentsTotal,
		globalRegistry.ExperimentsRunning,
		globalRegistry.ExperimentsCompleted,
		globalRegistry.ExperimentsFailed,
		globalRegistry.ExperimentsRecovered,
		globalRegistry.ExperimentDurationMs,
		globalRegistry.AttacksExecuted,
		globalRegistry.AttacksFailed,
		globalRegistry.RecoveriesExecuted,
		globalRegistry.RecoveriesFailed,
		globalRegistry.SchedulerRunning,
		globalRegistry.LastExperimentAt,
	)

	return globalRegistry
}

// ResetRegistry clears the global registry (mainly for testing)
func ResetRegistry() {
	globalRegistry = nil
}
