# ui

Frontend presentation layer, Templ templates, Tailwind CSS v4 styling, component registries, and static web assets.

## Purpose

Owns the user interface architecture of SiteCraft: compiling type-safe HTML templates, managing Tailwind v4 design tokens, registering pluggable page blocks, and embedding static assets into the binary.

## Ownership

- `ui/assets/` — Tailwind CSS entry point (`input.css`), runtime tokens (`tokens.css`), client scripts, icons, and images.
- `ui/components/` — Reusable Templ components (base primitives, headers, footers, drawers, modals, widgets).
- `ui/components/base/` — Zero-allocation Go Variants (CVA) button, badge, and card primitives.
- `ui/layouts/` — Master templates (`base.templ`) and SEO/meta tag helpers.
- `ui/pages/` — Top-level page views (home, blog, tracker, shop, invoice, vip, fleet, legal).
- `ui/registry/` — Dynamic section registry & block factory pattern (`RegisterBlock`).
- `ui/sections/` — Pre-built section blocks (hero, features, services, booking, quiz, steps, reviews, team, map, faq, etc.).
- `ui/embed.go` — Embedded asset file system (`//go:embed assets/*`).

## Local Contracts

- Type-safe templates: all HTML markup authored in `*.templ` files; never write raw unchecked HTML strings.
- Zero-Node.js in production: compile CSS and templates ahead of time (`templ generate`, `tailwindcss`).
- Design tokens: configure themes via `@theme` in `ui/assets/input.css` and runtime CSS variables in `ui/assets/css/tokens.css`.
- Go Variants (CVA) in `ui/components/base/`: use typed Go structs for variant resolution; use `tailwind-merge-go` only when resolving conflicting override classes.
- Alpine + HTMX compatibility: use `hx-alpine-compat` patterns so dynamic DOM replacements safely reinitialize Alpine state without memory leaks.
- CSS Data-Variants: prefer native Tailwind v4 data variants (`data-open:`, `data-active:`) triggered by Alpine data attributes over manual JavaScript class manipulation.
- Accessibility standard: adhere to ГОСТ Р 52872-2019 accessibility mode (`body.a11y-mode`) and honor `@media (prefers-reduced-motion: reduce)`.

## Work Guidance

- Re-generate Go template code after editing any `.templ` file: `templ generate`.
- Rebuild minified CSS on utility changes: `tailwindcss -i ./ui/assets/input.css -o ./ui/assets/styles.css --minify`.
- Register new section blocks in `ui/registry/registry.go` with fallback handling.
- Audit responsive layouts across mobile (390px), tablet (768px), and desktop (1440px) using the `layout-review` skill.
- Generate or align sections against `brief.md` using the `stitch-workflow` skill.

## Verification

- Generate Templ code: `templ generate`
- Run UI package tests: `go test -v ./ui/...`
- Rebuild production CSS: `tailwindcss -i ./ui/assets/input.css -o ./ui/assets/styles.css --minify`

## Child DOX Index

- [components/cinematic/AGENTS.md](components/cinematic/AGENTS.md) — 31 Zero-GSAP cinematic animation modules, scrollytelling mechanics, and 60 FPS performance contracts.
