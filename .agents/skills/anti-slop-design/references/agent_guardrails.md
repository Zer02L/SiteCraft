# Руководство по интеграции ограничений Anti-Slop в ИИ-агенты (Agent Guardrails)

Этот документ содержит **системные директивы, правила фильтрации и готовые промпт-блоки** для внедрения в системные промпты автономных ИИ-агентов (Claude, ChatGPT, v0, Cursor, Copilot, Antigravity, T3 Code).

Цель: **гарантировать исключение AI Slop на этапе генерации первого токена**.

---

## 🚫 1. Кодекс категорических запретов (Strict Negative Constraints)

Вставьте этот блок в системный промпт агента:

```markdown
### CRITICAL ANTI-AI-SLOP DIRECTIVE (MANDATORY EXECUTION RULES)

You are strictly prohibited from generating generic "AI Slop" web design patterns. You MUST adhere to these non-negotiable negative constraints:

1. NO PURPLE/CYAN SAAS GRADIENTS:
   - FORBIDDEN: `from-indigo-600 via-purple-600 to-pink-500`, `from-violet-500 to-cyan-400`, or purple ambient haze.
   - USE: Strict monochrome or neutral palette + ONE purposeful brand accent (e.g. ochre, cobalt, terracotta, emerald).

2. NO GRADIENT TEXT:
   - FORBIDDEN: `bg-clip-text text-transparent bg-gradient-to-...`.
   - USE: Solid, high-contrast, opaque typography (`text-neutral-900` or `text-white`).

3. NO AMBIENT BLUR ORBS / BLOBS:
   - FORBIDDEN: `blur-3xl`, `blur-2xl rounded-full absolute bg-purple-500/20`.
   - USE: Structural dividing borders (`border-t border-content/10`), crisp geometric grids, or authentic photography.

4. NO GLASSMORPHISM OVERKILL:
   - FORBIDDEN: Stacking `backdrop-blur-md bg-white/5 border border-white/10` across multiple cards.
   - USE: Solid opaque surface contrast (`bg-surface-muted`, crisp 1px borders).

5. NO "EYEBROW PILL" HERO CRUTCH:
   - FORBIDDEN: Inert pill badge directly above H1 (e.g., `✨ Introducing v2.0` or `FEATURE SPOTLIGHT` before `H1`).
   - RULE: If an eyebrow can be removed without losing sentence meaning, DELETE IT. Write a self-sufficient, punchy headline.

6. NO MONOTONOUS 3-CARD GRIDS:
   - FORBIDDEN: Row of 3 identical height boxes each with `[Icon in square] + [H3] + [2 lines of text] + [Learn more →]`.
   - USE: Asymmetric layouts, editorial rows, dense data tables, or 8/4 split Bento layouts.

7. NO RUSSIAN-DOLL CARD NESTING (BOX-IN-A-BOX):
   - FORBIDDEN: Card inside a card where both have borders, background fills, and rounded corners.
   - USE: Hairline dividers (`divide-y divide-content/10`) or simple whitespace padding.

8. NO SLUGGISH OR REFLOW ANIMATIONS:
   - FORBIDDEN: `transition: all`, `duration-500` / `duration-700` on interactive controls, animating `height`, `width`, `margin`, or `top/left`.
   - USE: Snappy GPU-only properties (`transform`, `opacity`) with `duration-150` to `duration-200 ease-out`. Always respect `prefers-reduced-motion`.

9. NO INVISIBLE CONTENT ON LOAD:
   - FORBIDDEN: `opacity-0 translate-y-8` concealing critical page content before JS executes.
   - RULE: Server-rendered or static content MUST be visible immediately without client JS.

10. NO MARKETING BUZZWORD SALAD:
    - FORBIDDEN WORDS: "Supercharge", "Unleash the power of", "Seamlessly integrate", "Next-generation platform", "In today's fast-paced world", "Elevate your business", "All-in-one solution".
    - USE: Concrete nouns and measurable verbs: What + Who + Measurable Outcome + Timeline + Cost/Warranty.

11. NO SYNTHETIC PROOF OR PLASTIC AVATARS:
    - FORBIDDEN: Invented metrics ("99.9% satisfaction", "10,000+ teams"), Midjourney plastic face avatars, dummy corporate logo walls.
    - RULE: Only use verified data provided in project brief. If no proof is supplied, omit the proof block entirely.

12. NO A11Y VIOLATIONS:
    - FORBIDDEN: `outline-none` or `focus:outline-none` without `focus-visible:ring-2`.
    - FORBIDDEN: Icon-only buttons lacking `aria-label` or `.sr-only`.
    - FORBIDDEN: Gray text with contrast lower than 4.5:1 against its background.
```

