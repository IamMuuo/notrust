package state

import (
	"sync"

	"github.com/iammuuo/notrust/internal/docker"
)

// Registry holds one TrackedContainer per container id. Safe for
// concurrent use.
type Registry struct {
	mu sync.Mutex
	// id -> TrackedContainer key value
	containers map[string]*TrackedContainer
}

func NewRegistry() *Registry {
	return &Registry{containers: make(map[string]*TrackedContainer)}
}

// Sync reconciles against a fresh poll from Engine.List. New containers
// are added as ACTIVE (or whatever Docker currently reports), containers
// Docker no longer reports at all are dropped entirely.
func (r *Registry) Sync(snapshots []docker.ContainerSnapshot) {
	r.mu.Lock()
	defer r.mu.Unlock()

	seen := make(map[string]struct{}, len(snapshots))
	for _, snap := range snapshots {
		seen[snap.ID] = struct{}{}

		tc, ok := r.containers[snap.ID]
		if !ok {
			r.containers[snap.ID] = &TrackedContainer{
				Snapshot: snap,
				Status:   statusFromDocker(snap.State),
			}
			continue
		}
		tc.Snapshot = snap
	}

	for id := range r.containers {
		if _, ok := seen[id]; !ok {
			delete(r.containers, id)
		}
	}
}

// Snapshot returns a copy of the slice, safe to range over without
// holding the lock. Each element is still a pointer into the registry
// though, so a caller mutating tc.Status is intentional and persists.
func (r *Registry) Snapshot() []*TrackedContainer {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]*TrackedContainer, 0, len(r.containers))
	for _, tc := range r.containers {
		out = append(out, tc)
	}
	return out
}

func statusFromDocker(dockerState string) Status {
	switch dockerState {
	case "paused":
		return StatusPaused
	case "exited", "dead":
		return StatusStopped
	default:
		return StatusActive
	}
}
