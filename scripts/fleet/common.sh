#!/usr/bin/env bash
# ==============================================================================
# SiteCraft Fleet Manager — Common definitions and helper functions
# ==============================================================================

BOLD="\033[1m"
GREEN="\033[0;32m"
CYAN="\033[0;36m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
RESET="\033[0m"

FLEET_SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$(cd "${FLEET_SCRIPTS_DIR}/../.." && pwd)"
FLEET_DIR="${BASE_DIR}/fleet_sites"
ENGINE_BIN="${BASE_DIR}/bin/sitecraft"

print_header() {
    echo -e "${CYAN}${BOLD}"
    echo "=========================================================="
    echo "       ⚡ SiteCraft Engine — Fleet of Binaries Manager     "
    echo "=========================================================="
    echo -e "${RESET}"
}

build_temp_binary() {
    local TARGET_BIN="${1:-${BASE_DIR}/bin/sitecraft.new}"
    mkdir -p "$(dirname "$TARGET_BIN")"
    echo -e "${CYAN}🔨 Компиляция бинарника в ${TARGET_BIN}...${RESET}"

    if command -v templ &> /dev/null && command -v tailwindcss &> /dev/null; then
        templ generate
        tailwindcss -i ui/assets/input.css -o ui/assets/styles.css --minify
        tailwindcss -i ui/assets/admin_input.css -o ui/assets/admin.css --minify
        go build -ldflags="-s -w" -o "$TARGET_BIN" .
    elif [ -f "${BASE_DIR}/devenv.nix" ] && command -v devenv &> /dev/null; then
        devenv shell -- bash -c "templ generate && tailwindcss -i ui/assets/input.css -o ui/assets/styles.css --minify && tailwindcss -i ui/assets/admin_input.css -o ui/assets/admin.css --minify && go build -ldflags='-s -w' -o '$TARGET_BIN' ."
    elif command -v go &> /dev/null; then
        go build -ldflags="-s -w" -o "$TARGET_BIN" .
    else
        echo -e "${RED}❌ Не найден компилятор Go для локальной сборки.${RESET}"
        exit 1
    fi
    chmod +x "$TARGET_BIN"
    echo -e "${GREEN}✓ Бинарник успешно скомпилирован: ${TARGET_BIN}${RESET}"
}

ensure_engine() {
    if [ ! -f "$ENGINE_BIN" ]; then
        echo -e "${YELLOW}⚠️ Бинарник движка $ENGINE_BIN не найден. Выполняю сборку...${RESET}"
        build_temp_binary "$ENGINE_BIN"
        echo -e "${GREEN}✓ Движок успешно собран в $ENGINE_BIN${RESET}"
    fi
}
