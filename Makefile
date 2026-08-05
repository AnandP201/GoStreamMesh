GO ?= go
SERVICE ?= ingestion-service

.PHONY: build build-all fmt test test-race vet check run-ingestion run-worker run-query clean

build:
	$(GO) build -o bin/$(SERVICE) ./cmd/$(SERVICE)

build-all:
	$(GO) build ./cmd/...

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

check: fmt test vet

run-ingestion:
	$(GO) run ./cmd/ingestion-service

run-worker:
	$(GO) run ./cmd/worker-service

run-query:
	$(GO) run ./cmd/query-service

clean:
	$(GO) clean
