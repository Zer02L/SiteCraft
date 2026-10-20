# cinematic

Zero-GSAP cinematic animation modules, scrollytelling mechanics, and high-performance visual effects.

## Purpose

Owns the collection of 31 cinematic animation components built with pure CSS, Motion One, and Alpine.js for storytelling, interactive showcases, and high-conversion landing pages.

## Ownership

- `ui/components/cinematic/*.templ` — All 31 cinematic animation components (dock nav, coverflow, sticky stack, horizontal scroll, cursor reactive, text scramble, ogl waves, etc.).
- `ui/components/cinematic/README.md` — Component catalog, visual descriptions, and API parameter references.

## Local Contracts

- Zero-GSAP standard: strictly NO GSAP or heavy animation runtimes; rely exclusively on modern CSS animations, Motion One, and Alpine.js.
- 60 FPS performance budget: animate hardware-accelerated properties (`transform` and `opacity`) only; avoid layout thrashing via `top`, `left`, `width`, or `height`.
- Accessibility & Reduced Motion: automatically disable continuous or high-frequency motion when `body.a11y-mode` is active or `@media (prefers-reduced-motion: reduce)` matches.
- DOM Swaps & Memory: cleanup and rebind IntersectionObservers, scroll listeners, and timers on `htmx:afterSwap` / `htmx:beforeCleanup` without memory leaks.

## Work Guidance

- Use the `story-review` skill to evaluate scrollytelling pacing, frame stability, and mobile responsiveness.
- Keep inline component scripts scoped and self-contained; avoid global `window` namespace pollution.
- Test fallback layouts on mobile viewports (390px) where advanced cursor or WebGL effects may be suppressed.

## Verification

- Generate Templ code: `templ generate`
- Run story audit skill: invoke `story-review` skill or manually verify 60 FPS in DevTools Performance tab.

## Child DOX Index

None.
