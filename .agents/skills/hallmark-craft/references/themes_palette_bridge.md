# Мост тем Hallmark в токены SiteCraft и presets/*.json

В референсах скилла **hallmark-craft** локально размещены **все 50 эталонных шаблонов/тем Hallmark**:
- Директория: [references/themes/](themes/)
- Реестр: [references/themes/index.json](themes/index.json)
- Для каждого шаблона доступен Markdown-файл со спецификацией (`<slug>.md`) и визуальный референс рендера (`<slug>.jpg`).

Hallmark оперирует концепцией «материала» (бумага, чернила, шрифт, радиусы, фактура). В SiteCraft эта концепция напрямую маппится в **Tailwind v4 Semantic Design Tokens** и отраслевые пресеты [presets/](../../../../presets/).

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

## 2. Полный каталог 50 шаблонов Hallmark по жанрам

### Editorial (Издательские, Журнальные, Ремесленные) — 20 шаблонов

_Идеально для медиа, лонгридов, юридических бюро, ремесленных брендов, культурных проектов._

| Шаблон | Спецификация | Превью | Display шрифт | Постура | Тон бумаги | Акцент | Ключевые ниши |
|---|---|---|---|---|---|---|---|
| **Almanac** | [almanac.md](themes/almanac.md) | [almanac.jpg](themes/almanac.jpg) | Hanken Grotesk | `restrained` | mid (L:84) | chromatic-magenta | almanac, yearbook, field guide |
| **Atelier** | [atelier.md](themes/atelier.md) | [atelier.jpg](themes/atelier.jpg) | Sentient | `restrained` | light (L:94.5) | cool | fashion, atelier, couture |
| **Broadside** | [broadside.md](themes/broadside.md) | [broadside.jpg](themes/broadside.jpg) | Six Caps | `committed` | light (L:92) | warm | auction, sale, market |
| **Brutal** | [brutal.md](themes/brutal.md) | [brutal.jpg](themes/brutal.jpg) | Libre Franklin | `committed` | light (L:97.5) | chromatic-yellow | brutalist, raw, bold |
| **Carnival** | [carnival.md](themes/carnival.md) | [carnival.jpg](themes/carnival.jpg) | Big Shoulders Display | `full-palette` | light (L:92) | chromatic-yellow | record label, podcast, comedy |
| **Editorial** | [editorial.md](themes/editorial.md) | [editorial.jpg](themes/editorial.jpg) | Bricolage Grotesque | `restrained` | light (L:97) | chromatic-green | magazine, editorial, feature |
| **Formwork** | [formwork.md](themes/formwork.md) | [formwork.jpg](themes/formwork.jpg) | Oswald | `restrained` | mid (L:74) | warm | construction, concrete, skatepark |
| **Garden** | [garden.md](themes/garden.md) | [garden.jpg](themes/garden.jpg) | Vollkorn | `committed` | light (L:95.8) | chromatic-green | farm, apiary, honey |
| **Grid** | [grid.md](themes/grid.md) | [grid.jpg](themes/grid.jpg) | Radio Canada Big | `full-palette` | light (L:97.5) | cool | identity, brand system, wayfinding |
| **Hedgerow** | [hedgerow.md](themes/hedgerow.md) | [hedgerow.jpg](themes/hedgerow.jpg) | Prata | `drenched` | dark (L:24) | chromatic-green | distillery, gin, brewery |
| **Ink** | [ink.md](themes/ink.md) | [ink.jpg](themes/ink.jpg) | Epilogue | `restrained` | light (L:99.5) | chromatic-teal | library, lending, community |
| **Loom** | [loom.md](themes/loom.md) | [loom.jpg](themes/loom.jpg) | Libre Caslon Display | `committed` | light (L:95) | cool | weaving, textile, linen |
| **Manifesto** | [manifesto.md](themes/manifesto.md) | [manifesto.jpg](themes/manifesto.jpg) | Anton | `committed` | dark (L:13) | warm | manifesto, statement, declaration |
| **Newsprint** | [newsprint.md](themes/newsprint.md) | [newsprint.jpg](themes/newsprint.jpg) | Zilla Slab | `restrained` | light (L:94) | warm | newspaper, broadsheet, journalism |
| **Riso** | [riso.md](themes/riso.md) | [riso.jpg](themes/riso.jpg) | Cabinet Grotesk | `committed` | light (L:97) | chromatic-pink | risograph, riso, print studio |
| **Solstice** | [solstice.md](themes/solstice.md) | [solstice.jpg](themes/solstice.jpg) | Yeseva One | `restrained` | light (L:94) | warm | farm, orchard, cider |
| **Specimen** | [specimen.md](themes/specimen.md) | [specimen.jpg](themes/specimen.jpg) | Bodoni Moda | `restrained` | light (L:96.5) | cool | typography, type specimen, foundry |
| **Sport** | [sport.md](themes/sport.md) | [sport.jpg](themes/sport.jpg) | Saira Condensed | `restrained` | light (L:98) | chromatic-volt | running, marathon, race |
| **Studio** | [studio.md](themes/studio.md) | [studio.jpg](themes/studio.jpg) | Hedvig Letters Serif | `restrained` | light (L:97.2) | warm | design studio, creative studio, branding |
| **Vellum** | [vellum.md](themes/vellum.md) | [vellum.jpg](themes/vellum.jpg) | Grenze Gotisch | `committed` | light (L:94) | cool | roastery, coffee, bakery |

