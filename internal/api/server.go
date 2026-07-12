package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"chaosguard/internal/api/handlers"
	"chaosguard/pkg/logger"
)

// Server coordinates the lifecycle of the Gin HTTP API server
type Server struct {
	mu       sync.Mutex
	port     int
	server   *http.Server
	running  bool
	stopChan chan struct{}
}

// NewServer creates a new API Server instance
func NewServer(port int, h *handlers.Handler) *Server {
	router := SetupRouter(h)

	return &Server{
		port: port,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		stopChan: make(chan struct{}),
	}
}

// Start boots the HTTP server in the background
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("API server is already running")
	}
	s.running = true
	s.stopChan = make(chan struct{})
	s.mu.Unlock()

	logger.Info("Starting REST API server on port %d", s.port)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err, "API server error during execution")
		}
	}()

	return nil
}

// Stop shuts down the API server gracefully
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}
	s.running = false
	close(s.stopChan)

	logger.Info("Stopping REST API server gracefully")
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		logger.Error(err, "Failed to shutdown REST API server gracefully")
		return err
	}

	logger.Info("REST API server stopped successfully")
	return nil
}

// IsRunning reports if the server is active
func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
