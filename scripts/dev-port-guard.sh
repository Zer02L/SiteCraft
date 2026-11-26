#!/usr/bin/env bash
# Pre-start guard для `mise run dev`: убирает осиротевший tmp/main прошлого air-сеанса,
# который держит порт и мешает новому серверу сделать bind.
set -euo pipefail

port="${PORT:-8080}"
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"

listener="$(ss -ltnpH "( sport = :${port} )" 2>/dev/null | sed -n 's/.*pid=\([0-9]*\).*/\1/p' | head -n1 || true)"
[[ -z "$listener" ]] && exit 0

exe="$(readlink "/proc/${listener}/exe" 2>/dev/null || true)"
exe="${exe% \(deleted\)}"
cwd="$(readlink -f "/proc/${listener}/cwd" 2>/dev/null || true)"

# Чужой процесс не трогаем — только наш dev-билд из этого репозитория.
if [[ "$exe" != "$root/tmp/main" || "$cwd" != "$root" ]]; then
  echo "[dev] Port ${port} is in use by PID ${listener} (${exe:-unknown}). Stop it manually." >&2
  exit 1
fi

echo "[dev] Stopping stale dev server PID ${listener} on :${port}"
kill "$listener" 2>/dev/null || true
for _ in {1..20}; do
  kill -0 "$listener" 2>/dev/null || break
  sleep 0.1
done
if kill -0 "$listener" 2>/dev/null; then
  kill -9 "$listener" 2>/dev/null || true
  sleep 0.2
fi
if ss -ltnH "( sport = :${port} )" | grep -q .; then
  echo "[dev] Port ${port} is still busy after cleanup." >&2
  exit 1
fi
