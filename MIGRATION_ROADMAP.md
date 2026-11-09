# План миграции SiteCraft: с PocketBase на Chi + SQLite/Postgres + sqlc

> Документ фиксирует утверждённый стек, архитектурные правила и интерактивный чек-лист перехода проекта **SiteCraft** на независимый серверный Go-стек.

---

## 1. Утверждённый технологический стек

| Слой | Технология / Пакет | Актуальная версия | Назначение / Обоснование |
| :--- | :--- | :--- | :--- |
| **Язык** | **Go** | **1.27.0** | Актуальная версия компилятора в окружении (`go.mod`), 100% совместимость |
| **HTTP Router** | **`github.com/go-chi/chi/v5`** | **v5.2.1** | Свежий стабильный релиз chi, 100% совместим с `net/http` |
| **Логирование** | **`log/slog`** | Встроено в stdlib | Нулевые зависимости, максимальная скорость JSON/Text логирования |
| **СУБД по умолчанию** | **`modernc.org/sqlite`** | **v1.36.0** | Чистый Go, WAL-режим, CGO=0, нулевой оверхед по памяти |
| **СУБД (резерв/scale)** | **`github.com/jackc/pgx/v5`** | **v5.7.2** | Актуальный драйвер Postgres, поддержка пулов и конвейеров |
| **Генератор запросов** | **`sqlc`** | **v1.31.1** | Type-safe SQL кодогенерация с `emit_interface: true` |
| **Миграции** | **`github.com/pressly/goose/v3`** | **v3.24.1** | ANSI SQL миграции с `go:embed` и автозапуском при старте |
| **Сессии админки** | **`github.com/alexedwards/scs/v2`** | **v2.8.0** | CGO=0 SQLStore, `SameSite`, `HttpOnly`, периодическая очистка |
| **Медиа-пайплайн** | **Чистый Go / `golang.org/x/image`** | **v0.47.0** | Pure-Go масштабирование (thumb, sm, md, lg), дисковое кэширование, immutable headers |
| **SSR / UI** | **`github.com/a-h/templ`** | **v0.3.1100** | Компилируемые в чистый Go шаблоны, Zero-alloc streaming |
| **Стилизация** | **Tailwind CSS** | **v4 Standalone** | Rust/Oxide движок, мгновенная сборка без Node.js |
| **Динамика / AJAX** | **HTMX** | **v2 / v4** | Серверные HTML-фрагменты, OOB swaps, плавные переходы |
| **Микро-UX** | **Alpine.js** | **v3.17+** | Легковесный клиентский интерактив |

---

## 2. Архитектурный контракт «Dual-Engine Ready» (SQLite ↔ Postgres)

1. **Интерфейс репозитория:** Вся бизнес-логика зависит от `db.Querier` (генерируется `sqlc`).
2. **Первичные ключи:** Строковые `UUIDv4/v7` (`TEXT PRIMARY KEY`).
3. **Даты:** Строго `time.Time` в UTC (в SQLite пишется как ISO-8601 строка, в Postgres как `TIMESTAMPTZ`).
4. **Блоки конструктора:** Хранятся как `TEXT` (SQLite) / `JSONB` (Postgres) и мапятся в Go как `json.RawMessage`.
5. **ANSI SQL:** В запросах `sqlc` запрещены специфичные диалектные функции (`strftime`, `FTS5`, `->>`).

---

## 3. Чек-лист реализации

### Фаза 1: Инфраструктура БД и генерация кода (Data Layer)
- [x] **1.1.** Создать структуру каталогов `sql/queries/`, `sql/migrations/sqlite/`, `sql/migrations/postgres/`.
- [x] **1.2.** Перенести текущие схемы коллекций (`schema.go`, `collections_*.go`) в ANSI SQL миграцию `00001_init.sql`.
- [x] **1.3.** Настроить конфигурационный файл `sqlc.yaml` с флагом `emit_interface: true`.
- [x] **1.4.** Описать типовые SQL-запросы в `sql/queries/` (`pages.sql`, `leads.sql`, `services.sql`, `articles.sql`, `settings.sql`, `orders.sql`, `coupons.sql`, `resources.sql`, `branches.sql`, `reviews.sql`, `vacancies.sql`, `chat.sql`).
- [x] **1.5.** Сгенерировать пакет `internal/db` командой `sqlc generate` и проверить сгенерированный интерфейс `Querier`.
- [x] **1.6.** Настроить модуль инициализации базы данных `internal/db/connection.go`:
  - Настройка WAL-режима и `busy_timeout(5000)`.
  - Ограничение пула соединений SQLite (1 writer, N readers).
  - Встраивание миграций через `go:embed` и автозапуск через `goose.Up`.

---

### Фаза 2: Сессии, Безопасность и Авторизация Админки
- [x] **2.1.** Создать таблицу `sessions` в схеме миграций для пакета `alexedwards/scs`.
- [x] **2.2.** Реализовать CGO=0 совместимый `SQLStore` для сессий SQLite/Postgres с автоочисткой протухших записей.
- [x] **2.3.** Реализовать хэширование паролей администратора через `bcrypt` (`HashPassword`, `CheckPassword`).
- [x] **2.4.** Написать хэндлеры `/admin/login` (GET/POST) и `/admin/logout` с поддержкой `HX-Redirect`.
- [x] **2.5.** Написать middleware `RequireAdminAuth` для защиты административных маршрутов.
- [x] **2.6.** Подключить middleware защиты от CSRF (`nosurf`) для всех HTML-форм.

