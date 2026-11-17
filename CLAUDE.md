# Project Guidelines

Refer to [AGENTS.md](./AGENTS.md) for core contracts, architecture, and DOX framework rules.

## Core Directives
- **Verification**: Run `make check` (runs Templ code generation, Tailwind CSS v4 compilation, `go vet ./...`, and `go test ./...`) before reporting task completion.
- **Style**: Follow Ponytail mode (Lazy Senior Dev Mode) — minimal diffs, zero over-engineering, reuse existing codebase, stdlib first.
- **Stack**: Go 1.27 + Chi v5 + modernc/sqlite (PostgreSQL-ready) + sqlc + goose + scs + Templ + HTMX + Alpine.js + Tailwind CSS v4. Single binary, zero Node.js runtime in production.
- **Templates**: Always run `templ generate` after editing `*.templ` files.
- **Styles Exclusion**: STRICT: NEVER open, inspect, view, or edit `ui/assets/styles.css` or `ui/assets/admin.css` — they are automatically generated build artifacts and strictly excluded from AI visibility. Edit `ui/assets/input.css`, `ui/assets/admin_input.css`, or `ui/assets/css/tokens.css` instead.
- **Cookbooks**: Check `.agents/cookbooks/` for canonical snippets (Templ CVA, HTMX handlers, secure forms).
