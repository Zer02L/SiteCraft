---
name: flat-to-3d
description: Architectural 3D scrollytelling skill for SiteCraft (Go + Templ + Tailwind v4 + Alpine.js + HTMX). Transforms flat 2D drawings (exploded axo, section sweep, open section, build time-lapse, schedule build, card model, sunlit elevation) into interactive 3D experiences with OGL WebGL, dolly-zoom, procedural shaders, Zero-CDN, and ГОСТ a11y compliance.
---

# Flat-to-3D Architectural Scrollytelling (SiteCraft Edition — OGL Engine)

Ты — **Senior Creative 3D & Scrollytelling Engineer** движка **SiteCraft**.
Твоя задача — проектировать, адаптировать и внедрять кинематографические интерактивные 3D-переходы из плоских архитектурных чертежей (аксонометрия, разрезы, календарный план, фасад, макет) в объемные пространственные сцены по методологии **Flat-to-3D**.

В качестве базового 3D-движка в SiteCraft строго используется **OGL** (`/assets/js/ogl.mjs`, ~39 KB gzip) — минималистичный WebGL-движок без раздутого Three.js рантайма, без сторонних CDN, с чистыми GLSL-шейдерами, поддержкой дизайн-токенов темы, ГОСТ a11y-доступностью и PageSpeed 100/100.

---

## 🏛️ КЛЮЧЕВАЯ КОНЦЕПЦИЯ: СЕМЬ ТЕХНИК FLAT-TO-3D

| # | Техника | Архитектурная метафора | Механика в OGL и камере |
|---|---|---|---|
| **01** | **Exploded House** | Взрыв-аксонометрия каркаса или здания | Слои смонтированы в иерархии `Transform` (плита, стены, остекление, планкен, стропила, кровля). Плоттерный скан вычерчивает чернила. При скролле слои последовательно опускаются по `position.y`. Долли-зум сужает дистанцию и расширяет угол линзы (1.5° → 60°), превращая чертеж в перспективу комнаты на уровне глаз, после чего начинается обход комнат. |
| **02** | **Section Sweep** | Архитектурный разрез с проявлением глубины | Чертеж сечения открывается в глубину по оси Z. Сметающий лазер/сканер (`uSweepZ`) в шейдере OGL переводит чертежные контуры в физические материалы (дерево, бетон, стекло), а пунктирные линии инсоляции — в объемные солнечные лучи. Камера заходит внутрь прямо сквозь разрез. |
| **03** | **Open Section** | Раскрытие разреза с фиксацией масштаба | Разрез строится плоттером. Объектив расширяется, сохраняя неизменный видимый размер линии реза (`dolly zoom lock`). Камера в OGL последовательно перемещается по этажам от плоскости сечения. |
| **04** | **Build Time-Lapse** | Поэтапный таймлапс стройки | Сцена начинается с листа из 6 чертежей последовательных фаз возведения. При скролле камера ныряет в первый чертеж прямо на стройплощадку, и здание растет неделя за неделей вплоть до сдачи объекта. |
| **05** | **Schedule Build** | Календарный график Ганта в 3D | Скроллинг интерактивного таймлайна диаграммы Ганта возводит здание балка за балкой с движением башенного крана через интерполяцию положений OGL `Transform`. |
| **06** | **Card Model** | Макет из белого картона на столе архитектора | Белый макет на ночном рабочем столе. Камера пикирует к масштабной фигурке человека, и картонная модель трансформируется в настоящий полноразмерный дом через радиальный шейдерный фронт `length(vWorld - uCenter) < uRadius`. |
| **07** | **Sunlit Elevation** | Главный фасад с расчетом инсоляции | Ортогональный чертеж фасада, где штриховка является математически точной тенью (`dot(vNormal, uSunDir)`). Сметающий скан «овеществляет» кирпичную кладку, солнце меняет азимут и высоту, дверь распахивается, и зритель входит в холл. |

---

## ⚡ ПОЧЕМУ OGL ВМЕСТО THREE.JS В SITECRAFT

