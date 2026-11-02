# Шпаргалка 17 композиционных секций (Tailwind v4 & Templ)

Композиции — это строительные блоки макроструктур. Страница объединяет от 3 до 6 разных композиций без повторения одного типа дважды подряд.

---

### 1. Split (Сплит 50/50 или 60/40)
* **Классы Tailwind v4:** `grid grid-cols-1 lg:grid-cols-12 gap-8 items-center`
* **Сетка:** Текст `lg:col-span-5` или `lg:col-span-6`, медиа `lg:col-span-7` или `lg:col-span-6`.
* **Правило:** Чередовать сторону медиа (слева/справа) не чаще одного раза за страницу.

### 2. Showcase (Полноэкранный акцент)
* **Классы:** `w-full max-w-7xl mx-auto px-4`
* **Сетка:** Крупный медиа-контейнер `aspect-16/9 lg:aspect-21/9 rounded-card overflow-hidden` + под ним 2 блока в пропорции `lg:grid-cols-12` (`col-span-8` и `col-span-4`).

### 3. Grid (Сетка равных сущностей)
* **Классы:** `grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6`
* **Особенность:** Карточки равной высоты (`flex flex-col justify-between`), выровненные заголовки, минимум 240px высота.

### 4. Bento (Асимметричная мозаика)
* **Классы:** `grid grid-cols-1 md:grid-cols-12 gap-4 lg:gap-6 auto-rows-[280px]`
* **Сетка:** Ведущий блок `md:col-span-8 md:row-span-2`, второстепенные блоки `md:col-span-4 md:row-span-1`.

### 5. Rows (Списки с тонкими разделителями)
* **Классы:** `divide-y divide-content/10 border-y border-content/10`
* **Строка:** `py-6 grid grid-cols-1 lg:grid-cols-12 gap-4 items-baseline` (заголовок `lg:col-span-4 font-semibold`, описание/цена `lg:col-span-8 text-content-muted`).

### 6. Table (Сравнительная таблица)
* **Классы:** `w-full text-left border-collapse text-sm`
* **Ячейки:** `py-4 px-4 border-b border-content/10 tabular-nums`

### 7. Figures (Счётчики и метрики)
* **Классы:** `grid grid-cols-2 lg:grid-cols-4 gap-8 divide-y lg:divide-y-0 lg:divide-x divide-content/10`
* **Цифра:** `text-4xl lg:text-5xl font-bold tracking-tight text-accent`

### 8. Steps (Пошаговый процесс)
* **Классы:** `grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-6 relative`
* **Элемент:** Номер шага (цифра в круге или крупный индекс), заголовок `text-base font-semibold`, описание.

### 9. Statement (Крупный типографический акцент)
* **Классы:** `py-20 max-w-4xl mx-auto text-center px-4`
* **Текст:** `text-2xl md:text-4xl lg:text-5xl font-medium tracking-tight leading-snug`

### 10. Band (Полоса во всю ширину)
* **Классы:** `w-full bg-surface-muted border-y border-content/10 py-16 px-4`
* **Назначение:** Смена фонового материала страницы («бумага 2» или контрастный цвет) под ключевой блок.

### 11. Form (Форма заявки / Контакты)
* **Классы:** `grid grid-cols-1 lg:grid-cols-12 gap-12 max-w-6xl mx-auto`
* **Сетка:** Контекст и гарантии `lg:col-span-5`, поля ввода `lg:col-span-7`.

### 12. Rail (Горизонтальный скролл)
* **Классы:** `flex gap-6 overflow-x-auto snap-x snap-mandatory pb-6 no-scrollbar`
* **Карточки:** `snap-start shrink-0 w-[280px] sm:w-[340px]`

### 13. Pinned (Липкая колонка + скролл)
* **Классы:** `grid grid-cols-1 lg:grid-cols-12 gap-12`
* **Левая колонка:** `lg:col-span-5 lg:sticky lg:top-24 self-start`
* **Правая колонка:** `lg:col-span-7 space-y-16`

### 14. Switcher (Интерактивный переключатель табов)
* **Классы:** `x-data="{ activeTab: 'monthly' }"`
* **Контрол:** Сегментированный переключатель `inline-flex p-1 bg-surface-muted rounded-pill`

### 15. Index (Плотный структурированный реестр)
* **Классы:** `space-y-2` (строки по 48-56px с мета-данными в колонках).

### 16. Schedule (Расписание / Сетка времени)
* **Классы:** `grid grid-cols-1 md:grid-cols-7 gap-3` (календарная сетка или слоты записи).

### 17. Offer (Блок моментальной покупки / Заказ)
* **Классы:** `max-w-4xl mx-auto p-8 rounded-card border border-content/15 bg-surface shadow-card`
* **Содержимое:** Фото/опции + актуальная цена + CTA кнопка + условия гарантии/доставки.
