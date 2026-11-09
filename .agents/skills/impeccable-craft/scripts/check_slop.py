#!/usr/bin/env python3
"""
SiteCraft Anti-Slop & Quality Detector (adapted from Impeccable by Paul Bakaus)
Deterministic static checker for Go Templ (*.templ) and Tailwind CSS v4.
Runs with zero LLM tokens and 0 dependencies.

Usage:
    python3 check_slop.py [path/to/dir/or/file] [--strict]
"""

import os
import sys
import re
from pathlib import Path

# Rule definitions
RULES = [
    {
        "id": "AI-SLOP-01",
        "name": "gradient-slop",
        "severity": "P1",
        "desc": "Дефолтные неоновые фиолетово-синие градиенты (purple/indigo/blue)",
        "pattern": re.compile(r'\b(from-(indigo|purple|violet)-\d{3}|to-(purple|indigo|violet)-\d{3}|via-(purple|indigo)-\d{3})\b'),
        "fix": "Используйте семантические токены из presets/*.json (bg-surface, bg-primary, text-accent) вместо шаблонных AI-градиентов."
    },
    {
        "id": "AI-SLOP-02",
        "name": "background-orbs",
        "severity": "P2",
        "desc": "Случайные размытые фоновые сферы (blur-2xl / blur-3xl)",
        "pattern": re.compile(r'\b(blur-2xl|blur-3xl)\b'),
        "fix": "Уберите случайные размытые пятна или замените на структурные сеточные линии / чистый фон."
    },
    {
        "id": "TOKEN-01",
        "name": "hardcoded-hex",
        "severity": "P1",
        "desc": "Захардкоженные HEX-цвета в утилитах Tailwind вместо семантических токенов",
        "pattern": re.compile(r'\b(bg|text|border|fill|stroke)-\[#[0-9a-fA-F]{3,8}\]'),
        "fix": "Замените жесткий HEX на семантические токены темы: bg-surface, text-content, border-content/10, text-primary, bg-accent."
    },
    {
        "id": "RHYTHM-01",
        "name": "cramped-padding",
        "severity": "P2",
        "desc": "Слишком тесный паддинг (p-1, p-2) для карточки или контейнера",
        "pattern": re.compile(r'class="[^"]*\b(rounded-(xl|2xl|3xl)|border|shadow)[^"]*\b(p-1|p-2|p-2\.5)\b[^"]*"'),
        "fix": "Увеличьте внутренний отступ карточки: минимум p-4 sm:p-6 md:p-8."
    },
    {
        "id": "A11Y-01",
        "name": "focus-outline-stripped",
        "severity": "P0",
        "desc": "Удаление outline без замены на focus-visible:ring (нарушение ГОСТ/WCAG)",
        "pattern": re.compile(r'class="[^"]*\bfocus:outline-none\b(?!.*focus-visible:ring)[^"]*"'),
        "fix": "Добавьте focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 для клавиатурной навигации."
    },
    {
        "id": "COPY-01",
        "name": "marketing-slop",
        "severity": "P1",
        "desc": "Абстрактный маркетинговый мусор (бан по AGENTS.md)",
        "pattern": re.compile(r'(инновационные\s+решения|лидеры\s+рынка|революционный\s+подход|высококачественн\w+\s+услуг\w|индивидуальный\s+подход\s+к\s+каждому|world-class\s+solutions|next-generation\s+platform)', re.IGNORECASE),
        "fix": "Замените пустые штампы на формулу: конкретная услуга + измеримая выгода/срок/стандарт (смета, ГОСТ, дни)."
    },
    {
        "id": "RESPONSIVE-01",
        "name": "fixed-width-overflow",
        "severity": "P0",
        "desc": "Фиксированная ширина w-[>360px] без max-w-full (риск скролла на мобилке 390px)",
        "pattern": re.compile(r'\bw-\[(3[7-9]\d|[4-9]\d{2}|\d{4,})px\](?!.*max-w-full)'),
        "fix": "Используйте w-full max-w-[...] или адаптивные префиксы (w-full md:w-[...]) во избежание горизонтального скролла."
    },
    {
        "id": "MOTION-01",
        "name": "slow-transition",
        "severity": "P2",
        "desc": "Слишком медленная анимация (>350ms), замедляющая взаимодействие",
        "pattern": re.compile(r'\bduration-(500|700|1000)\b'),
        "fix": "Ограничьте тактильные микроинтерфейсы диапазоном duration-150 .. duration-250 для отклика в 60 FPS."
    }
]

