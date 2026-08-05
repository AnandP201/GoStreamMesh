package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/AnandP201/GoStreamMesh/internal/config"
	"github.com/AnandP201/GoStreamMesh/internal/logger"
	"github.com/AnandP201/GoStreamMesh/internal/server"
)

const serviceName = "ingestion-service"

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.Load(serviceName, ":8080")
	if err != nil {
		slog.Error("invalid configuration", slog.String("error", err.Error()))
		return 1
	}

	log := logger.New(cfg.Service, cfg.Environment, cfg.LogLevel)
	health := server.NewHealth(cfg.Service)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err := server.Run(ctx, log, cfg, health.Handler(log), func() {
		health.SetReady(false)
	}); err != nil {
		log.Error("service stopped with an error", slog.String("error", err.Error()))
		return 1
	}
	return 0
}
