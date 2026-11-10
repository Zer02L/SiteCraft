#!/usr/bin/env bash
set -e

# ==============================================================================
# 🚀 SiteCraft Fleet Manager (Fleet of Binaries / Multi-Tenant Orchestrator)
# ==============================================================================
# Управление изолированными экземплярами сайтов для малого и среднего бизнеса.
# Архитектурный паттерн: 1 клиент = 1 изолированный процесс/контейнер + pb_data.
# ==============================================================================

BOLD="\033[1m"
GREEN="\033[0;32m"
CYAN="\033[0;36m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
RESET="\033[0m"

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FLEET_DIR="${BASE_DIR}/fleet_sites"
ENGINE_BIN="${BASE_DIR}/bin/sitecraft"

print_header() {
    echo -e "${CYAN}${BOLD}"
    echo "=========================================================="
    echo "       ⚡ SiteCraft Engine — Fleet of Binaries Manager     "
    echo "=========================================================="
    echo -e "${RESET}"
}

usage() {
    print_header
    echo -e "${BOLD}Использование:${RESET}"
    echo "  $0 create <site-id> --preset=<niche> [--theme=<theme>] [--port=<port>]"
    echo "  $0 start <site-id>"
    echo "  $0 stop <site-id>"
    echo "  $0 restart <site-id>"
    echo "  $0 list"
    echo "  $0 backup <site-id>"
    echo "  $0 podman-up"
    echo "  $0 podman-down"
    echo ""
    echo -e "${BOLD}Примеры:${RESET}"
    echo "  $0 create auto-master --preset=auto --theme=auto --port=8091"
    echo "  $0 create stomatolog-plus --preset=medical --theme=medical --port=8092"
    echo "  $0 list"
    echo "  $0 backup auto-master"
    echo "  $0 podman-up"
    exit 1
}

ensure_engine() {
    if [ ! -f "$ENGINE_BIN" ]; then
        echo -e "${YELLOW}⚠️ Бинарник движка $ENGINE_BIN не найден. Выполняю сборку...${RESET}"
        if command -v templ &> /dev/null && command -v tailwindcss &> /dev/null; then
            templ generate
            tailwindcss -i ui/assets/input.css -o ui/assets/styles.css --minify
            go build -ldflags="-s -w" -o bin/sitecraft .
        elif command -v devenv &> /dev/null; then
            devenv shell -- bash -c "templ generate && tailwindcss -i ui/assets/input.css -o ui/assets/styles.css --minify && go build -ldflags='-s -w' -o bin/sitecraft ."
        else
            echo -e "${RED}❌ Не найден templ/tailwindcss для локальной сборки. Соберите бинарник через Podman или devenv.${RESET}"
            exit 1
        fi
        echo -e "${GREEN}✓ Движок успешно собран в bin/sitecraft${RESET}"
    fi
}

cmd_create() {
    local SITE_ID="$1"
    shift || true

    if [ -z "$SITE_ID" ]; then
        echo -e "${RED}❌ Укажите идентификатор сайта: $0 create <site-id>${RESET}"
        exit 1
    fi

    local PRESET="legal"
    local THEME=""
    local PORT="8091"
    local ADMIN_EMAIL="admin@${SITE_ID}.local"
    local ADMIN_PASS="admin123456"

    while [[ "$#" -gt 0 ]]; do
        case $1 in
            --preset=*) PRESET="${1#*=}" ;;
            --theme=*) THEME="${1#*=}" ;;
            --port=*) PORT="${1#*=}" ;;
            --admin-email=*) ADMIN_EMAIL="${1#*=}" ;;
            --admin-pass=*) ADMIN_PASS="${1#*=}" ;;
            *) echo "Неизвестный флаг: $1"; exit 1 ;;
        esac
        shift
    done

    if [ -z "$THEME" ]; then
        THEME="$PRESET"
    fi

    local SITE_PATH="${FLEET_DIR}/${SITE_ID}"
    if [ -d "$SITE_PATH" ]; then
        echo -e "${RED}❌ Сайт с ID '${SITE_ID}' уже существует в ${SITE_PATH}${RESET}"
        exit 1
    fi
    echo -e "${CYAN}🚀 Создание нового изолированного инстанса '${SITE_ID}'...${RESET}"
    mkdir -p "${SITE_PATH}/data"
    mkdir -p "${SITE_PATH}/logs"

    # Запись метаданных сайта
    cat <<EOF > "${SITE_PATH}/site.env"
