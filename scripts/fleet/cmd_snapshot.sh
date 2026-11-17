#!/usr/bin/env bash
# ==============================================================================
# SiteCraft Fleet Manager — Snapshot, backup, export and rollback commands
# ==============================================================================

FLEET_SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -z "$BASE_DIR" ] || [ -z "$FLEET_DIR" ]; then
    # shellcheck source=scripts/fleet/common.sh
    source "${FLEET_SCRIPTS_DIR}/common.sh"
fi

if ! declare -f cmd_stop >/dev/null || ! declare -f cmd_start >/dev/null; then
    # shellcheck source=scripts/fleet/cmd_instance.sh
    source "${FLEET_SCRIPTS_DIR}/cmd_instance.sh"
fi

take_sqlite_snapshot() {
    local SITE_PATH="$1"
    local DB_FILE="${SITE_PATH}/data/sitecraft.db"
    local SNAPSHOT_DIR="${SITE_PATH}/data/snapshots"
    mkdir -p "${SNAPSHOT_DIR}"
    local TIMESTAMP
    TIMESTAMP="$(date +'%Y%m%d_%H%M%S')"
    local SNAPSHOT_FILE="${SNAPSHOT_DIR}/sitecraft_pre_update_${TIMESTAMP}.db"

    if [ -f "$DB_FILE" ]; then
        echo -e "📸 Создание обязательного преддеплойного снимка SQLite..."
        if command -v sqlite3 &> /dev/null; then
            sqlite3 "$DB_FILE" ".backup '${SNAPSHOT_FILE}'"
        else
            cp -p "$DB_FILE" "$SNAPSHOT_FILE"
            [ -f "${DB_FILE}-wal" ] && cp -p "${DB_FILE}-wal" "${SNAPSHOT_FILE}-wal"
            [ -f "${DB_FILE}-shm" ] && cp -p "${DB_FILE}-shm" "${SNAPSHOT_FILE}-shm"
        fi
        echo -e "${GREEN}✓ Снимок базы сохранен: ${CYAN}${SNAPSHOT_FILE}${RESET}"
        echo "$SNAPSHOT_FILE"
    else
        echo ""
    fi
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

cmd_rollback() {
    local SITE_ID="$1"
    local SITE_PATH="${FLEET_DIR}/${SITE_ID}"

    if [ ! -d "$SITE_PATH" ]; then
        echo -e "${RED}❌ Сайт '${SITE_ID}' не найден в ${FLEET_DIR}.${RESET}"
        exit 1
    fi

    local SNAPSHOT_DIR="${SITE_PATH}/data/snapshots"
    local LATEST_SNAPSHOT
    LATEST_SNAPSHOT="$(ls -t "${SNAPSHOT_DIR}"/sitecraft_pre_update_*.db 2>/dev/null | head -n 1 || true)"

    if [ -z "$LATEST_SNAPSHOT" ] || [ ! -f "$LATEST_SNAPSHOT" ]; then
        echo -e "${RED}❌ Снимков базы для отката не найдено в ${SNAPSHOT_DIR}.${RESET}"
        exit 1
    fi

    echo -e "${YELLOW}⏪ Выполняю откат инстанса '${SITE_ID}' к снимку:${RESET}"
    echo -e "  📍 ${CYAN}${LATEST_SNAPSHOT}${RESET}"

    # Остановка текущего процесса
    cmd_stop "$SITE_ID"

    # Восстановление базы данных
    cp -f "$LATEST_SNAPSHOT" "${SITE_PATH}/data/sitecraft.db"
    echo -e "${GREEN}✓ База данных восстановлена.${RESET}"

    # Откат бинарника если есть .prev
    if [ -f "${ENGINE_BIN}.prev" ]; then
        cp -f "${ENGINE_BIN}.prev" "${ENGINE_BIN}"
        chmod +x "${ENGINE_BIN}"
        echo -e "${GREEN}✓ Бинарник движка восстановлен из .prev${RESET}"
    fi

    # Перезапуск инстанса
    cmd_start "$SITE_ID"
    echo -e "${GREEN}${BOLD}✓ Инстанс '${SITE_ID}' успешно возвращен к стабильному состоянию!${RESET}"
}
