package server

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/AnandP201/GoStreamMesh/internal/config"
)

func TestRunGracefullyStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Config{
		HTTP: config.HTTPConfig{
			Address:           "127.0.0.1:0",
			ReadHeaderTimeout: time.Second,
			ReadTimeout:       time.Second,
			WriteTimeout:      time.Second,
			IdleTimeout:       time.Second,
		},
		ShutdownTimeout: time.Second,
	}

	shutdownStarted := make(chan struct{})
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := Run(ctx, logger, cfg, NewHealth("test").Handler(logger), func() {
		close(shutdownStarted)
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	select {
	case <-shutdownStarted:
	default:
		t.Fatal("beforeShutdown was not called")
	}
}
