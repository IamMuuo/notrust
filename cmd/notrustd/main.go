package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/iammuuo/notrust/internal/config"
	"github.com/iammuuo/notrust/internal/daemon"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	logLevelFlag := flag.String(
		"log-level",
		"info",
		"log level: debug, info, warn, error",
	)
	flag.Parse()

	var level slog.Level
	if err := level.UnmarshalText([]byte(*logLevelFlag)); err != nil {
		fmt.Fprintf(os.Stderr, "invalid log level %q: %v\n", *logLevelFlag, err)
		os.Exit(1)
	}

	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}),
	)
	slog.SetDefault(logger)

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	d := daemon.New(logger, cfg)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- d.Run(ctx) }()

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
