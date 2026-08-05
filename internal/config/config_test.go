package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	clearConfigEnvironment(t)

	cfg, err := Load("ingestion-service", ":8080")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Service != "ingestion-service" {
		t.Fatalf("Service = %q, want ingestion-service", cfg.Service)
	}
	if cfg.HTTP.Address != ":8080" {
		t.Fatalf("HTTP.Address = %q, want :8080", cfg.HTTP.Address)
	}
	if cfg.Ingestion.QueueCapacity != 10_000 {
		t.Fatalf("QueueCapacity = %d, want 10000", cfg.Ingestion.QueueCapacity)
	}
	if cfg.Ingestion.WorkerCount < 1 {
		t.Fatalf("WorkerCount = %d, want at least 1", cfg.Ingestion.WorkerCount)
	}
}

func TestLoadOverrides(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("HTTP_ADDRESS", ":9090")
	t.Setenv("SHUTDOWN_TIMEOUT", "30s")
	t.Setenv("INGESTION_QUEUE_CAPACITY", "25000")
	t.Setenv("INGESTION_WORKER_COUNT", "12")

	cfg, err := Load("worker-service", ":8081")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Environment != "production" || cfg.LogLevel != "debug" {
		t.Fatalf("unexpected environment settings: %+v", cfg)
	}
	if cfg.HTTP.Address != ":9090" {
		t.Fatalf("HTTP.Address = %q, want :9090", cfg.HTTP.Address)
	}
	if cfg.ShutdownTimeout != 30*time.Second {
		t.Fatalf("ShutdownTimeout = %s, want 30s", cfg.ShutdownTimeout)
	}
	if cfg.Ingestion.QueueCapacity != 25_000 || cfg.Ingestion.WorkerCount != 12 {
		t.Fatalf("unexpected ingestion settings: %+v", cfg.Ingestion)
	}
}

func TestLoadReportsAllInvalidValues(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv("APP_ENV", "staging")
	t.Setenv("LOG_LEVEL", "verbose")
	t.Setenv("SHUTDOWN_TIMEOUT", "never")
	t.Setenv("INGESTION_WORKER_COUNT", "0")
	t.Setenv("RABBITMQ_URL", "http://localhost:5672")
	t.Setenv("ELASTICSEARCH_URL", "localhost:9200")

	_, err := Load("ingestion-service", ":8080")
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}

	for _, expected := range []string{
		"APP_ENV",
		"LOG_LEVEL",
		"SHUTDOWN_TIMEOUT",
		"INGESTION_WORKER_COUNT",
		"RABBITMQ_URL",
		"ELASTICSEARCH_URL",
	} {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("Load() error %q does not mention %s", err, expected)
		}
	}
}

func clearConfigEnvironment(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"APP_ENV",
		"LOG_LEVEL",
		"HTTP_ADDRESS",
		"HTTP_READ_HEADER_TIMEOUT",
		"HTTP_READ_TIMEOUT",
		"HTTP_WRITE_TIMEOUT",
		"HTTP_IDLE_TIMEOUT",
		"SHUTDOWN_TIMEOUT",
		"RABBITMQ_URL",
		"ELASTICSEARCH_URL",
		"INGESTION_QUEUE_CAPACITY",
		"INGESTION_WORKER_COUNT",
		"INGESTION_ENQUEUE_TIMEOUT",
		"RABBITMQ_PUBLISH_TIMEOUT",
	} {
		t.Setenv(key, "")
	}
}
