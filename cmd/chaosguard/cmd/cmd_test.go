package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"chaosguard/pkg/config"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs(args)

	err := RootCmd.Execute()
	return buf.String(), err
}

func TestCLICommands(t *testing.T) {
	// Create a mock daemon server to handle CLI client requests during testing
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/runtime":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true,"data":{"state":"running"}}`))
		case "/runtime/stop":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true,"data":"Shutdown initiated"}`))
		case "/scheduler/status":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true,"data":{"running":true,"mode":"random"}}`))
		case "/containers":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true,"data":[{"id":"c1","name":"web-server","status":"running","is_monitored":true}]}`))
		case "/experiments":
			if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"success":true,"data":{"id":"exp1","status":"pending"}}`))
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"success":true,"data":[{"id":"exp1","target_container_id":"c1","status":"completed"}]}`))
			}
		case "/logs":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true,"data":["test log line 1","test log line 2"]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"success":false,"error":"not found"}`))
		}
	}))
	defer server.Close()

	// Parse the port from the mock server URL
	var port int
	_, err := fmt.Sscanf(server.URL, "http://127.0.0.1:%d", &port)
	if err != nil {
		_, err = fmt.Sscanf(server.URL, "http://localhost:%d", &port)
		if err != nil {
			t.Fatalf("failed to parse mock server port: %v", err)
		}
	}

	// Create temporary config file with mock server port
	tmpFile, err := os.CreateTemp("", "chaosguard-test-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	configContent := fmt.Sprintf(`dashboard:
  port: %d
database:
  path: ./chaosguard.db
`, port)

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("failed to write temp config content: %v", err)
	}
	_ = tmpFile.Close()

	// Override package variables to force CLI commands to load this test config
	cfgFile = tmpFile.Name()
	ActiveConfig = config.DefaultConfig()
	ActiveConfig.Dashboard.Port = port

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "version command",
			args:    []string{"version"},
			wantErr: false,
		},
		{
			name:    "doctor command",
			args:    []string{"doctor"},
			wantErr: false,
		},
		{
			name:    "config command",
			args:    []string{"config"},
			wantErr: false,
		},
		{
			name:    "status command",
			args:    []string{"status"},
			wantErr: false,
		},
		{
			name:    "stop command",
			args:    []string{"stop"},
			wantErr: false,
		},
		{
			name:    "dashboard command",
			args:    []string{"dashboard", "--open=false"},
			wantErr: false,
		},
		{
			name:    "report command",
			args:    []string{"report", "-f", "json"},
			wantErr: false,
		},
		{
			name:    "attack command missing target and type",
			args:    []string{"attack"},
			wantErr: true,
		},
		{
			name:    "attack command success simulation",
			args:    []string{"attack", "-t", "web-server", "-a", "pause"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executeCommand(tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("executeCommand(%v) returned error %v, wantErr %v", tt.args, err, tt.wantErr)
			}
		})
	}
}
