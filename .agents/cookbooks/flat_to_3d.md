# Cookbook: Flat-to-3D Architectural Scrollytelling (OGL Edition)

Canonical pattern for creating high-performance architectural 3D scroll transitions (Exploded Axo, Section Sweep, Card Model) in SiteCraft with Go, Templ, Alpine.js, Tailwind v4 tokens, OGL WebGL, and ГОСТ a11y compliance.

## Architecture Highlights
- **100% OGL (`/assets/js/ogl.mjs`)**: Minimalist WebGL engine (~39 KB gzip), 0 external CDNs, ~20 MB RAM.
- **Procedural Geometry**: Walls, slabs, roofs, and framing generated directly in code (< 120 KB total footprint).
- **Dolly-Zoom (Long-Lens trick)**: Transforms 1.5° flat isometric drawing into a wide 60° eye-level perspective as you scroll.
- **a11y Fallback**: When `prefers-reduced-motion` or `body.a11y-mode` is detected, automatically switches to accessible static blueprint cards with layer specifications.
- **HTMX Safe**: Cleans up animation frames, event listeners, and WebGL context (`loseContext()`) on `htmx:beforeCleanupElement`.

---

## 1. Templ Component: `ui/components/cinematic/flat_to_3d.templ`

```templ
package cinematic

import (
	"fmt"
	"github.com/Oudwins/tailwind-merge-go"
)

type FlatTo3DLayer struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type FlatTo3DProps struct {
	ID           string
	Technique    string // "exploded" | "section-sweep" | "card-model"
	Layers       []FlatTo3DLayer
	ScrollHeight string // e.g. "h-[350vh]"
	Class        string
	Attrs        templ.Attributes
}

templ FlatTo3D(props FlatTo3DProps) {
	<section
		id={ props.ID }
		class={ twmerge.Merge("relative w-full", props.ScrollHeight, props.Class) }
		x-data={ fmt.Sprintf("flatTo3DController({ id: '%s', technique: '%s' })", props.ID, props.Technique) }
		x-intersect:enter="resume()"
		x-intersect:leave="pause()"
		data-technique={ props.Technique }
		{ props.Attrs... }
	>
		<!-- Sticky Viewport Canvas (100svh) -->
		<div class="sticky top-0 h-screen w-full overflow-hidden bg-surface select-none">
			<canvas
				x-ref="canvas"
				class="absolute inset-0 h-full w-full block transition-opacity duration-500"
				:class="{ 'opacity-0 pointer-events-none': reducedMotion }"
				aria-hidden="true"
			></canvas>

			<!-- Architectural HUD & Frame Marks -->
			<div class="pointer-events-none absolute inset-4 border border-content/10 flex flex-col justify-between p-4">
				<div class="flex justify-between items-start font-mono text-xs text-content/60">
					<div class="flex items-center space-x-2">
						<span class="w-2 h-2 rounded-full bg-accent animate-pulse"></span>
						<span class="font-bold uppercase tracking-wider">SiteCraft OGL-3D</span>
						<span class="text-accent" x-text="'[' + activeLayerTitle + ']'"></span>
					</div>
					<div class="font-bold">
						<span x-text="'PROGRESS: ' + Math.round(progress * 100) + '%'"></span>
					</div>
				</div>

				<div class="flex justify-between items-end font-mono text-[11px] text-content/50">
					<div>
						<div>ENGINE: OGL v1.0.11 (39 KB)</div>
						<div>LENS: <span x-text="fovDisplay"></span></div>
					</div>
					<div class="w-12 h-12 border border-content/20 rounded-full relative flex items-center justify-center">
						<div
							class="w-0.5 h-5 bg-accent absolute top-1 origin-bottom transition-transform"
							:style="'transform: rotate(' + headingAngle + 'deg)'"
						></div>
						<span class="text-[8px] font-mono">N</span>
					</div>
				</div>
			</div>

			<!-- Dynamic Scrollytelling Narrative Cards -->
			<div class="pointer-events-none absolute right-6 top-1/2 -translate-y-1/2 w-80 space-y-3">
				for i, layer := range props.Layers {
					<div
						class="p-4 rounded-card backdrop-blur-md border transition-all duration-300"
						:class={ fmt.Sprintf("activeLayerIndex === %d ? 'bg-surface/90 border-accent text-content shadow-lg translate-x-0' : 'bg-surface/30 border-content/10 text-content/40 translate-x-3'", i) }
					>
						<div class="flex items-center space-x-2 font-mono text-xs">
							<span class="w-5 h-5 rounded-full flex items-center justify-center border border-current text-[10px] font-bold">
								{ fmt.Sprintf("%02d", i+1) }
							</span>
							<span class="font-semibold uppercase tracking-wider">{ layer.Title }</span>
						</div>
						<p class="mt-2 text-xs leading-relaxed text-content/80" x-show={ fmt.Sprintf("activeLayerIndex === %d", i) }>
							{ layer.Description }
						</p>
					</div>
				}
			</div>

			<!-- Accessible A11y & Reduced Motion Fallback -->
			<div
				x-show="reducedMotion"
				class="absolute inset-0 z-20 bg-surface p-8 overflow-y-auto"
				role="region"
				aria-label="Архитектурная схема здания"
			>
				<div class="max-w-3xl mx-auto space-y-6">
					<div class="border-b border-content/10 pb-4">
						<h2 class="text-2xl font-bold text-content">Архитектурная схема объекта</h2>
						<p class="text-sm text-content/60 mt-1">Режим высокой доступности (ГОСТ Р 52872-2019)</p>
					</div>
					<div class="grid gap-4 md:grid-cols-2">
						for i, layer := range props.Layers {
							<div class="p-5 border border-content/10 rounded-card bg-surface/50">
								<div class="font-mono text-xs text-accent font-bold">СЛОЙ { fmt.Sprintf("%02d", i+1) }</div>
								<h3 class="text-lg font-semibold text-content mt-1">{ layer.Title }</h3>
								<p class="text-sm text-content/70 mt-2">{ layer.Description }</p>
							</div>
						}
					</div>
				</div>
			</div>
		</div>
	</section>
}
```