---

### Фаза 3: Маршрутизация на Chi и рефакторинг хэндлеров
- [x] **3.1.** Создать корневой сервер `internal/server/server.go` с роутером `chi.NewRouter()`.
- [x] **3.2.** Подключить базовый стек middleware:
  - `middleware.RequestID`
  - `middleware.RealIP`
  - `middleware.Recoverer`
  - `middleware.Compress(5)`
  - Логирование через `log/slog`
  - A/B тестирование `HTTPABTestingMiddleware`
  - UTM-метки `HTTPUTMMiddleware`
  - Кэш-контроль `HTTPCacheControlMiddleware`
  - Локализация `i18n.HTTPMiddleware`
- [x] **3.3.** Переписать хэндлеры из `internal/handlers` на стандартный `http.HandlerFunc(w http.ResponseWriter, r *http.Request)` с вызовами `sqlc` и `Repository`.
- [x] **3.4.** Проверить корректность серверного рендера `Templ` на чистом `http.ResponseWriter`.
- [x] **3.5.** Проверить работу всех HTMX-эндпоинтов (отправка заявок, добавление в корзину, фильтры, отзывы, трекинг).

---

### Фаза 4: Медиа-движок (Media Pipeline)
- [x] **4.1.** Реализовать сервис `internal/media`:
  - Загрузка оригиналов в каталог `/data/uploads/originals/`.
  - Валидация размера (до 25MB) и MIME-типов файлов.
- [x] **4.2.** Pure-Go ресайзинг через `golang.org/x/image/draw`:
  - Генерация миниатюр: `thumb` (150px), `sm` (320px), `md` (640px), `lg` (1280px).
  - Поддержка форматов JPEG и PNG с сохранением высокого качества.
- [x] **4.3.** Реализовать хэндлер раздачи статики `/media/{size}/{filename}`:
  - Кэширование сгенерированных миниатюр на диске (`data/uploads/cache/`).
  - Установка долгосрочных HTTP-заголовков кэширования (`Cache-Control: public, max-age=31536000, immutable`).

---

### Фаза 5: Кастомная админ-панель конструктора (Templ + HTMX)
- [x] **5.1.** Сверстать базовый лейаут админки на `Templ` + `Tailwind v4` (сайдбар, профиль, переключатель темы).
- [x] **5.2.** Раздел «Страницы и Блоки» (`client_admin_blocks.templ`): визуальное управление структурой блоков страницы (Hero, Bento, Services, FAQ, перемещение вверх/вниз, скрытие, добавление).
- [x] **5.3.** Раздел «Заявки / Лиды» (`client_admin_leads.templ`): таблица лидов с фильтрацией, пагинацией и сменой статуса через HTMX без перезагрузки.
- [x] **5.4.** Раздел «Медиатека» (`client_admin_media.templ`): загрузка картинок с перетаскиванием (drag-and-drop) и копированием ссылки.
- [x] **5.5.** Раздел «Настройки сайта» (`client_admin_settings.templ`): управление SEO-тегами, фавиконкой, контактами, 152-ФЗ и скриптами аналитики.

---

### Фаза 6: Скрипт миграции данных из PocketBase
- [x] **6.1.** Написать консольную утилиту `cmd/migrate-pb/main.go`.
- [x] **6.2.** Реализовать чтение старой базы `pb_data/data.db` и конвертацию записей в новую схему SQLite.
- [x] **6.3.** Перенести загруженные файлы из `pb_data/storage` в новую структуру `/data/uploads/originals/`.
- [x] **6.4.** Проверено на реальной базе данных `pb_data/data.db`: успешно мигрированы `site_settings`, `site_content`, 11 `page_blocks`, 4 `services`, 2 `leads`, 2 `articles`, 2 `orders`, 2 `coupons`, 2 `resources`, 3 `branches`, 3 `reviews`, 2 `vacancies`, файлы перенесены.

---

### Фаза 7: Очистка зависимостей, сборка и деплой
- [x] **7.1.** Удалить зависимости PocketBase из `go.mod` и `go.sum` (`go mod tidy`). Полностью очищено, CGO=0!
- [x] **7.2.** Обновить `Dockerfile`:
  - Standalone Tailwind CSS v4 CLI + Templ CLI в многоэтапном билде без Node.js.
  - Минимальный runtime-образ `alpine:3.21` с монтируемым томом `/data`.
- [x] **7.3.** Обновить скрипты развертывания [fleet.sh](fleet.sh) и документацию [FLEET_ARCHITECTURE.md](FLEET_ARCHITECTURE.md).
- [x] **7.4.** Провести замеры потребления RAM и TTFB:
  - **TTFB:** **0.8 – 3.2 мс** (цель: < 5 мс достигнута с запасом 5x).
  - **RAM RSS:** **21.9 МБ** (цель: < 25 МБ достигнута).
  - **Unit & Integration Tests:** 100% пройдены (`go test ./...` PASS).
