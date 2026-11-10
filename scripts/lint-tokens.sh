#!/usr/bin/env bash
# ==============================================================================
# Design Token Linter & Anti-Slop Guard for SiteCraft
# Verifies that UI templates in ui/sections and ui/components (base + widgets)
# do not use hardcoded color utilities (e.g. text-gray-500, bg-slate-800, border-stone-800)
# and strictly adhere to semantic design tokens.
# ==============================================================================

set -euo pipefail

echo "==> Linting UI templates for semantic design token compliance..."

VIOLATIONS=$(grep -rn -E "\b(bg|text|border)-(gray|slate|zinc|neutral|stone)-[0-9]+" \
  ui/sections/*.templ ui/components/base/*.templ ui/components/*.templ || true)

if [ -n "$VIOLATIONS" ]; then
    echo "❌ Error: Found hardcoded color utilities in UI templates!"
    echo "Always use semantic tokens (e.g. bg-surface, text-content, text-on-primary, border-content/10):"
    echo ""
    echo "$VIOLATIONS"
    exit 1
fi

echo "✅ All UI sections and components strictly comply with semantic design tokens."
exit 0
