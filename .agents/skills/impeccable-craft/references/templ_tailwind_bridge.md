# Мост интеграции: Templ, Tailwind v4 и токены SiteCraft

Документ описывает технический мост между дизайн-системой **Impeccable** и архитектурным стеком **SiteCraft** (Go 1.27 + Templ + Tailwind CSS v4 + Alpine.js).

---

## 🏗️ 1. Архитектурный контраст: Impeccable Original vs SiteCraft

| Параметр | Impeccable (Original) | Impeccable-Craft (SiteCraft) |
|---|---|---|
| **Рантайм** | Node.js, `npx`, npm-пакеты | **Zero-Node.js** в продакшене, автономный Go-бинарник |
| **Шаблонизация** | React / JSX / Vue / TSX | **Go Templ (`*.templ`)** со строгой статической типизацией |
| **CSS-движок** | Tailwind v3 (`tailwind.config.js`) | **Tailwind v4** (директивы `@theme`, `@source` в CSS) |
| **Anti-Slop чекер** | JS/Rust CLI через Node.js | Легковесный скрипт `scripts/check_slop.py` (0 зависимостей) |
| **Контекст токенов** | `DESIGN.md` | `DESIGN.md` + синхронизация с `presets/*.json` и `tokens.css` |
| **Инспекция в IDE** | React DevTools | **DOM Source Mapping** (`data-component`, `data-source`, `data-field`) |

---

## 🎨 2. Семантические токены Tailwind CSS v4

В SiteCraft все цвета, радиусы и анимации управляются через `@theme` в `ui/assets/input.css` и CSS-переменные в `ui/assets/css/tokens.css`. 

### Семантическая таблица подстановок:

```css
/* ПРАВИЛО: Никогда не писать жесткие цвета! */

❌ Неправильно (AI Slop):
bg-[#0f172a] text-[#ffffff] border-[#334155]
from-indigo-500 to-purple-600
rounded-[18px]

✅ Правильно (SiteCraft Impeccable Tokens):
bg-surface text-content border-content/10
bg-primary text-primary-content
bg-surface-muted text-content-muted
rounded-(--radius-card)
rounded-(--radius-btn)
```

### Основные слои поверхности:
1. `bg-surface` — Базовый фон страницы или карточки верхнего уровня.
2. `bg-surface-muted` — Второстепенная поверхность (подложки, сайдбары, карточки в Bento).
3. `text-content` — Основной контрастный текст.
4. `text-content-muted` — Второстепенный текст, описания, подсказки.
5. `border-content/10` — Деликатная граница (10% непрозрачности контента), идеально сочетается как в светлой, так и в темной теме.
6. `bg-primary` / `bg-accent` — Главное действие, конверсионная кнопка или ключевой акцент.

---

## 🧩 3. Паттерн CVA (Class Variance Authority) в Templ

Вместо хаотичных наборов классов компоненты кнопок, бейджей и карточек строятся по каноническому паттерну CVA в Go:

```go
// Пример: ui/components/base/button.templ
package base

type ButtonVariant string
type ButtonSize string

const (
    BtnPrimary   ButtonVariant = "primary"
    BtnSecondary ButtonVariant = "secondary"
    BtnOutline   ButtonVariant = "outline"
    BtnGhost     ButtonVariant = "ghost"

    BtnSm ButtonSize = "sm"
    BtnMd ButtonSize = "md"
    BtnLg ButtonSize = "lg"
)

func ButtonClasses(variant ButtonVariant, size ButtonSize, extra string) string {
    base := "inline-flex items-center justify-center font-medium transition-all duration-200 ease-out focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 disabled:opacity-50 disabled:pointer-events-none rounded-(--radius-btn)"
    
    // Варианты стилизации
    var vClass string
    switch variant {
    case BtnPrimary:
        vClass = "bg-primary text-white hover:brightness-110 shadow-sm"
    case BtnSecondary:
        vClass = "bg-surface-muted text-content hover:bg-surface-muted/80"
    case BtnOutline:
        vClass = "border border-content/15 text-content hover:bg-content/5"
    case BtnGhost:
        vClass = "text-content hover:bg-content/5"
    }

    // Размеры
    var sClass string
    switch size {
    case BtnSm:
        sClass = "text-xs px-3 py-1.5 gap-1.5"
    case BtnMd:
        sClass = "text-sm px-4 py-2 gap-2"
    case BtnLg:
        sClass = "text-base px-6 py-3 gap-2.5"
    }

    return base + " " + vClass + " " + sClass + " " + extra
}
```

---

## 🎯 4. DOM Source Mapping

Обязательный контракт проекта SiteCraft: все компоненты и блоки обязаны содержать дата-атрибуты для прозрачной инспекции и редактирования:

```html
<section 
    class="py-16 md:py-24 bg-surface"
    data-component="hero-split"
    data-source="ui/sections/hero.templ"
>
    <div class="max-w-7xl mx-auto px-4 sm:px-6">
        <h1 
            class="text-4xl sm:text-5xl lg:text-6xl font-bold tracking-tight text-content"
            data-field="title"
        >
            Заголовок секции
        </h1>
    </div>
</section>
```

Это гарантирует, что агент и человек могут мгновенно сопоставить видимый в браузере DOM-элемент с точным файлом `.templ` в проекте.
