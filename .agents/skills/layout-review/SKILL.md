---
name: layout-review
description: Audits and validates responsive layout in SiteCraft across mobile (390px), tablet (768px), and desktop (1440px), tests dev-server (air/go), checks z-index collisions with headers/drawers/a11y-bar, and fixes Templ/Tailwind CSS issues.
---

# Layout Review (SiteCraft Edition)

Используй этот скилл для проверки адаптивности, отсутствия переполнений и корректности рендеринга шаблонов Templ и стилей Tailwind CSS v4 в движке SiteCraft.

## Ключевые брейкпоинты:
- **Mobile**: `390px` (iPhone 12/14/15)
- **Tablet**: `768px` (iPad Portrait)
- **Desktop**: `1440px` (HD/Retina Desktop)

---

## Порядок действий:

### 1. Подготовка и запуск dev-сервера:
- Скомпилируй шаблоны Templ и соберите стили:
  ```bash
  templ generate
  tailwindcss -i ./ui/assets/input.css -o ./ui/assets/styles.css
  ```
- Убедись, что запущен dev-сервер:
  - Через `air` (live-reload на порту `8080`), либо:
  - Через команду: `go run main.go serve --http=127.0.0.1:8090` (или бинарник в фоне).
- Проверь доступность страницы по адресу `http://127.0.0.1:8090` (или `http://127.0.0.1:8080`).

### 2. Аудит адаптивности и разметки (390px, 768px, 1440px):
- **Горизонтальный скролл (Horizontal Overflow):**
  - Выполни скрипт/инспекцию: `document.body.scrollWidth > window.innerWidth`.
  - Если есть утечки ширины: найди элементы с фиксированной шириной (`w-[...]`), отрицательными отступами (`-mx-...`) или неадаптивными flex/grid контейнерами без `min-w-0` / `overflow-hidden`.
- **Коллизии `z-index` и перекрытия:**
  - Проверь совместную работу интерактивных слоев:
    - Фиксированная шапка ([header.templ](../../ui/components/header.templ))
    - Плавающая панель контактов ([floating_contact.templ](../../ui/components/floating_contact.templ))
    - Панель доступности ГОСТ ([a11y_bar.templ](../../ui/components/a11y_bar.templ))
    - Дровер корзины ([cart_drawer.templ](../../ui/components/cart_drawer.templ)) и модальные окна
  - Убедись, что модалки и дроверы не перекрываются фиксированными баннерами.
- **Состояния Alpine.js & Tailwind v4 Data-Variants:**
  - Проверь работу селекторов `@custom-variant`: `data-open:`, `data-closed:`, `data-active:`.
  - Убедись, что нет багов с `x-cloak` и схлопыванием контента до инициализации Alpine.js.

### 3. Проверка тем и режима доступности:
- Переключи тему через селектор тем в шапке (светлая `main-day` ↔ темная `main-night`).
- Включи режим для слабовидящих (`body.a11y-mode`): проверь, что шрифты масштабируются, а контраст соответствует стандарту ГОСТ Р 52872-2019.

### 4. Автоматическое исправление:
- Внеси правки напрямую в соответствующие шаблоны `ui/sections/*.templ` или `ui/components/*.templ`.
- Если требуется корректировка глобальных стилей — обнови `ui/assets/input.css` или `ui/assets/css/*.css`.
- Запусти `templ generate` и перепроверь результат.
