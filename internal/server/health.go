package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync/atomic"
)

// Health tracks whether a process is able to accept work.
type Health struct {
	service string
	ready   atomic.Bool
}

func NewHealth(service string) *Health {
	health := &Health{service: service}
	health.ready.Store(true)
	return health
}

func (h *Health) SetReady(ready bool) {
	h.ready.Store(ready)
}

func (h *Health) Handler(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", h.handleLiveness)
	mux.HandleFunc("GET /health/ready", h.handleReadiness)
	mux.HandleFunc("GET /{$}", h.handleIndex)
	return requestLogger(logger, mux)
}

func (h *Health) handleLiveness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": h.service,
		"status":  "alive",
	})
}

func (h *Health) handleReadiness(w http.ResponseWriter, _ *http.Request) {
	if !h.ready.Load() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"service": h.service,
			"status":  "not_ready",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"service": h.service,
		"status":  "ready",
	})
}

func (h *Health) handleIndex(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": h.service,
		"status":  "running",
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.DebugContext(
			r.Context(),
			"http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_address", r.RemoteAddr),
		)
		next.ServeHTTP(w, r)
	})
}