---

### Modern-Minimal (Современный минимализм, Технологичные) — 12 шаблонов

_Идеально для B2B SaaS, финтеха, инжиниринга, архитектурных и девелоперских компаний._

| Шаблон | Спецификация | Превью | Display шрифт | Постура | Тон бумаги | Акцент | Ключевые ниши |
|---|---|---|---|---|---|---|---|
| **Blueprint** | [blueprint.md](themes/blueprint.md) | [blueprint.jpg](themes/blueprint.jpg) | Chakra Petch | `drenched` | mid (L:35) | neutral | architecture, architect, engineering |
| **Cobalt** | [cobalt.md](themes/cobalt.md) | [cobalt.jpg](themes/cobalt.jpg) | Familjen Grotesk | `restrained` | light (L:98.5) | cool | api, sdk, cli |
| **Console** | [console.md](themes/console.md) | [console.jpg](themes/console.jpg) | Host Grotesk | `restrained` | dark (L:17) | chromatic-amber | lab, laboratory, protocol |
| **Coral** | [coral.md](themes/coral.md) | [coral.jpg](themes/coral.jpg) | General Sans | `restrained` | light (L:97) | warm | saas, platform, billing |
| **Dose** | [dose.md](themes/dose.md) | [dose.jpg](themes/dose.jpg) | Schibsted Grotesk | `committed` | light (L:97.6) | chromatic-green | health, medication, pharmacy |
| **Dusk** | [dusk.md](themes/dusk.md) | [dusk.jpg](themes/dusk.jpg) | Wix Madefor Display | `committed` | dark (L:17.5) | cool | treasury, banking, fintech |
| **Furrow** | [furrow.md](themes/furrow.md) | [furrow.jpg](themes/furrow.jpg) | Red Hat Display | `committed` | light (L:96) | warm | farm, agriculture, land |
| **Rim** | [rim.md](themes/rim.md) | [rim.jpg](themes/rim.jpg) | Switzer | `committed` | dark (L:9) | warm | space, satellite, earth observation |
| **Terrazzo** | [terrazzo.md](themes/terrazzo.md) | [terrazzo.jpg](themes/terrazzo.jpg) | Syne | `restrained` | light (L:96) | warm | materials, library, architects |
| **Ultraviolet** | [ultraviolet.md](themes/ultraviolet.md) | [ultraviolet.jpg](themes/ultraviolet.jpg) | Outfit | `committed` | light (L:95) | cool | music, audio, stems |
| **Vantage** | [vantage.md](themes/vantage.md) | [vantage.jpg](themes/vantage.jpg) | Geologica | `restrained` | dark (L:13.5) | chromatic-lime | lidar, scan, survey |
| **Wayfind** | [wayfind.md](themes/wayfind.md) | [wayfind.jpg](themes/wayfind.jpg) | Chivo Mono | `committed` | light (L:95.5) | cool | transit, metro, rail |

---

### Atmospheric (Атмосферные, Темные, Светотеневые) — 9 шаблонов

_Идеально для AI/ML продуктов, студий звука и 3D, премиальных фотографов, ночных заведений._

| Шаблон | Спецификация | Превью | Display шрифт | Постура | Тон бумаги | Акцент | Ключевые ниши |
|---|---|---|---|---|---|---|---|
| **Bloom** | [bloom.md](themes/bloom.md) | [bloom.jpg](themes/bloom.jpg) | Ranade | `committed` | light (L:97) | warm | generative, ai creative, music |
| **Darkroom** | [darkroom.md](themes/darkroom.md) | [darkroom.jpg](themes/darkroom.jpg) | Sofia Sans Extra Condensed | `restrained` | dark (L:12.5) | chromatic-amber | photographer, photography, film |
| **Kiln** | [kiln.md](themes/kiln.md) | [kiln.jpg](themes/kiln.jpg) | Boska | `full-palette` | dark (L:13) | warm | kiln, ceramics, pottery |
| **Lumen** | [lumen.md](themes/lumen.md) | [lumen.jpg](themes/lumen.jpg) | Instrument Serif | `restrained` | dark (L:12.5) | warm | inference, model, llm |
| **Midnight** | [midnight.md](themes/midnight.md) | [midnight.jpg](themes/midnight.jpg) | Mona Sans | `committed` | dark (L:14) | cool | render, 3d, animation |
| **Refract** | [refract.md](themes/refract.md) | [refract.jpg](themes/refract.jpg) | Funnel Display | `restrained` | light (L:98.2) | cool | pool, lido, swim |
| **Ribbon** | [ribbon.md](themes/ribbon.md) | [ribbon.jpg](themes/ribbon.jpg) | Instrument Sans | `committed` | light (L:98.5) | chromatic-teal | sleep, sound, soundscape |
| **Spore** | [spore.md](themes/spore.md) | [spore.jpg](themes/spore.jpg) | Young Serif | `committed` | dark (L:15) | warm | grower, mushroom, fungi |
| **Terminal** | [terminal.md](themes/terminal.md) | [terminal.jpg](themes/terminal.jpg) | Martian Mono | `committed` | dark (L:12.5) | chromatic-green | cli, terminal, shell |

