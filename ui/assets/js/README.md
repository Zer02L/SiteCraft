# 📦 SiteCraft Client Runtime & JS Registry

Вся клиентская JavaScript-инфраструктура движка **SiteCraft** строится на принципах:
* **Zero External CDN Runtime:** Все критические библиотеки скомпилированы в бинарник Go (`ui.AssetsFS()`) и отдаются локально по адресу `/assets/js/*`.
* **100% Offline-Ready & Интранет-готовность:** Сайт не зависит от доступности зарубежных CDN (`unpkg`, `jsdelivr`, `cdnjs`), блокировок РКН или медленного внешнего DNS.
* **PageSpeed 100/100:** Минимальный суммарный объем JS по сети (**< 80 КБ gzip** на весь стек).

---

## 📋 Реестр библиотек и точные версии

| Файл | Библиотека | Точная версия | Размер (raw / gzip) | Назначение |
| :--- | :--- | :--- | :--- | :--- |
| [`htmx.min.js`](htmx.min.js) | **HTMX** | `4.0.0` | 36 KB / ~14 KB | Серверные AJAX-действия, подгрузка контента, формы без перезагрузки |
| [`hx-alpine-compat.js`](hx-alpine-compat.js) | **HTMX Alpine Extension** | `4.0.0` (ext) | 4.2 KB / ~1.5 KB | Бесшовное сохранение состояния Alpine при HTMX DOM-свопах |
| [`alpine.min.js`](alpine.min.js) | **Alpine.js** | `3.17.4` | 55 KB / ~15 KB | Локальное реактивное состояние (модалки, бургер-меню, корзина, табы) |
| [`alpine-mask.min.js`](alpine-mask.min.js) | **@alpinejs/mask** | `3.17.4` | 2.1 KB / ~0.9 KB | Маски ввода номеров телефонов (`+7 (999) 000-00-00`), дат и сумм |
| [`motion.min.js`](motion.min.js) | **Motion One** | `10.18.0` | 25 KB / ~3.8 KB | Аппаратный скролл-скраббинг, шкалы прогресса, параллакс |
| [`ogl.mjs`](ogl.mjs) | **OGL (Minimal WebGL)** | `1.0.11` | 130 KB / ~39 KB | Кинематографические 3D-волны, интерактивные шейдеры, glTF-модели |
| [`canvas-confetti.min.js`](canvas-confetti.min.js) | **Canvas Confetti** | `1.9.4` | 11 KB / ~2.8 KB | Салют/конфетти при конверсии: успешная заявка, оплата заказа |
| [`quill.min.js`](quill.min.js) | **Quill.js** | `2.0.2` | 205 KB / ~45 KB | Визуальный WYSIWYG-редактор статей блога в кабинете `/client` |

---

## 🚫 Антипаттерны: Чего КАТЕГОРИЧЕСКИ НЕЛЬЗЯ добавлять в проект

При разработке новых модулей и секций строго запрещено подключать следующие классы библиотек:

### 1. ⚠️ Монолитные 3D-движки в глобальном бандле (Three.js, Babylon.js)
* **Почему запрещено глобально:** Прямое подключение `three.js` в `<head>` весит ~600 КБ (~170 КБ gzip), немедленно инициализирует тяжелый WebGL-контекст, потребляет до 300 МБ RAM и сбивает Google PageSpeed до 60–80 баллов, вызывая зависания в мобильных браузерах и Telegram Mini Apps.
* **Базовый 3D-стек по умолчанию:**
  * Для процедурных фонов, волн, частиц и архитектурных проекций (Flat-to-3D) — **OGL** (`/assets/js/ogl.mjs`, всего ~39 КБ gzip).
  * Для карточек с наклонами и псевдо-3D параллакса — нативный CSS 3D (`transform-style: preserve-3d`).
* **Когда РАЗРЕШЕНО и РЕКОМЕНДОВАНО использовать Three.js:**
  * Специализированные интерактивные модули: конфигураторы товаров/домов, загрузка моделей glTF/GLB со скелетной анимацией (`AnimationMixer`), сложные PBR-материалы (`MeshPhysicalMaterial` с прозрачностью, шероховатостью, HDR картами окружения) и Raycasting со сложной геометрией.
