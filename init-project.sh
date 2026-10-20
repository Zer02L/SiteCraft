#!/usr/bin/env bash
set -e

# SiteCraft Starter Kit Interactive Initializer
# Запуск: ./init-project.sh
# Также поддерживаются аргументы: ./init-project.sh [--name="..."] [--preset="..."] [--theme="..."]

# Цвета для красивого вывода в терминале
BOLD="\033[1m"
GREEN="\033[0;32m"
CYAN="\033[0;36m"
YELLOW="\033[1;33m"
PURPLE="\033[0;35m"
RESET="\033[0m"

echo -e "${CYAN}${BOLD}"
echo "=========================================================="
echo "          ⚡ SiteCraft Interactive Initializer            "
echo "=========================================================="
echo -e "${RESET}"

# Значения по умолчанию
NAME=""
PRESET=""
THEME=""
PORT="8090"
ADMIN_EMAIL="admin@local.dev"
ADMIN_PASS="admin123456"
AUTORUN=""

# Парсинг опциональных CLI аргументов (если переданы)
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --name=*) NAME="${1#*=}" ;;
        --preset=*) PRESET="${1#*=}" ;;
        --theme=*) THEME="${1#*=}" ;;
        --port=*) PORT="${1#*=}" ;;
        --admin-email=*) ADMIN_EMAIL="${1#*=}" ;;
        --admin-pass=*) ADMIN_PASS="${1#*=}" ;;
        --autorun) AUTORUN="yes" ;;
        -h|--help)
            echo "Использование: $0 [опции]"
            echo "  --name=<имя>         Имя исполняемого файла (по умолчанию: sitecraft-app)"
            echo "  --preset=<пресет>    Нишевый пресет: legal, medical, auto, beauty"
            echo "  --theme=<тема>       Дизайн-тема: legal, medical, auto, beauty, tech, dark"
            echo "  --port=<порт>        Порт сервера (по умолчанию: 8090)"
            echo "  --admin-email=<email> Email суперпользователя"
            echo "  --admin-pass=<пароль> Пароль суперпользователя"
            echo "  --autorun            Автоматически запустить сервер после сборки"
            exit 0
            ;;
        *) echo "Неизвестный параметр: $1"; exit 1 ;;
    esac
    shift
done

# ----------------------------------------------------
# 1. Интерактивный выбор ниши / пресета
# ----------------------------------------------------
if [ -z "$PRESET" ]; then
    echo -e "${BOLD}Выберите сферу деятельности сайта (пресет структуры и контента):${RESET}"
    echo -e "  ${GREEN}1)${RESET} ⚖️  Юриспруденция / Нотариус / Адвокат    (${CYAN}legal${RESET})"
    echo -e "  ${GREEN}2)${RESET} 🏥 Медицинский центр / Клиника / УЗИ   (${CYAN}medical${RESET})"
    echo -e "  ${GREEN}3)${RESET} 🚗 Автосервис / СТО / Диагностика      (${CYAN}auto${RESET})"
    echo -e "  ${GREEN}4)${RESET} 🌸 Салон красоты / SPA / Стилисты      (${CYAN}beauty${RESET})"
    echo -e "  ${GREEN}5)${RESET} 🧪 Тестовый полигон всех 20 модулей    (${CYAN}test${RESET})"
    echo ""
    read -p "Введите номер [1-5] (по умолчанию 1): " PRESET_CHOICE

    case "$PRESET_CHOICE" in
        2) PRESET="medical" ;;
        3) PRESET="auto" ;;
        4) PRESET="beauty" ;;
        5) PRESET="test" ;;
        *) PRESET="legal" ;;
    esac
fi

# ----------------------------------------------------
# 2. Тема оформления (автоматически соответствует пресету, если не передана явно через аргументы)
# ----------------------------------------------------
if [ -z "$THEME" ]; then
    THEME="$PRESET"
fi

# ----------------------------------------------------
# 3. Имя бинарника / проекта
# ----------------------------------------------------
if [ -z "$NAME" ]; then
    DEFAULT_NAME="${PRESET}-site"
    echo ""
    read -p "Имя выходного бинарного файла [Enter для '${DEFAULT_NAME}']: " INPUT_NAME
    NAME="${INPUT_NAME:-$DEFAULT_NAME}"
fi

# ----------------------------------------------------
# 4. Учётная запись администратора
# ----------------------------------------------------
echo ""
echo -e "${BOLD}Настройка суперпользователя PocketBase:${RESET}"
read -p "Email админа [Enter для '${ADMIN_EMAIL}']: " INPUT_EMAIL
ADMIN_EMAIL="${INPUT_EMAIL:-$ADMIN_EMAIL}"

