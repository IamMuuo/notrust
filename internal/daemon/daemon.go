package daemon

import (
	"context"
	"log/slog"
	"time"
)

type Daemon struct {
	Logger   *slog.Logger
	Interval time.Duration
}

func New(logger *slog.Logger) *Daemon {
	return &Daemon{
		Logger:   logger,
		Interval: 3 * time.Second,
	}
}

func (d *Daemon) Run(ctx context.Context) error {
	ticker := time.NewTicker(d.Interval)
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
