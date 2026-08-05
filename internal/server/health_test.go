package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthEndpoints(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	health := NewHealth("test-service")
	handler := health.Handler(logger)

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{path: "/health/live", wantStatus: http.StatusOK, wantBody: `"status":"alive"`},
		{path: "/health/ready", wantStatus: http.StatusOK, wantBody: `"status":"ready"`},
		{path: "/", wantStatus: http.StatusOK, wantBody: `"status":"running"`},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if !strings.Contains(response.Body.String(), test.wantBody) {
				t.Fatalf("body = %q, want it to contain %q", response.Body.String(), test.wantBody)
			}
			if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", contentType)
			}
		})
	}
}

func TestReadinessCanBeDisabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	health := NewHealth("test-service")
	health.SetReady(false)
	request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	response := httptest.NewRecorder()

	health.Handler(logger).ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

func TestUnknownRouteReturnsNotFound(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	request := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	response := httptest.NewRecorder()

	NewHealth("test-service").Handler(logger).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}
