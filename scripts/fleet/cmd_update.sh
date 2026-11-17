#!/usr/bin/env bash
# ==============================================================================
# SiteCraft Fleet Manager — Safe deployment, validation and hot-reload commands
# ==============================================================================

FLEET_SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -z "$BASE_DIR" ] || [ -z "$FLEET_DIR" ]; then
    # shellcheck source=scripts/fleet/common.sh
    source "${FLEET_SCRIPTS_DIR}/common.sh"
fi

if ! declare -f take_sqlite_snapshot >/dev/null; then
    # shellcheck source=scripts/fleet/cmd_snapshot.sh
    source "${FLEET_SCRIPTS_DIR}/cmd_snapshot.sh"
fi

update_single_instance() {
    local SITE_ID="$1"
    local SITE_PATH="${FLEET_DIR}/${SITE_ID}"

    if [ ! -d "$SITE_PATH" ] || [ ! -f "$SITE_PATH/site.env" ]; then
        echo -e "${RED}❌ Сайт '${SITE_ID}' не найден в ${FLEET_DIR}.${RESET}"
        return 1
    fi

    # shellcheck source=/dev/null
    source "$SITE_PATH/site.env"
    local DB_FILE="${SITE_PATH}/data/sitecraft.db"

    # 1. Обязательный преддеплойный снимок базы
    local SNAPSHOT_FILE
    SNAPSHOT_FILE="$(take_sqlite_snapshot "$SITE_PATH")"

    # 2. Двухэтапная проверка: валидация схемы и конфигурации через --test-config
    echo -e "🔍 Двухэтапная проверка: валидация конфигурации (--test-config)..."
    local TEST_DB="${DB_FILE}"
    if [ ! -f "$TEST_DB" ]; then
        TEST_DB="${BASE_DIR}/pb_data/sitecraft.db"
    fi

    if ! "${BASE_DIR}/bin/sitecraft.new" --test-config --db="${TEST_DB}"; then
        echo -e "${RED}❌ ОШИБКА: Валидация --test-config для инстанса '${SITE_ID}' провалилась!${RESET}"
        echo -e "${YELLOW}🛡️ Защитный барьер: бинарник НЕ заменен, запущенный инстанс НЕ затронут.${RESET}"
        rm -f "${BASE_DIR}/bin/sitecraft.new"
        return 1
    fi
    echo -e "${GREEN}✓ Валидация --test-config успешно пройдена!${RESET}"

    # 3. Атомарная замена бинарника
    echo -e "🔄 Атомарная замена бинарника движка..."
    [ -f "${ENGINE_BIN}" ] && cp -p "${ENGINE_BIN}" "${ENGINE_BIN}.prev"
    mv -f "${BASE_DIR}/bin/sitecraft.new" "${ENGINE_BIN}"
    chmod +x "${ENGINE_BIN}"
    echo -e "${GREEN}✓ Бинарник ${ENGINE_BIN} атомарно обновлен.${RESET}"

    # 4. Fleet Hot-Reload / Graceful Restart
    local PID_FILE="${SITE_PATH}/app.pid"
    local WAS_RUNNING=0
    local OLD_PID=""
    if [ -f "$PID_FILE" ]; then
        OLD_PID="$(cat "$PID_FILE")"
        if kill -0 "$OLD_PID" 2>/dev/null; then
            WAS_RUNNING=1
        fi
    fi

    if command -v systemctl &> /dev/null && systemctl is-active --quiet "sitecraft@${SITE_ID}" 2>/dev/null; then
        echo -e "⚡ Перезагрузка сервиса systemd: systemctl reload sitecraft@${SITE_ID}..."
        systemctl reload "sitecraft@${SITE_ID}" || systemctl restart "sitecraft@${SITE_ID}"
        echo -e "${GREEN}✓ Сервис systemd sitecraft@${SITE_ID} успешно перезагружен без простоя.${RESET}"
    elif [ "$WAS_RUNNING" -eq 1 ]; then
        echo -e "⚡ Выполняю безопасный Hot-Reload процесса (PID: ${OLD_PID})..."

        # Мягкая отправка SIGHUP для инвалидации кэша
        kill -HUP "$OLD_PID" 2>/dev/null || true

        # Остановка старого процесса
        kill "$OLD_PID" 2>/dev/null || true
        local WAITED=0
        while kill -0 "$OLD_PID" 2>/dev/null && [ "$WAITED" -lt 6 ]; do
            sleep 0.5
            WAITED=$((WAITED + 1))
        done

        # Запуск нового бинарника
        setsid "${SITE_PATH}/run.sh" </dev/null > "${SITE_PATH}/logs/stdout.log" 2>&1 &
        local NEW_PID=$!
        disown "$NEW_PID" 2>/dev/null || true
        echo "$NEW_PID" > "$PID_FILE"

        sleep 1.5

        # Проверка жизнеспособности (Liveness Health-Check)
        local HEALTH_OK=0
        if kill -0 "$NEW_PID" 2>/dev/null; then
            if command -v curl &> /dev/null; then
                if curl -s -f -m 3 "http://127.0.0.1:${PORT}/health" &> /dev/null || curl -s -f -m 3 "http://127.0.0.1:${PORT}/" &> /dev/null; then
                    HEALTH_OK=1
                fi
            else
                HEALTH_OK=1
            fi
        fi

        if [ "$HEALTH_OK" -eq 1 ]; then
            echo -e "${GREEN}${BOLD}✓ Инстанс '${SITE_ID}' успешно обновлен и активен! PID: ${NEW_PID}${RESET}"
            echo -e "  👉 Публичный URL: ${CYAN}http://127.0.0.1:${PORT}${RESET}"
            echo -e "  👉 Health-Check:  ${CYAN}http://127.0.0.1:${PORT}/health${RESET}"
        else
            echo -e "${RED}🚨 КРИТИЧЕСКАЯ ОШИБКА: Новый процесс не отвечает на порту ${PORT}!${RESET}"
            echo -e "${YELLOW}⏪ Выполняю автоматический откат (Automatic Rollback)...${RESET}"

            kill "$NEW_PID" 2>/dev/null || true

            # Восстановление базы из снимка
            if [ -n "$SNAPSHOT_FILE" ] && [ -f "$SNAPSHOT_FILE" ]; then
                cp -f "$SNAPSHOT_FILE" "$DB_FILE"
                echo -e "  ✓ База данных восстановлена из снимка ${SNAPSHOT_FILE}"
            fi

            # Восстановление предыдущего бинарника
            if [ -f "${ENGINE_BIN}.prev" ]; then
                cp -f "${ENGINE_BIN}.prev" "${ENGINE_BIN}"
                chmod +x "${ENGINE_BIN}"
                echo -e "  ✓ Бинарник откачен к стабильной версии"
            fi

            # Перезапуск предыдущей версии
            setsid "${SITE_PATH}/run.sh" </dev/null > "${SITE_PATH}/logs/stdout.log" 2>&1 &
            local PREV_PID=$!
            disown "$PREV_PID" 2>/dev/null || true
            echo "$PREV_PID" > "$PID_FILE"
            echo -e "${YELLOW}⚠️ Инстанс восстановлен на стабильной версии (PID: $(cat "$PID_FILE")).${RESET}"
            return 1
        fi
    else
        echo -e "${YELLOW}ℹ️ Инстанс '${SITE_ID}' не был запущен. Бинарник обновлен и проверен.${RESET}"
        echo -e "  Для запуска: ${CYAN}./fleet.sh start ${SITE_ID}${RESET}"
    fi
    return 0
}

