# internal

Go backend core of SiteCraft: Chi v5 HTTP routing, SQLite/PostgreSQL data layer via sqlc and goose, SCS session authentication, bimg media pipeline, in-memory caching, i18n, events, business modules, and background notifications.

## Purpose

Owns all server-side application logic, data persistence, HTTP routing, real-time messaging, and multi-tenant fleet data management in a single compiled Go binary.

## Ownership

- `internal/auth/` — Session authentication, password hashing (`bcrypt`), and route protection middleware (`RequireAdminAuth`).
- `internal/cache/` — In-memory caching with `sync.RWMutex` and automatic invalidation on data mutations.
- `internal/data/` — JSON preset loader (`preset.go`) and repository interface (`sql_repository.go`) bridging `internal/db` with domain models.
- `internal/db/` — Database connection management (`connection.go`), goose migration runner, and sqlc-generated type-safe queries (`Querier`, `Queries`, models).
- `internal/events/` — Internal asynchronous event bus for decoupled component messaging.
- `internal/exporter/` — Static Site Generation (SSG) HTML/CSS exporter.
- `internal/handlers/` — Domain handlers implementing standard `http.HandlerFunc` (web, leads, booking, cart, checkout, chat, reviews, services, fleet, client admin, TMA).
- `internal/i18n/` — Multi-language catalog, URL/cookie/header negotiation, localized validation.
- `internal/media/` — Pure-Go and bimg image processing, thumbnail generation (thumb, sm, md, lg), and static media caching.
- `internal/models/` — Domain models and DTOs (booking, ecommerce, content, fleet, seo, activity, chat, b2b).
- `internal/modules/` — Business logic extensions (payment processing, etc.).
- `internal/notifications/` — Async notifications (Telegram Bot relay, CRM webhooks, iCal .ics reminders).
- `internal/server/` — Root Chi HTTP server, middleware stack (RequestID, RealIP, Slog, Recoverer, Compress, CSRF, UTM, A/B, i18n), route registration, and graceful shutdown.
- `internal/testutil/` — Test helpers and isolated database test fixtures (`NewTestApp(t)` with in-memory/temp SQLite).

## Local Contracts

- Go 1.27 standard library idioms; zero external CGO dependencies in default build (`modernc.org/sqlite`).
- Database Dual-Engine Ready: all DB interactions go through `db.Querier` or `data.Repository`. Schema changes are written as ANSI SQL migrations in `sql/migrations/sqlite/` and `sql/migrations/postgres/`.
- Concurrency & Thread safety: shared in-memory state must be guarded by `sync.RWMutex` or channel coordination. SQLite connections configured with 1 writer and N readers in WAL mode.
- Structured logging: use standard `log/slog` for operational, audit, and security logs.
- Security & Sanitization: apply `bluemonday` sanitization on untrusted input, validate honeypots on public forms, use bcrypt for password hashing, and protect admin endpoints with SCS sessions + CSRF tokens.
- Granular Cache-Control: adhere to caching policies (`public, max-age=...` for static/public cacheable routes; `private, no-cache, no-store` for authenticated `/admin` routes).
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
  3. Register route in Chi router (`internal/server/server.go`).
  4. Write a compact integration test using `testutil.NewTestApp(t)` and `httptest.NewRecorder()`.
- **Recipe: How to write an isolated database test**:
  1. Initialize app with `app := testutil.NewTestApp(t)`.
  2. Cleanup is automatic via `t.Cleanup` (`app.Close()`).
  3. Query or mutate data via `app.Repo` or `app.DB` with complete isolation and migrations applied.
- **Testing & E2E notes**: Drag-and-drop browser sorting and Alpine.js transitions/animations are verified via Go `httptest` server-side HTML partials rather than full browser E2E (Zero Node.js constraint).
- **Validation shortcuts & risks**: Validation in `review.go` checks byte length `len(name) < 2` (UTF-8 Cyrillic 1 char = 2 bytes); upgrade to `utf8.RuneCountInString` when single-character Unicode validation is required.

## Verification

- Run internal tests: `go test -v ./internal/...`
- Full package compilation: `go build -o /dev/null ./...`

## Child DOX Index

None. Flat Go package hierarchy owned collectively by `internal/AGENTS.md`.
