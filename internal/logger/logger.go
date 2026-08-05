package logger

import (
	"log/slog"
	"os"
)

// New returns a JSON logger with stable service metadata.
func New(service, environment, configuredLevel string) *slog.Logger {
	level := new(slog.LevelVar)
	level.Set(parseLevel(configuredLevel))

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler).With(
		slog.String("service", service),
		slog.String("environment", environment),
	)
}

func parseLevel(value string) slog.Level {
	switch value {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
