package state

import (
	"context"
	"log/slog"
	"time"

	"github.com/iammuuo/notrust/internal/docker"
)

// IdleChecker decides whether a container currently looks idle
type IdleChecker interface {
	IsIdle(ctx context.Context, c docker.ContainerSnapshot) (bool, error)
}

type Thresholds struct {
	PauseAfter time.Duration
	StopAfter  time.Duration
}

// Machine applies ACTIVE -> PAUSED -> STOPPED transitions across
// whatever the Registry currently holds.
type Machine struct {
	Registry   *Registry
	Engine     docker.Engine
	Idle       IdleChecker
	Thresholds Thresholds
	Logger     *slog.Logger
}

// Evaluate walks every tracked container once, applying at most one
// transition per container per call.
func (m *Machine) Evaluate(ctx context.Context) {
	now := time.Now()
	for _, tc := range m.Registry.Snapshot() {
		switch tc.Status {
		case StatusActive:
			m.evaluateActive(ctx, tc, now)
		case StatusPaused:
			m.evaluatePaused(ctx, tc, now)
		case StatusStopped:
			// terminal until an external wake, nothing to evaluate
		}
	}
}

func (m *Machine) evaluateActive(
	ctx context.Context,
	tc *TrackedContainer,
	now time.Time,
) {
	idle, err := m.Idle.IsIdle(ctx, tc.Snapshot)
	if err != nil {
		m.Logger.Error(
			"idle check failed",
			"container",
			tc.Snapshot.Name,
			"err",
			err,
		)
		return
	}
	if !idle {
		tc.IdleSince = time.Time{} // reset, it's active again
		return
	}
	if tc.IdleSince.IsZero() {
		tc.IdleSince = now // first tick we've seen it idle
		return
	}
	if now.Sub(tc.IdleSince) < m.Thresholds.PauseAfter {
		return // idle, but not long enough yet
	}

	if err := m.Engine.Pause(ctx, tc.Snapshot.ID); err != nil {
		m.Logger.Error(
			"pause failed",
			"container",
			tc.Snapshot.Name,
			"err",
			err,
		)
		return
	}
	m.Logger.Info("paused container", "container", tc.Snapshot.Name)
	tc.Status = StatusPaused
	tc.PausedAt = now
}

// Transitions a container from paused -> exited
func (m *Machine) evaluatePaused(
	ctx context.Context,
	tc *TrackedContainer,
	now time.Time,
) {
	if now.Sub(tc.PausedAt) < m.Thresholds.StopAfter {
		return
	}
	// TODO: (erick) find out whether this works fine or allow the user to configure
	if err := m.Engine.Stop(ctx, tc.Snapshot.ID, 10*time.Second); err != nil {
		m.Logger.Error("stop failed", "container", tc.Snapshot.Name, "err", err)
		return
	}
	m.Logger.Info("stopped container", "container", tc.Snapshot.Name)
	tc.Status = StatusStopped
}
