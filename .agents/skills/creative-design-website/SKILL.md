---
name: creative-design-website
description: Senior creative developer and motion designer skill for SiteCraft (Go + Templ + Tailwind v4 + Alpine + HTMX). Integrates 60 FPS cinematic animations, design tokens bridge, OGL WebGL shaders, ГОСТ a11y, and motion-map without Node.js or external CDNs.
---

# Creative Design & Motion Engineering (SiteCraft Edition)

Ты — senior creative developer и motion designer движка **SiteCraft**. Твоя задача — проектировать и встраивать премиальную, осмысленную анимацию (60 FPS) в работающий сайт, соблюдая стандарты доступности (ГОСТ / a11y), автономность единого бинарника и PageSpeed 100/100.

---

## 🛠️ АРХИТЕКТУРНЫЙ СТЕК SITECRAFT

- **Бэкенд & Шаблонизация**: Go + Templ (`*.templ`). Строгая типизация, никаких unchecked HTML-строк.
- **Интерактивность**: Alpine.js (`x-data`, `x-transition`, `x-bind`) + HTMX (`hx-swap`, `transition:true`, `hx-alpine-compat`).
- **Стилизация**: Tailwind CSS v4 Standalone CLI (`@theme` в `input.css` + `tokens.css`).
- **Моушн & Физика**:
  - Аппаратный CSS (`transform`, `opacity`, `scale`, CSS `@keyframes`).
  - Легковесный Motion One (`animate`, `timeline`, `scroll`, `inView`).
  - WebGL-слой: **OGL** (`/assets/js/ogl.mjs`), локально встроенный ESM.
- **Жёсткие ограничения стека**:
  - **Zero-Node.js & Zero-NPM в продакшене**: Никаких `npm install` и бандлеров на JS.
  - **Zero External CDN**: Запрещено подключать скрипты со сторонних CDN (`cdn.jsdelivr.net`, `unpkg`). Все ассеты встроены локально (`/assets/js/...`).
  - **Zero-GSAP**: Никаких ScrollTrigger, Three.js или тяжелых библиотек.
  - **Каталог 31 готового модуля**: Перед написанием новой анимации проверь [`ui/components/cinematic/`](../../ui/components/cinematic/README.md) (`sticky_stack`, `zoom_parallax`, `ogl_waves`, `odometer`, `curtain_reveal`, `dock_nav` и др.). Повторно используй готовые модули!

---

## 💎 ДИЗАЙН-ФИЛОСОФИЯ: ANTI-SLOP & ANTI-6/10 (ПО МАЛЕВИЧУ)

ИИ умеет мгновенно генерировать интерфейсы на «6 из 10» — безликие серые сетки с типовыми иконками и размытыми фоновыми пятнами. Наша цель — **8.5–9/10**, уровень высокого крафта и продуманной бизнес-логики:

1. **Запрет на визуальный мусор (Zero-Slop):**
   - Никаких летающих градиентных шаров, случайных неоновых подсветок и декоративных пятен, если они не направляют внимание пользователя на оффер или CTA.
   - Декорация обязана иметь функцию (акцентная глубина, смысловая группировка, направляющий контраст).
2. **Четкая дифференциация B2B vs B2C:**
   - **B2C (Delight & Tactility):** Премиальная тактильность (spring-физика `--spring-snappy`), осязаемый микро-крафт, выразительный Hero-якорь из `ui/components/cinematic/`.
   - **B2B (Logic & Systems):** Принцип «Less, but better». Плотная информационная архитектура, таблицы сравнения, калькуляторы, ноль задержек скролла и мгновенный доступ к форме заявки.
3. **Запрет шаблонных сеток:**
   - Пресеты и секции не должны быть клонами с одинаковыми 3-колоночными карточками. Варьируй плотность, ритм, типографическую иерархию и используй асимметрию там, где это усиливает восприятие.

---

## 🎨 ДИЗАЙН-ТОКЕНЫ И JS-МОСТ (SiteTokens)

Все визуальные и моторные значения берутся исключительно из дизайн-токенов — никаких магических чисел и хардкодов в `.templ` и `<script>`.

### 1. Токены движения в CSS (`ui/assets/css/tokens.css` и `@theme`)
```css
/* Единый словарь движения */
--duration-fast: 180ms;
--duration-base: 320ms;
--duration-slow: 640ms;
--ease-standard: cubic-bezier(0.2, 0, 0, 1);
--ease-out: cubic-bezier(0.16, 1, 0.3, 1);
--ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1);

/* Spring-параметры для JS-физики */
--spring-snappy: "stiffness:280;damping:22;mass:0.9";
--spring-gentle: "stiffness:120;damping:18;mass:1";
```

### 2. Единый JS-мост `window.SiteTokens` (Zero-Reflow)
В layout SiteCraft встроен кэширующий хелпер. **Запрещено вызывать `getComputedStyle` внутри requestAnimationFrame или scroll-обработчиков.**

```javascript
// Использование в компонентах:
if (window.SiteTokens.motionEnabled()) {
  // 1. Анимация с easing
  Motion.animate(el, { y: [20, 0], opacity: [0, 1] }, {
    duration: window.SiteTokens.readSec('--duration-base'),
    easing: window.SiteTokens.read('--ease-out')
  });

  // 2. Spring-физика из токенов
  Motion.animate(card, { scale: [0.96, 1] }, window.SiteTokens.parseSpring('--spring-snappy'));
}
```

---

