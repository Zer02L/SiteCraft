# Мост тем Hallmark в токены SiteCraft и presets/*.json

Hallmark оперирует концепцией «материала» (бумага, чернила, шрифт, фактура). В SiteCraft эта концепция напрямую маппится в **Tailwind v4 Semantic Design Tokens** и отраслевые пресеты [presets/](file:///home/zerg/Projects/web/SiteCreator/presets/).

---

## 1. Таблица маппинга токенов

| Роль Hallmark | Токен SiteCraft | Назначение в Tailwind v4 |
|---|---|---|
| `--color-paper` | `--color-surface` | Базовый фон страницы (`bg-surface`) |
| `--color-paper-2` | `--color-surface-muted` | Вторичный фон секций / Band / карточек (`bg-surface-muted`) |
| `--color-paper-3` | `--color-surface-subtle` | Акцентная подложка инпутов и бейджей |
| `--color-ink` | `--color-content` | Основной цвет заголовков и текста (`text-content`) |
| `--color-ink-2` | `--color-content-muted` | Второстепенный текст, подписи (`text-content-muted`) |
| `--color-rule` | `border-content/10` | Разделители и сетка |
| `--color-accent` | `--color-accent` | Кнопка целевого действия, активные ссылки (`bg-accent`, `text-accent`) |
| `--color-accent-ink` | `--color-accent-fg` | Текст поверх кнопки акцента |
| `--radius-card` | `--radius-card` | Радиус скругления карточек (из `ui/assets/css/tokens.css`) |
| `--radius-pill` | `--radius-badge` | Радиус капсул, бейджей и табов (`rounded-full`) |
| `--font-display` | `font-serif-title` или `@theme --font-*` | Шрифт акцентных заголовков |
| `--font-body` | `font-sans` (Inter) | Шрифт основного текста |

---

## 2. Популярные материалы Hallmark, адаптированные под пресеты SiteCraft

### 1. `vellum` (Антикварная бумага и ремесло)
* **Палитра:** Тёплый пергамент (#FBF7EE), чернильный уголь (#1C1A17), терракотовый акцент (#C85A32).
* **Шрифты:** Serif (Merriweather / Playfair) + Inter.
* **Маппинг в SiteCraft:** Пресет `presets/b2c_craft_coffee.json`, винотеки, ремесленные пекарни, юридические бутики.

### 2. `garden` (Органика и натуральность)
* **Палитра:** Овсяная бумага (#F4F4EE), тёмно-хвойные чернила (#16221A), шалфей и мох (#3D6B4F).
* **Шрифты:** Vollkorn / Merriweather + Hanken Grotesk.
* **Маппинг в SiteCraft:** Пресет `presets/b2c_beauty.json`, эко-товары, клиники превентивной медицины.

### 3. `studio` (Архитектурный минимализм и швейцарская сетка)
* **Палитра:** Холодный белый (#F8F9FA), глубокий графит (#111315), сигнальный оранжево-красный (#E63916).
* **Шрифты:** Grotesk (Bebas Neue / Inter) + моноширинные маргиналии.
* **Маппинг в SiteCraft:** Пресет `presets/b2b_repair.json`, архитектурные бюро, строительный B2B.

### 4. `coral` (Гуманный финтех и чистый SaaS)
* **Палитра:** Мягкий нейтральный серый (#F5F5F7), тёплый антрацит (#1E2022), живой коралловый сигнал (#EE5D43).
* **Шрифты:** General Sans / Inter.
* **Маппинг в SiteCraft:** Пресет `presets/b2b_cyber.json`, финансовые сервисы, B2B SaaS.

### 5. `pasteup` (Зин, фестиваль, плотная органика)
* **Палитра:** Газетная бумага (#EFEFE9), глубокий чёрный (#0D0D0D), лимонно-жёлтый или лаймовый акцент (#D4E600).
* **Шрифты:** Плотный акцидентный гротеск (Bebas Neue) + технический гротеск.
* **Маппинг в SiteCraft:** Мероприятия, творческие пространства, авторские лектории (`b2b_academy.json`).
