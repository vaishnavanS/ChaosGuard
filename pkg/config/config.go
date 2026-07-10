package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// DefaultConfigName is the default name for the configuration file
const DefaultConfigName = "chaosguard.yaml"

// Config holds all configuration parameters for ChaosGuard
type Config struct {
	SafeMode   bool             `mapstructure:"safe_mode" json:"safe_mode" yaml:"safe_mode"`
	Dashboard  DashboardConfig  `mapstructure:"dashboard" json:"dashboard" yaml:"dashboard"`
	Metrics    MetricsConfig    `mapstructure:"metrics" json:"metrics" yaml:"metrics"`
	Scheduler  SchedulerConfig  `mapstructure:"scheduler" json:"scheduler" yaml:"scheduler"`
	Containers ContainerFilter  `mapstructure:"containers" json:"containers" yaml:"containers"`
	Database   DatabaseConfig   `mapstructure:"database" json:"database" yaml:"database"`
}

type DashboardConfig struct {
	Port int `mapstructure:"port" json:"port" yaml:"port"`
}

type MetricsConfig struct {
	Port int `mapstructure:"port" json:"port" yaml:"port"`
}

type SchedulerConfig struct {
	Mode           string `mapstructure:"mode" json:"mode" yaml:"mode"`
	AttackInterval string `mapstructure:"attack_interval" json:"attack_interval" yaml:"attack_interval"`
	AttackDuration string `mapstructure:"attack_duration" json:"attack_duration" yaml:"attack_duration"`
}

type ContainerFilter struct {
	Include []string `mapstructure:"include" json:"include" yaml:"include"`
	Exclude []string `mapstructure:"exclude" json:"exclude" yaml:"exclude"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path" json:"path" yaml:"path"`
}

// DefaultConfig returns a configuration struct with pre-populated defaults
func DefaultConfig() *Config {
	return &Config{
		SafeMode: true,
		Dashboard: DashboardConfig{
			Port: 8080,
		},
		Metrics: MetricsConfig{
			Port: 2112,
		},
		Scheduler: SchedulerConfig{
			Mode:           "random",
			AttackInterval: "30s",
			AttackDuration: "10s",
		},
		Containers: ContainerFilter{
			Include: []string{},
			Exclude: []string{"postgres", "mysql", "redis", "mongodb", "prometheus", "grafana", "chaosguard"},
		},
		Database: DatabaseConfig{
			Path: "./chaosguard.db",
		},
	}
}

// WriteDefault writes the default configuration file to the specified path
func WriteDefault(path string) error {
	if path == "" {
		path = DefaultConfigName
	}

	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists at %s", path)
	}

	defaultContent := `safe_mode: true

dashboard:
  port: 8080

metrics:
  port: 2112

scheduler:
  mode: "random"
  attack_interval: "30s"
  attack_duration: "10s"

containers:
  include: []
  exclude:
    - "postgres"
    - "mysql"
    - "redis"
    - "mongodb"
    - "prometheus"
    - "grafana"
    - "chaosguard"

database:
  path: "./chaosguard.db"
`
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(path, []byte(defaultContent), 0644)
}

// Load loads the configuration from a file and overrides with environment variables
func Load(cfgFile string, onReload func(*Config)) (*Config, error) {
	v := viper.New()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.AddConfigPath(".")
		v.SetConfigName("chaosguard")
		v.SetConfigType("yaml")
	}

	// Environment variables
	v.SetEnvPrefix("CHAOSGUARD")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Load defaults into viper
	defaults := DefaultConfig()
	v.SetDefault("safe_mode", defaults.SafeMode)
	v.SetDefault("dashboard.port", defaults.Dashboard.Port)
	v.SetDefault("metrics.port", defaults.Metrics.Port)
	v.SetDefault("scheduler.mode", defaults.Scheduler.Mode)
	v.SetDefault("scheduler.attack_interval", defaults.Scheduler.AttackInterval)
	v.SetDefault("scheduler.attack_duration", defaults.Scheduler.AttackDuration)
	v.SetDefault("containers.include", defaults.Containers.Include)
	v.SetDefault("containers.exclude", defaults.Containers.Exclude)
	v.SetDefault("database.path", defaults.Database.Path)

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) && cfgFile != "" {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Hot reload
	if onReload != nil && v.ConfigFileUsed() != "" {
		v.OnConfigChange(func(e fsnotify.Event) {
			var newConfig Config
			if err := v.Unmarshal(&newConfig); err == nil {
				if err := newConfig.Validate(); err == nil {
					onReload(&newConfig)
				}
			}
		})
		v.WatchConfig()
	}

	return &config, nil
}

// Validate checks the configuration for semantic correctness
func (c *Config) Validate() error {
	if c.Dashboard.Port <= 0 || c.Dashboard.Port > 65535 {
		return fmt.Errorf("dashboard port %d is out of range [1-65535]", c.Dashboard.Port)
	}
	if c.Metrics.Port <= 0 || c.Metrics.Port > 65535 {
		return fmt.Errorf("metrics port %d is out of range [1-65535]", c.Metrics.Port)
	}

	switch c.Scheduler.Mode {
	case "random", "round-robin", "sequential", "manual":
		// valid
	default:
		return fmt.Errorf("invalid scheduler mode: %s (must be random, round-robin, sequential, or manual)", c.Scheduler.Mode)
	}

	interval, err := time.ParseDuration(c.Scheduler.AttackInterval)
	if err != nil {
		return fmt.Errorf("invalid attack_interval: %w", err)
	}
	if interval < 1*time.Second {
		return fmt.Errorf("attack_interval must be at least 1s, got %s", c.Scheduler.AttackInterval)
	}

	duration, err := time.ParseDuration(c.Scheduler.AttackDuration)
	if err != nil {
		return fmt.Errorf("invalid attack_duration: %w", err)
	}
	if duration < 1*time.Second {
		return fmt.Errorf("attack_duration must be at least 1s, got %s", c.Scheduler.AttackDuration)
	}

	if duration >= interval {
		return fmt.Errorf("attack_duration (%s) must be shorter than attack_interval (%s)", c.Scheduler.AttackDuration, c.Scheduler.AttackInterval)
	}

	if c.Database.Path == "" {
		return errors.New("database path cannot be empty")
	}

	return nil
}
