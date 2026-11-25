#!/usr/bin/env python3
"""
Modern Web Guidance Static Linter for SiteCraft (Zero-Dependency)
Scans Go Templ (*.templ), HTML, CSS, and JS files for legacy web anti-patterns,
missing modern platform standards (Baseline), Core Web Vitals risks,
and opportunities to replace heavy JavaScript with native HTML5/CSS.

Usage:
    python3 modern_web_lint.py [path/to/scan] [--strict] [--json]
"""

import os
import sys
import re
import json
from pathlib import Path
from typing import List, Dict, Any

# Rule definitions
RULES = [
    # --- 1. Modals, Popovers & Top Layer (MW-MOD) ---
    {
        "id": "MW-MOD-01",
        "name": "legacy-modal-overlay",
        "severity": "P1",
        "category": "Modals",
        "desc": "Устаревший оверлей модального окна (fixed inset-0 z-50) вместо нативного <dialog>",
        "pattern": re.compile(r'class="[^"]*\bfixed\s+inset-0\b[^"]*\b(z-(40|50|999)|z-\[\d+\])\b[^"]*\b(bg-black/\d+|bg-neutral-\d+/\d+)\b[^"]*"'),
        "fix": "Используйте нативный тег <dialog> с методом .showModal(), стилизацией ::backdrop и CSS @starting-style. Браузер сам управляет Top Layer, фокусом и Esc."
    },
    {
        "id": "MW-MOD-02",
        "name": "manual-click-outside-popover",
        "severity": "P2",
        "category": "Popovers",
        "desc": "Ручной Alpine/JS перехват клика вне элемента (@click.outside) для всплывающего меню или тултипа",
        "pattern": re.compile(r'@click\.outside\s*=\s*"[^"]*(isOpen|show|open)\s*=\s*false"'),
        "fix": "Используйте нативный Popover API (атрибут popover=\"auto\" и popovertarget=\"...\"). Браузер автоматически обрабатывает light dismiss без JS."
    },

    # --- 2. Performance & Core Web Vitals (MW-PERF) ---
    {
        "id": "MW-PERF-01",
        "name": "hero-image-missing-fetchpriority",
        "severity": "P1",
        "category": "Performance",
        "desc": "Главное изображение первого экрана (hero / banner) без атрибута fetchpriority=\"high\"",
        "pattern": re.compile(r'<img\b(?![^>]*\bfetchpriority=["\']high["\'])[^>]*(class="[^"]*\b(hero|banner|lead-image)\b[^"]*"|alt="[^"]*(hero|главн)[^"]*")[^>]*>', re.IGNORECASE),
        "fix": "Добавьте fetchpriority=\"high\" и loading=\"eager\" к LCP-изображению первого экрана для ускорения Largest Contentful Paint."
    },
    {
        "id": "MW-PERF-02",
        "name": "legacy-js-lazyload",
        "severity": "P0",
        "category": "Performance",
        "desc": "Устаревшие JS-библиотеки ленивой загрузки (lazysizes, data-src) вместо нативного loading=\"lazy\"",
        "pattern": re.compile(r'\b(data-src\s*=|class="[^"]*\blazysizes\b[^"]*"|\blazyload\.js\b)'),
        "fix": "Удалите JS-библиотеку ленивой загрузки. Используйте нативные HTML-атрибуты loading=\"lazy\" и decoding=\"async\"."
    },
    {
        "id": "MW-PERF-03",
        "name": "blocking-render-webfont-import",
        "severity": "P1",
        "category": "Performance",
        "desc": "Блокирующий рендеринг @import внешних шрифтов Google Fonts в CSS",
        "pattern": re.compile(r'@import\s+url\(["\']https?://fonts\.googleapis\.com/[^"\']+["\']\);'),
        "fix": "Локализуйте шрифты в woff2 (Zero External CDN) или используйте <link rel=\"preload\" as=\"font\" type=\"font/woff2\" crossorigin> с font-display: swap."
    },

    # --- 3. Layout & Sizing Standards (MW-LYT) ---
    {
        "id": "MW-LYT-01",
        "name": "window-resize-listener",
        "severity": "P2",
        "category": "Layout",
        "desc": "Слушатель ресайза окна (@resize.window / window.addEventListener('resize')) вместо контейнерных запросов",
        "pattern": re.compile(r'(@resize\.window|window\.addEventListener\([\'"]resize[\'"])'),
        "fix": "Используйте CSS Container Queries (@container, @sm:, @md: в Tailwind v4). Компоненты должны адаптироваться к родителю, а не к экрану."
    },
    {
        "id": "MW-LYT-02",
        "name": "aspect-ratio-padding-hack",
        "severity": "P2",
        "category": "Layout",
        "desc": "Устаревший хак соотношения сторон через паддинг (pt-[56.25%], pb-[56.25%])",
        "pattern": re.compile(r'\b(p[tb]-\[56\.25%\]|p[tb]-\[75%\]|p[tb]-\[100%\])\b'),
        "fix": "Используйте современное нативное CSS-свойство aspect-ratio (в Tailwind: aspect-video, aspect-square или aspect-[16/9])."
    },

    # --- 4. Motion & Discrete Transitions (MW-ANI) ---
    {
        "id": "MW-ANI-01",
        "name": "heavy-anim-library",
        "severity": "P0",
        "category": "Motion",
        "desc": "Подключение тяжелых внешних библиотек анимации (GSAP, TweenMax, Anime.js, Velocity.js)",
        "pattern": re.compile(r'\b(gsap|TweenMax|ScrollTrigger|anime\.js|velocity\.js)\b'),
        "fix": "В SiteCraft действует правило Zero-GSAP. Используйте нативный CSS (transform, opacity), CSS Scroll-Driven Animations или Motion One (~3 KB)."
    },
    {
        "id": "MW-ANI-02",
        "name": "reflow-property-animation",
        "severity": "P1",
        "category": "Motion",
        "desc": "Анимация свойств reflow (width, height, top, left, margin, padding) вместо transform",
        "pattern": re.compile(r'class="[^"]*\b(transition-(all|\[[^\]]*(width|height|top|left|margin|padding)[^\]]*\])|style="[^"]*transition:[^"]*(width|height|top|left|margin|padding))\b[^"]*"'),
        "fix": "Анимируйте исключительно аппаратные GPU-свойства: transform (translate, scale) и opacity для поддержания стабильных 60 FPS."
    },

    # --- 5. Forms & Inputs (MW-FRM) ---
    {
        "id": "MW-FRM-01",
        "name": "tel-input-missing-autocomplete",
        "severity": "P2",
        "category": "Forms",
        "desc": "Поле ввода телефона (type=\"tel\") без атрибута autocomplete=\"tel\" или inputmode=\"tel\"",
        "pattern": re.compile(r'<input\b(?![^>]*\bautocomplete=["\']tel["\'])[^>]*\btype=["\']tel["\'][^>]*>'),
        "fix": "Добавьте autocomplete=\"tel\" и inputmode=\"tel\" для облегчения автозаполнения браузером и вызова удобной цифровой клавиатуры."
    },
    {
        "id": "MW-FRM-02",
        "name": "password-input-missing-autocomplete",
        "severity": "P2",
        "category": "Forms",
        "desc": "Поле ввода пароля (type=\"password\") без атрибута autocomplete (current-password или new-password)",
        "pattern": re.compile(r'<input\b(?![^>]*\bautocomplete=["\'](current-password|new-password)["\'])[^>]*\btype=["\']password["\'][^>]*>'),
        "fix": "Укажите autocomplete=\"current-password\" или autocomplete=\"new-password\" для поддержки нативных менеджеров паролей и Passkeys."
    },

    # --- 6. Accessibility & Focus States (MW-A11Y) ---
    {
        "id": "MW-A11Y-01",
        "name": "stripped-focus-outline-without-ring",
        "severity": "P0",
        "category": "A11y",
        "desc": "Срезание outline-none без видимого компенсирующего кольца фокуса focus-visible:ring-",
        "pattern": re.compile(r'class="[^"]*\b(focus:outline-none|outline-none)\b(?![^"]*\bfocus(-visible)?:ring)[^"]*"'),
        "fix": "Запрещено скрывать фокус клавиатуры (ГОСТ Р 52872-2019 / WCAG 2.4.7). Добавьте focus-visible:ring-2 focus-visible:ring-primary."
    },
    {
        "id": "MW-A11Y-02",
        "name": "icon-button-missing-accessible-label",
        "severity": "P0",
        "category": "A11y",
        "desc": "Кнопка только с иконкой без текста и без aria-label / aria-labelledby",
        "pattern": re.compile(r'<button\b(?![^>]*\b:?aria-label=["\'{])[^>]*>\s*(?:<svg\b|@icons\.)(?:(?!</?button\b).)*?(?:</svg>|\))\s*</button>', re.DOTALL),
        "fix": "Кнопки-иконки обязаны иметь понятный скринридерам атрибут aria-label=\"...\" (ГОСТ Р 52872-2019)."
    }
]

