# Матрица веб-стандартов Baseline в SiteCraft

Спецификация **Baseline** (W3C WebDX) разделяет возможности веб-платформы на:
* **Baseline Widely Available**: поддерживается всеми основными браузерами (Chrome, Firefox, Safari, Edge) не менее 30 месяцев.
* **Baseline Newly Available**: поддерживается стабильными релизами всех основных браузеров (2024–2025 гг.).
* **Limited Availability**: находится в стадии стандартизации или имеет частичную кроссбраузерную поддержку.

Ниже приведена матрица сопоставления фич платформы с архитектурными решениями движка **SiteCraft** (Go + Templ + Tailwind v4 + Alpine + HTMX).

---

## 📊 Сводная таблица возможностей Baseline

| Возможность | Статус Baseline | Поддержка браузеров | Назначение в SiteCraft | Заменяемый Legacy-паттерн |
|---|---|---|---|---|
| `<dialog>` & `::backdrop` | 🟢 **Widely Available** (2022) | Chrome 37+, Safari 15.4+, Firefox 98+, Edge 79+ | Модальные окна, подтверждения действий, быстрые заявки. | `fixed inset-0 z-50` дивы, кастомные фокус-трапы на JS. |
| CSS Container Queries (`@container`) | 🟢 **Widely Available** (2023) | Chrome 105+, Safari 16.0+, Firefox 110+, Edge 105+ | Адаптивные Bento-карточки, сайдбар-виджеты в Tailwind v4 (`@container`, `@sm:`). | Медиазапросы ширины экрана, `ResizeObserver` в JS. |
| CSS `:has()` Parent Selector | 🟢 **Widely Available** (2023) | Chrome 105+, Safari 15.4+, Firefox 121+, Edge 105+ | Стилизация предков по состоянию потомков (активные чекбоксы, валидация). | Навешивание классов `has-active-child` через JavaScript. |
| CSS Subgrid | 🟢 **Widely Available** (2023) | Chrome 117+, Safari 16.0+, Firefox 71+, Edge 117+ | Идеальное табличное выравнивание заголовков и цен в карточках тарифов. | Хакерские flex-распорки, выравнивание через JS. |
| Popover API (`popover`) | 🟡 **Newly Available** (2024) | Chrome 114+, Safari 17.0+, Firefox 125+, Edge 114+ | Тултипы, контекстные меню, выпадающие фильтры каталога. | Popper.js, Tippy.js, `@click.outside` в Alpine.js. |
| CSS `@starting-style` | 🟡 **Newly Available** (2024) | Chrome 117+, Safari 17.5+, Firefox 129+, Edge 117+ | Анимация появления элементов из `display: none` и закрытия `<dialog>`. | `x-transition` в Alpine, костыли с `setTimeout` в JS. |
| `transition-behavior: allow-discrete` | 🟡 **Newly Available** (2024) | Chrome 117+, Safari 17.5+, Firefox 129+, Edge 117+ | Плавные переходы дискретных CSS-свойств (`display`, `visibility`). | Сложные JS state machines для анимации размонтирования. |
| View Transitions API (Same-doc) | 🟡 **Newly Available** (2024) | Chrome 111+, Safari 18.0+, Firefox 135+ | Бесшовные анимации смены экранов при HTMX-свопах (`transition:true`). | SPA-роутеры, Barba.js, тяжелые кастомные переходы. |
| CSS Anchor Positioning | 🟡 **Newly Available** (2024) | Chrome 125+, Edge 125+ (Safari / Firefox в разработке) | Привязка тултипов и выпадающих списков к кнопке без абсолютных координат. | Вычисление координат через `getBoundingClientRect()`. |
| CSS `:user-invalid` / `:user-valid` | 🟢 **Widely Available** (2023) | Chrome 119+, Safari 16.4+, Firefox 88+, Edge 119+ | Эргономичная валидация форм (подсветка только после клика/ввода). | JS `onblur` / `oninput` обработчики с флагами dirty/touched. |
| `fetchpriority="high"` | 🟢 **Widely Available** (2024) | Chrome 101+, Safari 17.2+, Firefox 132+, Edge 101+ | Мгновенная приоритизация LCP Hero-картинки первого экрана. | Прелоады всех картинок подряд в `<head>`. |
| `content-visibility: auto` | 🟢 **Widely Available** (2023) | Chrome 85+, Safari 18.0+, Firefox 125+, Edge 85+ | Изоляция рендера блоков за пределами экрана (ускорение длинных страниц). | Кастомный виртуальный скроллинг на JS для каталогов. |

---

## 🎯 Стратегия прогрессивного улучшения (Progressive Enhancement)

В SiteCraft все компоненты проектируются по принципу многоуровневой надежности:

```
[Уровень 3: Премиальный UX] — View Transitions, Popover API, CSS Anchor Positioning
             ▲
[Уровень 2: Современная база] — <dialog>, @container, :has(), :user-invalid, Subgrid
             ▲
[Уровень 1: Чистый No-JS]    — Семантический HTML5, чистый CSS Tailwind v4, HTMX ссылки
```

1. **Если браузер поддерживает Popover API**: меню открывается нативно, всплывает в Top Layer без z-index коллизий, Esc и Light Dismiss работают аппаратно.
2. **Если браузер устарел**: Alpine.js хелпер обеспечивает мягкий фолбэк с переключением видимости.
3. **Если JavaScript отключен**: базовые формы и ссылки отправляются стандартными HTTP POST/GET запросами, сохраняя 100% работоспособность сайта.
