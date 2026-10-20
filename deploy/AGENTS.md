# deploy

Deployment manifests, process management units, reverse proxy configurations, fleet orchestration, and Zero-RAM monitoring.

## Purpose

Owns operational infrastructure contracts, daemon configurations, security hardening for reverse proxies, multi-tenant fleet orchestration, and lightweight health watchdog monitoring.

## Ownership

- `deploy/nginx/` — Production Nginx reverse proxy configuration (`sitecraft.conf`) with Brotli, SSL, security headers, and WebSocket routing.
- `deploy/systemd/` — Systemd service unit (`sitecraft.service`) configuring restart policies and non-root execution.
- `deploy/monitoring/` — Zero-RAM watchdog script (`watchdog.sh`) and environment configuration (`.watchdog.env.example`).
- Root deployment assets:
  - `fleet.sh` — Bash CLI orchestrating multi-tenant binary instances on a single host.
  - `Caddyfile` — Automatic TLS reverse proxy configuration.
  - `Dockerfile`, `docker-compose.yml` — Containerized build and run definitions.

## Local Contracts

- Single binary runtime: compiled binary runs autonomously with an embedded SQLite database in WAL mode.
- Non-root user: systemd and container services must run under an unprivileged `sitecraft` user.
- Reverse proxy headers: must forward `X-Real-IP`, `X-Forwarded-For`, `X-Forwarded-Proto`, and support HTTP upgrade for WebSocket / SSE (`/api/realtime`, chat widget).
- Zero-RAM watchdog: `watchdog.sh` runs as a cron job, checks HTTP endpoints and resource thresholds, and alerts via Telegram without persistent background daemon memory consumption.

## Work Guidance

- Verify shell syntax with `bash -n` on all `.sh` scripts before committing changes.
- Ensure `fleet.sh` commands (`start`, `stop`, `restart`, `status`, `backup`, `update`) remain idempotent and preserve SQLite database files.
- Nginx and Caddy configs must enforce TLS 1.2+, HSTS, X-Content-Type-Options, and frame options.

## Verification

- Shell syntax validation: `bash -n deploy/monitoring/watchdog.sh fleet.sh init-project.sh`

## Child DOX Index

None.
