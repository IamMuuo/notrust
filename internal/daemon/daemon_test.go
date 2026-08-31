package daemon

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/iammuuo/notrust/internal/config"
	"github.com/iammuuo/notrust/internal/docker"
	"github.com/iammuuo/notrust/internal/state"
)

func TestRun_StopsOnContextCancel(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(
		slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	engine, err := docker.NewEngine()
	if err != nil {
		t.Fatalf("docker engine creation errored: %s", err.Error())
	}

	d := New(
		logger,
		&config.Config{PollInterval: 5 * time.Millisecond},
		engine,
		state.NeverIdle{},
	)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() { errCh <- d.Run(ctx) }()

	time.Sleep(30 * time.Millisecond) // let a couple of ticks happen
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal(
			"Run did not return after context cancel — possible goroutine leak",
		)
	}

	out := buf.String()
	t.Log(out)
	if !strings.Contains(out, "stopping poller") {
		t.Error("expected shutdown log line")
	}
}
