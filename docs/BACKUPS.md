# PostgreSQL Backup & Restore

## Automated Daily Backup

```bash
# backup script (run via cron daily)
pg_dump -h localhost -U postgres ratelimiter -F c -f /backups/ratelimiter_$(date +%Y%m%d_%H%M%S).dump
```

## Restore

```bash
pg_restore -h localhost -U postgres -d ratelimiter --clean /backups/ratelimiter_20260715_120000.dump
```

## Docker Compose

The `deploy/docker/docker-compose.yml` includes a `pg-backup` service that runs `pg_dump` daily and stores archives in `./backups/`.

## Retention

- Keep daily backups for 30 days
- Keep weekly backups for 6 months
- Keep monthly backups for 2 years

## Verification

Restore to a staging environment and run:
```bash
go test ./internal/database/... -run TestBackupRestore -v
```
