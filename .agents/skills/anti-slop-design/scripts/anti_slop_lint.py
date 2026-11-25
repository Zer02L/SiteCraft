#!/usr/bin/env python3
"""
Anti-Slop Web Design Static Linter (Zero-Dependency)
Scans HTML, Go Templ, JSX, TSX, Vue, Svelte, Astro, and CSS files
for AI Slop web design patterns, accessibility traps, and generic LLM clichés.

Usage:
    python3 anti_slop_lint.py [path/to/scan] [--strict] [--json]
"""

import os
import sys
import re
import json
from pathlib import Path
from typing import List, Dict, Any

# Rule definitions
RULES = [
    # --- 1. Color & Lighting Slop (AS-CLR) ---
    {
        "id": "AS-CLR-01",
        "name": "indigo-purple-gradient",
        "severity": "P1",
        "category": "Color",
        "desc": "Шаблонный неоновый фиолетово-синий SaaS-градиент",
        "pattern": re.compile(r'\b(from-(indigo|purple|violet)-\d{3}|to-(purple|indigo|violet)-\d{3}|via-(purple|indigo|violet)-\d{3})\b'),
        "fix": "Замените на строгую монохромную палитру с ОДНИМ функциональным акцентом бренда из токенов темы."
    },
    {
        "id": "AS-CLR-02",
        "name": "gradient-clipped-text",
        "severity": "P1",
        "category": "Color",
        "desc": "Градиентная заливка текста (bg-clip-text text-transparent)",
        "pattern": re.compile(r'class="[^"]*\bbg-clip-text\b[^"]*\btext-transparent\b[^"]*"|class="[^"]*\btext-transparent\b[^"]*\bbg-clip-text\b[^"]*"'),
        "fix": "Используйте сплошной контрастный цвет шрифта (text-content, text-neutral-900 / text-white). Не прячьте слабый копирайтинг за градиент."
    },
    {
        "id": "AS-CLR-03",
        "name": "ambient-blur-orbs",
        "severity": "P2",
        "category": "Color",
        "desc": "Размытые фоновые сферы и пятна (blur-2xl / blur-3xl)",
        "pattern": re.compile(r'\b(blur-2xl|blur-3xl)\b'),
        "fix": "Уберите случайные размытые пятна. Замените на структурные линии сетки (border-t border-content/10) или честный чистый фон."
    },
    {
        "id": "AS-CLR-04",
        "name": "glassmorphism-stack",
        "severity": "P2",
        "category": "Color",
        "desc": "Чрезмерный псевдо-стекломорфизм (backdrop-blur с полупрозрачным фоном и рамкой)",
        "pattern": re.compile(r'class="[^"]*\bbackdrop-blur-(sm|md|lg|xl)\b[^"]*\b(bg-white/\d+|bg-black/\d+|bg-neutral-\d+/\d+)\b[^"]*"'),
        "fix": "Используйте непрозрачные или умеренно контрастные поверхности (bg-surface-muted, border-content/10) без мыльного размытия."
    },
    {
        "id": "AS-CLR-05",
        "name": "hardcoded-hex-color",
        "severity": "P1",
        "category": "Tokens",
        "desc": "Захардкоженный HEX-код в утилитах Tailwind вместо семантических дизайн-токенов",
        "pattern": re.compile(r'\b(bg|text|border|fill|stroke)-\[#[0-9a-fA-F]{3,8}\]'),
        "fix": "Используйте семантические токены темы: bg-surface, text-content, border-border, text-accent."
    },

    # --- 2. Structural & Layout Slop (AS-LYT) ---
    {
        "id": "AS-LYT-04",
        "name": "cramped-rounded-card",
        "severity": "P2",
        "category": "Layout",
        "desc": "Судорожный радиус скругления (rounded-2xl/3xl) при зажатом внутреннем паддинге (p-1/p-2/p-3)",
        "pattern": re.compile(r'class="[^"]*\b(rounded-(2xl|3xl)|rounded-\[3[2-9]px\])[^"]*\b(p-1|p-2|p-2\.5|p-3)\b[^"]*"'),
        "fix": "Соблюдайте пропорцию: для rounded-2xl требуется минимум p-6, для rounded-3xl — p-8..p-10."
    },
    {
        "id": "AS-LYT-06",
        "name": "rigid-fixed-width-overflow",
        "severity": "P0",
        "category": "Layout",
        "desc": "Жесткая ширина w-[>360px] без max-w, вызывающая горизонтальный скролл на экранах 360-390px",
        "pattern": re.compile(r'(?<![-:\w])w-\[(3[7-9]\d|[4-9]\d{2}|\d{4,})px\](?!.*max-w)'),
        "fix": "Используйте w-full max-w-[...] или адаптивные брейкпоинты (w-full md:w-[...])."
    },

    # --- 3. Copywriting & Marketing Slop (AS-CPY) ---
    {
        "id": "AS-CPY-01",
        "name": "ai-marketing-cliches",
        "severity": "P1",
        "category": "Copy",
        "desc": "Шаблонный маркетинговый мусор и штампы генеративных LLM",
        "pattern": re.compile(
            r'(инновационные\s+решения|лидеры\s+рынка|революционный\s+подход|раскройте\s+потенциал|'
            r'в\s+современном\s+(быстро\s+меняющемся\s+)?мире|индивидуальный\s+подход\s+к\s+каждому|'
            r'высококачественн\w+\s+услуг\w|экосистема\s+будущего|supercharge\s+your|'
            r'unleash\s+the\s+power\s+of|seamlessly\s+integrat\w+|elevate\s+your|'
            r'next-generation\s+platform|all-in-one\s+solution|in\s+today\'s\s+fast-paced\s+world)',
            re.IGNORECASE
        ),
        "fix": "Замените пустые штампы на формулу прямого действия: Что делаем + Срок + Гарантия + Цена/Результат."
    },
    {
        "id": "AS-CPY-02",
        "name": "template-arrow-punctuation",
        "severity": "P2",
        "category": "Copy",
        "desc": "Приклеенная шаблонная стрелка (→) на кнопках и ссылках",
        "pattern": re.compile(r'(узнать\s+больше|подробнее|заказать|купить|начать|learn\s+more|get\s+started|read\s+more)\s*(→|->|&rarr;)', re.IGNORECASE),
        "fix": "Используйте чистый глагол действия без стрелки: «Рассчитать смету», «Каталог проектов», «Оформить заказ»."
    },

    # --- 4. Motion & Performance Slop (AS-MOT) ---
    {
        "id": "AS-MOT-01",
        "name": "invisible-load-trap",
        "severity": "P1",
        "category": "Motion",
        "desc": "Скрытие контента за opacity-0 translate-y-8 без серверного рендеринга (страница пустая без JS)",
        "pattern": re.compile(r'class="[^"]*\b(opacity-0)\b[^"]*\b(translate-y-[4-9]|translate-y-1\d)\b(?!.*no-js)[^"]*"'),
        "fix": "Контент должен быть виден сразу при загрузке. Анимации должны обогащать появление, а не скрывать верстку от пользователя и поисковиков."
    },
    {
        "id": "AS-MOT-02",
        "name": "sluggish-transition-duration",
        "severity": "P2",
        "category": "Motion",
        "desc": "Слишком медленная анимация (>350ms) или transition-all, создающая ощущение задержки отклика",
        "pattern": re.compile(r'\b(duration-(500|700|1000)|transition-all)\b'),
        "fix": "Используйте тактильный отклик duration-150..duration-200 ease-out на аппаратных свойствах transform и opacity."
    },

    # --- 5. Accessibility Violations (AS-A11Y) ---
    {
        "id": "AS-A11Y-01",
        "name": "stripped-keyboard-focus",
        "severity": "P0",
        "category": "A11y",
        "desc": "Удаление outline без кольца focus-visible:ring (блокирует навигацию с клавиатуры по ГОСТ/WCAG)",
        "pattern": re.compile(r'class="[^"]*\b(focus:outline-none|outline-none)\b(?!.*focus-visible:ring)[^"]*"'),
        "fix": "Добавьте явное кольцо клавиатурного фокуса: focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2."
    },
    {
        "id": "AS-A11Y-02",
        "name": "unlabeled-icon-button",
        "severity": "P1",
        "category": "A11y",
        "desc": "Кнопка содержит только иконку SVG без доступной текстовой метки",
        "pattern": re.compile(r'<button(?![^>]*(aria-label|aria-labelledby))[^>]*>\s*<svg(?![^>]*(title|aria-label))[^>]*>(?:(?!<\/button>).)*<\/button>', re.DOTALL),
        "fix": "Добавьте атрибут aria-label=\"Описание действия\" или скрытый спан <span class=\"sr-only\">...</span>."
    }
]

