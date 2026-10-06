# ⚡ SiteCraft Starter Kit

**SiteCraft** — универсальный модульный стартовый набор (Boilerplate) для создания сверхбыстрых коммерческих сайтов, каталогов услуг и лендингов со встроенной панелью управления на стеке **Go 1.27 + PocketBase + Templ + HTMX + Tailwind CSS**.

![PageSpeed](https://img.shields.io/badge/PageSpeed-100%2F100-success?style=flat-square)
![Binary Size](https://img.shields.io/badge/Single%20Binary-~25MB-blue?style=flat-square)
![RAM Usage](https://img.shields.io/badge/RAM%20Usage-20--40MB-orange?style=flat-square)
![Zero JS](https://img.shields.io/badge/Zero--JS-HTMX%20Foundation-purple?style=flat-square)

---

## 🎯 Ключевые преимущества

1. **Единый бинарник (~25 МБ):** Веб-сервер, база данных SQLite, панель управления (`/_/`), шаблоны и ассеты скомпилированы в один файл.
2. **Нулевой рантайм Node.js в продакшене:** Стили и шаблоны компилируются заранее при сборке.
3. **In-Memory Cache (sync.RWMutex):** Отдача страниц за **< 10 мс (TTFB)** с автоматической инвалидацией при редактировании данных в админке.
4. **Модульный конструктор секций (Block Reordering):** Управление порядком и видимостью блоков на странице прямо из PocketBase.
5. **Встроенный модуль онлайн-оплаты:** Поддержка **ЮKassa** и встроенный Mock Sandbox для тестирования.
6. **Антиспам и безопасность:** Невидимый Honeypot для ботов и серверная XSS-санитизация через `bluemonday`.
7. **Бизнес-модули "из коробки":**
   * Панель доступности для слабовидящих по **ГОСТ Р 52872-2019** (микро-скрипт Vanilla JS < 1.5 КБ).
   * Запись на приём по слотам времени или простая лидген-форма.
   * Асинхронные уведомления в **Telegram Bot** и **CRM Webhooks** (amoCRM / Битрикс24).
   * Поддержка 5 дизайн-тем: `legal`, `medical`, `auto`, `tech`, `dark`.

---

## 🚀 Быстрый старт

### 1. Сборка проекта через интерактивный CLI

```bash
# Инициализация и сборка проекта под нужную нишу (legal, medical, auto, beauty)
./init-project.sh --name="my-site" --preset=medical --theme=medical
```

Скрипт автоматически:
* Компилирует Templ шаблоны (`templ generate`).
* Минифицирует Tailwind стили с семантическими дизайн-токенами (`tailwindcss --minify`).
* Собирает автономный бинарный файл в `bin/my-site`.

### 2. Запуск сервера

```bash
./bin/my-site serve --http=127.0.0.1:8090 --preset=medical
```

* **Сайт:** [http://127.0.0.1:8090](http://127.0.0.1:8090)
* **Панель управления PocketBase:** [http://127.0.0.1:8090/_/](http://127.0.0.1:8090/_/) *(задайте email и пароль супер-админа при первом входе)*.

---

## 📦 Готовые нишевые пресеты

В каталоге `presets/` подготовлены готовые структуры данных:
* `presets/legal.json` — Юридические компании, адвокаты, нотариусы.
* `presets/medical.json` — Медицинские клиники, стоматологии, диагностика.
* `presets/auto.json` — Автосервисы, СТО, детейлинг-центры.
* `presets/beauty.json` — Салоны красоты, барбершопы, спа-центры.

---

## 🛠 Разработка (Live-Reload)

При наличии [devenv](https://devenv.sh):
```bash
devenv shell
dev # запуск air для автопересборки при изменении .go, .templ и .css файлов
```

Ручная сборка:
```bash
templ generate
tailwindcss -i ui/assets/input.css -o ui/assets/styles.css --minify
go test ./...
go build -ldflags="-s -w" -o bin/sitecraft .
```

---

## 🐳 Развертывание в Docker

В репозитории подготовлены мультистейдж `Dockerfile` и `docker-compose.yml`:

```bash
docker compose up -d --build
```
Все данные (база SQLite и загруженные файлы) сохраняются в томе `pb_data`. Для бэкапа достаточно скопировать эту папку.

---

## 📚 Документация

* **[GUIDELINES.md](GUIDELINES.md)** — Полное руководство разработчика, устройство In-Memory кэша, дизайн-токены Tailwind, платежи, санитизация и подробное сравнение со стеками WordPress, Tilda и Next.js.
* **[SITECRAFT_ARCHITECTURE_PLAN.md](SITECRAFT_ARCHITECTURE_PLAN.md)** — Архитектурная спецификация (RFC), обоснование технических решений и дорожная карта.