read -p "Пароль админа [Enter для '${ADMIN_PASS}']: " INPUT_PASS
ADMIN_PASS="${INPUT_PASS:-$ADMIN_PASS}"

# ----------------------------------------------------
# 5. Сводка конфигурации
# ----------------------------------------------------
echo ""
echo -e "${YELLOW}${BOLD}=== Параметры сборки ===${RESET}"
echo -e "  📦 Пресет ниши:    ${GREEN}$PRESET${RESET}"
echo -e "  🎨 Дизайн-тема:    ${GREEN}$THEME${RESET}"
echo -e "  📁 Имя бинарника:  ${GREEN}bin/$NAME${RESET}"
echo -e "  👤 Email админа:   ${CYAN}$ADMIN_EMAIL${RESET}"
echo -e "  🔑 Пароль админа:  ${CYAN}$ADMIN_PASS${RESET}"
echo -e "========================="
echo ""

# Проверка окружения сборки: Devenv или локальные утилиты
RUN_PREFIX=""
if ! command -v templ &> /dev/null || ! command -v tailwindcss &> /dev/null; then
    if command -v devenv &> /dev/null; then
        echo -e "${PURPLE}ℹ️ Утилиты сборки будут запущены через devenv shell...${RESET}"
        RUN_PREFIX="devenv shell --"
    fi
fi

# ----------------------------------------------------
# 6. Процесс компиляции
# ----------------------------------------------------
echo -e "🔨 ${BOLD}[1/4] Компиляция Templ шаблонов...${RESET}"
if [ -n "$RUN_PREFIX" ]; then
    $RUN_PREFIX templ generate
else
    export PATH="$(go env GOPATH 2>/dev/null)/bin:$PATH"
    templ generate
fi

echo -e "🎨 ${BOLD}[2/4] Минификация Tailwind CSS...${RESET}"
if [ -n "$RUN_PREFIX" ]; then
    $RUN_PREFIX tailwindcss -i ui/assets/input.css -o ui/assets/styles.css --minify
else
    tailwindcss -i ui/assets/input.css -o ui/assets/styles.css --minify
fi

echo -e "⚙️  ${BOLD}[3/4] Компиляция единого бинарника Go (Single Binary)...${RESET}"
mkdir -p bin
if [ -n "$RUN_PREFIX" ]; then
    $RUN_PREFIX go build -ldflags="-s -w" -o "bin/$NAME" .
else
    go build -ldflags="-s -w" -o "bin/$NAME" .
fi

echo -e "👤 ${BOLD}[4/4] Инициализация базы данных и пресета...${RESET}"
"./bin/$NAME" superuser upsert "$ADMIN_EMAIL" "$ADMIN_PASS" >/dev/null 2>&1 || true
"./bin/$NAME" apply-preset "$PRESET" --theme="$THEME" >/dev/null 2>&1 || true

echo ""
echo -e "${GREEN}${BOLD}✅ Проект успешно собран и настроен для [${PRESET}] (тема: ${THEME})!${RESET}"
echo -e "Исполняемый файл: ${CYAN}bin/$NAME${RESET}"
echo ""

# ----------------------------------------------------
# 7. Запуск сервера (по запросу)
# ----------------------------------------------------
if [ -z "$AUTORUN" ]; then
    echo -e "Желаете запустить сервер прямо сейчас? (${GREEN}http://127.0.0.1:${PORT}${RESET})"
    read -p "[Y/n]: " LAUNCH_CHOICE
    case "$LAUNCH_CHOICE" in
        [nN][oO]|[nN])
            AUTORUN="no"
            ;;
        *)
            AUTORUN="yes"
            ;;
    esac
fi

if [ "$AUTORUN" = "yes" ]; then
    echo ""
    echo -e "🚀 ${BOLD}Запуск сервера SiteCraft...${RESET}"
    echo -e "  🌐 Публичный сайт:   ${CYAN}http://127.0.0.1:${PORT}${RESET}"
    echo -e "  ⚙️  Панель управления: ${CYAN}http://127.0.0.1:${PORT}/_/${RESET}"
    echo -e "     Логин:  ${BOLD}$ADMIN_EMAIL${RESET}"
    echo -e "     Пароль: ${BOLD}$ADMIN_PASS${RESET}"
    echo -e "  (Для остановки нажмите Ctrl+C)"
    echo ""
    exec "./bin/$NAME" serve --http="127.0.0.1:${PORT}" --preset="$PRESET" --theme="$THEME"
else
    echo "Для ручного запуска выполните:"
    echo -e "  ${BOLD}./bin/$NAME serve --http=127.0.0.1:${PORT} --preset=$PRESET --theme=$THEME${RESET}"
    echo ""
    echo "Панель управления PocketBase: http://127.0.0.1:${PORT}/_/"
fi
