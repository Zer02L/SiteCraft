#!/usr/bin/env bash
# ==============================================================================
# SiteCraft Fleet Manager — Podman / Docker container orchestration
# ==============================================================================

FLEET_SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -z "$BASE_DIR" ] || [ -z "$FLEET_DIR" ]; then
    # shellcheck source=scripts/fleet/common.sh
    source "${FLEET_SCRIPTS_DIR}/common.sh"
fi

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