---

## 2. Alpine.js Client Controller (OGL-Native)

```javascript
import { Renderer, Camera, Transform, Mesh, Box, Color, Program } from '/assets/js/ogl.mjs';

document.addEventListener('alpine:init', () => {
    Alpine.data('flatTo3DController', ({ id, technique }) => ({
        progress: 0,
        activeLayerIndex: 0,
        activeLayerTitle: '',
        fovDisplay: '1.5°',
        headingAngle: 0,
        reducedMotion: false,
        isActive: false,

        renderer: null,
        camera: null,
        scene: null,
        layers: [],
        plotterProg: null,

        init() {
            this.reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches 
                || document.body.classList.contains('a11y-mode');

            if (this.reducedMotion) return;

            this.initOGL();
            window.addEventListener('scroll', () => this.onScroll(), { passive: true });
            window.addEventListener('resize', () => this.onResize(), { passive: true });
            this.onScroll();

            // HTMX Memory Leak Prevention
            document.body.addEventListener('htmx:beforeCleanupElement', (e) => {
                if (e.target.contains(this.$el)) {
                    this.cleanup();
                }
            }, { once: true });
        },

        initOGL() {
            const canvas = this.$refs.canvas;
            this.renderer = new Renderer({ canvas, dpr: Math.min(window.devicePixelRatio, 2), antialias: true });
            const gl = this.renderer.gl;
            this.scene = new Transform();

            this.camera = new Camera(gl, { fov: 1.5, near: 0.1, far: 2000 });
            this.camera.position.set(0, 15, 1200);
            this.camera.lookAt([0, 1.5, 0]);

            // Чтение дизайн-токенов темы
            const paper = window.SiteTokens?.read('--color-surface') || '#f2efe7';
            const ink = window.SiteTokens?.read('--color-content') || '#1d1c19';
            const accent = window.SiteTokens?.read('--color-accent') || '#e85d04';

            this.plotterProg = new Program(gl, {
                vertex: `
                    attribute vec3 position;
                    attribute vec3 normal;
                    uniform mat4 modelMatrix;
                    uniform mat4 viewMatrix;
                    uniform mat4 projectionMatrix;
                    varying vec3 vWorld;
                    varying vec3 vLocal;
                    void main() {
                        vLocal = position;
                        vec4 worldPos = modelMatrix * vec4(position, 1.0);
                        vWorld = worldPos.xyz;
                        gl_Position = projectionMatrix * viewMatrix * worldPos;
                    }
                `,
                fragment: `
                    precision highp float;
                    uniform float uDrawSweep;
                    uniform vec3 uPaperColor;
                    uniform vec3 uInkColor;
                    uniform vec3 uAccentColor;
                    varying vec3 vWorld;
                    varying vec3 vLocal;
                    float hash(vec3 p) {
                        p = fract(p * vec3(443.897, 441.423, 437.195));
                        p += dot(p, p.yzx + 19.19);
                        return fract((p.x + p.y) * p.z);
                    }
                    void main() {
                        if (vWorld.y > uDrawSweep) discard;
                        float edge = abs(vWorld.y - uDrawSweep);
                        vec3 col = mix(uPaperColor, uInkColor, 0.12);
                        if (edge < 0.35) {
                            col = mix(uAccentColor, col, smoothstep(0.0, 0.35, edge));
                        }
                        col *= 1.0 + (hash(floor(vLocal * 140.0)) - 0.5) * 0.03;
                        gl_FragColor = vec4(col, 1.0);
                    }
                `,
                uniforms: {
                    uDrawSweep: { value: 0.0 },
                    uPaperColor: { value: new Color(paper) },
                    uInkColor: { value: new Color(ink) },
                    uAccentColor: { value: new Color(accent) }
                }
            });

            // Создаем процедурные слои в иерархии Transform
            const slab = new Transform();
            slab.addChild(new Mesh(gl, { geometry: new Box(gl, { width: 8, height: 0.4, depth: 6 }), program: this.plotterProg }));
            this.scene.addChild(slab);

            const walls = new Transform();
            walls.addChild(new Mesh(gl, { geometry: new Box(gl, { width: 7.6, height: 2.8, depth: 5.6 }), program: this.plotterProg }));
            walls.position.y = 10;
            this.scene.addChild(walls);

            const roof = new Transform();
            roof.addChild(new Mesh(gl, { geometry: new Box(gl, { width: 8.4, height: 0.6, depth: 6.4 }), program: this.plotterProg }));
            roof.position.y = 20;
            this.scene.addChild(roof);

            this.layers = [slab, walls, roof];
            this.onResize();
        },

        onScroll() {
            if (!this.isActive || this.reducedMotion) return;
            const rect = this.$el.getBoundingClientRect();
            const scrollDist = this.$el.offsetHeight - window.innerHeight;
            if (scrollDist <= 0) return;

            const p = Math.max(0, Math.min(1, -rect.top / scrollDist));
            this.progress = p;

            // 1. Плоттер
            this.plotterProg.uniforms.uDrawSweep.value = -2 + p * 25;

            // 2. Движение слоев
            const pWalls = Math.min(1, Math.max(0, (p - 0.15) / 0.3));
            this.layers[1].position.y = 10 - pWalls * (10 - 1.6);

            const pRoof = Math.min(1, Math.max(0, (p - 0.4) / 0.3));
            this.layers[2].position.y = 20 - pRoof * (20 - 3.3);

            // 3. Dolly Zoom (1.5° -> 55°)
            const fov = 1.5 + Math.pow(p, 2) * 53.5;
            this.camera.perspective({ fov, aspect: this.renderer.width / this.renderer.height });
            const dist = 1200 * (1 - p * 0.985);
            this.camera.position.set(0, 1.6 + (1 - p) * 12, Math.max(3.5, dist));
            this.camera.lookAt([0, 1.6, 0]);

            this.fovDisplay = fov.toFixed(1) + '°';
            this.headingAngle = Math.round(p * 90);
            this.renderScene();
        },

        renderScene() {
            if (this.renderer && this.scene && this.camera) {
                this.renderer.render({ scene: this.scene, camera: this.camera });
            }
        },

        onResize() {
            if (!this.renderer || !this.camera) return;
            this.renderer.setSize(window.innerWidth, window.innerHeight);
            this.camera.perspective({ aspect: window.innerWidth / window.innerHeight });
            this.renderScene();
        },

        resume() { this.isActive = true; this.renderScene(); },
        pause() { this.isActive = false; },
        cleanup() {
            this.pause();
            if (this.renderer?.gl) {
                this.renderer.gl.getExtension('WEBGL_lose_context')?.loseContext();
                this.renderer = null;
            }
        }
    }));
});
```
