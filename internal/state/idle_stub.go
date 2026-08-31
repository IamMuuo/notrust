package state

import (
	"context"

	"github.com/iammuuo/notrust/internal/docker"
)

// NeverIdle is a placeholder IdleChecker that always reports active.
// Swap for the real metrics-based checker once internal/metrics exists.
// Until then this keeps the machine safe to run nothing gets paused
// on a guess.
type NeverIdle struct{}

func (NeverIdle) IsIdle(
	ctx context.Context,
	c docker.ContainerSnapshot,
) (bool, error) {
	return false, nil
}
