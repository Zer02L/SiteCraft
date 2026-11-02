# cinematic

Zero-GSAP cinematic animation modules, scrollytelling mechanics, and high-performance visual effects.

## Purpose

Owns the collection of 33 cinematic animation components built with pure CSS, Motion One, Alpine.js, and OGL WebGL for storytelling, interactive showcases, and high-conversion landing pages.

## Ownership

- `ui/components/cinematic/*.templ` — All 33 cinematic animation components (dock nav, coverflow, sticky stack, horizontal scroll, cursor reactive, text scramble, ogl waves, flat to 3d, etc.).
- `ui/components/cinematic/README.md` — Component catalog, visual descriptions, and API parameter references.

## Local Contracts

- Zero-GSAP standard: strictly NO GSAP or heavy animation runtimes; rely exclusively on modern CSS animations, Motion One, and Alpine.js.
- 60 FPS performance budget: animate hardware-accelerated properties (`transform` and `opacity`) only; avoid layout thrashing via `top`, `left`, `width`, or `height`.
- Motion Tokens & Zero-Reflow Bridge: all durations, easings, and spring physics must be sourced from CSS tokens via `window.SiteTokens` (`readSec`, `parseSpring`); never call `getComputedStyle` inside RAF or scroll handlers.
- Accessibility & Reduced Motion: automatically disable continuous or high-frequency motion when `body.a11y-mode` is active or `@media (prefers-reduced-motion: reduce)` matches (`window.SiteTokens.motionEnabled()`).
- WCAG 2.2.2 (Pause/Stop/Hide): any looping animation or media running longer than 5 seconds with autoplay must provide user controls to pause or suppress the motion.
- DOM Swaps & Memory: cleanup and rebind IntersectionObservers, scroll listeners, and timers on `htmx:afterSwap` / `htmx:beforeCleanup` without memory leaks.

## Work Guidance

- Use the `story-review` skill to evaluate scrollytelling pacing, frame stability, and mobile responsiveness.
- Use the `creative-design-website` skill for choreographing new cinematic sections and generating `motion-map.md`.
- Keep inline component scripts scoped and self-contained; avoid global `window` namespace pollution.
- Test fallback layouts on mobile viewports (390px) where advanced cursor or WebGL effects may be suppressed.

## Verification

- Generate Templ code: `templ generate`
- Run story audit skill: invoke `story-review` skill or manually verify 60 FPS in DevTools Performance tab.

## Child DOX Index

None.
