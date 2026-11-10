# ui

Frontend presentation layer, Templ templates, Tailwind CSS v4 styling, component registries, and static web assets.

## Purpose

Owns the user interface architecture of SiteCraft: compiling type-safe HTML templates, managing Tailwind v4 design tokens, registering pluggable page blocks, and embedding static assets into the binary.

## Ownership

- `ui/assets/` — Tailwind CSS entry point (`input.css`), runtime tokens (`tokens.css`), client scripts, icons, and `images/`.
- `ui/components/` — Reusable Templ components (base primitives, headers, footers, drawers, modals, widgets).
- `ui/components/base/` — Zero-allocation Go Variants (CVA) button, badge, card, input, and modal primitives.
- `ui/components/cinematic/` — 31 Zero-GSAP cinematic animation and scrollytelling modules.
- `ui/components/icons/` — Self-contained type-safe Lucide SVG icon components compiled via Templ.
- `ui/components/cinematic_2/` — Screen 1 & Screen 2 components adapted from jhill.com (cinematic video hero and tilted ArtStation marquee collage).
- `ui/layouts/` — Master templates (`base.templ`) and SEO/meta tag helpers.
- `ui/pages/` — Top-level page views (home, blog, tracker, shop, invoice, vip, fleet, legal, client admin).
- `ui/registry/` — Dynamic section registry (`BlockRegistry`), generic payload parser (`ParsePayload[T]`), and block factory pattern (`RenderBlock`).
- `ui/sections/` — Pre-built section blocks and variants (hero, features, services, booking, quiz, steps, reviews, team, map, faq, etc.).
- `ui/embed.go` — Embedded asset file system (`//go:embed all:assets`).

## Local Contracts

- Type-safe templates: all HTML markup authored in `*.templ` files; never write raw unchecked HTML strings. STRICT: Never edit generated `*_templ.go` files directly.
- DOM Source Mapping: `RenderBlock` in `ui/registry/` must wrap dynamic blocks in `BlockWrapper` (`<div class="contents">` with `data-block` and `data-source`), and base components (`ui/components/base/`) must preserve `data-component` and `data-source` attributes for seamless T3 Code preview inspection.
- Zero-Node.js in production: compile CSS and templates ahead of time (`templ generate`, `tailwindcss`).
- Design tokens & Semantic markup: configure themes via `@theme` in `ui/assets/input.css` and runtime CSS variables in `ui/assets/css/tokens.css`. Always use semantic utility classes (`bg-surface`, `text-content`, `text-accent`, `rounded-card`, `rounded-btn`) rather than hardcoded colors or pixel radii.
- Payload parsing: use generic `registry.ParsePayload[T any]` to parse block payloads without handwriting repetitive unmarshalers.
- Section variants: maintain variants as named functions in the owning package (e.g. `hero.CyberHero`, `sections.ReviewsSlider`) to avoid excessive micro-packages.
- Go Variants (CVA) in `ui/components/base/`: use typed Go structs for variant resolution; use `tailwind-merge-go` only when resolving conflicting override classes.
- Alpine + HTMX compatibility: use `hx-alpine-compat` patterns so dynamic DOM replacements safely reinitialize Alpine state without memory leaks.
- CSS Data-Variants: prefer native Tailwind v4 data variants (`data-open:`, `data-active:`) triggered by Alpine data attributes over manual JavaScript class manipulation.
- Accessibility standard: adhere to ГОСТ Р 52872-2019 accessibility mode (`body.a11y-mode`) and honor `@media (prefers-reduced-motion: reduce)`.
- Anti-Slop Guard in UI: All components must use semantic tokens from `presets/*.json` (never default purple/indigo gradients or floating blur orbs). Elements with `[data-reveal]` must remain accessible and visible without JavaScript via `<noscript>` progressive enhancement. Animations must be restricted to `transform` and `opacity` (150–200ms) with zero layout thrashing.

## Work Guidance

- **Recipe: How to add a new landing section block**:
  1. Create component in `ui/sections/<name>.templ` using semantic Tailwind v4 design tokens (`bg-surface`, `text-content`, etc.).
  2. Run `templ generate` to compile template to Go.
  3. Register the block in `ui/registry/registry.go` (payload struct + factory mapping in `RenderBlock`).
  4. Add default block data to JSON presets in `presets/` and run `make check`.
- Re-generate Go template code after editing any `.templ` file: `templ generate`.
- Rebuild minified CSS on utility changes: `tailwindcss -i ./ui/assets/input.css -o ./ui/assets/styles.css --minify`.
- Register new section blocks in `ui/registry/registry.go` with fallback handling.
- Audit responsive layouts across mobile (390px), tablet (768px), and desktop (1440px) using the `layout-review` skill.
- Generate architectural landing layouts using the `hallmark-craft` skill (`.agents/skills/hallmark-craft/SKILL.md`).
- Generate or align sections against `brief.md` using the `stitch-workflow` skill.

## Verification

- Generate Templ code: `templ generate`
- Run UI package tests: `go test -v ./ui/...`
- Rebuild production CSS: `tailwindcss -i ./ui/assets/input.css -o ./ui/assets/styles.css --minify`

## Child DOX Index

- [components/cinematic/AGENTS.md](components/cinematic/AGENTS.md) — 31 Zero-GSAP cinematic animation modules, scrollytelling mechanics, and 60 FPS performance contracts.
- [components/cinematic_2/AGENTS.md](components/cinematic_2/AGENTS.md) — Screen 1 & Screen 2 components from jhill.com (100svh video hero + tilted ArtStation marquee collage).
