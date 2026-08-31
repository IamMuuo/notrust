package metrics

import (
	"context"
	"fmt"
	"sync"

	"github.com/iammuuo/notrust/internal/docker"
)

type Thresholds struct {
	CPUPercent float64 // e.g. 0.5
	NetBytes   uint64  // e.g. 1024
}

// Collector implements state.IdleChecker. Network stats from Docker are
// cumulative since container start, so we track the last observed
// total ourselves to compute a per tick delta.
type Collector struct {
	Engine     docker.Engine
	Thresholds Thresholds
	mu         sync.Mutex
	lastNet    map[string]uint64 // container id -> last cumulative rx+tx
}

func NewCollector(engine docker.Engine, t Thresholds) *Collector {
	return &Collector{
		Engine:     engine,
		Thresholds: t,
		lastNet:    make(map[string]uint64),
	}
}

func (c *Collector) IsIdle(
	ctx context.Context,
	snap docker.ContainerSnapshot,
) (bool, error) {
	stats, err := c.Engine.Stats(ctx, snap.ID)
	if err != nil {
		return false, fmt.Errorf("stats for %s: %w", snap.Name, err)
	}

	if stats.CPUPercent >= c.Thresholds.CPUPercent {
		return false, nil
	}

	total := stats.NetRxBytes + stats.NetTxBytes
	c.mu.Lock()
	prev, seen := c.lastNet[snap.ID]
	c.lastNet[snap.ID] = total
	c.mu.Unlock()

	if !seen {
		return false, nil // no baseline yet, don't guess on the first look
	}
	var delta uint64
	if total > prev {
		delta = total - prev
	}
	if delta >= c.Thresholds.NetBytes {
		return false, nil
	}

	established, err := hasEstablishedConnections(snap.Ports)
	if err != nil {
		return false, fmt.Errorf(
			"checking connections for %s: %w",
			snap.Name,
			err,
		)
	}
	return !established, nil
}
