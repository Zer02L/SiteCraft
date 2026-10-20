#!/usr/bin/env bash
# ==============================================================================
# 🛡️ SiteCraft Server Watchdog (Zero-RAM Linux VPS Monitor)
# Мониторинг: Диск, RAM, CPU Load, статус процесса SiteCraft
# Уведомления: Telegram Bot API с антиспам-кулдауном
# Потребление ресурсов: 0 МБ RAM (запускается по cron раз в 10-15 минут на ~0.1 сек)
# ==============================================================================

set -euo pipefail

# --- Конфигурация (можно переопределить через файл .watchdog.env рядом со скриптом) ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.watchdog.env"

if [[ -f "${ENV_FILE}" ]]; then
    # shellcheck disable=SC1090
    source "${ENV_FILE}"
fi

TG_BOT_TOKEN="${TG_BOT_TOKEN:-}"
TG_CHAT_ID="${TG_CHAT_ID:-}"

# Пороговые значения
DISK_THRESHOLD="${DISK_THRESHOLD:-85}"          # Процент заполнения диска (%)
MEM_THRESHOLD="${MEM_THRESHOLD:-10}"            # Минимум свободной RAM (%)
LOAD_MULTIPLIER="${LOAD_MULTIPLIER:-2}"         # Множитель Load Average к количеству CPU-ядер
SERVICE_NAME="${SERVICE_NAME:-sitecraft}"       # Имя systemd сервиса для проверки

# Файлы состояния для антиспама (не более 1 алерта в 2 часа по одной проблеме)
STATE_DIR="/tmp/sitecraft_watchdog"
mkdir -p "${STATE_DIR}"
ALERT_COOLDOWN=$((2 * 3600)) # 2 часа в секундах

send_telegram() {
    local message="$1"
    if [[ -z "${TG_BOT_TOKEN}" || -z "${TG_CHAT_ID}" ]]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] [STDOUT WARNING]: ${message}"
        return 0
    fi

    curl -s -m 10 -X POST "https://api.telegram.org/bot${TG_BOT_TOKEN}/sendMessage" \
        -d chat_id="${TG_CHAT_ID}" \
        -d text="${message}" \
        -d parse_mode="HTML" > /dev/null || true
}

should_alert() {
    local check_key="$1"
    local state_file="${STATE_DIR}/${check_key}.last"
    local now
    now=$(date +%s)

    if [[ -f "${state_file}" ]]; then
        local last_sent
        last_sent=$(cat "${state_file}")
        if (( now - last_sent < ALERT_COOLDOWN )); then
            return 1 # Кулдаун еще активен, спамить не нужно
        fi
    fi

    echo "${now}" > "${state_file}"
    return 0
}

clear_alert() {
    local check_key="$1"
    local state_file="${STATE_DIR}/${check_key}.last"
    if [[ -f "${state_file}" ]]; then
        rm -f "${state_file}"
        return 0 # Была ошибка, теперь восстановилось
    fi
    return 1 # Ошибки и не было
}

HOSTNAME="$(hostname)"
IP_ADDR="$(hostname -I 2>/dev/null | awk '{print $1}' || echo '127.0.0.1')"

# ==============================================================================
# 1. Проверка свободного места на диске (Root partition /)
# ==============================================================================
DISK_USAGE=$(df -P / | awk 'NR==2 {gsub("%","",$5); print $5}')
DISK_AVAIL_GB=$(df -h -P / | awk 'NR==2 {print $4}')

if [[ "${DISK_USAGE}" -gt "${DISK_THRESHOLD}" ]]; then
    if should_alert "disk"; then
        send_telegram "⚠️ <b>[${HOSTNAME} (${IP_ADDR})] Заканчивается место на диске!</b>%0A%0AЗаполнено: <b>${DISK_USAGE}%</b>%0AОсталось: <b>${DISK_AVAIL_GB}</b>%0AПорог: ${DISK_THRESHOLD}%%0A<i>Рекомендация: проверьте логи в /var/log или медиафайлы pb_data.</i>"
    fi
