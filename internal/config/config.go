package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	EnvironmentDevelopment = "development"
	EnvironmentTest        = "test"
	EnvironmentProduction  = "production"
)

// Config contains process-wide settings shared by all GoStreamMesh services.
type Config struct {
	Service          string
	Environment      string
	LogLevel         string
	HTTP             HTTPConfig
	ShutdownTimeout  time.Duration
	RabbitMQURL      string
	ElasticsearchURL string
	Ingestion        IngestionConfig
}

type HTTPConfig struct {
	Address           string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type IngestionConfig struct {
	QueueCapacity  int
	WorkerCount    int
	EnqueueTimeout time.Duration
	PublishTimeout time.Duration
}

// Load reads configuration from the environment and applies safe local
// defaults. Invalid values are returned together so startup failures are easy
// to diagnose.
func Load(service, defaultHTTPAddress string) (Config, error) {
	cfg := Config{
		Service:     service,
		Environment: envString("APP_ENV", EnvironmentDevelopment),
		LogLevel:    strings.ToLower(envString("LOG_LEVEL", "info")),
		HTTP: HTTPConfig{
			Address:           envString("HTTP_ADDRESS", defaultHTTPAddress),
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		ShutdownTimeout:  15 * time.Second,
		RabbitMQURL:      envString("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		ElasticsearchURL: envString("ELASTICSEARCH_URL", "http://localhost:9200"),
		Ingestion: IngestionConfig{
			QueueCapacity:  10_000,
			WorkerCount:    max(runtime.GOMAXPROCS(0), 1),
			EnqueueTimeout: 50 * time.Millisecond,
			PublishTimeout: 5 * time.Second,
		},
	}

	var errs []error
	cfg.HTTP.ReadHeaderTimeout = envDuration("HTTP_READ_HEADER_TIMEOUT", cfg.HTTP.ReadHeaderTimeout, &errs)
	cfg.HTTP.ReadTimeout = envDuration("HTTP_READ_TIMEOUT", cfg.HTTP.ReadTimeout, &errs)
	cfg.HTTP.WriteTimeout = envDuration("HTTP_WRITE_TIMEOUT", cfg.HTTP.WriteTimeout, &errs)
	cfg.HTTP.IdleTimeout = envDuration("HTTP_IDLE_TIMEOUT", cfg.HTTP.IdleTimeout, &errs)
	cfg.ShutdownTimeout = envDuration("SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout, &errs)
	cfg.Ingestion.EnqueueTimeout = envDuration("INGESTION_ENQUEUE_TIMEOUT", cfg.Ingestion.EnqueueTimeout, &errs)
	cfg.Ingestion.PublishTimeout = envDuration("RABBITMQ_PUBLISH_TIMEOUT", cfg.Ingestion.PublishTimeout, &errs)
	cfg.Ingestion.QueueCapacity = envPositiveInt("INGESTION_QUEUE_CAPACITY", cfg.Ingestion.QueueCapacity, &errs)
	cfg.Ingestion.WorkerCount = envPositiveInt("INGESTION_WORKER_COUNT", cfg.Ingestion.WorkerCount, &errs)

	if cfg.Service == "" {
		errs = append(errs, errors.New("service name must not be empty"))
	}
	if cfg.HTTP.Address == "" {
		errs = append(errs, errors.New("HTTP_ADDRESS must not be empty"))
	}
	if !isOneOf(cfg.Environment, EnvironmentDevelopment, EnvironmentTest, EnvironmentProduction) {
		errs = append(errs, fmt.Errorf("APP_ENV must be development, test, or production, got %q", cfg.Environment))
	}
	if !isOneOf(cfg.LogLevel, "debug", "info", "warn", "error") {
		errs = append(errs, fmt.Errorf("LOG_LEVEL must be debug, info, warn, or error, got %q", cfg.LogLevel))
	}
	validateURL("RABBITMQ_URL", cfg.RabbitMQURL, []string{"amqp", "amqps"}, &errs)
	validateURL("ELASTICSEARCH_URL", cfg.ElasticsearchURL, []string{"http", "https"}, &errs)

	return cfg, errors.Join(errs...)
}

func envString(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func envDuration(key string, fallback time.Duration, errs *[]error) time.Duration {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		*errs = append(*errs, fmt.Errorf("%s must be a positive duration, got %q", key, value))
		return fallback
	}
	return parsed
}

func envPositiveInt(key string, fallback int, errs *[]error) int {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		*errs = append(*errs, fmt.Errorf("%s must be a positive integer, got %q", key, value))
		return fallback
	}
	return parsed
}

func validateURL(key, value string, allowedSchemes []string, errs *[]error) {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || !isOneOf(parsed.Scheme, allowedSchemes...) {
		*errs = append(*errs, fmt.Errorf("%s must be an absolute URL with one of schemes %s, got %q", key, strings.Join(allowedSchemes, ", "), value))
	}
}

func isOneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