| Критерий | OGL (`/assets/js/ogl.mjs`) | Three.js |
|---|---|---|
| **Вес в сети (gzip)** | **~39 КБ** (встроен локально в проект) 🏆 | ~170–210 КБ (раздувает бандл) ❌ |
| **Архитектурный контракт** | Разрешен и рекомендован в `ui/assets/js/` 🏆 | В списке нежелательных антипаттернов ⚠️ |
| **Потребление RAM** | **~15–25 МБ** 🏆 | 120–250 МБ (микрофризы на мобилках) ❌ |
| **Кастомные шейдеры** | Прямой нативный GLSL `Program` без костылей 🏆 | Требует строкового патчинга `onBeforeCompile` ❌ |
| **Контроль WebGL** | Полный контроль над буферами, VAO и очисткой контекста 🏆 | Тяжелый скрытый граф материалов ❌ |

---

## 📐 МАТЕМАТИКА И АЛГОРИТМЫ FLAT-TO-3D НА OGL

### 1. Формула Dolly-Zoom («Эффект Вертиго») на камере OGL
Превращение плоского изометрического чертежа в объемную перспективу комнаты:
$$w = 2 \cdot d \cdot \tan\left(\frac{\theta}{2}\right) = \text{const}$$

В OGL это реализуется через метод `camera.perspective({ fov, aspect })`:
```javascript
// Нормализованный прогресс скролла p в диапазоне [0, 1]
const p = Math.max(0, Math.min(1, scrollProgress));

// Плавный переход FOV от 1.5° (телеобъектив ~1.2 км) до 60° (глаз человека)
const fov = 1.5 + Math.pow(p, 2) * 58.5;
const aspect = gl.canvas.width / gl.canvas.height;

// Сохраняем видимый размер кадра на чертеже
const targetDist = 1200 / Math.tan((fov * Math.PI) / 360);

camera.perspective({ fov, aspect });
camera.position.set(0, 1.6 + (1 - p) * 15, targetDist * (1 - p * 0.95));
camera.lookAt([0, 1.6, 0]);
```

### 2. Функция оконного скроллинга `win(p, a, b)`
```javascript
export const clamp = (v, min = 0, max = 1) => Math.min(max, Math.max(min, v));
export const win = (p, a, b) => clamp((p - a) / (b - a));
export const easeOutCubic = t => 1 - Math.pow(1 - t, 3);
```

### 3. Плоттерный шейдер и процедурный шум бумаги в OGL `Program`
Никаких строковых патчей Three.js — чистый, компилируемый за 2 мс шейдер OGL:
```javascript
import { Program, Color } from '/assets/js/ogl.mjs';

export function createPlotterProgram(gl, { paperColor, inkColor, accentColor }) {
    return new Program(gl, {
        vertex: `
            attribute vec3 position;
            attribute vec3 normal;
            uniform mat4 modelMatrix;
            uniform mat4 viewMatrix;
            uniform mat4 projectionMatrix;

            varying vec3 vWorld;
            varying vec3 vNormal;
            varying vec3 vLocal;

            void main() {
                vLocal = position;
                vNormal = normal;
                vec4 worldPos = modelMatrix * vec4(position, 1.0);
                vWorld = worldPos.xyz;
                gl_Position = projectionMatrix * viewMatrix * worldPos;
            }
        `,
        fragment: `
            precision highp float;

            uniform float uDrawSweep;   // фронт сканирования плоттера (Y или Z)
            uniform vec3 uPaperColor;   // цвет бумаги
            uniform vec3 uInkColor;     // цвет контуров чернил
            uniform vec3 uAccentColor;  // оранжевый маркер наконечника рейсфедера
            uniform float uGrain;       // интенсивность фактуры бумаги

            varying vec3 vWorld;
            varying vec3 vNormal;
            varying vec3 vLocal;

            // Процедурный шум бумаги (без внешних текстур)
            float hashPaper(vec3 p) {
                p = fract(p * vec3(443.897, 441.423, 437.195));
                p += dot(p, p.yzx + 19.19);
                return fract((p.x + p.y) * p.z);
            }

            void main() {
                // Отсекаем не начерченную плоттером часть
                if (vWorld.y > uDrawSweep) discard;

                // Акцентная светящаяся полоса на острие черчения
                float edge = abs(vWorld.y - uDrawSweep);
                vec3 col = mix(uPaperColor, uInkColor, 0.12);

                if (edge < 0.35) {
                    col = mix(uAccentColor, col, smoothstep(0.0, 0.35, edge));
                }

                // Наложение процедурного зерна
                float grain = (hashPaper(floor(vLocal * 150.0)) - 0.5) * uGrain;
                col *= (1.0 + grain);

                gl_FragColor = vec4(col, 1.0);
            }
        `,
        uniforms: {
            uDrawSweep: { value: 0.0 },
            uPaperColor: { value: new Color(paperColor || '#f2efe7') },
            uInkColor: { value: new Color(inkColor || '#1d1c19') },
            uAccentColor: { value: new Color(accentColor || '#e85d04') },
            uGrain: { value: 0.025 }
        }
    });
}
```

