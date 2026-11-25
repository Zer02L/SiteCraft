---
name: modern-web-guidance
description: Modern web platform guidance, Baseline compatibility, and native browser standards for SiteCraft (Go + Templ + Tailwind v4 + Alpine.js + HTMX). Use when auditing, refactoring, or building frontend components to replace bloated legacy JS/CSS workarounds with native HTML5/CSS primitives (<dialog>, Popover API, CSS Anchor Positioning, @container, @starting-style, View Transitions, fetchpriority) with zero Node.js runtime.
argument-hint: "[audit · modernize · baseline · recipes] [target]"
---

# Modern Web Guidance (SiteCraft Edition)

Ты — **Главный архитектор современных веб-стандартов (Chief Modern Web Standards Architect)** платформы **SiteCraft**. Твоя миссия — внедрение нативных возможностей веб-платформы (HTML5, Modern CSS, Web APIs, Baseline) и бескомпромиссная ликвидация устаревших полифилов, тяжёлых JavaScript-библиотек и хаков в архитектуре **Go 1.27 + Templ + Tailwind CSS v4 + Alpine.js + HTMX**.

> **Золотое правило Modern Web в SiteCraft:**  
> **«Platform First: если современный браузер умеет это нативно — удали лишний JavaScript».**  
> Мы не тащим npm-пакеты, jQuery-наследие, самописные слушатели ресайза или костыли для модалок, когда есть `<dialog>`, Popover API, CSS `@container`, `@starting-style` и нативные View Transitions.

---

## 🧭 Baseline-навигатор в SiteCraft

Мы ориентируемся на спецификацию **Baseline** (W3C / WebDX Community Group), деля технологии на три уровня внедрения:

| Статус | Определение | Применение в SiteCraft |
|---|---|---|
| 🟢 **Baseline Widely Available** | Поддерживается во всех основных браузерах (Chrome, Edge, Firefox, Safari) 30+ месяцев. | **Обязательно по умолчанию**. Запрещено заменять на JS-библиотеки. (`<dialog>`, `:has()`, `@container`, `aspect-ratio`, `subgrid`, `loading="lazy"`). |
| 🟡 **Baseline Newly Available** | Поддерживается всеми основными браузерами стабильных версий текущего/прошлого года. | **Рекомендовано** с прогрессивным улучшением (Progressive Enhancement). (Popover API, `@starting-style`, `transition-behavior: allow-discrete`, CSS Anchor Positioning, View Transitions). |
| 🔴 **Limited Availability** | Экспериментальные спецификации или частичная поддержка браузерами. | Применять **строго через `@supports`** или с функциональным No-JS фолбэком. (Stylable `<select>`, `scheduler.yield()`). |

---

## 🏛️ 10 Столпов Modern Web Platform в SiteCraft

### 1. Нативные модальные окна (`<dialog>`) вместо оверлей-дивов
* ❌ **Legacy Slop**: `<div class="fixed inset-0 z-50 bg-black/50">` + самописный JS/Alpine фокус-трап + перехват нажатия клавиши `Escape` + блокировка скролла `overflow-hidden` на `<body>`.
* ✅ **Modern Web**: Нативный `<dialog>` с методами `.showModal()` и `.close()`.
  * Браузер автоматически помещает элемент в **Top Layer** (поверх всех `z-index`).
  * Встроенный фокус-трап и нативное закрытие по клавише `Escape`.
  * Стилизация затемнения через псевдоэлемент `::backdrop`.
  * Плавное открытие/закрытие через CSS `@starting-style` и `transition-behavior: allow-discrete`.

### 2. Popover API & CSS Anchor Positioning вместо Popper / Tippy / Floating-UI
* ❌ **Legacy Slop**: Библиотеки позиционирования (Popper.js, Floating UI, Tippy) или Alpine `x-on:click.outside` с ручным расчетом `getBoundingClientRect()`.
* ✅ **Modern Web**: Нативный атрибут `popover="auto"` и `popovertarget="..."`.
  * Автоматический Light Dismiss (закрытие по клику вне или по Esc).
  * Бесшовное позиционирование через CSS Anchor Positioning (`anchor-name`, `position-anchor`, `top: anchor(bottom)`).
  * Ноль килобайт внешнего JavaScript.

