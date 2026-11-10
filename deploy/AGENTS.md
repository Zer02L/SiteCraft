# deploy

Deployment manifests, process management units, reverse proxy configurations, fleet orchestration, and Zero-RAM monitoring.

## Purpose

Owns operational infrastructure contracts, daemon configurations, security hardening for reverse proxies, multi-tenant fleet orchestration, and lightweight health watchdog monitoring.

## Ownership

- `deploy/nginx/` — Production Nginx reverse proxy configuration (`sitecraft.conf`) with Brotli, SSL, security headers, and WebSocket routing.
- `deploy/systemd/` — Systemd service unit (`sitecraft.service`) configuring restart policies and non-root execution.
- `deploy/podman/` — Rootless Podman Quadlet container unit (`sitecraft.container`) and deployment guides.
- `deploy/monitoring/` — Zero-RAM watchdog script (`watchdog.sh`) and environment configuration (`.watchdog.env.example`).
- Root deployment assets:
  - `fleet.sh` — Bash CLI orchestrating multi-tenant binary instances and Podman container lifecycle.
  - `Caddyfile` — Automatic TLS reverse proxy configuration.
  - `Containerfile`, `compose.yaml` (and compatibility `Dockerfile`, `docker-compose.yml`) — Rootless OCI container definitions for Podman.

## Local Contracts

- Single binary runtime: compiled binary runs autonomously with an embedded SQLite database in WAL mode.
- Non-root user: systemd, Podman Quadlet, and container services must run under an unprivileged `sitecraft` user (UID/GID 10001).
- Podman Rootless OCI: persistent volumes must be mounted with SELinux label flags (`:Z`) pointing to `/data`.
- Reverse proxy headers: must forward `X-Real-IP`, `X-Forwarded-For`, `X-Forwarded-Proto`, and support HTTP upgrade for WebSocket / SSE (`/api/realtime`, chat widget).
- Zero-RAM watchdog: `watchdog.sh` runs as a cron job, checks HTTP endpoints and resource thresholds, and alerts via Telegram without persistent background daemon memory consumption.

## Work Guidance

- Verify shell syntax with `bash -n` on all `.sh` scripts before committing changes.
- Ensure `fleet.sh` commands (`start`, `stop`, `restart`, `status`, `backup`, `podman-up`, `podman-down`) remain idempotent and preserve SQLite database files.
- Nginx and Caddy configs must enforce TLS 1.2+, HSTS, X-Content-Type-Options, and frame options.

## Verification & Future Roadmap

- Shell syntax validation: `bash -n deploy/monitoring/watchdog.sh fleet.sh init-project.sh`
- Future DevOps Quality Gates (Backlog):
  - `ansible-lint` — static analysis and idempotency verification of Ansible playbooks and roles for host and fleet configuration.
  - `kubeconform` — schema-based validation of Kubernetes manifests and CRDs in CI/CD before deployment.

## Child DOX Index

None.
