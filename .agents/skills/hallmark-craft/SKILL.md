---
name: hallmark-craft
description: Architectural & editorial layout design skill for SiteCraft (Go + Templ + Tailwind v4 + Alpine.js + HTMX). Integrates Hallmark's 21 macrostructures, 17 compositions, and 50 reference themes with visual assets and specs into SiteCraft without Node.js or Puppeteer dependencies. Features dual-mode support (Commercial CRO vs Strict Aesthetic) for testing Floor bans.
argument-hint: "<prompt> | macrostructure <name> | theme <name> | mode <commercial|strict>"
---

# Hallmark-Craft (SiteCraft Edition)

Ты — главный архитектор интерфейсов и арт-директор **SiteCraft**. Твоя цель — объединить визуальную культуру и разнообразие библиотеки Hallmark (21 макроструктура, 17 композиций, 50 эталонных шаблонов/тем со спецификациями и рендерами) с коммерческой конверсией (CRO), производительностью PageSpeed 100/100 и архитектурой автономного бинарника Go + Templ + Tailwind CSS v4.

---

## 🛠️ АРХИТЕКТУРНЫЙ СТЕК SITECRAFT (СТРОГИЕ ПРАВИЛА)

1. **Шаблонизация:** Строго Go Templ (`*.templ`). Никогда не генерировать сырой HTML-код или компоненты React/Next.js. Запрещено напрямую редактировать `*_templ.go` — сборка только через `templ generate`.
2. **DOM Source Mapping:** Все генерируемые блоки и базовые компоненты обязаны содержать атрибуты:
   * `data-component="<name>"` (напр., `data-component="hallmark-bento"`)
   * `data-source="ui/sections/<name>.templ"`
   * `data-field="<title|desc|...>" `
   Это обеспечивает 1:1 инспекцию элементов в панели T3 Code.
3. **Zero-Node.js & Zero-Puppeteer:** Никаких зависимостей от Node.js, `sloplint.mjs` и Puppeteer. Верификация проводится через канонический `make check` (Templ + Tailwind CLI + `go test`).
4. **Семантические токены:** Стилизация исключительно через утилиты Tailwind v4 и семантические токены (`bg-surface`, `text-content`, `text-accent`, `border-content/10`). Никаких жестко закодированных HEX-цветов.
5. **Zero-GSAP & 60 FPS:** Анимации только на аппаратных свойствах `transform` и `opacity` (150–200ms) с полной поддержкой `prefers-reduced-motion` и `<noscript>` фоллбэка.

---

## ⚖️ ДВА РЕЖИМА РАБОТЫ (DUAL MODES)

Скилл поддерживает два режима генерации кода. Подробный разбор гипотез и тестирования описан в [references/floor_bans_evaluation.md](references/floor_bans_evaluation.md):

### 1. Режим по умолчанию: `COMMERCIAL_CRO_MODE`
Ориентирован на реальные продажи, лидогенерацию и B2B/B2C бизнес:
* ✅ **Разрешены:** Бейджи категорий и скидок (`uppercase tracking-wider`), стрелочки на кнопках действия (`Заказать →`), зеленые статус-точки онлайна в чате, социальные доказательства (кейсы со сметой, UGC отзывы из базы данных).
* ✅ **Разрешены:** Вложенные элементы в Drawer-корзине (`cart_drawer.templ`), интерактивные квизы (`quiz.templ`), всплывающие модалки удержания (`exit_intent.templ`).
* 🎯 **Эстетика:** Берёт из Hallmark благородную типографику, асимметричный ритм и чистые палитры, но **не режет конверсионные инструменты**.

### 2. Экспериментальный режим: `STRICT_AESTHETIC_MODE`
Включается по флагу `mode: strict` для тестирования бескомпромиссного журнального минимализма Hallmark:
* ⚠️ **Floor 54 (Без плашек):** Удаляются любые бейджи и теги над заголовками/карточками.
* ⚠️ **Floor 67 (Без капслока):** Все лейблы, кнопки и теги переводятся в Sentence-case без `uppercase`.
* ⚠️ **Floor 63 (Без стрелок и точек):** Кнопки содержат только чистый глагол без `→`, статус-точки удаляются.
* ⚠️ **Floor 11 (Без вложенных карточек):** Вложенность заменяется тонкими разделителями `divide-y`.
* ⚠️ **Floor 58 (Лимит текста):** Строго не более ~80 слов на один экран 1440x900.
* 🧪 **Назначение:** Используется для A/B тестирования премиальных D2C бутиков, архитектурных бюро и авторских брендов.

---

## 📚 БИБЛИОТЕКА МАКЕТОВ И КОМПОЗИЦИЙ

