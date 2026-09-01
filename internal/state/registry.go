package state

import (
	"log/slog"
	"sync"
	"time"

	"github.com/iammuuo/notrust/internal/docker"
)

type Registry struct {
	mu         sync.Mutex
	containers map[string]*TrackedContainer
	logger     *slog.Logger
}

func NewRegistry(logger *slog.Logger) *Registry {
	return &Registry{
		containers: make(map[string]*TrackedContainer),
		logger:     logger,
	}
}

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

		// Docker's real state is the source of truth. If it no longer
		// matches what we last recorded, something outside notrust
		// changed it: docker start, docker unpause, a compose restart,
		// a human. Reconcile and reset timers, otherwise a container
		// that comes back to life keeps a stale PausedAt/IdleSince from
		// its previous life and can get immediately re-paused or
		// re-stopped on the very next tick.
		if newStatus := statusFromDocker(snap.State); newStatus != tc.Status {
			if r.logger != nil {
				r.logger.Info("container state changed externally",
					"container", snap.Name, "from", tc.Status, "to", newStatus)
			}
			tc.Status = newStatus
			tc.IdleSince = time.Time{}
			tc.PausedAt = time.Time{}
		}
	}

	for id := range r.containers {
		if _, ok := seen[id]; !ok {
			delete(r.containers, id)
		}
	}
}

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