### 4. Чертежные контуры и ребра в OGL (`WireMesh`)
Для создания эстетики чертежа тушью используй встроенный в OGL класс `WireMesh` или передавай геометрии в режиме `gl.LINES`:
```javascript
import { Mesh, WireMesh, Box } from '/assets/js/ogl.mjs';

// Каркас стены: основное тело + четкие контурные ребра чертежа
const boxGeom = new Box(gl, { width: 4, height: 2.8, depth: 0.3 });
const wallMesh = new Mesh(gl, { geometry: boxGeom, program: plotterProgram });
const wireEdges = new WireMesh(gl, { geometry: boxGeom, program: lineProgram });
wallMesh.addChild(wireEdges);
```

---

## 🛠️ ЭТАЛОННАЯ ИНТЕГРАЦИЯ В SITECRAFT (TEMPL + ALPINE.JS)

### 1. Templ-компонент (`ui/components/cinematic/flat_to_3d.templ`)
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
}

templ FlatTo3D(props FlatTo3DProps) {
	<section
		id={ props.ID }
		class={ twmerge.Merge("relative w-full", props.ScrollHeight, props.Class) }
		x-data={ fmt.Sprintf("flatTo3DController({ id: '%s', technique: '%s' })", props.ID, props.Technique) }
		x-intersect:enter="resume()"
		x-intersect:leave="pause()"
		data-technique={ props.Technique }
	>
		<!-- 100svh Sticky Viewport -->
		<div class="sticky top-0 h-screen w-full overflow-hidden bg-surface select-none">
			<!-- OGL Canvas -->
			<canvas
				x-ref="canvas"
				class="absolute inset-0 h-full w-full block transition-opacity duration-500"
				:class="{ 'opacity-0 pointer-events-none': reducedMotion }"
				aria-hidden="true"
			></canvas>

			<!-- Чертежная CAD-рамка -->
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
					<!-- Мини-радар ориентации -->
					<div class="w-12 h-12 border border-content/20 rounded-full relative flex items-center justify-center">
						<div
							class="w-0.5 h-5 bg-accent absolute top-1 origin-bottom transition-transform"
							:style="'transform: rotate(' + headingAngle + 'deg)'"
						></div>
						<span class="text-[8px] font-mono">N</span>
					</div>
				</div>
			</div>

			<!-- Информационные плашки шагов -->
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

			<!-- ГОСТ Р 52872-2019 / Reduced Motion Fallback -->
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

### 2. Клиентский OGL-контроллер на Alpine.js
```javascript
import { Renderer, Camera, Transform, Mesh, Box, Color } from '/assets/js/ogl.mjs';

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
        rafId: null,

        init() {
            this.reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches 
                || document.body.classList.contains('a11y-mode');

            if (this.reducedMotion) return;

            this.initOGL();
            window.addEventListener('scroll', () => this.onScroll(), { passive: true });
            window.addEventListener('resize', () => this.onResize(), { passive: true });
            this.onScroll();

            // Безопасная очистка памяти при HTMX DOM-swaps
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

            // Чтение цветов через мост дизайн-токенов SiteTokens
            const paper = window.SiteTokens?.read('--color-surface') || '#f2efe7';
            const ink = window.SiteTokens?.read('--color-content') || '#1d1c19';
            const accent = window.SiteTokens?.read('--color-accent') || '#e85d04';

            this.plotterProg = createPlotterProgram(gl, { paperColor: paper, inkColor: ink, accentColor: accent });

            // Процедурная сборка слоев (Фундамент, Стены, Кровля)
            this.buildLayers(gl);
            this.onResize();
        },

        buildLayers(gl) {
            // Слой 1: Плита
            const slabNode = new Transform();
            const slabMesh = new Mesh(gl, { geometry: new Box(gl, { width: 8, height: 0.4, depth: 6 }), program: this.plotterProg });
            slabNode.addChild(slabMesh);
            this.scene.addChild(slabNode);

            // Слой 2: Стены
            const wallsNode = new Transform();
            const wallsMesh = new Mesh(gl, { geometry: new Box(gl, { width: 7.6, height: 2.8, depth: 5.6 }), program: this.plotterProg });
            wallsNode.addChild(wallsMesh);
            wallsNode.position.y = 10; // начальное смещение аксонометрии
            this.scene.addChild(wallsNode);

            // Слой 3: Кровля
            const roofNode = new Transform();
            const roofMesh = new Mesh(gl, { geometry: new Box(gl, { width: 8.4, height: 0.6, depth: 6.4 }), program: this.plotterProg });
            roofNode.addChild(roofMesh);
            roofNode.position.y = 20; // парит высоко
            this.scene.addChild(roofNode);

            this.layers = [
                { node: slabNode, baseY: 0.2, floatY: 0.2 },
                { node: wallsNode, baseY: 1.6, floatY: 10 },
                { node: roofNode, baseY: 3.3, floatY: 20 }
            ];
        },

        onScroll() {
            if (!this.isActive || this.reducedMotion) return;
            const rect = this.$el.getBoundingClientRect();
            const scrollDistance = this.$el.offsetHeight - window.innerHeight;
            if (scrollDistance <= 0) return;

            const p = Math.max(0, Math.min(1, -rect.top / scrollDistance));
            this.progress = p;

            this.updateScene(p);
            this.renderScene();
        },

        updateScene(p) {
            // 1. Анимация плоттера
            this.plotterProg.uniforms.uDrawSweep.value = -2 + p * 25;

            // 2. Приземление слоев поочередно
            const pWalls = win(p, 0.15, 0.45);
            this.layers[1].node.position.y = 10 - pWalls * (10 - 1.6);

            const pRoof = win(p, 0.4, 0.7);
            this.layers[2].node.position.y = 20 - pRoof * (20 - 3.3);

            // 3. Dolly Zoom камеры: 1.5° -> 55°
            const fov = 1.5 + Math.pow(p, 2) * 53.5;
            this.camera.perspective({ fov, aspect: this.renderer.width / this.renderer.height });
            const dist = 1200 * (1 - p * 0.985);
            this.camera.position.set(0, 1.6 + (1 - p) * 12, Math.max(3.5, dist));
            this.camera.lookAt([0, 1.6, 0]);

            // HUD
            this.fovDisplay = fov.toFixed(1) + '°';
            this.headingAngle = Math.round(p * 90);
            this.activeLayerIndex = p < 0.35 ? 0 : p < 0.65 ? 1 : 2;
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

        resume() {
            this.isActive = true;
            this.renderScene();
        },

        pause() {
            this.isActive = false;
        },

        cleanup() {
            this.pause();
            if (this.renderer && this.renderer.gl) {
                this.renderer.gl.getExtension('WEBGL_lose_context')?.loseContext();
                this.renderer = null;
            }
        }
    }));
});
```

---

## 🎯 СТАРТОВЫЕ ПРОМПТЫ ДЛЯ АГЕНТОВ (OGL-EDITION)

### Промпт: 01-Exploded-House (Каркасные и модульные дома / Префаб)
```text
Спроектируй скроллинг-секцию Flat-to-3D на движке OGL (/assets/js/ogl.mjs) для компании модульных домов.
Архитектурный сюжет:
1. Открывается чертежная аксонометрия: дом разобран на 3 парящих слоя (фундаментная плита, клееный каркас со стенами, фальцевая кровля).
2. Плоттерный скан на OGL (Program с uDrawSweep и процедурным шумом бумаги) чертит контуры снизу вверх.
3. При скролле слои последовательно садятся на свои места.
4. Долли-зум на камере OGL (1.5° -> 55°) приближает зрителя к крыльцу на уровень глаз.
Все цвета через дизайн-токены var(--color-surface), var(--color-content), var(--color-accent).
Обязателен a11y-фоллбек при prefers-reduced-motion и очистка памяти при htmx:beforeCleanupElement.
```

---

## 📋 ЧЕК-ЛИСТ ПРИЕМКИ
1. [ ] **Движок**: Использован исключительно OGL (`/assets/js/ogl.mjs`), никаких внешних CDN и Three.js.
2. [ ] **Вес страницы**: < 100–120 КБ (PageSpeed 100/100).
3. [ ] **0% CPU вне экрана**: Рендеринг вызывается только по событиям скролла или паузится через `x-intersect`.
4. [ ] **Zero Memory Leak**: При вызове `htmx:beforeCleanupElement` вызывается `loseContext()`.
5. [ ] **Доступность**: При `body.a11y-mode` или `prefers-reduced-motion` сцена переключается в двумерный режим экспликации.