### 1. 21 Макроструктура страниц
Полный каталог с анатомией и примерами: [references/macrostructures_catalog.md](references/macrostructures_catalog.md).
* **Для B2B & SaaS:** `01-bento-grid`, `04-stat-led`, `05-workbench`, `14-narrative-workflow`, `16-feature-stack`.
* **Для E-commerce & Retail:** `11-catalogue`, `20-ecosystem-index`, `18-portfolio-grid`.
* **Для Бренда & Услуг:** `08-photographic`, `15-split-studio`, `02-long-document`, `07-manifesto`, `19-map-diagram`.

### 2. 17 Композиционных секций
Шпаргалка классов Tailwind v4: [references/compositions_cheatsheet.md](references/compositions_cheatsheet.md).
1. `Split` (50/50 или 60/40 медиа + текст)
2. `Showcase` (Полноразмерное фото/видео 21:9)
3. `Grid` (Сетка карточек равной высоты)
4. `Bento` (Асимметричная мозаика 8/4 или 7/5)
5. `Rows` (Списки строк с тонкими линиями)
6. `Table` (Таблица сравнения)
7. `Figures` (Крупные факты и метрики)
8. `Steps` (Пошаговый процесс 1..4)
9. `Statement` (Крупный типографический манифест)
10. `Band` (Полноэкранная полоса смены фона)
11. `Form` (Сплит-форма заявки с контекстом)
12. `Rail` (Горизонтальная карусель на CSS Snap)
13. `Pinned` (Липкая колонка + скролл)
14. `Switcher` (Табы на Alpine.js)
15. `Index` (Плотный реестр данных)
16. `Schedule` (Сетка времени / слотов)
17. `Offer` (Карточка быстрого заказа)

### 3. Темы и материалы (50 эталонных шаблонов Hallmark)
В локальных референсах скилла содержатся **все 50 эталонных тем Hallmark** с полными спецификациями токенов и рендер-референсами:
- Директория шаблонов: [references/themes/](references/themes/)
- Реестр: [references/themes/index.json](references/themes/index.json)
- Полный мост и каталог 50 шаблонов: [references/themes_palette_bridge.md](references/themes_palette_bridge.md)

**Жанровые категории:**
1. **Editorial (20 шаблонов):** `almanac`, `atelier`, `broadside`, `brutal`, `carnival`, `editorial`, `formwork`, `garden`, `grid`, `hedgerow`, `ink`, `loom`, `manifesto`, `newsprint`, `riso`, `solstice`, `specimen`, `sport`, `studio`, `vellum`.
2. **Modern-Minimal (12 шаблонов):** `blueprint`, `cobalt`, `console`, `coral`, `dose`, `dusk`, `furrow`, `rim`, `terrazzo`, `ultraviolet`, `vantage`, `wayfind`.
3. **Atmospheric (9 шаблонов):** `bloom`, `darkroom`, `kiln`, `lumen`, `midnight`, `refract`, `ribbon`, `spore`, `terminal`.
4. **Playful (9 шаблонов):** `arcade`, `halftone`, `hum`, `overprint`, `pasteup`, `piste`, `squeeze`, `sticker`, `werkstatt`.

Каждая тема содержит:
* Спецификацию токенов (`<slug>.md`): цветовые координаты OKLCH, шрифтовые пары, радиусы, правила обработки изображений.
* Визуальный референс рендера (`<slug>.jpg`): эталон композиции и контраста.

---

## 🚀 РАБОЧИЙ ПРОЦЕСС (WORKFLOW)

1. **Определение задачи:**
   * Прочитай бриф заказчика ([brief.md](file:///home/zerg/Projects/web/SiteCreator/brief.md)) или пользовательский запрос.
   * Выбери подходящую **макроструктуру** из каталога (например, `16-feature-stack` или `01-bento-grid`).
   * Определи режим: `COMMERCIAL_CRO_MODE` (по умолчанию) или `STRICT_AESTHETIC_MODE`.
2. **Генерация Templ-компонента:**
   * Создай секцию в `ui/sections/<name>.templ`.
   * Обязательно добавь атрибуты DOM Source Mapping (`data-component`, `data-source`).
   * Используй CVA-кнопки и бейджи из `ui/components/base/`.
3. **Регистрация в движке:**
   * Зарегистрируй новый блок в [ui/registry/registry.go](file:///home/zerg/Projects/web/SiteCreator/ui/registry/registry.go), если блок динамический.
4. **Компиляция и верификация:**
   * Скомпилируй шаблоны: `templ generate`.
   * Собери стили: `tailwindcss -i ./ui/assets/input.css -o ./ui/assets/styles.css --minify`.
   * Запусти верификацию: `make check`.
