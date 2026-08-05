package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/AnandP201/GoStreamMesh/internal/config"
)

// Run serves HTTP until the context is canceled or the listener fails.
func Run(
	ctx context.Context,
	logger *slog.Logger,
	cfg config.Config,
	handler http.Handler,
	beforeShutdown func(),
) error {
	httpServer := &http.Server{
		Addr:              cfg.HTTP.Address,
		Handler:           handler,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	listener, err := net.Listen("tcp", cfg.HTTP.Address)
	if err != nil {
		return fmtError("listen", err)
	}

	listenError := make(chan error, 1)

	go func() {
		logger.Info("http server started", slog.String("address", listener.Addr().String()))
		listenError <- httpServer.Serve(listener)
	}()

	select {
	case err := <-listenError:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmtError("listen", err)

	case <-ctx.Done():

		if beforeShutdown != nil {
			beforeShutdown()
		}

		logger.Info("shutdown signal received")

		shutdownContext, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := httpServer.Shutdown(shutdownContext); err != nil {
			return fmtError("graceful shutdown", err)
		}

		logger.Info("http server stopped")
		return nil
	}
}

func fmtError(operation string, err error) error {
	return fmt.Errorf("%s: %w", operation, err)
}
