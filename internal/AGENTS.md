# internal

Go backend core of SiteCraft: PocketBase embedding, SQLite data layer, HTTP handlers, middleware, caching, i18n, events, business modules, and background notifications.

## Purpose

Owns all server-side application logic, data persistence, HTTP routing, real-time messaging, and multi-tenant fleet data management in a single compiled Go binary.

## Ownership

- `internal/cache/` — In-memory caching with `sync.RWMutex` and automatic invalidation on PocketBase admin mutations.
- `internal/data/` — PocketBase collections, schemas, migrations, hooks, domain queries (site, articles, booking, leads, reviews, services), and JSON preset loader.
- `internal/events/` — Internal asynchronous event bus for decoupled component messaging.
- `internal/exporter/` — Static Site Generation (SSG) HTML/CSS exporter.
- `internal/handlers/` — HTTP routers, middleware (i18n, rate limiting, auth, cache-control), and HTMX/JSON endpoints (leads, booking, cart, checkout, chat, reviews, services, fleet, client admin, TMA).
- `internal/i18n/` — Multi-language catalog, URL/cookie/header negotiation, localized validation.
- `internal/models/` — Domain models and DTOs (booking, ecommerce, content, fleet, seo, activity, chat, b2b).
- `internal/modules/` — Business logic extensions (payment processing, etc.).
- `internal/notifications/` — Async notifications (Telegram Bot relay, CRM webhooks, iCal .ics reminders).
- `internal/bot/` — Telegram bot interaction handlers.
- `internal/testutil/` — Test helpers and isolated PocketBase test fixtures (`NewTestApp(t)`).

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
- **Recipe: How to add an HTMX / HTTP endpoint**:
  1. Define request DTO, validation, and sanitization (honeypot + `bluemonday`) in `internal/handlers/<domain>.go`.
  2. Implement handler returning a partial Templ component (e.g. `component.Render(ctx, w)` with HTTP 200/422).
  3. Register route in router (`internal/handlers/routes.go` or module setup).
  4. Write a compact integration test using `testutil.NewTestApp(t)` and `httptest.NewRecorder()`.
- **Recipe: How to write an isolated PocketBase test**:
  1. Initialize app with `app := testutil.NewTestApp(t)`.
  2. Cleanup is automatic via `t.Cleanup` (`app.ResetBootstrapState()`).
  3. Query or mutate collections via `app.FindCollectionByNameOrId(...)` without risk of database locking.
- **Testing & E2E notes**: Drag-and-drop browser sorting and Alpine.js transitions/animations are verified via Go `httptest` server-side HTML partials rather than full browser E2E (Zero Node.js constraint).
- **Validation shortcuts & risks**: Validation in `review.go` checks byte length `len(name) < 2` (UTF-8 Cyrillic 1 char = 2 bytes); upgrade to `utf8.RuneCountInString` when single-character Unicode validation is required.

## Verification

- Run internal tests: `go test -v ./internal/...`
- Full package compilation: `go build -o /dev/null ./...`

## Child DOX Index

None. Flat Go package hierarchy owned collectively by `internal/AGENTS.md`.