def scan_file(file_path: Path):
    violations = []
    try:
        content = file_path.read_text(encoding="utf-8")
    except Exception as e:
        return [{"severity": "P0", "rule": "READ-FAIL", "desc": f"Не удалось прочитать файл: {e}", "line": 0, "fix": ""}]

    lines = content.splitlines()

    # Rule: Check DOM Source Mapping for section templates
    if "ui/sections/" in str(file_path):
        if "data-component=" not in content or "data-source=" not in content:
            violations.append({
                "severity": "P1",
                "rule": "MAPPING-01",
                "name": "missing-dom-mapping",
                "desc": "Секция не содержит атрибутов data-component или data-source",
                "line": 1,
                "fix": 'Добавьте в корневой тег: data-component="section-name" data-source="' + str(file_path.relative_to(file_path.parents[2] if len(file_path.parents) > 2 else file_path.parent)) + '"'
            })

    # Line-by-line checks
    for idx, line in enumerate(lines, start=1):
        # Ignore comments
        if "//" in line and "class=" not in line:
            stripped = line.strip()
            if stripped.startswith("//") or stripped.startswith("<!--"):
                continue

        # Check in-code disable tag
        if "impeccable-disable" in line or "data-impeccable-ignore" in line:
            continue

        for rule in RULES:
            match = rule["pattern"].search(line)
            if match:
                violations.append({
                    "severity": rule["severity"],
                    "rule": rule["id"],
                    "name": rule["name"],
                    "desc": f"{rule['desc']} (найдено: '{match.group(0)}')",
                    "line": idx,
                    "fix": rule["fix"]
                })

    return violations

def main():
    target = Path(sys.argv[1]) if len(sys.argv) > 1 else Path("ui")
    if not target.exists():
        print(f"Путь {target} не найден.")
        sys.exit(1)

    files_to_scan = []
    if target.is_file():
        files_to_scan.append(target)
    else:
        files_to_scan.extend(target.glob("**/*.templ"))

    if not files_to_scan:
        print(f"Шаблоны *.templ в {target} не найдены.")
        sys.exit(0)

    total_violations = 0
    p0_count = 0
    p1_count = 0
    p2_count = 0

    print(f"\n🔍 [Impeccable-Craft] Сканирование {len(files_to_scan)} шаблонов на Anti-Slop & Quality...\n")

    for f in sorted(files_to_scan):
        results = scan_file(f)
        if results:
            print(f"📄 \033[1m{f}\033[0m")
            for r in results:
                total_violations += 1
                sev = r["severity"]
                if sev == "P0":
                    p0_count += 1
                    color = "\033[31;1m" # Red
                elif sev == "P1":
                    p1_count += 1
                    color = "\033[33;1m" # Yellow
                else:
                    p2_count += 1
                    color = "\033[36m"   # Cyan

                print(f"  {color}[{sev}] {r['rule']} ({r['name']})\033[0m L{r['line']}: {r['desc']}")
                if r["fix"]:
                    print(f"      \033[90m↳ Решение: {r['fix']}\033[0m")
            print()

    print("—" * 60)
    print(f"Итог: {total_violations} замечаний (P0: {p0_count}, P1: {p1_count}, P2: {p2_count})")

    if p0_count > 0:
        print("❌ Обнаружены критические P0 дефекты (a11y/overflow). Требуется исправление.")
        sys.exit(2)
    elif p1_count > 0:
        print("⚠️ Обнаружены P1 замечания (AI-slop / жесткие цвета). Рекомендуется применить /polish или /distill.")
        sys.exit(0)
    else:
        print("✨ Дизайн чист! AI-slop и грубых дефектов верстки не обнаружено.")
        sys.exit(0)

if __name__ == "__main__":
    main()
