---
name: stitch-workflow
description: Guides the workflow for Google Stitch CLI tailored to SiteCraft (Go, Templ, Tailwind v4), including generating designs from brief.md, converting HTML to Templ components, exporting assets, and syncing tokens into tokens.css and presets.
---

# Stitch Workflow (SiteCraft Edition)

Используй этот скилл при проектировании, генерации и интеграции дизайн-макетов из Stitch CLI в движок SiteCraft (**Go + Templ + Tailwind CSS v4 + Alpine.js + HTMX**).

## Архитектурные правила интеграции:
1. **Без сырого HTML в продакшене:** все сгенерированные интерфейсы конвертируются в типобезопасные шаблоны **Templ** (`ui/sections/*.templ` или `ui/components/*.templ`).
2. **Дизайн-система Tailwind v4:** цвета и шрифты не хардкодятся в инлайн-стилях, а связываются с CSS-токенами движка (`var(--color-primary)`, `var(--color-accent)`, `var(--radius-card)` и др.).
3. **Регистрация в Block Registry:** новые секции должны подключаться через фабрику `ui/registry/registry.go`.

---

## Порядок действий:

### 1. Авторизация в Stitch:
- Выполни команду: `stitch login` (открывает окно авторизации Google).

### 2. Генерация концепций на основе `brief.md`:
- Прочитай [brief.md](../../brief.md) (целевая аудитория, главы сторителлинга, визуальный стиль).
- Сформируй промпт для Stitch CLI на основе брифа.
- Запроси 4–8 вариантов каждого экрана для выбора идеального решения под сторителлинг.

### 3. Экспорт и распределение ассетов:
- Выполни экспорт: `stitch export`.
- Перемести медиа-файлы (изображения, SVG, иконки) в `ui/assets/images/` или `public/`.
- Зарегистрируй статические ассеты с учетом `ui/embed.go`.

### 4. Конвертация макета в компоненты SiteCraft:
- Преобразуй экспортированный HTML в синтаксис Templ:
  - Секции страниц размещай в `ui/sections/<section_name>.templ`.
  - Повторно используемые элементы — в `ui/components/base/` (используя `@base.Button`, `@base.Card`, `@base.Badge`) или `ui/components/`.
  - Для векторных иконок используй локальный пакет `ui/components/icons` вместо тяжелых инлайн-SVG.
- Зарегистрируй новый блок в `ui/registry/registry.go` через `registry.RegisterBlock("<type>", ...)`.

### 5. Синхронизация дизайн-токенов:
- Обнови палитру и радиусы в `ui/assets/css/tokens.css` и `@theme` в `ui/assets/input.css`.
- При необходимости зафиксируй настройки в нишевом пресете `presets/<preset-name>.json` (или `site_settings` PocketBase).

### 6. Компиляция и сборка:
- Выполни генерацию шаблонов и стилей:
  ```bash
  templ generate
  tailwindcss -i ./ui/assets/input.css -o ./ui/assets/styles.css --minify
  ```
