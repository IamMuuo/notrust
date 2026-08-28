package daemon

import (
	"context"
	"log/slog"
	"time"

	"github.com/iammuuo/notrust/internal/config"
)

type Daemon struct {
	Logger *slog.Logger
	Cfg    *config.Config
}

func New(logger *slog.Logger, config *config.Config) *Daemon {
	return &Daemon{
		Logger: logger,
		Cfg:    config,
	}
}

func (d *Daemon) Run(ctx context.Context) error {
	ticker := time.NewTicker(d.Cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.Logger.Debug("polling docker containers.")
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