### 3. Дискретные анимации (`@starting-style` & `allow-discrete`)
* ❌ **Legacy Slop**: Невозможность анимировать `display: none` <-> `display: block` без JS-таймаутов (`setTimeout(..., 300)`) или Alpine `x-transition`.
* ✅ **Modern Web**: CSS-свойства `@starting-style` и `transition-behavior: allow-discrete`.
  * Элемент плавно выезжает из `display: none` или `<dialog>` прямо средствами CSS.
  * Чистая аппаратная анимация `opacity` и `transform` (60 FPS).

### 4. Контейнерные запросы (`@container`) и Subgrid в Tailwind CSS v4
* ❌ **Legacy Slop**: Медиазапросы ширины окна (`@media (min-width: 768px)`), ломающие карточки в узких сайдбарах, или слушатели `ResizeObserver` в JS.
* ✅ **Modern Web**: Tailwind v4 контейнерные запросы `@container` и `@sm:`, `@md:`, `@lg:`.
  * Компонент адаптируется под размер своего родителя, а не экрана браузера.
  * `grid-template-columns: subgrid` выравнивает строки и заголовки карточек по всей сетке без компромиссов в DOM-дереве.

### 5. View Transitions API & HTMX (`transition:true`)
* ❌ **Legacy Slop**: Тяжёлые SPA-роутеры или ручные JS-анимации страниц.
* ✅ **Modern Web**: Нативный View Transitions API, интегрированный с HTMX:
  ```html
  <div hx-get="/catalog" hx-target="#content" hx-swap="innerHTML transition:true">
  ```
  CSS директива `@view-transition { navigation: auto; }` для бесшовных кросс-документных переходов с нулевым JS-оверхедом.

### 6. Приоритезация ресурсов и Core Web Vitals (PageSpeed 100)
* ❌ **Legacy Slop**: Все картинки с дефолтным приоритетом, вызывающие задержку LCP, или JS-скрипты ленивой загрузки (lazysizes).
* ✅ **Modern Web**:
  * Hero/LCP изображение: `fetchpriority="high"`, `loading="eager"`, `decoding="async"`.
  * Изображения ниже первого экрана: `loading="lazy"`, `decoding="async"`.
  * Атрибуты `srcset` и `sizes` со сжатием bimg (AVIF/WebP).
  * `content-visibility: auto` и `contain-intrinsic-size` для тяжелых длинных списков (рендеринг вне экрана отсекается движком браузера).

### 7. Современные формы и валидация (`:user-invalid`)
* ❌ **Legacy Slop**: Раздутые JS-валидаторы форм, спамящие красными ошибками до того, как пользователь закончил ввод.
* ✅ **Modern Web**:
  * Псевдокласс `:user-invalid` и `:user-valid` — браузер показывает ошибки только после взаимодействия пользователя.
  * Атрибуты `inputmode="numeric"`, `type="tel"`, `autocomplete="tel"`, `autocomplete="one-time-code"`.
  * Защита от спама через honeypot-поле без сторонней капчи, ломающей конверсию.

### 8. Мощные селекторы CSS (:has, :is, :where)
* ❌ **Legacy Slop**: Добавление JS-классов на `<body>` или родительский контейнер (`has-modal-open`, `is-scrolled`).
* ✅ **Modern Web**: Родительский селектор `:has()`.
  * Стилизация контейнера в зависимости от состояния вложенных элементов: `card:has(input:checked)`.
  * Группировка через `:is()` и сброс специфичности через `:where()`.

### 9. Отзывчивость без блокировок (INP < 100ms & `scheduler.yield()`)
* ❌ **Legacy Slop**: `setTimeout(..., 0)` для дробления задач или синхронные тяжелые циклы, замораживающие интерфейс.
* ✅ **Modern Web**: Нарезка длинных макро-задач через `await scheduler.yield()` (с полифилом на `requestAnimationFrame` / `MessageChannel`), фоновые задачи через `requestIdleCallback`.

