package state

import (
	"time"

	"github.com/iammuuo/notrust/internal/docker"
)

type Status string

const (
	StatusActive  Status = "active"
	StatusPaused  Status = "paused"
	StatusStopped Status = "stopped"
)

// TrackedContainer pairs a docker snapshot with notrust's own view of
// where that container sits in the ACTIVE -> PAUSED -> STOPPED lifecycle.
type TrackedContainer struct {
	Snapshot docker.ContainerSnapshot
	Status   Status

	// IdleSince is when we first observed this container idle while
	// ACTIVE. Zero value means "not currently idle".
	IdleSince time.Time

	// PausedAt is when we transitioned to PAUSED, used to evaluate
	// the PAUSED -> STOPPED escalation.
	PausedAt time.Time
}
