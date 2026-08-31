package daemon

import (
	"context"
	"log/slog"
	"time"

	"github.com/iammuuo/notrust/internal/config"
	"github.com/iammuuo/notrust/internal/docker"
	"github.com/iammuuo/notrust/internal/state"
)

type Daemon struct {
	Logger   *slog.Logger
	Cfg      *config.Config
	Engine   docker.Engine
	Registry *state.Registry
	Machine  *state.Machine
}

func New(
	logger *slog.Logger,
	cfg *config.Config,
	engine docker.Engine,
	idle state.IdleChecker,
) *Daemon {
	registry := state.NewRegistry()
	machine := &state.Machine{
		Registry: registry,
		Engine:   engine,
		Idle:     idle,
		Thresholds: state.Thresholds{
			PauseAfter: cfg.Idle.PauseAfter,
			StopAfter:  cfg.Idle.StopAfter,
		},
		Logger: logger,
	}
	return &Daemon{
		Logger:   logger,
		Cfg:      cfg,
		Engine:   engine,
		Registry: registry,
		Machine:  machine,
	}
}

func (d *Daemon) Run(ctx context.Context) error {
	ticker := time.NewTicker(d.Cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			containers, err := d.Engine.List(ctx)
			if err != nil {
				d.Logger.Error("listing containers", "err", err)
				continue
			}
			d.Registry.Sync(containers)
			d.Machine.Evaluate(ctx)
		case <-ctx.Done():
			d.Logger.Info("stopping poller...")
			d.shutDown()
			return nil
		}
	}
}

func (d *Daemon) shutDown() {
	d.Logger.Debug("shutting down notrustd")
}
