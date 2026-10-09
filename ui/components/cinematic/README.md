# 🎬 Cinematic Site Components (Полная коллекция 31 модуль — Zero-GSAP Edition)

Абсолютно **все 31 кинематографических модуля** адаптированы под архитектуру **SiteCraft** (**Go + Templ + Tailwind CSS + Alpine.js + HTMX**).

Все 31 модуль работают **БЕЗ внешних скриптов, БЕЗ GSAP и БЕЗ npm** (0 КБ сторонних зависимостей).

---

## 📑 Полная матрица 31 модуля

### 🌀 1. Scroll-Driven модули (10 из 10)
| # | Модуль | Файл Templ | Реализация без GSAP |
|---|---|---|---|
| 01 | **Text Mask Reveal** | [`text_mask.templ`](text_mask.templ) | **Motion One (~3.8 KB)** • Непрерывный субпиксельный скраббинг |
| 02 | **Sticky Stack Narrative** | [`sticky_stack.templ`](sticky_stack.templ) | CSS `position: sticky` + нативный `IntersectionObserver` |
| 03 | **Layered Zoom Parallax** | [`zoom_parallax.templ`](zoom_parallax.templ) | **Motion One (~3.8 KB)** • Глубокий 3D-зум и разлёт плашек |
| 04 | **Horizontal Scroll Hijack** | [`horizontal_scroll.templ`](horizontal_scroll.templ) | **Motion One (~3.8 KB)** • Плавный перевод в горизонталь |
| 05 | **Continuous Scrub Timeline**| [`continuous_scrub.templ`](continuous_scrub.templ) | **Motion One (~3.8 KB)** • Непрерывный скраббинг карточки и шкалы |
| 06 | **Sticky Card Stack** | [`sticky_cards.templ`](sticky_cards.templ) | Чистый CSS `position: sticky` со смещением `top` |
| 07 | **Scroll SVG Draw** | [`svg_draw.templ`](svg_draw.templ) | CSS `stroke-dashoffset` + динамический скролл-прогресс |
| 08 | **Curtain Reveal** | [`curtain_reveal.templ`](curtain_reveal.templ) | Sticky 100dvh + встречный `translateX` створок |
| 09 | **Split Screen Scroll** | [`split_scroll.templ`](split_scroll.templ) | Противоположный `translateY` левой и правой колонок |
| 10 | **Scroll Color Shift** | [`color_shift.templ`](color_shift.templ) | `IntersectionObserver` + переключение атрибута `data-theme` |

---

### 🖱️ 2. Cursor & Hover модули (8 из 8)
| # | Модуль | Файл Templ | Реализация |
|---|---|---|---|
| 11 | **Cursor-Reactive** | [`cursor_reactive.templ`](cursor_reactive.templ) | 3D-наклон карточки (Tilt) + магнитная кнопка |
| 12 | **Accordion Slider** | [`accordion_slider.templ`](accordion_slider.templ) | Pure CSS flex transition (`hover:flex-[3.5]`) |
| 13 | **Cursor Image Reveal** | [`image_reveal.templ`](image_reveal.templ) | Интерактивный слайдер сравнения «До / После» |
| 14 | **Hover Image Trail** | [`image_trail.templ`](image_trail.templ) | Шлейф из всплывающих карточек за курсором |
| 15 | **3D Flip Cards** | [`flip_cards.templ`](flip_cards.templ) | Pure CSS 3D (`transform-style: preserve-3d`) |
| 16 | **Magnetic Repel Grid** | [`magnetic_grid.templ`](magnetic_grid.templ) | Плитки сетки динамически отталкиваются от курсора |
| 17 | **Spotlight Border Cards** | [`spotlight_border.templ`](spotlight_border.templ) | Радиальная подсветка границ карточек за курсором |
| 18 | **Drag-to-Pan Grid** | [`drag_pan.templ`](drag_pan.templ) | Бесконечный перетаскиваемый холст (Miro/Figma style) |

---

### 👆 3. Click & Tap модули (6 из 6)
| # | Модуль | Файл Templ | Реализация |
|---|---|---|---|
| 19 | **View Transition Morphing** | [`view_transitions.templ`](view_transitions.templ) | Морфинг карточки в модалку через нативный View Transition API |
| 20 | **Particle Explosion Button**| [`particle_button.templ`](particle_button.templ) | Салют искр/частиц при клике на Canvas |
| 21 | **Odometer Counter** | [`odometer.templ`](odometer.templ) | Накат цифр при скролле во вьюпорт |
| 22 | **3D Coverflow Carousel** | [`coverflow.templ`](coverflow.templ) | Apple Coverflow карусель с 3D-перспективой |
| 23 | **Dynamic Island Nav** | [`dynamic_island.templ`](dynamic_island.templ) | Морфинг плашки уведомлений на Alpine.js |
| 24 | **macOS Dock Nav** | [`dock_nav.templ`](dock_nav.templ) | Плавное увеличение иконок при приближении мыши |

---

### 🔮 4. Ambient & Auto модули (7 из 7)
| # | Модуль | Файл Templ | Реализация |
|---|---|---|---|
| 25 | **Text Scramble Decode** | [`text_scramble.templ`](text_scramble.templ) | Раскодирование символов («матрица») на RAF |
| 26 | **Kinetic Marquee** | [`kinetic_marquee.templ`](kinetic_marquee.templ) | Бесконечная бегущая строка на GPU (`@keyframes`) |
| 27 | **Mesh Gradient** | [`mesh_gradient.templ`](mesh_gradient.templ) | Живой фон с плавающими размытыми цветными орбами |
| 28 | **Circular Text Path** | [`circular_text.templ`](circular_text.templ) | Вращающийся по кругу текст вокруг иконки |
| 29 | **Glitch Effect** | [`glitch_text.templ`](glitch_text.templ) | Киберпанк RGB-сплит текста на чистом CSS |
| 30 | **Typewriter Effect** | [`typewriter.templ`](typewriter.templ) | Печатающийся текст с мигающей кареткой |
| 31 | **Gradient Stroke Text** | [`gradient_stroke.templ`](gradient_stroke.templ) | Контурный текст с бегущим градиентом на CSS |

---

## ♿ Архитектурные стандарты SiteCraft
1. **Zero External Runtime:** 0 КБ внешних скриптов, 0 зависимостей от сторонних CDN или npm.
2. **ГОСТ a11y & Reduced Motion:** При включении режима доступности или `@media (prefers-reduced-motion)` агрессивные анимации автоматически выключаются.
3. **HTMX Совместимость:** Обработчики событий поддерживают `htmx:afterSwap`.
4. **Интеграция с темами:** Цвета завязаны на CSS-переменные `--color-primary`, `--color-accent`, `bg-surface`, `text-content`.
