#!/usr/bin/env bash
set -e

# ==============================================================================
# 🚀 SiteCraft Fleet Manager (Fleet of Binaries / Multi-Tenant Orchestrator)
# ==============================================================================
# Управление изолированными экземплярами сайтов для малого и среднего бизнеса.
# Архитектурный паттерн: 1 клиент = 1 изолированный процесс/контейнер + pb_data.
# ==============================================================================

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FLEET_SCRIPTS_DIR="${BASE_DIR}/scripts/fleet"

# shellcheck source=scripts/fleet/common.sh
source "${FLEET_SCRIPTS_DIR}/common.sh"
# shellcheck source=scripts/fleet/cmd_instance.sh
source "${FLEET_SCRIPTS_DIR}/cmd_instance.sh"
# shellcheck source=scripts/fleet/cmd_snapshot.sh
source "${FLEET_SCRIPTS_DIR}/cmd_snapshot.sh"
# shellcheck source=scripts/fleet/cmd_podman.sh
source "${FLEET_SCRIPTS_DIR}/cmd_podman.sh"
# shellcheck source=scripts/fleet/cmd_update.sh
source "${FLEET_SCRIPTS_DIR}/cmd_update.sh"

usage() {
    print_header
    echo -e "${BOLD}Использование:${RESET}"
    echo "  $0 create <site-id> --preset=<niche> [--theme=<theme>] [--port=<port>]"
    echo "  $0 start <site-id>"
    echo "  $0 stop <site-id>"
    echo "  $0 restart <site-id>"
    echo "  $0 update <site-id>|all|--engine   # Безопасное обновление (--test-config + снимок БД + Hot-Reload)"
    echo "  $0 rollback <site-id>              # Откат к предыдущему стабильному бинарнику и снимку БД"
    echo "  $0 status <site-id>                # Проверка статуса процесса и /health эндпоинта"
    echo "  $0 list"
    echo "  $0 backup <site-id>"
    echo "  $0 export [<site-id>] [<out-dir>]"
    echo "  $0 podman-up"
    echo "  $0 podman-down"
    echo ""
    echo -e "${BOLD}Примеры безопасного деплоя и обновлений:${RESET}"
    echo "  $0 update auto-master              # Безопасный деплой для конкретного сайта"
    echo "  $0 update all                      # Обновление всего пула сайтов с проверкой"
    echo "  $0 update --engine                 # Компиляция и валидация только бинарника"
    echo "  $0 rollback auto-master            # Экстренный откат инстанса при сбое"
    echo "  $0 status auto-master              # Проверка здоровья /health"
    echo "  $0 list"
    exit 1
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
    update)
        shift
        cmd_update "$@"
        ;;
    rollback)
        cmd_rollback "$2"
        ;;
    status)
        cmd_status "$2"
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