# Supported file extensions
EXTENSIONS = {".templ", ".html", ".css", ".js"}
EXCLUDE_DIRS = {".git", "node_modules", "vendor", "tmp", "dist", "bin", ".agents"}
EXCLUDE_FILES = {"styles.css", "admin.css", "styles.min.css", "motion.min.js", "three.module.min.js", "ogl.mjs", "GLTFLoader.min.js", "OrbitControls.min.js"}

def scan_file(file_path: Path) -> List[Dict[str, Any]]:
    """Scan a single file against all modern web rules."""
    violations = []
    if file_path.name in EXCLUDE_FILES:
        return violations

    try:
        content = file_path.read_text(encoding="utf-8", errors="ignore")
    except Exception as e:
        return violations

    lines = content.splitlines()

    for rule in RULES:
        pattern = rule["pattern"]
        if rule.get("pattern").flags & re.DOTALL:
            # Multi-line match
            for m in pattern.finditer(content):
                start_pos = m.start()
                line_no = content[:start_pos].count("\n") + 1
                matched_text = m.group(0).strip().replace("\n", " ")
                if len(matched_text) > 80:
                    matched_text = matched_text[:77] + "..."
                violations.append({
                    "file": str(file_path),
                    "line": line_no,
                    "rule_id": rule["id"],
                    "name": rule["name"],
                    "severity": rule["severity"],
                    "category": rule["category"],
                    "desc": rule["desc"],
                    "match": matched_text,
                    "fix": rule["fix"]
                })
        else:
            # Line-by-line match
            for line_no, line in enumerate(lines, 1):
                m = pattern.search(line)
                if m:
                    matched_text = m.group(0).strip()
                    if len(matched_text) > 80:
                        matched_text = matched_text[:77] + "..."
                    violations.append({
                        "file": str(file_path),
                        "line": line_no,
                        "rule_id": rule["id"],
                        "name": rule["name"],
                        "severity": rule["severity"],
                        "category": rule["category"],
                        "desc": rule["desc"],
                        "match": matched_text,
                        "fix": rule["fix"]
                    })

    return violations

