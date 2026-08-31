package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/iammuuo/notrust/internal/config"
	"github.com/iammuuo/notrust/internal/docker"
)

type Daemon struct {
	Logger *slog.Logger
	Cfg    *config.Config
	Engine docker.Engine
}

func New(logger *slog.Logger, config *config.Config) *Daemon {
	engine, err := docker.NewEngine()
	if err != nil {
		panic(err)
	}
	return &Daemon{
		Logger: logger,
		Cfg:    config,
		Engine: engine,
	}
}

func (d *Daemon) Run(ctx context.Context) error {
	ticker := time.NewTicker(d.Cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			containers, _ := d.Engine.List(ctx)
			fmt.Printf("%30s %30s\n", "Name", "State")
			for _, container := range containers {
				fmt.Printf("%30s %30s\n", container.Name, container.State)
			}
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
