# ⚡ SiteCraft Engine & Starter Kit

**SiteCraft** — универсальный модульный движок и стартовый набор для создания сверхбыстрых коммерческих сайтов, каталогов услуг и лендингов малого и среднего бизнеса (SMB) на стеке **Go 1.27 + PocketBase + Templ + HTMX + Alpine.js + Tailwind CSS**.

![PageSpeed](https://img.shields.io/badge/PageSpeed-100%2F100-success?style=flat-square)
![Binary Size](https://img.shields.io/badge/Single%20Binary-~25MB-blue?style=flat-square)
![RAM Usage](https://img.shields.io/badge/RAM%20Usage-20--35MB-orange?style=flat-square)
![Interactive UI](https://img.shields.io/badge/UI-HTMX%20%2B%20Alpine.js-purple?style=flat-square)
![Multi-Tenant](https://img.shields.io/badge/Multi--Tenant-Fleet%20of%20Binaries-emerald?style=flat-square)

---

## 🎯 Ключевые преимущества

1. **Единый автономный бинарник (~25 МБ):** Веб-сервер, база данных SQLite, панель управления (`/_/`), шаблоны и ассеты скомпилированы в один файл.
2. **Нулевой рантайм Node.js в продакшене:** Стили и шаблоны компилируются заранее при сборке.
3. **In-Memory Cache (sync.RWMutex):** Отдача страниц за **< 10 мс (TTFB)** с автоматической инвалидацией при редактировании данных в админке.
4. **Связка HTMX + Alpine.js:** 
   * **Alpine.js + @alpinejs/mask** — мгновенные клиентские микро-интеракции (мобильное бургер-меню, маски номеров телефонов, календарь).
   * **HTMX** — серверные AJAX-действия без перезагрузки страниц (отправка заявок, lazy-loading карты, проверка занятых слотов).
5. **Динамический Theming Engine (управление цветами и шрифтами из PocketBase):** Возможность задавать `primary_color`, `accent_color`, `background_color` и `font_family` на лету из панели управления без пересборки Tailwind CSS!
6. **Реестр компонентов (Block Registry / Factory Pattern):** Расширяемая модульная архитектура (`ui/pages/registry.go`). Любые новые секции легко регистрируются через `RegisterBlock(type, renderer)`.
7. **Генератор статических сайтов (Static SSG Exporter):** Выгрузка сайта в чистый плоский HTML/CSS для бесплатного хостинга на GitHub Pages, Cloudflare Pages или S3.
8. **Оркестратор Fleet of Binaries (`./fleet.sh`):** Мультисайтовое развертывание для агентств и студий. 1 клиент = 1 изолированная база SQLite + личная админка. Потребление ~25 МБ RAM на сайт (до 50 сайтов на VPS за $10).
9. **Бизнес-модули "из коробки":**
   * **B2B-трекер заказов в реальном времени (`/track`):** Отслеживание готовности авто/техники с живым SSE/HTMX обновлением прогресс-бара и онлайн-оплатой через ЮKassa.
   * **SEO-Блог и База знаний (`/blog`):** Экспертные статьи, живой поиск HTMX, динамический `sitemap.xml` и счетчик просмотров.
   * **Telegram Mini Apps (TMA) E-commerce & Paywall:** Запуск магазина (`/shop`) с нативной кнопкой Telegram MainButton, беспарольная авторизация HMAC-SHA256 и закрытый VIP-клуб (`/vip`).
   * Онлайн-оплата **ЮKassa** со встроенным Sandbox.
   * Панель доступности для слабовидящих по **ГОСТ Р 52872-2019** (микро-скрипт Vanilla JS < 1.5 КБ).
   * Интерактивная запись по слотам времени или простая лидген-форма.
   * Невидимый Honeypot для ботов и серверная XSS-санитизация через `bluemonday`.
   * Асинхронные уведомления в **Telegram Bot** и **CRM Webhooks** (amoCRM / Битрикс24).

---

## 🏛 Архитектура проекта

### 1. Архитектура ядра и потока данных (Engine Architecture)

```mermaid
flowchart TD
    Client["Пользователь / Браузер / Telegram App"]
    
    subgraph ReverseProxy["Обратный прокси (Опционально)"]
        Caddy["Caddy / Nginx (SSL & Маршрутизация доменов)"]
    end

    subgraph Binary["Единый автономный бинарник SiteCraft (~25 МБ)"]
        direction TB
        
        Router["Echo HTTP Router & Middleware\n(Honeypot, BlueMonday Sanitize, Static Embed)"]
        
        subgraph BusinessHandlers["Бизнес-хэндлеры (internal/handlers)"]
            H_Web["Web Handler (/ /blog /shop)"]
            H_Track["B2B Tracker (/track)"]
            H_TMA["TMA Mini App & VIP Club (/vip)"]
            H_Lead["Lead Generator & Slots (/booking)"]
            H_Pay["Payment Handler (/payment)"]
        end
        
        subgraph EventBusLayer["Событийная шина & Сервисы (internal/events & internal/notifications)"]
            EBus["In-Memory Event Bus (bus.go)\n(LeadCreatedEvent, OrderPaidEvent)"]
            Subscribers["Subscribers / Notifier Interface\n(TelegramNotifier, WebhookNotifier)"]
            PayGateway["PaymentProvider Interface\n(YooKassa, MockProvider)"]
        end

        subgraph RenderingLayer["Шаблонизация & UI (ui/pages & ui/sections)"]
            Registry["Block Registry / Factory Pattern\n(Hero, Services, Booking, FAQ, Gallery, etc.)"]
            Templ["Templ Go Templates (Type-safe HTML)"]
            ClientLibs["HTMX (AJAX) + Alpine.js (Микро-интеракции)"]
        end

        subgraph CoreData["Слой данных & Кэш"]
            Cache["In-Memory Cache (sync.RWMutex)\n< 10ms TTFB"]
            PB["Embedded PocketBase (Go Framework)"]
            SQLite[("Встроенная SQLite БД\n(pb_data/data.db)")]
        end

        SSG["Static SSG Exporter\n(Flat HTML/CSS в ./dist)"]
    end

    subgraph ExternalServices["Внешние сервисы"]
        TG["Telegram Bot (Алерты заявок)"]
        CRM["CRM Webhooks (amoCRM / Битрикс24)"]
        Yoo["ЮKassa API (Эквайринг)"]
    end

    Client --> Caddy
    Caddy --> Router
    Client -.-> Router

    Router --> BusinessHandlers
    H_Web & H_Track & H_TMA & H_Lead --> Registry
    Registry --> Templ
    Templ --> ClientLibs

    H_Lead -.->|Publish Event| EBus
    H_Pay -.->|Publish Event| EBus
    EBus --> Subscribers
    H_Pay --> PayGateway

    BusinessHandlers --> Cache
    Cache --> PB
    PB --> SQLite
    PB -.->|Инвалидация кэша OnRecordChange| Cache

    Subscribers -.->|Асинхронные уведомления| TG
    Subscribers -.->|Webhooks| CRM
    PayGateway -.->|API Запросы| Yoo
    
    PB --> SSG
```

### 2. Мультисайтовый парк (Fleet of Binaries)

```mermaid
flowchart LR
    Admin["Веб-студия / Администратор"] -->|./fleet.sh| FleetCLI["CLI Fleet Orchestrator"]
    
    FleetCLI --> SiteA["Site 1: Auto Service\nПорт :8091 (SQLite A)"]
    FleetCLI --> SiteB["Site 2: Pizza Delivery\nПорт :8092 (SQLite B)"]
    FleetCLI --> SiteC["Site 3: VIP Law Club\nПорт :8093 (SQLite C)"]
    FleetCLI --> SiteD["Site N: Dental Clinic\nПорт :8094 (SQLite N)"]

    CaddyProxy["Caddy Reverse Proxy"] -->|auto.site.ru| SiteA
    CaddyProxy -->|pizza.site.ru| SiteB
    CaddyProxy -->|law.site.ru| SiteC
    CaddyProxy -->|clinic.site.ru| SiteD
```

---

## 🚀 Управление парком сайтов (Fleet of Binaries)

Для создания и обслуживания десятков клиентских сайтов используйте встроенный CLI-оркестратор [fleet.sh](fleet.sh):

```bash
# 1. Создать изолированный сайт клиента с нишевым пресетом
./fleet.sh create auto-service   --preset=repair   --theme=auto    --port=8091  # СТО с B2B-трекером заказов
./fleet.sh create pizza-craft    --preset=delivery --theme=warm    --port=8092  # Доставка еды + TMA корзина
./fleet.sh create vip-lawyers    --preset=academy  --theme=legal   --port=8093  # Закрытый VIP-клуб / курсы
./fleet.sh create dental-clinic  --preset=medical  --theme=medical --port=8094  # Стоматология с записью

# 2. Запуск и управление инстансами в фоне
./fleet.sh start auto-express
./fleet.sh list                 # Таблица статусов (SITE_ID, PRESET, PORT, STATUS, PID)
./fleet.sh restart auto-express
./fleet.sh stop auto-express

# 3. Экспорт сайта клиента в статический HTML/CSS
./fleet.sh export auto-express ./dist-auto

# 4. Резервное копирование и передача клиенту «под ключ»
./fleet.sh backup auto-express  # Создает автономный архив backup_*.tar.gz с базой SQLite и файлами
```

Подробное описание архитектуры читайте в **[FLEET_ARCHITECTURE.md](FLEET_ARCHITECTURE.md)**.

---

## 🎨 Динамический Theming Engine & Модульная CSS-архитектура

Движок поддерживает гибкую многоуровневую систему оформления:

### 1. Семантические CSS-токены и модульные стили (`ui/assets/css/`)
Стили декомпозированы на модули для удобной поддержки и расширения:
* [`tokens.css`](ui/assets/css/tokens.css) — семантические переменные в формате RGB triplets (`--color-primary`, `--color-accent`, `--color-surface`, `--color-content`) с поддержкой Tailwind alpha-value (`bg-primary/90`, `text-content`, `border-content/10`).
* [`components.css`](ui/assets/css/components.css) — типографика, кнопки, карточки и утилитарные классы.
* [`a11y.css`](ui/assets/css/a11y.css) — стили режимов доступности (ГОСТ Р 52872-2019): контрастные схемы (черный/белый, белый/черный, синий/голубой), управление размером шрифта, кернингом и скрытием изображений.
* [`animations.css`](ui/assets/css/animations.css) — ключевые кадры и анимации интерфейса.

### 2. Сезонные темы и режим День / Ночь
В шапку сайта ([header.templ](ui/components/header.templ)) встроен интерактивный переключатель (Alpine.js) с сохранением состояния в `localStorage`:
* **Базовые палитры:** Основная (`main`), Весна (`spring`), Лето (`summer`), Осень (`autumn`), Зима (`winter`).
* **Режим День / Ночь:** Мгновенное переключение между светлой и темной версией любой выбранной темы (`main-day`, `main-night`, `winter-night` и т.д.).
* **Специальные режимы:** Режим для людей с особенностями цветовосприятия (`colorblind` / high-contrast) и нишевые пресеты (`cyber`, `legal`, `medical`, `auto`).

### 3. Управление палитрой из админки PocketBase (`site_settings`)
Помимо встроенных CSS-тем, доступна кастомизация на лету без пересборки Tailwind CSS:
* `primary_color`: базовый цвет бренда (например, `#0d9488` или `13 148 136`)
* `accent_color`: акцентный цвет кнопок и плашек (например, `#f59e0b` или `245 158 11`)
* `background_color`: фон страниц сайта
* `font_family`: семейство шрифта Google Fonts (например, `Inter`, `Roboto`, `Montserrat`)

Стили динамически инжектируются в `:root` шаблона [base.templ](ui/layouts/base.templ).

---

## 🧩 Реестр компонентов (Block Registry)

Архитектура рендеринга секций построена на паттерне Factory / Registry в [ui/pages/registry.go](ui/pages/registry.go):

```go
// Регистрация кастомного блока (плагина)
pages.RegisterBlock("video_banner", func(b models.PageBlock, ctx pages.RenderContext) templ.Component {
    return components.VideoBanner(b.Payload)
})
```

Все блоки страницы упорядочиваются и настраиваются через коллекцию `page_blocks` в админке PocketBase.

---

## ⚡ Событийная шина и слабая связанность (Event-Driven Bus)

Бизнес-хэндлеры не вызывают напрямую внешние сервисы. Вместо этого используется внутрипроцессная шина событий (`internal/events/bus.go`):

```go
// Публикация события заявки из любого модуля
eventBus.Publish(ctx, events.LeadCreatedEvent{
    Settings: settings,
    Lead:     leadModel,
})
```

* **Подписчики (`internal/notifications/notifications.go`):** Telegram-бот и CRM-вебхуки подписываются на события изолированно, не блокируя HTTP-ответ посетителю сайта.
* **Интерфейс `Notifier`:** поддерживает композитный паттерн (`CompositeNotifier`) для одновременной отправки в несколько каналов.
* **Интерфейс `PaymentProvider`:** абстрагирует платежные системы (ЮKassa, Тинькофф, тестовый Mock-провайдер) с внедрением зависимостей через опции конструктора `NewPaymentHandler`.

---

## 📦 Статический экспорт (Static SSG)

Вы можете использовать SiteCraft не только как серверное приложение, но и как генератор статики:

```bash
./bin/sitecraft export --out=./dist --preset=medical
```
Команда собирает оптимизированный `index.html`, минифицированные стили и все медиа-ассеты в папку `./dist`, готовую к размещению на GitHub Pages, Vercel, Netlify или Cloudflare.

---

## 🏗 Архитектура шаблонов и последовательность компонентов

В SiteCraft каждый пресет задает свою уникальную воронку конверсии и порядок секций:

| Ниша / Пресет | Стратегия воронки | Последовательность секций и компонентов |
| :--- | :--- | :--- |
| **Юридический / Нотариус**<br>`presets/legal.json` | **Легитимность и регламент:** официальный статус, расписание и правовые стандарты. | `Hero` → `Schedule` → `Features` → `Services` → `Steps` → `Booking` → `FAQ` → `Map` |
| **Медицинский центр**<br>`presets/medical.json` | **Доверие к врачам:** оборудование, профильные врачи и подготовка к анализам. | `Hero` → `Features` → `Services` → `Team` → `Booking` → `Reviews` → `FAQ` → `Schedule` → `Map` |
| **Автосервис / СТО**<br>`presets/auto.json`, `presets/repair.json` | **Прайс-лист, быстрый лидген и B2B-трекер:** прозрачные цены на ТО, гарантия 12 месяцев. | `Hero` → `Services` → `Booking` → `Features` → `Steps` → `Reviews` → `FAQ` → `Map` → `Schedule` |
| **Салон красоты & SPA**<br>`presets/beauty.json` | **Visual First (Эстетика):** портфолио работ, бьюти-меню и карточки топ-мастеров. | `Hero` → `Gallery` → `Services` → `Features` → `Team` → `Reviews` → `Booking` → `FAQ` → `Map` |
| **Киберпанк & Web3**<br>`presets/cyber.json` | **Futuristic Experience:** темная high-tech эстетика, CyberHero, шрифт Bebas Neue. | `CyberHero` → `Features` → `Services` → `Booking` → `FAQ` |
| **Доставка еды / Dark Kitchen**<br>`presets/delivery.json` | **Food E-commerce & TMA:** меню, акции, онлайн-корзина с интеграцией Telegram WebApp. | `Hero` → `Services` → `Features` → `Steps` → `Reviews` → `Booking` → `FAQ` |
| **Академия & Онлайн-курсы**<br>`presets/academy.json` | **EdTech & Paywall:** программа обучения, лекторы, подписка на VIP-клуб. | `Hero` → `Features` → `Services` → `Team` → `Steps` → `Reviews` → `Booking` → `FAQ` |

---

## 📊 Сравнение с аналогами

| Критерий | SiteCraft (Go + PocketBase + HTMX) | WordPress (PHP + MySQL) | Tilda / Webflow (No-Code SaaS) | Next.js + Strapi (Node.js Fullstack) |
|---|---|---|---|---|
| **Потребление RAM** | **~20–35 МБ** | 250–600 МБ | Облако (SaaS) | 350–800 МБ |
| **TTFB (отклик)** | **< 10 мс** | 150–500 мс | 100–300 мс | 50–150 мс |
| **Сложность деплоя** | **1 бинарник** (или Docker) | LAMP/LEMP стек + плагины | 0 (кнопка в облаке) | Node.js рантайм + БД + Redis |
| **Встроенная админка** | **Да (PocketBase)** | Да (тяжелая CMS) | Да (визуальный редактор) | Да (в Headless CMS) |
| **Стоимость владения** | **$0** (до 50 сайтов на VPS за $10) | $5–15/мес за сайт | $12–25/мес за сайт | $15–30/мес |
| **Контроль данных** | **100% On-Premise (SQLite)** | 100% On-Premise | 0% (Vendor lock-in) | 100% On-Premise |
| **Интерактивность UI** | **HTMX + Alpine.js** | jQuery / тяжелые плагины | Базовые виджеты | React/Vue гидратация |
| **PageSpeed** | **95–100** | 50–80 | 60–85 | 85–95 |

---

## 🛠 Быстрый старт одного сайта

### Интерактивная сборка
```bash
./init-project.sh
```
Скрипт предложит выбрать сферу бизнеса, тему оформления и учетные данные суперпользователя.

### Ручной запуск бинарника
```bash
./bin/my-site serve --http=127.0.0.1:8090 --preset=medical
```
* **Сайт:** [http://127.0.0.1:8090](http://127.0.0.1:8090)
* **Панель управления PocketBase:** [http://127.0.0.1:8090/_/](http://127.0.0.1:8090/_/)
  * **Логин по умолчанию:** `admin@local.dev`
  * **Пароль по умолчанию:** `admin123456`

---

## 🐳 Развертывание в Docker

```bash
docker compose up -d --build
```
Все данные (база SQLite и загруженные файлы) сохраняются в томе `pb_data`.

---

## 📚 Документация

* **[FLEET_ARCHITECTURE.md](FLEET_ARCHITECTURE.md)** — Архитектура мультисайтового парка инстансов (Fleet of Binaries), интеграция с Caddy/Nginx.
* **[123.md](123.md)** — Спецификация и статус внедрения модулей движка.
* **[GUIDELINES.md](GUIDELINES.md)** — Руководство разработчика, кэширование, платежи и безопасность.
* **[SITECRAFT_ARCHITECTURE_PLAN.md](SITECRAFT_ARCHITECTURE_PLAN.md)** — Архитектурный RFC и первоначальный план проекта.