def scan_path(target_path: Path) -> List[Dict[str, Any]]:
    """Scan a file or recursively scan a directory."""
    all_violations = []

    if target_path.is_file():
        if target_path.suffix in EXTENSIONS:
            all_violations.extend(scan_file(target_path))
    elif target_path.is_dir():
        for root, dirs, files in os.walk(target_path):
            dirs[:] = [d for d in dirs if d not in EXCLUDE_DIRS]
            for file in files:
                p = Path(root) / file
                if p.suffix in EXTENSIONS:
                    all_violations.extend(scan_file(p))

    return all_violations

def main():
    target = sys.argv[1] if len(sys.argv) > 1 and not sys.argv[1].startswith("--") else "."
    strict_mode = "--strict" in sys.argv
    json_mode = "--json" in sys.argv

    path = Path(target)
    if not path.exists():
        print(f"Ошибка: путь '{target}' не существует.", file=sys.stderr)
        sys.exit(1)

    violations = scan_path(path)

    if json_mode:
        print(json.dumps({
            "target": str(path),
            "total_violations": len(violations),
            "violations": violations
        }, indent=2, ensure_ascii=False))
        return

    # Terminal output
    print(f"\n🌐 Modern Web Guidance Linter (SiteCraft Edition)")
    print(f"   Цель проверки: {path.resolve()}\n")

    if not violations:
        print("✅ Поздравляем! 0 нарушений. Код полностью соответствует современным веб-стандартам и Baseline.\n")
        sys.exit(0)

    p0_count = sum(1 for v in violations if v["severity"] == "P0")
    p1_count = sum(1 for v in violations if v["severity"] == "P1")
    p2_count = sum(1 for v in violations if v["severity"] == "P2")

    for v in violations:
        sev = v["severity"]
        color_code = "\033[91m" if sev == "P0" else "\033[93m" if sev == "P1" else "\033[96m"
        reset_code = "\033[0m"

        print(f"{color_code}[{sev}] {v['rule_id']}: {v['name']}{reset_code} ({v['category']})")
        print(f"  Файл: {v['file']}:{v['line']}")
        print(f"  Суть: {v['desc']}")
        print(f"  Фрагмент: {v['match']}")
        print(f"  💡 Решение: {v['fix']}\n")

    print("-" * 65)
    print(f"Итого нарушений: {len(violations)} (P0: {p0_count}, P1: {p1_count}, P2: {p2_count})")

    if p0_count > 0:
        print("❌ Блокирующие дефекты (P0): устраните нарушения доступности и тяжелые библиотеки перед релизом.")
        sys.exit(1)
    elif strict_mode and (p1_count > 0 or p2_count > 0):
        print("❌ В режиме --strict любые замечания P1 и P2 приводят к ошибке.")
        sys.exit(1)

    sys.exit(0)

if __name__ == "__main__":
    main()
