package docker

import (
	"context"
	"time"
)

// ContainerSnapshot is decoupled from the Docker SDK's
// types. Only fields notrust actually cares about. A Docker API bump
// only ever touches this package, nothing downstream.
type ContainerSnapshot struct {
	ID     string
	Name   string
	Image  string
	State  string // "running", "paused", "exited", ...
	Labels map[string]string
	Ports  []PortBinding
}

type PortBinding struct {
	ContainerPort int
	HostPort      int
	Protocol      string // "tcp" or "udp"
}

type StatsSample struct {
	CPUPercent float64
	NetRxBytes uint64
	NetTxBytes uint64
}

// Engine is the narrow surface notrust needs from Docker. internal/state
// and internal/metrics depend on this interface, never on the SDK
// directly, so they can be tested against docker/fake.Engine without a
// real daemon.
type Engine interface {
	List(ctx context.Context) ([]ContainerSnapshot, error)
	Inspect(ctx context.Context, id string) (ContainerSnapshot, error)

	Pause(ctx context.Context, id string) error
	Unpause(ctx context.Context, id string) error
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string, timeout time.Duration) error
	Stats(ctx context.Context, id string) (StatsSample, error)
}
