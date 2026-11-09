# ⚡ SiteCraft Engine — Fleet of Binaries Architecture

Архитектура мультисайтового развертывания по модели **Fleet of Binaries (Изолированные инстансы)** для малого и среднего бизнеса.

---

## 🎯 Архитектурный принцип

Каждый клиент/сайт получает:
1. **Изолированную базу данных SQLite** (`data/sitecraft.db` в режиме WAL) или подключение к PostgreSQL.
2. **Изолированное хранилище файлов** (`data/uploads/originals` + `data/uploads/cache`).
3. **Автономную панель управления SiteCraft Admin** (`http://domain:port/admin`).
4. **Собственные секреты и настройки** (токены Telegram, ЮKassa, webhook CRM) в таблице `site_settings`.
5. **Выделенный порт или socket**, направленный через обратный прокси (Nginx, Caddy, Traefik).

```
                      ┌──────────────────────┐
                      │    Nginx / Caddy     │
                      │  (SSL / Reverse Proxy)│
                      └──────────┬───────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │ (autoservice.ru)      │ (stomatolog.ru)       │ (notary.ru)
         ▼                       ▼                       ▼
┌──────────────────┐   ┌──────────────────┐   ┌──────────────────┐
│ SiteCraft #1     │   │ SiteCraft #2     │   │ SiteCraft #3     │
│ Port 8091        │   │ Port 8092        │   │ Port 8093        │
│ RAM: ~15MB       │   │ RAM: ~15MB       │   │ RAM: ~15MB       │
│ Fleet: auto-01   │   │ Fleet: clinic-02 │   │ Fleet: legal-03  │
└────────┬─────────┘   └────────┬─────────┘   └────────┬─────────┘
         │                      │                      │
         ▼                      ▼                      ▼
┌──────────────────┐   ┌──────────────────┐   ┌──────────────────┐
│  /data (DB1)     │   │  /data (DB2)     │   │  /data (DB3)     │
└──────────────────┘   └──────────────────┘   └──────────────────┘
```

---

## 🛠 Управление через `fleet.sh`

В репозиторий встроен CLI-инструмент [fleet.sh](fleet.sh).

### 1. Создание нового сайта клиента
```bash
./fleet.sh create auto-express --preset=auto --theme=auto --port=8091
./fleet.sh create dental-clinic --preset=medical --theme=medical --port=8092
./fleet.sh create law-firm      --preset=legal-swiss --theme=legal --port=8093
```
*Доступные пресеты:* `notary`, `legal`, `legal-swiss`, `legal-british`, `legal-milano`, `legal-corporate`, `medical`, `auto`, `repair`, `beauty`, `cyber`, `delivery`, `academy`.

### 2. Запуск и управление
```bash
# Запустить инстанс в фоновом режиме
./fleet.sh start auto-express

# Посмотреть состояние всех сайтов (порт, статус, PID)
./fleet.sh list

# Остановить сайт
./fleet.sh stop auto-express

# Перезапустить
./fleet.sh restart auto-express
```

### 3. Экспорт в статический сайт (SSG)
```bash
# Генерация плоского HTML/CSS для размещения на GitHub Pages, Cloudflare Pages или S3
./fleet.sh export auto-express ./dist-auto
```

### 4. Резервное копирование и передача клиенту («под ключ»)
```bash
./fleet.sh backup auto-express
```
Скрипт генерирует единый автономный архив `.tar.gz`, включающий всю SQLite базу и загруженный клиентом медиаконтент. Вы можете передать этот архив клиенту или перенести на любой VPS в одну команду.

---

## 🖥 Веб-панель Fleet Master Dashboard (`/fleet` и `/client/tab/fleet`)

Помимо консольного скрипта `fleet.sh`, для веб-студий и агентств доступен наглядный веб-интерфейс:

1. **Мониторинг в реальном времени:** Список всех инстансов сайтов, выделенные порты, текущий статус (🟢 Активен, 🔴 Остановлен), системный PID и объем потребляемой памяти RAM.
2. **Управление в 1 клик:**
   * Запуск (`start`), остановка (`stop`), перезапуск (`restart`) и бэкап (`backup`) без подключения по SSH через API `POST /api/fleet/action`.
3. **Развертывание новых клиентов из браузера:**
   * Форма создания инстанса (`POST /api/fleet/create`) с указанием ID сайта, выбором пресета ниши, темы оформления и автоматическим выделением свободного порта.
4. **Интеграция с клиентским кабинетом:**
   * Доступен как самостоятельный пульт `/fleet`, так и в виде отдельной вкладки `/client/tab/fleet` в панели управления владельца инстанса (при включенном флаге `enable_fleet` в `site_settings`).

---

## 🌐 Пример конфигурации Caddy (Reverse Proxy с авто-SSL)

```caddy
autoservice-pro.ru {
    reverse_proxy 127.0.0.1:8091
}

dental-premium.ru {
    reverse_proxy 127.0.0.1:8092
}
```

---

## ⚡ Преимущества перед классическим SaaS Multi-Tenant

* **Zero Leaks:** Исключена вероятность программной утечки данных одного клиента другому.
* **Независимые бэкапы:** Восстановление базы одного клиента не затрагивает остальных.
* **Масштабируемость без оверхеда:** Потребление RAM одного сайта составляет всего **~20–35 МБ**. На VPS за $6–10 (2 ГБ RAM) комфортно работают **40–50 клиентских сайтов**.