SUPPORTED_EXTENSIONS = {'.html', '.htm', '.templ', '.jsx', '.tsx', '.vue', '.svelte', '.astro', '.css'}

def scan_file(file_path: Path) -> List[Dict[str, Any]]:
    violations = []
    try:
        content = file_path.read_text(encoding="utf-8")
    except Exception as e:
        return [{
            "id": "SYS-01",
            "name": "read-error",
            "severity": "P0",
            "category": "System",
            "desc": f"Не удалось прочитать файл: {e}",
            "file": str(file_path),
            "line": 0,
            "fix": "Проверьте права доступа и кодировку файла."
        }]

    lines = content.splitlines()

    for rule in RULES:
        pattern = rule["pattern"]
        if rule.get("multiline"):
            for match in pattern.finditer(content):
                start_pos = match.start()
                line_no = content.count('\n', 0, start_pos) + 1
                violations.append({
                    "id": rule["id"],
                    "name": rule["name"],
                    "severity": rule["severity"],
                    "category": rule["category"],
                    "desc": rule["desc"],
                    "file": str(file_path),
                    "line": line_no,
                    "snippet": match.group(0)[:80].replace('\n', ' '),
                    "fix": rule["fix"]
                })
        else:
            for idx, line in enumerate(lines, 1):
                match = pattern.search(line)
                if match:
                    violations.append({
                        "id": rule["id"],
                        "name": rule["name"],
                        "severity": rule["severity"],
                        "category": rule["category"],
                        "desc": rule["desc"],
                        "file": str(file_path),
                        "line": idx,
                        "snippet": match.group(0)[:80],
                        "fix": rule["fix"]
                    })

    return violations

