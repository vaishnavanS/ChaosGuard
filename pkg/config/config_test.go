package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.SafeMode {
		t.Error("expected SafeMode to be true by default")
	}
	if cfg.Dashboard.Port != 8080 {
		t.Errorf("expected dashboard port 8080, got %d", cfg.Dashboard.Port)
	}
	if cfg.Metrics.Port != 2112 {
		t.Errorf("expected metrics port 2112, got %d", cfg.Metrics.Port)
	}
	if cfg.Scheduler.Mode != "random" {
		t.Errorf("expected scheduler mode 'random', got %s", cfg.Scheduler.Mode)
	}
}

func TestWriteDefaultAndLoad(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "chaosguard-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()
	os.Remove(tmpFile.Name()) // remove so WriteDefault can write it

	err = WriteDefault(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to write default config: %v", err)
	}

	cfg, err := Load(tmpFile.Name(), nil)
	if err != nil {
		t.Fatalf("failed to load written config: %v", err)
	}

	if !cfg.SafeMode {
		t.Error("expected SafeMode to be true")
	}
	if cfg.Scheduler.AttackInterval != "30s" {
		t.Errorf("expected attack interval '30s', got %s", cfg.Scheduler.AttackInterval)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
	}{
		{
			name:    "valid defaults",
			modify:  func(c *Config) {},
			wantErr: false,
		},
		{
			name: "invalid dashboard port",
			modify: func(c *Config) {
				c.Dashboard.Port = 70000
			},
			wantErr: true,
		},
		{
			name: "invalid scheduler mode",
			modify: func(c *Config) {
				c.Scheduler.Mode = "invalid_mode"
			},
			wantErr: true,
		},
		{
			name: "invalid attack interval duration",
			modify: func(c *Config) {
				c.Scheduler.AttackInterval = "30x"
			},
			wantErr: true,
		},
		{
			name: "invalid attack duration duration",
			modify: func(c *Config) {
				c.Scheduler.AttackDuration = "10x"
			},
			wantErr: true,
		},
		{
			name: "attack duration longer than interval",
			modify: func(c *Config) {
				c.Scheduler.AttackInterval = "10s"
				c.Scheduler.AttackDuration = "20s"
			},
			wantErr: true,
		},
		{
			name: "too short attack interval",
			modify: func(c *Config) {
				c.Scheduler.AttackInterval = "500ms"
				c.Scheduler.AttackDuration = "100ms"
			},
			wantErr: true,
		},
		{
			name: "empty db path",
			modify: func(c *Config) {
				c.Database.Path = ""
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modify(cfg)
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnvironmentOverride(t *testing.T) {
	os.Setenv("CHAOSGUARD_SAFE_MODE", "false")
	os.Setenv("CHAOSGUARD_DASHBOARD_PORT", "9090")
	defer func() {
		os.Unsetenv("CHAOSGUARD_SAFE_MODE")
		os.Unsetenv("CHAOSGUARD_DASHBOARD_PORT")
	}()

	cfg, err := Load("", nil)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	if cfg.SafeMode {
		t.Error("expected SafeMode override to be false")
	}
	if cfg.Dashboard.Port != 9090 {
		t.Errorf("expected dashboard port override to be 9090, got %d", cfg.Dashboard.Port)
	}
}
