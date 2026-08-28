package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/iammuuo/notrust/internal/daemon"
)

func main() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	d := daemon.New(logger)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- d.Run(ctx)
	}()

	slog.Info("starting notrustd daemon...")

	select {
	case <-ctx.Done():
		slog.Info("received shutdown signal, waiting for daemon to stop")
	case err := <-done:
		if err != nil {
			slog.Error("daemon exited unexpectedly", "err", err)
			os.Exit(1)
		}
		slog.Info("goodbye cruel world!")
		return
	}

	if err := <-done; err != nil {
		slog.Error("daemon shutdown returned error", "err", err)
		os.Exit(1)
	}
	slog.Info("goodbye cruel world!")
}