else
    if clear_alert "disk"; then
        send_telegram "✅ <b>[${HOSTNAME}] Дисковое пространство в норме:</b> ${DISK_USAGE}% (свободно ${DISK_AVAIL_GB})"
    fi
fi

# ==============================================================================
# 2. Проверка оперативной памяти (Available RAM)
# ==============================================================================
if command -v free >/dev/null 2>&1; then
    MEM_STATS=$(free -m | awk '/Mem:/ {print $2, $7}')
    TOTAL_MEM=$(echo "${MEM_STATS}" | awk '{print $1}')
    AVAIL_MEM=$(echo "${MEM_STATS}" | awk '{print $2}')
    
    if [[ "${TOTAL_MEM}" -gt 0 ]]; then
        AVAIL_PCT=$(( AVAIL_MEM * 100 / TOTAL_MEM ))
        if [[ "${AVAIL_PCT}" -lt "${MEM_THRESHOLD}" ]]; then
            if should_alert "memory"; then
                send_telegram "🚨 <b>[${HOSTNAME} (${IP_ADDR})] Критически мало памяти RAM!</b>%0A%0AДоступно: <b>${AVAIL_PCT}%</b> (${AVAIL_MEM} MB из ${TOTAL_MEM} MB)%0AПорог: ${MEM_THRESHOLD}%%0A<i>Рекомендация: проверьте утечки процессов командой btop.</i>"
            fi
        else
            if clear_alert "memory"; then
                send_telegram "✅ <b>[${HOSTNAME}] Память RAM стабилизировалась:</b> доступно ${AVAIL_PCT}% (${AVAIL_MEM} MB)"
            fi
        fi
    fi
fi

# ==============================================================================
# 3. Проверка нагрузки на процессор (Load Average 1m)
# ==============================================================================
CPU_CORES=$(grep -c ^processor /proc/cpuinfo 2>/dev/null || echo 1)
LOAD_1M=$(awk '{print $1}' /proc/loadavg 2>/dev/null || echo 0)
MAX_LOAD=$(( CPU_CORES * LOAD_MULTIPLIER ))

# Сравнение чисел с плавающей точкой в bash
if awk "BEGIN {exit !(${LOAD_1M} > ${MAX_LOAD})}"; then
    if should_alert "cpu"; then
        send_telegram "🔥 <b>[${HOSTNAME} (${IP_ADDR})] Высокая нагрузка CPU!</b>%0A%0ALoad Average: <b>${LOAD_1M}</b> (ядер: ${CPU_CORES}, порог: ${MAX_LOAD})%0A<i>Сервер находится под повышенной нагрузкой.</i>"
    fi
else
    if clear_alert "cpu"; then
        send_telegram "✅ <b>[${HOSTNAME}] Нагрузка CPU вернулась в норму:</b> Load ${LOAD_1M}"
    fi
fi

# ==============================================================================
# 4. Проверка статуса сервиса SiteCraft (Systemd)
# ==============================================================================
if command -v systemctl >/dev/null 2>&1 && systemctl is-enabled "${SERVICE_NAME}" >/dev/null 2>&1; then
    if ! systemctl is-active --quiet "${SERVICE_NAME}"; then
        if should_alert "service"; then
            send_telegram "💥 <b>[${HOSTNAME} (${IP_ADDR})] СЕРВИС УПАЛ: ${SERVICE_NAME}</b>%0A%0AСтатус: <b>INACTIVE / FAILED</b>%0A<i>Попытка автоматического перезапуска...</i>"
            # Попытка автоматического перезапуска
            systemctl restart "${SERVICE_NAME}" || true
        fi
    else
        if clear_alert "service"; then
            send_telegram "✅ <b>[${HOSTNAME}] Сервис ${SERVICE_NAME} снова активен и работает.</b>"
        fi
    fi
fi
