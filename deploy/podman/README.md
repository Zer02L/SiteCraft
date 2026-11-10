# 🦭 Развертывание SiteCraft в Podman

В проекте используется **Podman** в качестве основного контейнерного рантайма (Rootless OCI).

## Преимущества Podman
- **Rootless безопасность:** контейнеры выполняются без root-привилегий и без постоянного фонового демона (`dockerd`).
- **Совместимость со стандартами OCI:** сборка через канонический `Containerfile`.
- **Systemd Quadlet:** прямая декларативная интеграция с systemd через `.container` спецификации без костылей и внешних оберток.
- **SELinux изоляция:** тома монтируются с флагом `:Z` для безопасного разграничения доступа к данным SQLite.

---

## Варианты запуска

### 1. Быстрый запуск через Compose
```bash
# Через podman compose:
podman compose up -d --build

# Либо через podman-compose:
podman-compose up -d --build
```
Для остановки:
```bash
podman compose down
```

### 2. Сборка и запуск через CLI
```bash
# Сборка образа
podman build -t sitecraft:latest -f Containerfile .

# Запуск rootless контейнера с сохранением данных
podman run -d \
  --name sitecraft \
  -p 8090:8090 \
  -v sitecraft_data:/data:Z \
  --restart unless-stopped \
  sitecraft:latest
```

### 3. Автозапуск через Systemd Quadlet (Production)
Скопируйте файл `sitecraft.container` в каталог systemd пользователя:
```bash
mkdir -p ~/.config/containers/systemd
cp deploy/podman/sitecraft.container ~/.config/containers/systemd/

# Перезагрузите systemd генераторы
systemctl --user daemon-reload

# Запустите сервис
systemctl --user start sitecraft.service
systemctl --user enable sitecraft.service

# Просмотр логов
journalctl --user -u sitecraft.service -f
```
