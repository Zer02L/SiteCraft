# Cookbook: Isolated On-Demand Three.js Module (SiteCraft Edition)

Canonical pattern for integrating Three.js for specialized 3D tasks (interactive glTF/GLB product configurators, skeletal rigging, and realistic PBR materials) without compromising PageSpeed 100/100, Zero-External-CDN, or ГОСТ a11y compliance.

---

## 🎯 Когда использовать Three.js вместо OGL
- **OGL (`/assets/js/ogl.mjs`, ~39 KB)** — базовый движок SiteCraft для процедурных фонов, волн, частиц, плоттерных эффектов и аксонометрических скролл-трансформаций (Flat-to-3D).
- **Three.js (`/assets/js/vendor/three.module.min.js`)** — подключается **только On-Demand** для:
  1. Загрузки сложных 3D-моделей в формате glTF/GLB с Draco/Meshopt компрессией.
  2. Скелетной анимации (`AnimationMixer`, скиннинг, морфинг).
  3. Физически корректных материалов PBR (`MeshPhysicalMaterial` с прозрачностью `transmission`, лакировкой `clearcoat`, шероховатостью и IBL картами окружения).
  4. Интерактивного Raycasting с пространственным ускорением (`three-mesh-bvh`) для выбора деталей кликом.

---

## 🏛️ 5 Железных правил Three.js в SiteCraft

1. **Zero-Global (Dynamic Import Only)**: Three.js никогда не подключается в `<head>` страницы. Загрузка происходит строго через динамический `import()` по вхождению вьюпорта (`x-intersect.once`) или по клику пользователя.
2. **Zero External CDN**: Файлы `three.module.min.js`, `GLTFLoader.min.js`, `OrbitControls.min.js` вендорятся локально в `ui/assets/js/vendor/` и отдаются бинарником Go (`ui.AssetsFS()`).
3. **Demand-Driven Rendering**: Запрещен непрерывный цикл `requestAnimationFrame` на 60/120 FPS при статичной сцене! Кадр рендерится только при перемещении камеры, смене детали или активной интерполяции. При скролле за пределы экрана RAF замораживается (`x-intersect:leave`).
4. **Strict Memory Cleanup on HTMX (`dispose`)**: Обязательная подписка на событие `htmx:beforeCleanupElement` с рекурсивным освобождением буферов геометрий, текстур, материалов и вызовом `renderer.forceContextLoss()`.
5. **a11y & DevicePixelRatio Cap**: Разрешение рендера ограничивается `Math.min(window.devicePixelRatio, 2)` (защита мобильных GPU от перегрева на Retina 3x), обязателен семантический HTML-дублер и поддержка `@media (prefers-reduced-motion: reduce)`.

---

## 1. Templ Компонент: `ui/components/cinematic/three_model_viewer.templ`

```templ
package cinematic

import (
	"fmt"
	"github.com/Oudwins/tailwind-merge-go"
)

type ThreeViewerProps struct {
	ID          string
	ModelURL    string // e.g. "/assets/models/house.glb"
	PosterURL   string // e.g. "/assets/img/house-poster.webp"
	Title       string
	Description string
	Class       string
	Attrs       templ.Attributes
}

templ ThreeModelViewer(props ThreeViewerProps) {
	<div
		id={ props.ID }
		class={ twmerge.Merge("relative w-full aspect-16/10 rounded-card overflow-hidden bg-surface-subtle border border-content/10 select-none", props.Class) }
		x-data={ fmt.Sprintf("threeViewerController({ id: '%s', modelUrl: '%s' })", props.ID, props.ModelURL) }
		x-intersect.margin.200px.once="initViewer()"
		x-intersect:enter="resume()"
		x-intersect:leave="pause()"
		role="region"
		aria-label={ props.Title }
		{ props.Attrs... }
	>
		<!-- 1. Фоновый постер (быстрый LCP до загрузки 3D) -->
		<img
			x-show="!loaded"
			src={ props.PosterURL }
			alt={ props.Title }
			class="absolute inset-0 w-full h-full object-cover transition-opacity duration-500"
			loading="lazy"
		/>

		<!-- 2. Индикатор загрузки WebGL -->
		<div
			x-show="loading"
			class="absolute inset-0 flex flex-col items-center justify-center bg-surface/80 backdrop-blur-xs gap-3 z-10"
		>
			<div class="w-8 h-8 rounded-full border-2 border-accent border-t-transparent animate-spin"></div>
			<span class="text-xs font-mono text-content-muted" x-text="'Загрузка 3D-модели: ' + progress + '%'"></span>
		</div>

		<!-- 3. WebGL Canvas -->
		<canvas
			x-ref="canvas"
			class="absolute inset-0 w-full h-full block cursor-grab active:cursor-grabbing transition-opacity duration-700"
			:class="{ 'opacity-0': !loaded, 'opacity-100': loaded }"
		></canvas>

		<!-- 4. Семантический a11y-дублер и доступный HUD (ГОСТ) -->
		<div class="pointer-events-none absolute inset-x-4 bottom-4 flex justify-between items-end z-20">
			<div class="bg-surface/90 backdrop-blur-md border border-content/10 p-3 rounded-card text-xs max-w-sm pointer-events-auto">
				<div class="font-bold text-content">{ props.Title }</div>
				<p class="text-content-muted mt-1 leading-snug">{ props.Description }</p>
				<div class="mt-2 flex gap-2">
					<button
						type="button"
						@click="rotateStep(-90)"
						class="px-2 py-1 bg-surface-subtle border border-content/10 rounded text-[11px] hover:border-accent text-content focus-visible:ring-2 focus-visible:ring-accent"
						aria-label="Повернуть модель влево на 90 градусов"
					>
						⟲ Влево
					</button>
					<button
						type="button"
						@click="rotateStep(90)"
						class="px-2 py-1 bg-surface-subtle border border-content/10 rounded text-[11px] hover:border-accent text-content focus-visible:ring-2 focus-visible:ring-accent"
						aria-label="Повернуть модель вправо на 90 градусов"
					>
						Вправо ⟳
					</button>
					<button
						type="button"
						@click="resetCamera()"
						class="px-2 py-1 bg-surface-subtle border border-content/10 rounded text-[11px] hover:border-accent text-content focus-visible:ring-2 focus-visible:ring-accent"
						aria-label="Сбросить положение камеры"
					>
						Сброс
					</button>
				</div>
			</div>

			<!-- Бейдж статуса -->
			<div class="hidden sm:flex items-center gap-2 font-mono text-[10px] text-content-muted bg-surface/80 backdrop-blur-xs px-2.5 py-1 rounded border border-content/10">
				<span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
				<span>PBR 3D READY</span>
			</div>
		</div>
	</div>
}
```