---

### Playful (Игровые, Яркие, Фестивальные) — 9 шаблонов

_Идеально для инди-игр, фестивалей, арт-пространств, креативных студий, молодежных D2C брендов._

| Шаблон | Спецификация | Превью | Display шрифт | Постура | Тон бумаги | Акцент | Ключевые ниши |
|---|---|---|---|---|---|---|---|
| **Arcade** | [arcade.md](themes/arcade.md) | [arcade.jpg](themes/arcade.jpg) | Silkscreen and Bungee | `full-palette` | dark (L:14) | chromatic-ember | game, indie, game studio |
| **Halftone** | [halftone.md](themes/halftone.md) | [halftone.jpg](themes/halftone.jpg) | Ultra | `committed` | light (L:96.2) | warm | drinks, soda, food brand |
| **Hum** | [hum.md](themes/hum.md) | [hum.jpg](themes/hum.jpg) | Chillax | `full-palette` | light (L:97) | chromatic-yellow | daily, habit, streak |
| **Overprint** | [overprint.md](themes/overprint.md) | [overprint.jpg](themes/overprint.jpg) | Anybody | `full-palette` | light (L:91.5) | cool | venue, gigs, live music |
| **Paste-up** | [pasteup.md](themes/pasteup.md) | [pasteup.jpg](themes/pasteup.jpg) | League Gothic | `committed` | mid (L:73) | chromatic-yellow | zine, fair, market |
| **Piste** | [piste.md](themes/piste.md) | [piste.jpg](themes/piste.jpg) | Barlow Condensed | `committed` | light (L:98.5) | cool | ski, mountain, resort |
| **Squeeze** | [squeeze.md](themes/squeeze.md) | [squeeze.jpg](themes/squeeze.jpg) | Gabarito | `drenched` | mid (L:41) | warm | fruit, food box, grocer |
| **Sticker** | [sticker.md](themes/sticker.md) | [sticker.jpg](themes/sticker.jpg) | Bowlby One SC | `full-palette` | light (L:95.7) | chromatic-magenta | stickers, print, merch |
| **Werkstatt** | [werkstatt.md](themes/werkstatt.md) | [werkstatt.jpg](themes/werkstatt.jpg) | Jost | `full-palette` | mid (L:84) | cool | school, programme, workshop |

---

## 3. Как использовать 50 шаблонов Hallmark в SiteCraft

1. **Подбор шаблона по брифу:**
   - Внимательно изучи `brief.md` или запрос заказчика.
   - Найди наиболее подходящий шаблон из 50 по ключевым словам и жанру (например: `almanac` для справочников/баз знаний, `blueprint` для архитектурных и строительных компаний, `vellum` для кофеен и крафта, `coral` для SaaS).
2. **Изучение спецификации и референса:**
   - Открой `references/themes/<slug>.md`, чтобы узнать точные цветовые координаты OKLCH, типографические правила (трекинг, интерлиньяж, регистр), правила обработки фотографий и допустимые элементы.
   - Посмотри изображение рендера `references/themes/<slug>.jpg` для сверки визуального ритма и композиции.
3. **Перенос токенов в SiteCraft:**
   - Переведи значения из блока `:root` в `presets/<preset_name>.json` и Tailwind v4 `@theme`:
     ```json
     "palette": {
       "surface": "oklch(94% 0.015 80)",
       "content": "oklch(20% 0.02 80)",
       "accent": "oklch(45% 0.18 350)"
     }
     ```
   - Задай правила скругления (`--radius-card`, `--radius-pill`) и разделителей (`border-content/10`).
4. **Режим генерации:**
   - В режиме `COMMERCIAL_CRO_MODE` используй типографику и палитру шаблона, сохраняя конверсионные триггеры (бейджи, шаги, кнопки быстрого заказа).
   - В режиме `STRICT_AESTHETIC_MODE` соблюдай строгие ограничения Floor bans Hallmark (никаких плашек, капслока, стрелочек и теней).