---

## 💎 2. Созидательный кодекс (Positive Craft Heuristics)

Вместо шаблонных генераций требуйте от агента соблюдения следующих 5 правил мастерства:

### Правило 1: Закон иерархии 1-к-3 (1-to-3 Visual Hierarchy)
На любом экране устройства (390px, 768px, 1440px) должен быть:
* **Один главный фокус (Hero Focal Point):** центральный оффер, главный скриншот или ключевая цифра.
* **Два вспомогательных якоря:** навигация и primary CTA.
* **Всё остальное — фоновый порядок:** функциональные отступы, строгие разделители. Никаких 5 мерцающих бейджей и 4 разноцветных кнопок.

### Правило 2: Редакционный ритм (Editorial Rhythm)
* Чередуйте плотность: разреженная секция манифеста сменяется плотной таблицей характеристик, затем идет полноэкранный фото-блок.
* Заменяйте однотипные карточки на:
  * `Rows` — список с тонкими линиями и выровненными метаданными;
  * `Split 60/40` — слева интерактивная деталь, справа развернутый аргумент;
  * `Stat-led` — крупное число с текстовой расшифровкой без иконки.

### Правило 3: Принцип 60-30-10 в цвете
* 60% — доминирующая поверхность (холст: `bg-surface`).
* 30% — структурный контраст (фон секций, разделители, текст: `bg-surface-muted`, `border-border`, `text-content`).
* 10% — единственный акцент бренда (`bg-accent` для primary действия и ключевой метрики).

### Правило 4: Пропорциональный отступ (Spatial Balance)
* Радиус скругления определяет минимальный паддинг:
  * `rounded-md` (4–6px) → `p-3` или `p-4`;
  * `rounded-xl` (12px) → `p-6`;
  * `rounded-2xl` (16–20px) → `p-8`;
  * `rounded-3xl` (24–32px) → `p-10 md:p-12`.
* Никогда не допускайте `rounded-3xl p-2`.

### Правило 5: Честный копирайтинг прямого действия
* Вместо «Инновационная платформа для оптимизации ваших задач»:
* Пишите: «Анализ смет и спецификаций в PDF за 30 секунд с выгрузкой в 1С».

---

## 🛠️ 3. Чек-лист самопроверки агента перед выдачей кода

Перед тем как вернуть сгенерированный код пользователю, агент обязан проверить:
- [ ] 1. Присутствует ли хоть один класс `from-indigo` / `to-purple` / `blur-3xl`? *(Если да — удалить).*
- [ ] 2. Есть ли градиент на тексте (`bg-clip-text`)? *(Если да — заменить на сплошной цвет).*
- [ ] 3. Есть ли «бровь»-бейдж над H1/H2? *(Если заголовок понятен без бейджа — удалить бейдж).*
- [ ] 4. Спрятан ли контент за `opacity-0` без серверного рендеринга? *(Если да — сделать видимым).*
- [ ] 5. Есть ли кнопки без фокуса клавиатуры (`focus-visible`)? *(Если да — добавить).*
- [ ] 6. Все ли интерактивные элементы реагируют за 150–200ms? *(Если стоят duration-500+ — сократить).*
- [ ] 7. Проходит ли мобильная верстка тест на 360px без горизонтальной прокрутки?