---

## 2. Клиентский контроллер Alpine.js: `ui/assets/js/modules/three-viewer.js`

```javascript
/**
 * Изолированный On-Demand Three.js контроллер для SiteCraft.
 * Спроектирован с учетом PageSpeed 100/100, Demand-Rendering и HTMX Cleanup.
 */
document.addEventListener('alpine:init', () => {
  Alpine.data('threeViewerController', ({ id, modelUrl }) => ({
    id,
    modelUrl,
    loading: false,
    loaded: false,
    progress: 0,
    reducedMotion: false,

    // Внутренние ссылки (не реактивные)
    _three: null,
    _renderer: null,
    _scene: null,
    _camera: null,
    _controls: null,
    _model: null,
    _rafId: null,
    _needsRender: false,
    _isActive: false,

    init() {
      this.reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

      // Регистрация очистки ресурсов перед свопом страницы в HTMX
      document.body.addEventListener('htmx:beforeCleanupElement', (evt) => {
        if (evt.target.contains(this.$el)) {
          this.destroy();
        }
      }, { once: true });
    },

    // 1. Ленивая загрузка Three.js только при вхождении во вьюпорт
    async initViewer() {
      if (this.loaded || this.loading) return;
      this.loading = true;

      try {
        // Динамический импорт локально скомпилированных ES модулей (Zero CDN)
        const [THREE, { GLTFLoader }, { OrbitControls }] = await Promise.all([
          import('/assets/js/vendor/three.module.min.js'),
          import('/assets/js/vendor/GLTFLoader.min.js'),
          import('/assets/js/vendor/OrbitControls.min.js')
        ]);

        this._three = THREE;
        this._setupScene(THREE, OrbitControls);
        await this._loadModel(GLTFLoader);

        this.loaded = true;
        this.loading = false;
        this._isActive = true;
        this.requestRender();
      } catch (err) {
        console.error('[SiteCraft 3D] Failed to load Three.js module:', err);
        this.loading = false;
      }
    },

    _setupScene(THREE, OrbitControls) {
      const canvas = this.$refs.canvas;
      const width = this.$el.clientWidth;
      const height = this.$el.clientHeight;

      // Сцена
      this._scene = new THREE.Scene();

      // Камера
      this._camera = new THREE.PerspectiveCamera(45, width / height, 0.1, 100);
      this._camera.position.set(0, 1.5, 4);

      // Освещение (PBR Key-Fill-Rim)
      const ambientLight = new THREE.AmbientLight(0xffffff, 0.8);
      this._scene.add(ambientLight);

      const dirLight = new THREE.DirectionalLight(0xffffff, 1.5);
      dirLight.position.set(5, 10, 7);
      this._scene.add(dirLight);

      // Рендерер с ограничением DPR (защита мобильных устройств)
      this._renderer = new THREE.WebGLRenderer({
        canvas,
        alpha: true,
        antialias: true,
        powerPreference: 'high-performance'
      });
      this._renderer.setSize(width, height);
      this._renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
      this._renderer.toneMapping = THREE.ACESFilmicToneMapping;
      this._renderer.toneMappingExposure = 1.0;

      // Контроллер камеры
      this._controls = new OrbitControls(this._camera, canvas);
      this._controls.enableDamping = true;
      this._controls.dampingFactor = 0.05;
      this._controls.maxPolarAngle = Math.PI / 2 + 0.05; // запрет ухода под землю
      this._controls.minDistance = 1.5;
      this._controls.maxDistance = 10;
      this._controls.autoRotate = !this.reducedMotion;
      this._controls.autoRotateSpeed = 1.2;

      // Demand-Rendering: рендерим ТОЛЬКО при взаимодействии!
      this._controls.addEventListener('change', () => this.requestRender());

      // ResizeObserver для адаптивности
      this._resizeObserver = new ResizeObserver(() => this._onResize());
      this._resizeObserver.observe(this.$el);
    },

    _loadModel(GLTFLoader) {
      return new Promise((resolve, reject) => {
        const loader = new GLTFLoader();
        loader.load(
          this.modelUrl,
          (gltf) => {
            this._model = gltf.scene;

            // Центрирование модели
            const box = new this._three.Box3().setFromObject(this._model);
            const center = box.getCenter(new this._three.Vector3());
            this._model.position.sub(center);

            this._scene.add(this._model);
            resolve();
          },
          (xhr) => {
            if (xhr.lengthComputable) {
              this.progress = Math.round((xhr.loaded / xhr.total) * 100);
            }
          },
          (err) => reject(err)
        );
      });
    },

    _onResize() {
      if (!this._renderer || !this._camera) return;
      const width = this.$el.clientWidth;
      const height = this.$el.clientHeight;
      this._camera.aspect = width / height;
      this._camera.updateProjectionMatrix();
      this._renderer.setSize(width, height);
      this.requestRender();
    },

    // 2. Рендеринг по требованию (Demand-Driven Rendering)
    requestRender() {
      if (!this._isActive || !this._renderer) return;

      if (!this._needsRender) {
        this._needsRender = true;
        this._rafId = requestAnimationFrame(() => {
          this._needsRender = false;
          if (this._controls.autoRotate || this._controls.enableDamping) {
            this._controls.update();
          }
          this._renderer.render(this._scene, this._camera);

          // Если авто-вращение активно, продолжаем кадр
          if (this._controls.autoRotate) {
            this.requestRender();
          }
        });
      }
    },

    // a11y управление с клавиатуры
    rotateStep(degrees) {
      if (!this._controls) return;
      const rad = (degrees * Math.PI) / 180;
      this._camera.position.applyAxisAngle(new this._three.Vector3(0, 1, 0), rad);
      this._controls.update();
      this.requestRender();
    },

    resetCamera() {
      if (!this._camera || !this._controls) return;
      this._camera.position.set(0, 1.5, 4);
      this._controls.target.set(0, 0, 0);
      this._controls.update();
      this.requestRender();
    },

    // Остановка при скролле вне экрана (0% CPU/GPU)
    resume() {
      this._isActive = true;
      if (this.loaded) this.requestRender();
    },

    pause() {
      this._isActive = false;
      if (this._rafId) cancelAnimationFrame(this._rafId);
    },

    // 3. Строгая очистка памяти при HTMX свопе (Zero Memory Leak)
    destroy() {
      this.pause();
      if (this._resizeObserver) {
        this._resizeObserver.disconnect();
      }

      if (this._controls) {
        this._controls.dispose();
      }

      if (this._scene) {
        this._scene.traverse((node) => {
          if (node.isMesh) {
            if (node.geometry) node.geometry.dispose();
            if (node.material) {
              const materials = Array.isArray(node.material) ? node.material : [node.material];
              materials.forEach((mat) => {
                // Очистка всех подключенных текстур
                Object.keys(mat).forEach((prop) => {
                  if (mat[prop] && typeof mat[prop].dispose === 'function') {
                    mat[prop].dispose();
                  }
                });
                mat.dispose();
              });
            }
          }
        });
      }

      if (this._renderer) {
        this._renderer.dispose();
        this._renderer.forceContextLoss();
        this._renderer.domElement.remove();
      }

      this._scene = null;
      this._camera = null;
      this._renderer = null;
      this._three = null;
    }
  }));
});
```

---

## 3. Пайплайн оптимизации 3D-ассетов (Asset Pipeline)

Файл `.glb` размером > 5 МБ разрушает скорость первого взаимодействия. Перед добавлением файла в проект выполни оптимизацию через `gltf-transform`:

```bash
# Установка cli (разово в среде разработки)
npm install -g @gltf-transform/cli

# Оптимизация модели: Draco/Meshopt компрессия геометрии + WebP сжатие текстур
gltf-transform optimize input_model.glb output_model.glb \
  --draco \
  --texture-compress webp \
  --texture-size 1024
```

### Бюджеты для 3D-ассетов в SiteCraft:
- **Размер GLB файла**: до **1.5–2.5 МБ** (с Draco/Meshopt).
- **Разрешение текстур**: не более **1024x1024** для материалов деталей, **2048x2048** только для главного кузова/фасада.
- **Количество треугольников (Polycount)**: до 100 000 полигонов на всю модель.