cmd_update() {
    local TARGET="$1"
    shift || true

    if [ -z "$TARGET" ]; then
        echo -e "${RED}❌ Укажите цель обновления:${RESET}"
        echo "   $0 update <site-id>    # Обновить конкретный инстанс"
        echo "   $0 update all          # Обновить все инстансы Fleet"
        echo "   $0 update --engine     # Скомпилировать и проверить только бинарник"
        exit 1
    fi

    print_header
    echo -e "${CYAN}${BOLD}🚀 Безопасный деплой и Fleet Hot-Reload${RESET}\n"

    # Шаг 1: Компиляция временного бинарника
    local NEW_BIN="${BASE_DIR}/bin/sitecraft.new"
    build_temp_binary "$NEW_BIN"

    if [ "$TARGET" = "--engine" ] || [ "$TARGET" = "engine" ]; then
        echo -e "🔍 Проверка бинарника через --test-config..."
        if ! "$NEW_BIN" --test-config; then
            echo -e "${RED}❌ Валидация --test-config провалилась!${RESET}"
            rm -f "$NEW_BIN"
            exit 1
        fi
        [ -f "${ENGINE_BIN}" ] && cp -p "${ENGINE_BIN}" "${ENGINE_BIN}.prev"
        mv -f "$NEW_BIN" "${ENGINE_BIN}"
        chmod +x "${ENGINE_BIN}"
        echo -e "${GREEN}${BOLD}✓ Бинарник движка успешно обновлен и верифицирован: ${ENGINE_BIN}${RESET}"
        return
    fi

    if [ "$TARGET" = "all" ] || [ "$TARGET" = "--all" ]; then
        if [ ! -d "$FLEET_DIR" ] || [ -z "$(ls -A "$FLEET_DIR" 2>/dev/null)" ]; then
            echo -e "${YELLOW}ℹ️ Нет созданных клиентских сайтов в ${FLEET_DIR}. Обновляю только движок...${RESET}"
            [ -f "${ENGINE_BIN}" ] && cp -p "${ENGINE_BIN}" "${ENGINE_BIN}.prev"
            mv -f "$NEW_BIN" "${ENGINE_BIN}"
            chmod +x "${ENGINE_BIN}"
            echo -e "${GREEN}✓ Бинарник успешно обновлен.${RESET}"
            return
        fi

        # Проверка --test-config перед обновлением сайтов
        if ! "$NEW_BIN" --test-config; then
            echo -e "${RED}❌ Предварительная валидация --test-config провалилась!${RESET}"
            rm -f "$NEW_BIN"
            exit 1
        fi

        [ -f "${ENGINE_BIN}" ] && cp -p "${ENGINE_BIN}" "${ENGINE_BIN}.prev"
        mv -f "$NEW_BIN" "${ENGINE_BIN}"
        chmod +x "${ENGINE_BIN}"

        for site in "$FLEET_DIR"/*; do
            if [ -d "$site" ] && [ -f "$site/site.env" ]; then
                local S_ID
                S_ID="$(basename "$site")"
                echo -e "\n----------------------------------------------------------"
                echo -e "📦 Обновление инстанса: ${BOLD}${S_ID}${RESET}"
                echo -e "----------------------------------------------------------"
                take_sqlite_snapshot "$site" >/dev/null
                local PID_FILE="$site/app.pid"
                if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
                    # shellcheck source=/dev/null
                    source "$site/site.env"
                    local P
                    P="$(cat "$PID_FILE")"
                    kill -HUP "$P" 2>/dev/null || true
                    kill "$P" 2>/dev/null || true
                    sleep 1
                    setsid "${site}/run.sh" </dev/null > "${site}/logs/stdout.log" 2>&1 &
                    local RE_PID=$!
                    disown "$RE_PID" 2>/dev/null || true
                    echo "$RE_PID" > "$PID_FILE"
                    echo -e "${GREEN}✓ Инстанс '${S_ID}' перезапущен (PID: $(cat "$PID_FILE")).${RESET}"
                else
                    echo -e "${YELLOW}ℹ️ Инстанс '${S_ID}' остановлен. Готов к запуску.${RESET}"
                fi
            fi
        done
        echo -e "\n${GREEN}${BOLD}✓ Все инстансы Fleet успешно обновлены!${RESET}"
        return
    fi

    # Обновление конкретного инстанса
    update_single_instance "$TARGET"
}
