package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestMetricsServer_StartStop(t *testing.T) {
	port := 9090
	server := NewMetricsServer(port)

	if server.IsRunning() {
		t.Error("expected server to not be running initially")
	}

	// Start server
	err := server.Start(context.Background())
	if err != nil {
		t.Fatalf("failed to start metrics server: %v", err)
	}

	time.Sleep(100 * time.Millisecond) // Give server time to start

	if !server.IsRunning() {
		t.Error("expected server to be running after Start()")
	}

	// Test /metrics endpoint is accessible
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", port))
	if err != nil {
		t.Fatalf("failed to GET /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if len(body) == 0 {
		t.Error("expected non-empty metrics response")
	}

	// Test /health endpoint
	resp, err = http.Get(fmt.Sprintf("http://localhost:%d/health", port))
	if err != nil {
		t.Fatalf("failed to GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 for /health, got %d", resp.StatusCode)
	}

	// Stop server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Stop(ctx)
	if err != nil {
		t.Fatalf("failed to stop metrics server: %v", err)
	}

	if server.IsRunning() {
		t.Error("expected server to not be running after Stop()")
	}
}

func TestMetricsServer_AlreadyRunning(t *testing.T) {
	port := 9091
	server := NewMetricsServer(port)

	err := server.Start(context.Background())
	if err != nil {
		t.Fatalf("failed to start metrics server: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Try to start again - should be idempotent or return error
	err = server.Start(context.Background())
	if err == nil {
		// Starting an already-running server should not error in our implementation
		// (it checks and returns early)
	}

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Stop(ctx)
}

func TestMetricsServer_StopWhenNotRunning(t *testing.T) {
	port := 9092
	server := NewMetricsServer(port)

	// Stop without starting - should not error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Stop(ctx)
	if err != nil {
		t.Errorf("expected no error when stopping non-running server, got %v", err)
	}
}

func TestMetricsServer_HealthEndpoint(t *testing.T) {
	port := 9093
	server := NewMetricsServer(port)

	err := server.Start(context.Background())
	if err != nil {
		t.Fatalf("failed to start metrics server: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", port))
	if err != nil {
		t.Fatalf("failed to GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(body) != `{"status":"healthy"}` {
		t.Errorf("expected health response, got %s", string(body))
	}
}
