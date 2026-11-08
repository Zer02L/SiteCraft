# Подборка дизайн-скилов для Claude из видео

> **Источник:** [YouTube: 클로드 디자인 용도별 GitHub 스킬 7가지 추천 (AI LABS)](https://youtu.be/Ysr7oNDajJI)  
> **Тема:** 7 специализированных GitHub-скилов (Agent Skills) для улучшения UI/UX, анимаций и структуры веб-интерфейсов при разработке с AI-ассистентами (Claude Code, Cursor, Windsurf и др.).

---

## 8 рекомендуемых дизайн-скилов

### 1. `landing-page-design` (Разработка конверсионных лендингов)
* **Репозиторий:** [elayadesign/ai-design-skills](https://github.com/elayadesign/ai-design-skills)
* **Автор:** Elaya (`@elayadesigns`)
* **Назначение:** Создание структурированных, ориентированных на конверсию целевых страниц (SaaS, презентации продуктов, списки ожидания). Принуждает агента сначала сформировать смысловой каркас, заголовки, преимущества и FAQ, а также жестко задает токены типографики, цвета и отступы, предотвращая шаблонный «AI slop».
* **Установка:**
  ```bash
  npx skills add elayadesign/ai-design-skills
  ```

---

### 2. `emilkowalski/skills` (Интерфейсная анимация и микроинтердействия)
* **Репозиторий:** [emilkowalski/skills](https://github.com/emilkowalski/skills)
* **Автор:** Emil Kowalski (создатель Sonner, экс-Vercel/Linear)
* **Назначение:** Набор скилов по дизайн-инжинирингу и плавным анимациям (`animate`, `apple-design`, `review-animations`, `emil-design-eng`). Задает правильные кривые плавности (cubic-bezier, ease-out вместо некорректного ease-in на вход), микроинтердействия, тайминги до 300ms и стандарты Apple Human Interface.
* **Установка:**
  ```bash
  npx skills@latest add emilkowalski/skills
  ```

---

### 3. `web-design-engineer` (Комплексный веб-дизайн и стилистика)
* **Репозиторий:** [ConardLi/garden-skills](https://github.com/ConardLi/garden-skills) (скил `web-design-engineer`)
* **Автор:** Conard Li
* **Назначение:** Превращает агента в первоклассного design-engineer. Содержит фреймворк «5-dial Design Read» и готовые эстетические рецепты (Linear-стиль, Stripe Press, Bloomberg и др.), синхронизируя фронтенд-код с дизайн-системами и компонентной архитектурой.
* **Установка:**
  ```bash
  npx skills add ConardLi/garden-skills -s web-design-engineer
  ```

---

### 4. `tastemaker` (Стилизация по референсам и анти-AI шаблоны)
* **Репозиторий:** [codeswithroh/tastemaker](https://github.com/codeswithroh/tastemaker)
* **Автор:** Roh / codeswithroh
* **Назначение:** Анализ референсных изображений и предпочтений разработчика для переноса эстетики в код. Устраняет типичные AI-артефакты (случайные эмодзи, фиолетовые градиенты, непроверенный контраст) и выстраивает гармоничную цветовую палитру, выверенные шрифтовые пары и ритм отступов.
* **Установка:**
  ```bash
  npx skills add codeswithroh/tastemaker
  ```

---

### 5. `better-interface` / `interface-review` (Аудит UI/UX и дизайн-ревью)
* **Репозиторий:** [jakubkrehel/skills](https://github.com/jakubkrehel/skills)
* **Автор:** Jakub Krehel
* **Назначение:** Системный аудит интерфейса и фронтенд-кодовой базы. Включает специализированные подмодули: `better-ui`, `better-typography`, `better-colors`, `better-layout`, `better-accessibility` и `better-writing`. Формирует ранжированный отчет об ошибках UX и визуальной иерархии с вердиктами Block/Approve.
* **Установка:**
  ```bash
  npx skills add jakubkrehel/skills
  ```

---

### 6. `visual-critique` / `designer-skills` (Визуальная критика по скриншотам)
* **Репозиторий:** [Owl-Listener/designer-skills](https://github.com/Owl-Listener/designer-skills)
* **Назначение:** Комплексная система для дизайнеров и фронтендеров. Модуль `visual-critique` (команда `/visual-critique:critique-screen`) оценивает скриншоты готового экрана с точки зрения профессионального арт-директора, выявляя проблемы выравнивания, читаемости и предлагая точечные CSS-правки.
* **Установка:**
  ```bash
  # Плагин для Claude Code:
  /plugin marketplace add Owl-Listener/designer-skills
  ```

---

### 7. `build-awwwards-quality-sites` / `Skills` (Уровень Awwwards и визуал)
* **Репозиторий:** [MengTo/Skills](https://github.com/MengTo/Skills)
* **Автор:** Meng To (Design+Code)
* **Назначение:** Набор практических скилов для создания выразительных интерфейсов уровня Awwwards: смелые сетки, конвертация видеозаписей экрана в подробные промпты («Video to Super Prompt»), декомпозиция интерактивных состояний («HTML to Interaction Prompts») и разбор референсных посадочных страниц.
* **Установка:**
  ```bash
  npx skills add MengTo/Skills
  ```

---

### 8. `flat-to-3d` (Архитектурный 3D-скроллинг и трансформация чертежей)
* **Репозиторий:** [Sam1983Aing/flat-to-3d](https://github.com/Sam1983Aing/flat-to-3d)
* **Автор:** Samnang Aing
* **Назначение:** Трансформация плоской архитектурной графики (аксонометрия, разрезы, планшеты стадий, график Ганта, макет на столе, инсоляционный фасад) в кинематографический 3D-скроллинг при помощи техники «длинной линзы» (Dolly Zoom: 1.5° → 60°), процедурных шейдеров рейсфедера/бумаги и 100% процедурной геометрии в коде (< 150 КБ, 0 внешних тяжелых 3D-моделей).
* **Адаптация в SiteCraft:**
  - Локальный скил: `.agents/skills/flat-to-3d/SKILL.md`
  - Кукбук: `.agents/cookbooks/flat_to_3d.md`
  - 100% offline (Zero-CDN), поддержка токенов `SiteTokens`, ГОСТ a11y & `prefers-reduced-motion` фоллбек и HTMX cleanup.

---

### 9. `hallmark-craft` (Адаптация 21 макроструктуры и тем Hallmark под SiteCraft)
* **Назначение:** Архитектурная компоновка страниц на базе библиотеки Hallmark (21 макроструктура, 17 композиций, 102 темы), переведённая на стек Go + Templ + Tailwind v4 + Alpine.js + HTMX без зависимостей от Node.js и Puppeteer.
* **Поддержка двух режимов (для тестирования банов):**
  - `COMMERCIAL_CRO_MODE` (по умолчанию): сохраняет бейджи, стрелки, конверсионные B2B/B2C модули, вложенные карточки в корзинах и социальные доказательства.
  - `STRICT_AESTHETIC_MODE`: экспериментальный режим для проверки гипотез ультра-чистого журнального минимализма Hallmark Floor (без бейджей, без капслока, без стрелок).
* **Локальный скил:** `.agents/skills/hallmark-craft/SKILL.md`
* **Каталог макроструктур:** `.agents/skills/hallmark-craft/references/macrostructures_catalog.md`
* **Лаборатория тестов Floor:** `.agents/skills/hallmark-craft/references/floor_bans_evaluation.md`

---

### 10. `stageshow-craft` (Faux Product UI, мокапы и интерактивные демо с курсором)
* **Оригинал:** [benjitaylor/skills](https://github.com/benjitaylor/skills/tree/main) (скилл `stageshow`)
* **Автор оригинала:** Benji Taylor (`@benjitaylor`)
* **Назначение:** Создание кинематографичных мокапов интерфейсов продуктов (Hero-шоты SaaS, окна приложений с кнопками macOS, мобильные телефоны с Dynamic Island, терминалы, Bento-сетки фичей). Исключает «AI slop» за счет бюджета реального текста (3–7 строк, остальное — аккуратные плашки-плейсхолдеры `PlaceholderBar`), принципа естественного среза краев (`Edge Bleed`), нависающих сателлитов (`MockSatellite`) и интерактивного виртуального курсора (`MockCursor`).
* **Адаптация в SiteCraft:**
  - **Локальный скил:** `.agents/skills/stageshow-craft/SKILL.md`
  - **Кукбук:** `.agents/cookbooks/stageshow_craft.md`
  - **Templ-компоненты:** `ui/components/cinematic/stageshow.templ` (`StageshowStage`, `MockWindow`, `MockPhone`, `MockTerminal`, `PlaceholderBar`, `MockRow`, `MockSatellite`, `MockCursor`)
  - **Архитектура:** 100% Zero-Node.js, Zero-Playwright, автономный Go + Templ + Tailwind v4 + Alpine.js, ГОСТ a11y & `prefers-reduced-motion` фоллбэк.


Если вы не используете CLI `skills`, структуру можно подключить вручную:

1. Склонируйте нужный репозиторий во временную папку:
   ```bash
   git clone https://github.com/elayadesign/ai-design-skills.git /tmp/ai-design-skills
   ```
2. Скопируйте нужный скил в директорию скилов вашего проекта (например, `.claude/skills/` или `.agents/skills/`):
   ```bash
   mkdir -p .claude/skills
   cp -r /tmp/ai-design-skills/skills/landing-page-design .claude/skills/
   ```

> 💡 **Совет из видео:** Не активируйте слишком много дизайн-скилов одновременно в одном запросе, чтобы их дизайн-системы и правила не конфликтовали между собой. Выбирайте 1 профильный скил под текущую задачу (например, `landing-page-design` для структуры, `emilkowalski/skills` для доводки анимаций).
