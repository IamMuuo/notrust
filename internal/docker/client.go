package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type dockerEngine struct {
	cli *client.Client
}

func NewEngine() (Engine, error) {
	cli, err := client.New(
		client.FromEnv,
	)
	if err != nil {
		return nil, fmt.Errorf("creating docker client: %w", err)
	}
	return &dockerEngine{cli: cli}, nil
}

func (e *dockerEngine) List(ctx context.Context) ([]ContainerSnapshot, error) {
	// All: true — without it, docker only returns running containers,
	// and a paused/stopped container silently vanishes from view,
	// which breaks the PAUSED -> STOPPED escalation later.
	containers, err := e.cli.ContainerList(
		ctx,
		client.ContainerListOptions{All: true},
	)
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}

	snapshots := make([]ContainerSnapshot, 0, len(containers.Items))
	for _, c := range containers.Items {
		snapshots = append(snapshots, fromSummary(c))
	}
	return snapshots, nil
}

func (e *dockerEngine) Inspect(
	ctx context.Context,
	id string,
) (ContainerSnapshot, error) {
	info, err := e.cli.ContainerInspect(ctx, id, client.ContainerInspectOptions{
		Size: true,
	})
	if err != nil {
		return ContainerSnapshot{}, fmt.Errorf(
			"inspecting container %s: %w",
			id,
			err,
		)
	}

	return fromInspect(info.Container), nil
}

func (e *dockerEngine) Pause(ctx context.Context, id string) error {
	if _, err := e.cli.ContainerPause(
		ctx,
		id,
		client.ContainerPauseOptions{},
	); err != nil {
		return fmt.Errorf("pausing container %s: %w", id, err)
	}
	return nil
}

func (e *dockerEngine) Unpause(ctx context.Context, id string) error {
	if _, err := e.cli.ContainerUnpause(
		ctx,
		id,
		client.ContainerUnpauseOptions{},
	); err != nil {
		return fmt.Errorf("unpausing container %s: %w", id, err)
	}
	return nil
}

func (e *dockerEngine) Start(ctx context.Context, id string) error {
	if _, err := e.cli.ContainerStart(
		ctx,
		id,
		client.ContainerStartOptions{},
	); err != nil {
		return fmt.Errorf("starting container %s: %w", id, err)
	}
	return nil
}

func (e *dockerEngine) Stop(
	ctx context.Context,
	id string,
	timeout time.Duration,
) error {
	secs := int(timeout.Seconds())
	if _, err := e.cli.ContainerStop(
		ctx,
		id,
		client.ContainerStopOptions{Timeout: &secs},
	); err != nil {
		return fmt.Errorf("stopping container %s: %w", id, err)
	}
	return nil
}

// fromSummary maps the SDK's list-response type. Note this is a
// different shape than Inspect's response — see fromInspect below.
func fromSummary(c container.Summary) ContainerSnapshot {
	name := ""
	if len(c.Names) > 0 {
		name = strings.TrimPrefix(c.Names[0], "/")
	}

	ports := make([]PortBinding, 0, len(c.Ports))
	for _, p := range c.Ports {
		ports = append(ports, PortBinding{
			ContainerPort: int(p.PrivatePort),
			HostPort:      int(p.PublicPort),
			Protocol:      p.Type,
		})
	}

	return ContainerSnapshot{
		ID:     c.ID,
		Name:   name,
		Image:  c.Image,
		State:  string(c.State),
		Labels: c.Labels,
		Ports:  ports,
	}
}

// fromInspect maps the SDK's inspect-response type. Deliberately
// separate from fromSummary — ContainerList and ContainerInspect
// return two genuinely different SDK types with different field
// shapes (e.g. ports live in NetworkSettings.Ports here, not in a
// flat Ports slice), so a single shared mapper doesn't cleanly work.
func fromInspect(c container.InspectResponse) ContainerSnapshot {
	return ContainerSnapshot{
		ID:     c.ID,
		Name:   strings.TrimPrefix(c.Name, "/"),
		Image:  c.Config.Image,
		State:  string(c.State.Status),
		Labels: c.Config.Labels,
		// Ports omitted here for brevity — map from
		// c.NetworkSettings.Ports (a nat.PortMap) the same way,
		// worth its own small helper once wake proxy needs it.
	}
}
