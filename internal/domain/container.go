package domain

import "time"

// Container represents a target container managed by ChaosGuard.
type Container struct {
	ID          string    `json:"id"`           // Docker container ID
	Name        string    `json:"name"`         // Container name
	Image       string    `json:"image"`        // Docker image name
	Status      string    `json:"status"`       // Container status (e.g. running, paused, stopped)
	IsMonitored bool      `json:"is_monitored"` // Flag indicating if container is monitored/targetable
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ContainerRepository defines the data storage operations for Container entities.
type ContainerRepository interface {
	Get(id string) (*Container, error)
	List() ([]*Container, error)
	Save(container *Container) error
	Create(container *Container) error
	Update(container *Container) error
	Delete(id string) error
	UpdateState(id string, state string) error
}

// ContainerController defines runtime operations to control and inspect Docker containers.
type ContainerController interface {
	Discover() ([]*Container, error)
	Pause(id string) error
	Unpause(id string) error
	Stop(id string) error
	Start(id string) error
	Restart(id string) error
	Kill(id string) error
	Inspect(id string) (*Container, error)
}
