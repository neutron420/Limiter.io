# Deployment & Rollback

## Prerequisites

- Docker & Docker Compose
- PostgreSQL 16
- Redis 7
- Kafka 3.x
- Go 1.22+

## Build

```bash
docker compose -f deploy/docker/docker-compose.yml build
```

## Deploy

```bash
# Apply migrations
docker compose run --rm api ./api migrate

# Start services
docker compose -f deploy/docker/docker-compose.yml up -d
```

## Rollback

```bash
# Roll back to previous version
docker compose -f deploy/docker/docker-compose.yml down
git checkout <previous-tag>
docker compose -f deploy/docker/docker-compose.yml build
docker compose -f deploy/docker/docker-compose.yml up -d

# Roll back database (if needed)
pg_restore -h localhost -U postgres -d ratelimiter --clean /backups/last_stable.dump
```

## Health Checks

```bash
curl http://localhost:8080/status
curl http://localhost:8080/health
```

## Monitoring

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3001
- Kafka UI: http://localhost:8085

## Verification Post-Deploy

```bash
# Run smoke tests
go test ./internal/handlers/... -run TestRegisterLoginFlow -v
go test ./internal/handlers/... -run TestCreateProjectFlow -v

# Verify analytics pipeline
curl -X GET http://localhost:8080/api/v1/projects/test-id/analytics
```