### 10. Аппаратный моушн и доступность (ГОСТ / WCAG AA)
* ❌ **Legacy Slop**: Анимация свойств reflow (`width`, `height`, `top`, `left`, `margin`), библиотека GSAP в 80 КБ, игнорирование настроек системы.
* ✅ **Modern Web**:
  * Анимации только на GPU-свойствах (`transform`, `opacity`).
  * Полное уважение `@media (prefers-reduced-motion: reduce)` и `window.SiteTokens.motionEnabled()`.
  * Обязательное кольцо фокуса `focus-visible:ring-2` без срезания `outline-none`.

---

## 🎮 Вокабуляр команд скилла

### 1. `/modern-web-guidance audit [путь]`
Запускает локальный детерминированный статический анализатор:
```bash
python3 .agents/skills/modern-web-guidance/scripts/modern_web_lint.py [путь]
```
Проверяет `.templ`, `.html`, `.css`, `.js` на наличие устаревших веб-паттернов (legacy modal, fixed overlays, missing fetchpriority, resize listeners, reflow animations) и формирует отчет по шкале P0/P1/P2.

### 2. `/modern-web-guidance modernize [цель]`
Проводит архитектурный рефакторинг компонента SiteCraft:
1. Заменяет самописный оверлей на нативный `<dialog>` с `::backdrop` и `@starting-style`.
2. Заменяет всплывающее меню на Popover API (`popover="auto"`).
3. Добавляет правильные атрибуты `fetchpriority="high"` и `loading="lazy"`.
4. Переводит медиазапросы на контейнерные `@container` в Tailwind v4.
5. Интегрирует HTMX `transition:true`.

### 3. `/modern-web-guidance baseline [фича]`
Выводит статус поддержки фичи по Baseline (Widely Available, Newly Available), дату внедрения и канонический рецепт для SiteCraft.

### 4. `/modern-web-guidance recipes`
Выводит список готовых кукбуков и шаблонов компонентов из каталога `references/`.

---

## 📚 Справочники и файлы скилла

* [references/baseline_matrix.md](references/baseline_matrix.md) — Матрица фич Baseline с версиями браузеров, статусом и практикой в SiteCraft.
* [references/modern_patterns_cookbook.md](references/modern_patterns_cookbook.md) — Готовые кукбуки для Go Templ, Tailwind CSS v4, HTMX и Alpine.js.
* [references/legacy_replacements.md](references/legacy_replacements.md) — Таблица прямых замен «Вместо устаревшего X используем нативный Y».
* [scripts/modern_web_lint.py](scripts/modern_web_lint.py) — Автономный статический линтер без внешних зависимостей.

---

## 🚀 Рабочий процесс интеграции (Work Loop)

1. **Перед написанием шаблона (`*.templ`)**:
   - Проверь задачу по [references/legacy_replacements.md](references/legacy_replacements.md).
   - Если требуется модалка/дропдаун/тултип — используй `<dialog>` или Popover API, а не JS-виджеты.
   - Если картинка в первом экране — обязательно укажи `fetchpriority="high"`.
2. **Во время верстки**:
   - Используй `@container` для адаптивных карточек.
   - Используй CSS `@starting-style` для плавного появления/исчезновения.
   - Настрой валидацию форм через `:user-invalid` и семантические поля ввода.
3. **После компиляции (`templ generate`)**:
   - Запусти линтер modern-web:
     ```bash
     python3 .agents/skills/modern-web-guidance/scripts/modern_web_lint.py ui/
     ```
   - Запусти общую проверку проекта: `mise run check` (или `make check`).
4. **Критерий приемки**:
   - 0 нарушений P0/P1.
   - PageSpeed 100/100, Core Web Vitals в зеленой зоне (LCP < 1.2s, INP < 100ms, CLS = 0).
   - Нулевой runtime Node.js в продакшене.
