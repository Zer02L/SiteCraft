# internal

Go backend core of SiteCraft: PocketBase embedding, SQLite data layer, HTTP handlers, middleware, caching, i18n, events, business modules, and background notifications.

## Purpose

Owns all server-side application logic, data persistence, HTTP routing, real-time messaging, and multi-tenant fleet data management in a single compiled Go binary.

## Ownership

- `internal/cache/` — In-memory caching with `sync.RWMutex` and automatic invalidation on PocketBase admin mutations.
- `internal/data/` — PocketBase collections, schemas, migrations, hooks, and JSON preset loader.
- `internal/events/` — Internal asynchronous event bus for decoupled component messaging.
- `internal/exporter/` — Static Site Generation (SSG) HTML/CSS exporter.
- `internal/handlers/` — HTTP routers, middleware (i18n, rate limiting, auth, cache-control), and HTMX/JSON endpoints (leads, booking, cart, checkout, chat, reviews, fleet, client admin, TMA).
- `internal/i18n/` — Multi-language catalog, URL/cookie/header negotiation, localized validation.
- `internal/models/` — Domain models and DTOs (booking, ecommerce, content, fleet, seo, activity, chat, b2b).
- `internal/modules/` — Business logic extensions (payment processing, etc.).
- `internal/notifications/` — Async notifications (Telegram Bot relay, CRM webhooks, iCal .ics reminders).
- `internal/bot/` — Telegram bot interaction handlers.

## Local Contracts

- Go 1.27 standard library idioms; zero external CGO dependencies beyond PocketBase (`modernc.org/sqlite`).
- Concurrency & Thread safety: shared in-memory state must be guarded by `sync.RWMutex` or channel coordination.
- Structured logging: use standard `log/slog` for operational, audit, and security logs.
- Security & Sanitization: apply `bluemonday` sanitization on untrusted input, validate honeypots on public forms, use bcrypt for password hashing.
- Granular Cache-Control: adhere to caching policies (`public, max-age=...` for static/public cacheable routes; `private, no-cache, no-store` for authenticated `/client` routes).
- HTMX Responses: handlers return partial Templ components or OOB swaps for dynamic interactions.

## Work Guidance

- Grep all callers (`callers`) before modifying function signatures, route handlers, or service interfaces.
- Prioritize platform and Go stdlib capabilities before introducing external dependencies.
- Any non-trivial business logic, parser, auth check, or money-related calculation must include a focused unit test.
- Mark intentional shortcuts clearly: `// shortcut: <limit>, <upgrade condition>`.
- Never compromise trust boundary validation, error handling that prevents data loss, or user data privacy.

## Verification

- Run internal tests: `go test -v ./internal/...`
- Full package compilation: `go build -o /dev/null ./...`

## Child DOX Index

None. Flat Go package hierarchy owned collectively by `internal/AGENTS.md`.