* **5 железных правил эффективного использования Three.js в SiteCraft:**
  1. **Zero-Global / Lazy Dynamic Import**: Запрещено подключать Three.js в layout страницы. Импорт выполняется исключительно через динамический `import('/assets/js/vendor/three.module.min.js')` только в момент, когда 3D-контейнер приближается к экрану (`x-intersect.once`) или по клику пользователя.
  2. **Zero External CDN (100% Offline-Ready)**: Файлы `three.module.min.js` и нужные аддоны (`GLTFLoader.min.js`, `OrbitControls.min.js`) вендорятся локально в `ui/assets/js/vendor/` и раздаются сервером SiteCraft.
  3. **Demand-Driven Rendering (Рендеринг по событию)**: Запрещен непрерывный цикл `requestAnimationFrame` на статичных сценах! Отрисовка кадра `renderer.render(scene, camera)` вызывается только при вращении камеры пользователем, смене цвета/детали или во время активной интерполяции. При выходе холста из вьюпорта RAF обязательно останавливается (`x-intersect:leave`).
  4. **Strict Memory Cleanup on HTMX (`dispose`)**: WebGL-ресурсы не удаляются сборщиком мусора JS автоматически. При смене экрана по `htmx:beforeCleanupElement` обязателен рекурсивный вызов `.geometry.dispose()`, `.material.dispose()`, `.texture.dispose()` и `renderer.forceContextLoss()`.
  5. **GPU Caps & a11y**: Принудительное ограничение разрешения `Math.min(window.devicePixelRatio, 2)` (защита от перегрева на Retina 3x), поддержка `@media (prefers-reduced-motion: reduce)` и наличие семантического HTML-дублера для скринридеров (ГОСТ).
  * Подробный канонический паттерн и примеры см. в [`.agents/cookbooks/threejs_module.md`](../../.agents/cookbooks/threejs_module.md).

### 2. ❌ GSAP (GreenSock) и ScrollTrigger
* **Почему запрещено:** Весит более 65 КБ gzip, имеет лицензионные ограничения для коммерческих продуктов, постоянно держит слушатели скролла в Main Thread.
* **Что использовать вместо этого:**
  * **Motion One** (`/assets/js/motion.min.js`, всего ~3.8 КБ gzip) — использует нативный Web Animations API (WAAPI), выполняет анимации прямо в потоке композитора GPU без просадок FPS.
  * Нативные CSS-анимации Tailwind v4 (`animate-fade-in`, `@keyframes`).

### 3. ❌ Устаревшие JS-комбайны (Lodash, Underscore, jQuery, Moment.js, Day.js)
* **Почему запрещено:** В современном JavaScript ES2022+ есть нативные `structuredClone`, `Array.prototype.flat`, `Object.fromEntries`, а вся сложная бизнес-логика дат, фильтраций и валют в SiteCraft выполняется на **бэкенде Go** (`time.Time`).
* **Что использовать вместо этого:** Нативный чистый Vanilla JS.

### 4. ❌ Сторонние библиотеки тяжелых графиков (Chart.js, ApexCharts, Highcharts)
* **Почему запрещено:** Бандлы от 100 до 350 КБ ради 1–2 статических графиков в админке или на странице отчетов.
* **Что использовать вместо этого:** Генерация чистого векторного SVG прямо в шаблоне Templ на сервере с последующим кэшированием в `sync.RWMutex`.

### 5. ❌ Прямые внешние CDN-ссылки (`unpkg.com`, `jsdelivr.net`, `cdnjs`)
* **Почему запрещено:**
  * Риск недоступности при блокировках или падении стороннего CDN.
  * Дополнительные задержки на DNS-lookup и TLS Handshake к сторонним хостам.
  * Нарушение принципа единого автономного бинарника.
* **Правило:** Все внешние библиотеки должны сохраняться в `ui/assets/js/` и отдаваться через собственный сервер SiteCraft.

### 6. ❌ Игнорирование очистки памяти при HTMX DOM-swaps
* **Почему опасно:** При смене содержимого страниц через `hx-swap` глобальные таймеры (`setInterval`), анимационные циклы (`requestAnimationFrame`) и WebGL-контексты остаются висеть в памяти браузера, создавая утечки (memory leak).
* **Правило:** Всегда вешать очистку на событие `htmx:beforeCleanupElement`:
  ```javascript
  document.body.addEventListener('htmx:beforeCleanupElement', (e) => {
      if (e.target.contains(container)) {
          cancelAnimationFrame(animId);
          renderer.gl.getExtension('WEBGL_lose_context')?.loseContext();
      }
  }, { once: true });
  ```