def main():
    target_path = Path(sys.argv[1]) if len(sys.argv) > 1 and not sys.argv[1].startswith("--") else Path(".")
    strict_mode = "--strict" in sys.argv
    json_output = "--json" in sys.argv

    files_to_scan = []
    if target_path.is_file():
        files_to_scan.append(target_path)
    elif target_path.is_dir():
        for root, dirs, files in os.walk(target_path):
            # Ignore hidden and build directories
            dirs[:] = [d for d in dirs if not d.startswith(".") and d not in ("node_modules", "vendor", "dist", "build", "public")]
            for f in files:
                p = Path(root) / f
                if p.suffix in SUPPORTED_EXTENSIONS:
                    files_to_scan.append(p)
    else:
        print(f"Путь не найден: {target_path}", file=sys.stderr)
        sys.exit(1)

    all_violations = []
    for f in files_to_scan:
        all_violations.extend(scan_file(f))

    if json_output:
        print(json.dumps({
            "scanned_files": len(files_to_scan),
            "total_violations": len(all_violations),
            "strict_mode": strict_mode,
            "violations": all_violations
        }, indent=2, ensure_ascii=False))
    else:
        print(f"\n🔍 Просканировано файлов: {len(files_to_scan)}")
        if not all_violations:
            print("✨ Безупречно! AI Slop паттерны не обнаружены (Чистый Craft дизайн).\n")
            sys.exit(0)

        p0_count = sum(1 for v in all_violations if v["severity"] == "P0")
        p1_count = sum(1 for v in all_violations if v["severity"] == "P1")
        p2_count = sum(1 for v in all_violations if v["severity"] == "P2")

        print(f"⚠️ Обнаружено дефектов: {len(all_violations)} (P0 Блокеры: {p0_count}, P1 Ошибки: {p1_count}, P2 Замечания: {p2_count})\n")
        print(f"{'УРОВЕНЬ':<6} | {'ID':<10} | {'ФАЙЛ:СТРОКА':<35} | {'ОПИСАНИЕ'}")
        print("-" * 90)

        for v in all_violations:
            loc = f"{Path(v['file']).name}:{v['line']}"
            print(f"{v['severity']:<6} | {v['id']:<10} | {loc:<35} | {v['desc']}")
            print(f"       -> Фрагмент: {v['snippet']}")
            print(f"       -> Решение:  {v['fix']}\n")

    # Determine exit code
    has_blockers = any(v["severity"] in ("P0", "P1") for v in all_violations)
    if strict_mode and all_violations:
        sys.exit(1)
    elif has_blockers:
        sys.exit(1)
    else:
        sys.exit(0)

if __name__ == "__main__":
    main()
