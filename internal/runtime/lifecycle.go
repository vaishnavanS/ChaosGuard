package runtime

import "sync"

// State represents the current lifecycle phase of the ChaosGuard application.
type State int

const (
	StateStarting State = iota
	StateRunning
	StateStopping
	StateStopped
)

func (s State) String() string {
	switch s {
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// Lifecycle tracks and exposes the application runtime state for operators and future APIs.
type Lifecycle struct {
	mu    sync.RWMutex
	state State
}

// NewLifecycle creates a lifecycle manager in the stopped state.
func NewLifecycle() *Lifecycle {
	return &Lifecycle{state: StateStopped}
}

// SetState updates the current lifecycle state.
func (l *Lifecycle) SetState(state State) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.state = state
}

// State returns the current lifecycle state.
func (l *Lifecycle) State() State {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.state
}
