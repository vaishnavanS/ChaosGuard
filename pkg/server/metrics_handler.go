package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"chaosguard/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsServer serves Prometheus metrics on a configurable port
type MetricsServer struct {
	mu       sync.Mutex
	port     int
	server   *http.Server
	running  bool
	stopChan chan struct{}
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(port int) *MetricsServer {
	return &MetricsServer{
		port:     port,
		stopChan: make(chan struct{}),
	}
}

// Start begins serving metrics on the configured port
func (ms *MetricsServer) Start(ctx context.Context) error {
	ms.mu.Lock()
	if ms.running {
		ms.mu.Unlock()
		return fmt.Errorf("metrics server is already running")
	}
	ms.running = true
	ms.stopChan = make(chan struct{})
	ms.mu.Unlock()

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", ms.handleHealth)

	ms.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", ms.port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("Starting metrics server on port %d", ms.port)

	go func() {
		if err := ms.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err, "Metrics server error")
		}
	}()

	return nil
}

// Stop halts the metrics server
func (ms *MetricsServer) Stop(ctx context.Context) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.running {
		return nil
	}

	ms.running = false
	close(ms.stopChan)

	if ms.server != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := ms.server.Shutdown(shutdownCtx); err != nil {
			logger.Error(err, "Error shutting down metrics server")
			return err
		}
	}

	logger.Info("Metrics server stopped")
	return nil
}

// IsRunning returns whether the metrics server is running
func (ms *MetricsServer) IsRunning() bool {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.running
}

// handleHealth is a simple health check endpoint
func (ms *MetricsServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy"}`)
}