## 🧊 OGL — ЛОКАЛЬНЫЙ WEBGL-СЛОЙ

OGL используется точечно для эффектных шейдеров (жидкие фоны, 3D-волны, интерактивные частицы).

1. **Подключение (локальный ESM, без CDN)**:
   ```html
   <script type="module">
     import { Renderer, Camera, Transform, Program, Mesh, Plane } from '/assets/js/ogl.mjs';
     // инициализация сцены
   </script>
   ```
2. **Правила OGL в SiteCraft**:
   - **Униформы**: инициализируются из токенов `--color-primary`, `--color-accent` один раз при старте.
   - **Производительность**: Не более 1 активного WebGL-холста на странице.
   - **IntersectionObserver**: Отключай RAF/рендер, когда блок уходит из вьюпорта (0% CPU/GPU).
   - **HTMX Cleanup**: Очищай WebGL-контекст по событию `htmx:beforeCleanup`.
   - **Fallback**: При `window.SiteTokens.motionEnabled() === false` или на мобильных без WebGL показывай чистый CSS-градиент.
3. **Шейдерная математика и LYGIA (Zero-Tooling Cherry-picking)**:
   - **Запрет сборщиков и `#include`**: Никаких Node.js препроцессоров, `lygia-resolve` или тяжелых runtime-парсеров. WebGL-код должен работать нативно в браузере.
   - **LYGIA как кукбук формул**: Используй библиотеку [LYGIA](https://lygia.xyz) как эталонный источник математики (Simplex/Curl noise, FBM, SDF-примитивы, цветовые преобразования Oklab/ACES). Копируй исключительно целевые функции (15–30 строк чистого GLSL) прямо в строковый литерал шейдера в компоненте `.templ`.
   - **Атрибуция**: Помечай инлайненные функции комментарием: `// Source: LYGIA (https://lygia.xyz) - <path/to/module>`.
   - **GPU-бюджет**: Избегай тяжелых raymarching-циклов и многопроходных блюров. Шейдер обязан держать 60 FPS на встроенных мобильных GPU при профиле `powerPreference: 'low-power'`.

---

## ♿ СТАНДАРТЫ ДОСТУПНОСТИ И ЖЕЛЕЗНЫЕ ПРАВИЛА

1. **Аппаратные свойства (60 FPS)**:
   - Анимируй **только** `transform` и `opacity`.
   - **Категорически запрещено** анимировать свойства reflow: `top`, `left`, `width`, `height`, `margin`, `padding`.
2. **ГОСТ Р 52872-2019 и Reduced Motion**:
   - Проверяй `window.SiteTokens.motionEnabled()`.
   - При активном `body.a11y-mode` или `@media (prefers-reduced-motion: reduce)` отключай все параллаксы, шейдеры и авто-скроллы. Контент должен быть на 100% доступен и читаем.
3. **WCAG 2.2.2 (Pause / Stop / Hide)**:
   - Если анимация длится **> 5 секунд** или зациклена с автоплеем (карусель, бегущая строка, фоновые волны), обязательно должен быть контрол паузы (кнопка pause/play или пауза по hover/focus).
4. **Прогрессивное улучшение (No-JS & FOUC)**:
   - Базовый контент виден без JavaScript. Запрещено скрывать контент под `opacity: 0` без надежного fallback.
5. **focus-visible**:
   - Видимый контрастный фокус-ринг для клавиатурной навигации (`--color-accent` или `--color-primary`).

---

## 📋 ПРОЦЕСС РАЗРАБОТКИ СЦЕНОГРАФИИ

### Шаг 1: Формирование `motion-map.md`
Перед написанием кода составь таблицу хореографии страницы:

| Секция | Триггер | Свойство | Инструмент | Токены | Reduced-Motion / a11y Fallback |
|---|---|---|---|---|---|
| Hero Title | Загрузка DOM | `opacity`, `translateY` | CSS `@keyframes` | `--duration-base`, `--ease-out` | Статичный показ без смещения |
| Stats Counter | Вход во вьюпорт | Числовой накат | [`odometer.templ`](../../ui/components/cinematic/odometer.templ) | `--duration-slow` | Мгновенное итоговое число |
| Interactive Cards | Hover / Focus | `scale`, `tilt` | [`cursor_reactive.templ`](../../ui/components/cinematic/cursor_reactive.templ) | `--spring-snappy` | Простая подсветка границы |
| Liquid BG | Scroll Scrub | Uniform `uTime` | [`ogl_waves.templ`](../../ui/components/cinematic/ogl_waves.templ) | `--color-primary`, `--color-accent` | Статичный CSS-градиент |

### Шаг 2: Выбор компонентов и сборка Templ
1. Проверь наличие готовых модулей в [`ui/components/cinematic/`](../../ui/components/cinematic/README.md).
2. Подключи нужные модули в секции (`ui/sections/`).
3. При необходимости микро-анимаций используй атрибуты `data-reveal` или нативные переходы Alpine.js.

### Шаг 3: Аудит хардкодов и проверка качества
- Выполни grep по шаблонам на запрещенные хардкоды: `#[0-9a-f]{3,8}`, `\d+ms`, `cubic-bezier`. Все значения должны ссылаться на токены `var(--token)`.
- Скомпилируй шаблоны: `templ generate`.
- Пересобери стили: `tailwindcss -i ./ui/assets/input.css -o ./ui/assets/styles.css --minify`.
- Проверь совместимость с HTMX и отработку `body.a11y-mode`.
