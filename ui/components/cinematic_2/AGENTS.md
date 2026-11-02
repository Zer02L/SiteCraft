# cinematic_2

Cinematic 3D artist showcase components faithfully adapting Screen 1 (Hero turntable video) and Screen 2 (tilted ArtStation collage carousel) from https://www.jhill.com/.

## Purpose

Owns the high-fidelity UI components replicating the dark-aesthetic landing experience of jhill.com: full-bleed 100svh looping character turntable video hero and the infinite-scrolling tilted diagonal portfolio collage.

## Ownership

- `ui/components/cinematic_2/types.go` — Go models (`HeroProps`, `CollageProps`, `CollageItem`, `ShowcaseProps`) and canonical preset defaults.
- `ui/components/cinematic_2/hero_screen.templ` — Screen 1: 100svh video turntable hero, accessible play/pause controls, brand typography, and header navigation.
- `ui/components/cinematic_2/portfolio_collage.templ` — Screen 2: Tilted multi-track marquee carousel with seamless loop and ArtStation link.
- `ui/components/cinematic_2/showcase.templ` — Combined composite component rendering Screen 1 and Screen 2 in sequence.
- `ui/components/cinematic_2/cinematic_2_test.go` — Compact unit tests verifying markup rendering and accessibility attributes.

## Local Contracts

- Zero-GSAP & Zero-Node.js: All animations driven purely by CSS transforms, Alpine.js, and passive scroll listeners.
- 60 FPS performance: Hardware-accelerated transforms (`translate3d`, `rotate`) on composite layers; no layout reflows in tick loops.
- ГОСТ Р 52872-2019 & WCAG 2.2.2: Autoplay video provides an accessible play/pause toggle button; respects `@media (prefers-reduced-motion: reduce)` and `body.a11y-mode` by auto-pausing video and resetting collage rotation to 0deg.
- Infinite Seamless Loop: Collage tracks duplicate items horizontally and clamp offsets smoothly around the midpoint.
- Tailwind v4 integration: Semantic styling (`bg-black`, `text-white`, `border-white/10`, `font-sans`, `font-condensed`).

## Work Guidance

- Compile Templ templates after modifications: `templ generate`.
- Run package unit tests: `go test -v ./ui/components/cinematic_2/...`.
- Rebuild production styles: `tailwindcss -i ./ui/assets/input.css -o ./ui/assets/styles.css --minify`.

## Verification

- `go test -v ./ui/components/cinematic_2/...`
- `make check`

## Child DOX Index

None.