SITE_ID=${SITE_ID}
PRESET=${PRESET}
THEME=${THEME}
PORT=${PORT}
ADMIN_EMAIL=${ADMIN_EMAIL}
CREATED_AT=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
EOF

    ensure_engine

    echo -e "📦 Применение пресета [${PRESET}] и инициализация SQLite..."
    "$ENGINE_BIN" apply-preset "${PRESET}" --theme="${THEME}" --db="${SITE_PATH}/data/sitecraft.db" > "${SITE_PATH}/logs/init.log" 2>&1

    # Генерация автономного скрипта запуска для инстанса
    cat <<EOF > "${SITE_PATH}/run.sh"
#!/usr/bin/env bash
cd "${BASE_DIR}"
exec "${ENGINE_BIN}" serve --port=":${PORT}" --db="${SITE_PATH}/data/sitecraft.db"
EOF
    chmod +x "${SITE_PATH}/run.sh"

    echo -e "${GREEN}${BOLD}✓ Экземпляр '${SITE_ID}' успешно создан!${RESET}"
    echo -e "  📍 Директория данных: ${CYAN}${SITE_PATH}/data${RESET}"
    echo -e "  🌐 Порт:              ${CYAN}${PORT}${RESET}"
    echo -e "  👤 Админка:           ${CYAN}http://127.0.0.1:${PORT}/admin${RESET}"
    echo -e "  🚀 Запуск инстанса:   ${YELLOW}./fleet.sh start ${SITE_ID}${RESET}"
}

cmd_start() {
    local SITE_ID="$1"
    local SITE_PATH="${FLEET_DIR}/${SITE_ID}"

    if [ ! -d "$SITE_PATH" ]; then
        echo -e "${RED}❌ Сайт '${SITE_ID}' не найден в ${FLEET_DIR}.${RESET}"
        exit 1
    fi

    source "${SITE_PATH}/site.env"

    local PID_FILE="${SITE_PATH}/app.pid"
    if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
        echo -e "${YELLOW}⚠️ Сайт '${SITE_ID}' уже запущен (PID: $(cat "$PID_FILE")).${RESET}"
        return
    fi

    ensure_engine

    echo -e "⚡ Запуск сайта '${SITE_ID}' на порту ${PORT}..."
    nohup "${SITE_PATH}/run.sh" > "${SITE_PATH}/logs/stdout.log" 2>&1 &
    local NEW_PID=$!
    echo "$NEW_PID" > "$PID_FILE"

    sleep 1
    if kill -0 "$NEW_PID" 2>/dev/null; then
        echo -e "${GREEN}✓ Сайт '${SITE_ID}' успешно запущен! PID: ${NEW_PID}${RESET}"
        echo -e "  👉 Публичный URL:  ${CYAN}http://127.0.0.1:${PORT}${RESET}"
        echo -e "  👉 Админка:        ${CYAN}http://127.0.0.1:${PORT}/admin${RESET}"
    else
        echo -e "${RED}❌ Не удалось запустить процесс. Проверьте ${SITE_PATH}/logs/stdout.log${RESET}"
    fi
}

cmd_stop() {
    local SITE_ID="$1"
    local SITE_PATH="${FLEET_DIR}/${SITE_ID}"
    local PID_FILE="${SITE_PATH}/app.pid"

    if [ ! -f "$PID_FILE" ]; then
        echo -e "${YELLOW}ℹ️ PID файл для '${SITE_ID}' отсутствует.${RESET}"
        return
    fi

    local PID
    PID="$(cat "$PID_FILE")"
    if kill -0 "$PID" 2>/dev/null; then
        echo -e "Остановка '${SITE_ID}' (PID: $PID)..."
        kill "$PID" || true
        rm -f "$PID_FILE"
        echo -e "${GREEN}✓ Процесс остановлен.${RESET}"
    else
        echo -e "${YELLOW}Процесс не был активен. Удаляю PID-файл.${RESET}"
        rm -f "$PID_FILE"
    fi
}

