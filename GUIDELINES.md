# 📘 SiteCraft Starter Kit — Руководство разработчика (Guidelines)

> **SiteCraft** — это универсальный модульный стартовый набор (Boilerplate) для создания сверхбыстрых коммерческих сайтов и лендингов с панелью управления на стеке **Go 1.27 + PocketBase + Templ + HTMX + Tailwind CSS**.

---

## 📑 Содержание

1. [Архитектурная концепция](#1-архитектурная-концепция)
2. [Структура проекта](#2-структура-проекта)
3. [Быстрый старт и развертывание](#3-быстрый-старт-и-развертывание)
4. [Система тем и дизайн-токенов Tailwind](#4-система-тем-и-дизайн-токенов-tailwind)
5. [In-Memory кэширование (Sync.RWMutex)](#5-in-memory-кэширование-syncrwmutex)
6. [Управление нишами и пресетами](#6-управление-нишами-и-пресетами)
7. [Формы, валидация и антиспам](#7-формы-валидация-и-антиспам)
8. [Модуль доступности для слабовидящих (ГОСТ)](#8-модуль-доступности-для-слабовидящих-гост)
9. [Центр уведомлений (Telegram и CRM Webhooks)](#9-центр-уведомлений-telegram-и-crm-webhooks)
10. [Docker и продакшен-деплой](#10-docker-и-продакшен-деплой)

---

## 1. Архитектурная концепция

* **Единый бинарник (~25 МБ):** Веб-сервер, база данных SQLite, админка PocketBase (`/_/`), шаблоны Templ и статика упаковываются в один файл.
* **Нулевой рантайм Node.js в продакшене:** Стили и шаблоны компилируются заранее (`templ generate`, `tailwindcss -i ... -o ... --minify`).
* **Zero-JS фундамент (Progressive Enhancement):** Вся базовая функциональность работает через серверный рендеринг и HTMX-запросы.
* **RAM 20–40 МБ:** Позволяет запускать десятки независимых клиентских инстансов на одном VPS за 5$/мес.
* **TTFB < 10 мс:** Быстрая отдача HTML страниц за счет in-memory кэширования настроек и контента.

---

## 2. Структура проекта

```
.
├── cmd/                          # Точки входа (при необходимости расширения)
├── internal/
│   ├── cache/                    # Потокобезопасный in-memory кэш (RWMutex)
│   ├── data/                     # Схема SQLite, хуки PocketBase, сидер пресетов
│   ├── handlers/                 # HTTP-валидаторы форм и модульные тесты
│   ├── models/                   # Модели Go (SiteSettings, SiteContent, Lead, ServiceItem)
│   └── notifications/            # Асинхронные пуши в Telegram и CRM Webhook
├── presets/                      # Готовые JSON-пресеты (legal, medical, auto, beauty)
├── public/                       # Скриншоты и медиа-документация
├── ui/
│   ├── assets/                   # input.css, скомпилированный styles.css, картинки
│   ├── components/               # Переиспользуемые Templ-компоненты
│   │   ├── a11y_bar.templ        # Панель для слабовидящих (ГОСТ, Vanilla JS)
│   │   ├── booking.templ         # Форма лидгена (simple) или календарь слотов (slots)
│   │   ├── footer.templ          # Подвал сайта
│   │   ├── header.templ          # Навбар, a11y тумблер и переключатель тем
│   │   ├── hero.templ            # Главный экран
│   │   ├── map.templ             # Ленивая Яндекс Карта (hx-trigger="revealed")
│   │   ├── schedule.templ        # График работы и телефоны
│   │   └── services.templ        # Сетка услуг и прайс-лист
│   ├── layouts/
│   │   └── base.templ            # Базовый HTML5-каркас, a11y движок, SEO
│   └── pages/
│       └── home.templ            # Главная страница (сборка из компонентов)
├── Dockerfile                    # Multi-stage сборка минимального Docker-образа
├── docker-compose.yml            # Сервис с монтированием sqlite-базы в том pb_data
├── init-project.sh               # Интерактивный CLI-скрипт сборки нового проекта
├── main.go                       # Инициализация PocketBase и HTTP-роутов
└── tailwind.config.js            # Семантические токены цветов с поддержкой тем
```

---

## 3. Быстрый старт и развертывание

### Требования
* Go **1.27+**
* `templ` CLI (`go install github.com/a-h/templ/cmd/templ@latest`)
* `tailwindcss` CLI

### Создание и сборка проекта через CLI:
```bash
./init-project.sh --name="my-site" --preset=medical --theme=medical
```

Скрипт автоматически:
1. Компилирует шаблоны Templ (`templ generate`).
2. Собирает оптимизированный CSS (`tailwindcss --minify`).
3. Собирает бинарник `bin/my-site`.

### Запуск сервера:
```bash
./bin/my-site serve --http=127.0.0.1:8090 --preset=medical
```

* **Сайт:** `http://127.0.0.1:8090`
* **Админка PocketBase:** `http://127.0.0.1:8090/_/` (создайте первого супер-админа при первом входе).

---

## 4. Система тем и дизайн-токенов Tailwind

Для исключения конфликта между чисткой неиспользуемых классов Tailwind (Tree-shaking) и динамическими темами используются **семантические токены CSS-переменных**.

В шаблонах применяются только семантические классы:
* `bg-primary`, `bg-primary-dark`, `bg-primary-light`
* `text-accent`, `bg-accent`, `hover:bg-accent-hover`
* `bg-surface`, `bg-surface-muted`
* `text-content`, `text-content-muted`

Физические палитры настраиваются через атрибут `[data-theme="..."]` в `ui/assets/input.css`:

```css
/* МЕДИЦИНА / КЛИНИКА */
html[data-theme="medical"] {
  --color-primary: 13 148 136;
  --color-primary-dark: 15 118 110;
  --color-accent: 14 165 233;
  --color-surface: 248 250 252;
}

/* АВТОСЕРВИС */
html[data-theme="auto"] {
  --color-primary: 225 29 72;
  --color-primary-dark: 159 18 57;
  --color-accent: 245 158 11;
  --color-surface: 248 250 252;
}
```

Готовые встроенные темы: `legal`, `medical`, `auto`, `tech`, `dark`, `colorblind`.

---

## 5. In-Memory кэширование (Sync.RWMutex)

Для предотвращения нагрузки на SQLite при каждом HTTP-запросе данные настроек (`site_settings`), контента (`site_content`) и услуг (`services`) кэшируются в оперативной памяти сервера через `internal/cache/cache.go`.

### Инвалидация кэша:
Кэш автоматически сбрасывается в реальном времени при любых изменениях в панели управления PocketBase благодаря хукам в `internal/data/schema.go`:

```go
app.OnRecordAfterUpdateSuccess("site_settings", "site_content", "services").BindFunc(func(e *core.RecordEvent) error {
    invalidateByCollection(e.Record.Collection().Name)
    return e.Next()
})
```
* **Время выборки настроек:** **0 мс (O(1) в RAM)**.

---

## 6. Управление нишами и пресетами

Пресеты применяются **один раз на этапе старта проекта** или сидирования, не ломая структуру работающего сайта.

Каждый пресет хранится в `presets/<name>.json`:
* `presets/legal.json` — Нотариальные конторы, адвокаты, юридические бюро.
* `presets/medical.json` — Клиники, стоматологии, диагностические центры.
* `presets/auto.json` — Автосервисы, СТО, детейлинг-центры.
* `presets/beauty.json` — Салоны красоты, барбершопы, спа.

Применить пресет можно при запуске флагом:
```bash
./bin/my-site serve --preset=auto
```

---

## 7. Формы, валидация и антиспам

### Два режима работы (`booking_mode`):
1. **`simple` (Zero-JS Лидген):** Имя + Телефон + Вопрос (идеально для опта, автосервисов, B2B).
2. **`slots` (Интерактивная запись):** Выбор даты и слота времени из календаря со статусом занятости.

### Антиспам Honeypot:
В форме присутствует скрытое поле-ловушка для спам-ботов:
```html
<div class="hidden" aria-hidden="true">
    <label>URL: <input type="text" name="company_website_url" tabindex="-1" autocomplete="off"/></label>
</div>
```
Если робот заполняет это поле, сервер блокирует отправку заявки с ошибкой 422 без показа капчи реальным клиентам.

### Гибкая структура лидов:
Таблица `leads` хранит фиксированные поля (`name`, `phone`, `email`, `status`, `source`, `booking_date`, `booking_time`) и универсальный JSON-объект `payload`, куда попадают любые произвольные данные из форм без изменения схемы SQLite.

---

## 8. Модуль доступности для слабовидящих (ГОСТ)

Режим доступности по ГОСТ Р 52872-2019 работает на **нативном Vanilla JS (< 1.5 КБ)** без Alpine.js:
* 3 размера шрифта (`normal`, `large`, `xlarge`).
* 3 контрастные схемы (`black-on-white`, `white-on-black`, `blue-on-cyan`).
* Увеличение межбуквенного интервала.
* Полное скрытие визуальных изображений (`img`, `iframe`).
* Состояние запоминается в `localStorage`.

Тумблер `enable_a11y` в `site_settings` позволяет включить или выключить модуль в один клик.

---

## 9. Центр уведомлений (Telegram и CRM Webhooks)

Модуль `internal/notifications/notifications.go` отправляет уведомления о новых заявках в фоновом режиме (горутины) без задержки ответа пользователю:

1. **Telegram:** HTML-сообщение в указанный чат/канал с деталями записи и кастомными полями `payload`.
2. **CRM Webhook:** POST-запрос с JSON-моделью заявки на URL внешней системы (amoCRM, Битрикс24, n8n, Make).

Настройки задаются через переменные в админке `site_settings`:
* `telegram_bot_token`
* `telegram_chat_id`
* `webhook_url`

---

## 10. Docker и продакшен-деплой

Для развертывания подготовлены [`Dockerfile`](file:///home/zerg/Projects/web/lawyer/Dockerfile) и [`docker-compose.yml`](file:///home/zerg/Projects/web/lawyer/docker-compose.yml).

### Запуск в Docker:
```bash
docker compose up -d --build
```

### Бекапы и данные:
Все данные (SQLite база данных и загруженные файлы) сохраняются в том `pb_data`. Для создания бекапа достаточно скопировать директорию `pb_data`.
