package docker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/moby/moby/client"
)

type statsJSON struct {
	CPUStats    cpuStats                `json:"cpu_stats"`
	PreCPUStats cpuStats                `json:"precpu_stats"`
	Networks    map[string]networkStats `json:"networks"`
}

type cpuStats struct {
	CPUUsage struct {
		TotalUsage uint64 `json:"total_usage"`
	} `json:"cpu_usage"`
	SystemUsage uint64 `json:"system_cpu_usage"`
	OnlineCPUs  uint32 `json:"online_cpus"`
}

type networkStats struct {
	RxBytes uint64 `json:"rx_bytes"`
	TxBytes uint64 `json:"tx_bytes"`
}

func (e *dockerEngine) Stats(
	ctx context.Context,
	id string,
) (StatsSample, error) {
	resp, err := e.cli.ContainerStats(
		ctx,
		id,
		client.ContainerStatsOptions{Stream: false},
	) // false = one snapshot, not a stream
	if err != nil {
		return StatsSample{}, fmt.Errorf("fetching stats for %s: %w", id, err)
	}
	defer resp.Body.Close()

	var raw statsJSON
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return StatsSample{}, fmt.Errorf("decoding stats for %s: %w", id, err)
	}

	var rx, tx uint64
	for _, n := range raw.Networks {
		rx += n.RxBytes
		tx += n.TxBytes
	}

	return StatsSample{
		CPUPercent: cpuPercent(raw),
		NetRxBytes: rx,
		NetTxBytes: tx,
	}, nil
}

// cpuPercent uses the same formula the docker CLI itself uses for
// `docker stats`. A non-streaming call still populates both cpu_stats
// and precpu_stats from one internal sample, so this works off a
// single request.
func cpuPercent(s statsJSON) float64 {
	cpuDelta := float64(
		s.CPUStats.CPUUsage.TotalUsage,
	) - float64(
		s.PreCPUStats.CPUUsage.TotalUsage,
	)
	systemDelta := float64(
		s.CPUStats.SystemUsage,
	) - float64(
		s.PreCPUStats.SystemUsage,
	)
	if cpuDelta <= 0 || systemDelta <= 0 {
		return 0
	}
	cpus := float64(s.CPUStats.OnlineCPUs)
	if cpus == 0 {
		cpus = 1
	}
	return (cpuDelta / systemDelta) * cpus * 100.0
}