cmd_list() {
    print_header
    echo -e "${BOLD}Список сайтов в Fleet:${RESET}"
    echo "----------------------------------------------------------------------"
    printf "%-18s | %-10s | %-6s | %-10s | %-8s\n" "SITE ID" "PRESET" "PORT" "STATUS" "PID"
    echo "----------------------------------------------------------------------"

    if [ ! -d "$FLEET_DIR" ] || [ -z "$(ls -A "$FLEET_DIR" 2>/dev/null)" ]; then
        echo "  (нет созданных сайтов)"
        echo "----------------------------------------------------------------------"
        return
    fi

    for site in "$FLEET_DIR"/*; do
        if [ -d "$site" ] && [ -f "$site/site.env" ]; then
            source "$site/site.env"
            local PID_FILE="$site/app.pid"
            local STATUS="${RED}STOPPED${RESET}"
            local CURR_PID="-"

            if [ -f "$PID_FILE" ]; then
                local P
                P="$(cat "$PID_FILE")"
                if kill -0 "$P" 2>/dev/null; then
                    STATUS="${GREEN}RUNNING${RESET}"
                    CURR_PID="$P"
                fi
            fi

            printf "%-18s | %-10s | %-6s | %-19b | %-8s\n" "$SITE_ID" "$PRESET" "$PORT" "$STATUS" "$CURR_PID"
        fi
    done
    echo "----------------------------------------------------------------------"
}

cmd_backup() {
    local SITE_ID="$1"
    local SITE_PATH="${FLEET_DIR}/${SITE_ID}"

    if [ ! -d "$SITE_PATH" ]; then
        echo -e "${RED}❌ Сайт '${SITE_ID}' не найден.${RESET}"
        exit 1
    fi

    local BACKUP_NAME="backup_${SITE_ID}_$(date +'%Y%m%d_%H%M%S').tar.gz"
    local BACKUP_FILE="${BASE_DIR}/${BACKUP_NAME}"

    echo -e "📦 Создание архива сайта '${SITE_ID}'..."
    tar -czf "$BACKUP_FILE" -C "${FLEET_DIR}" "${SITE_ID}"
    echo -e "${GREEN}✓ Резервная копия готова: ${CYAN}${BACKUP_FILE}${RESET}"
    echo -e "  (Содержит всю базу SQLite, загруженные файлы клиентов и конфигурацию)."
}

cmd_export() {
    local SITE_ID="$1"
    local OUT_DIR="${2:-./dist}"

    ensure_engine

    if [ -n "$SITE_ID" ] && [ -d "${FLEET_DIR}/${SITE_ID}" ]; then
        echo -e "📦 Экспорт статики для инстанса '${SITE_ID}'..."
        "$ENGINE_BIN" export --dir="${FLEET_DIR}/${SITE_ID}/pb_data" --out="${OUT_DIR}"
    else
        echo -e "📦 Экспорт статики текущего сайта..."
        "$ENGINE_BIN" export --out="${OUT_DIR}"
    fi
    echo -e "${GREEN}✓ Статический сайт готов в папке: ${CYAN}${OUT_DIR}${RESET}"
}

cmd_podman_up() {
    print_header
    echo -e "${CYAN}🦭 Запуск SiteCraft в Podman (Rootless OCI)...${RESET}"
    if command -v podman &> /dev/null; then
        if podman compose version &> /dev/null 2>&1; then
            podman compose up -d --build
        elif command -v podman-compose &> /dev/null; then
            podman-compose up -d --build
        else
            echo -e "${YELLOW}podman compose/podman-compose не найдены. Выполняю сборку и запуск через podman CLI...${RESET}"
            podman build -t sitecraft:latest -f Containerfile .
            podman run -d --name sitecraft -p 8090:8090 -v sitecraft_data:/data:Z --restart unless-stopped sitecraft:latest
        fi
        echo -e "${GREEN}✓ Сервис SiteCraft запущен в Podman (порт http://127.0.0.1:8090).${RESET}"
    else
        echo -e "${RED}❌ Утилита podman не найдена в системе.${RESET}"
        exit 1
    fi
}

cmd_podman_down() {
    print_header
    echo -e "${CYAN}🦭 Остановка сервиса SiteCraft в Podman...${RESET}"
    if command -v podman &> /dev/null; then
        if podman compose version &> /dev/null 2>&1; then
            podman compose down
        elif command -v podman-compose &> /dev/null; then
            podman-compose down
        else
            podman stop sitecraft 2>/dev/null || true
            podman rm sitecraft 2>/dev/null || true
        fi
        echo -e "${GREEN}✓ Сервис остановлен.${RESET}"
    else
        echo -e "${RED}❌ Утилита podman не найдена в системе.${RESET}"
        exit 1
    fi
}

case "$1" in
    create)
        shift
        cmd_create "$@"
        ;;
    start)
        cmd_start "$2"
        ;;
    stop)
        cmd_stop "$2"
        ;;
    restart)
        cmd_stop "$2"
        cmd_start "$2"
        ;;
    list)
        cmd_list
        ;;
    backup)
        cmd_backup "$2"
        ;;
    export)
        cmd_export "$2" "$3"
        ;;
    podman-up|docker-up)
        cmd_podman_up
        ;;
    podman-down|docker-down)
        cmd_podman_down
        ;;
    *)
        usage
        ;;
esac
