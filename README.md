# GoStreamMesh

GoStreamMesh is a high-throughput log ingestion platform written in Go. The
planned pipeline accepts logs over HTTP, uses bounded worker pools to publish
them to RabbitMQ, batches downstream processing, and indexes events into
Elasticsearch for search and analysis.

The platform is designed around:

- Bounded concurrency and explicit backpressure.
- At-least-once delivery with idempotent event IDs.
- RabbitMQ retries and dead-letter queue handling.
- Elasticsearch Bulk API indexing.
- Prometheus metrics, Grafana dashboards, and Kibana log exploration.
- Integration, failure, and sustained load testing.

## Project Status

Steps 1 and 2 are complete:

- Requirements and measurable acceptance criteria are documented.
- All three service commands compile and run.
- Environment configuration is validated at startup.
- Services emit structured JSON logs and shut down gracefully.
- Liveness and readiness endpoints are available.
- Build, test, and container foundations are in place.

RabbitMQ publishing, ingestion APIs, and Elasticsearch indexing are planned for
the following implementation steps.

## Services

| Service             | Default address | Responsibility                          |
| ------------------- | --------------- | --------------------------------------- |
| `ingestion-service` | `:8080`         | Accept and publish log events           |
| `worker-service`    | `:8081`         | Consume, batch, retry, and index events |
| `query-service`     | `:8082`         | Search indexed log events               |

The current executable baseline exposes:

- `GET /`
- `GET /health/live`
- `GET /health/ready`

## Local Development

Go 1.26 or newer is required.

```bash
go test ./...
go vet ./...
go run ./cmd/ingestion-service
```

Run the other processes with:

```bash
go run ./cmd/worker-service
go run ./cmd/query-service
```

Copy the values from `.env.example` into your local environment as needed.
Every setting has a development default, so no environment file is required
for the current service skeleton.

The Makefile provides equivalent commands:

```bash
make check
make run-ingestion
make run-worker
make run-query
```

## Container Build

The generic Dockerfile builds any service command:

```bash
docker build \
  --build-arg SERVICE=ingestion-service \
  -f deployments/docker/Dockerfile \
  -t gostreammesh/ingestion-service:dev .
```
