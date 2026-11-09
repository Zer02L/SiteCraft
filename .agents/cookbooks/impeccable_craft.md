# Cookbook: Impeccable-Craft (Frontend Craft & Anti-Slop in SiteCraft)

Каноническое руководство по применению скилла **Impeccable-Craft** для создания и ревизии интерфейсов на стеке **Go 1.27 + Templ + Tailwind CSS v4 + Alpine.js + PocketBase**.

---

## 1. Быстрый чеклист перед сдачей верстки

Перед тем как завершить работу над компонентом или страницей, выполните 5 обязательных шагов:

1. **Запустить статический чекер качества:**
   ```bash
   python3 .agents/skills/impeccable-craft/scripts/check_slop.py ui/sections/your_section.templ
   ```
2. **Проверить цветовую дисциплину:**
   * Отсутствие HEX-кодов `bg-[#...]` или `text-[#...]`.
   * Использование семантических токенов `bg-surface`, `text-content`, `border-content/10`, `accent`.
3. **Проверить мобильную адаптивность (390px):**
   * Все контейнеры имеют `w-full max-w-[...]` (никаких жестких `w-[500px]`).
   * Карточки на мобильных устройствах не имеют отрицательных отступов, вызывающих горизонтальный скролл.
4. **Проверить доступность (ГОСТ / WCAG):**
   * Все интерактивные кнопки имеют кольцо фокуса `focus-visible:ring-2 focus-visible:ring-primary`.
   * Контрастность текста не ниже 4.5:1.
5. **Компиляция:**
   ```bash
   templ generate
   make check
   ```

---

## 2. Канонический паттерн Templ: Секция с идеальным крафтом

Пример реализации секции с соблюдением всех правил Impeccable-Craft (семантические токены, сбалансированная типографика, отсутствие AI-slop, DOM Source Mapping):

```templ
package sections

import (
	"sitecraft/ui/components/base"
)

type ValueFeature struct {
	Title       string
	Description string
	Metric      string
	Tag         string
}

templ ValueSection(features []ValueFeature) {
	<section
		class="py-16 sm:py-20 lg:py-28 bg-surface border-t border-content/10 relative overflow-hidden"
		data-component="value-section"
		data-source="ui/sections/value.templ"
	>
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<!-- Заголовок с выверенной иерархией и ограниченной шириной -->
			<div class="max-w-2xl mb-12 sm:mb-16">
				<span class="inline-block text-xs font-semibold uppercase tracking-wider text-primary mb-3">
					Прозрачные условия
				</span>
				<h2 
					class="text-3xl sm:text-4xl lg:text-5xl font-bold tracking-tight text-content leading-tight"
					data-field="title"
				>
					Фиксированная смета без скрытых доплат
				</h2>
				<p 
					class="mt-4 text-base sm:text-lg text-content-muted leading-relaxed"
					data-field="subtitle"
				>
					Каждый этап работ фиксируется в договоре с гарантийными обязательствами по ГОСТ.
				</p>
			</div>

			<!-- Асимметричная сетка без эффекта «матрешки» -->
			<div class="grid grid-cols-1 md:grid-cols-3 gap-6 lg:gap-8">
				for _, feat := range features {
					<div 
						class="flex flex-col justify-between p-6 sm:p-8 rounded-(--radius-card) bg-surface-muted/40 border border-content/10 hover:border-content/20 transition-all duration-200 ease-out"
					>
						<div>
							<div class="flex items-center justify-between gap-4 mb-4">
								<span class="text-xs font-medium px-2.5 py-1 rounded-full bg-content/5 text-content-muted border border-content/10">
									{ feat.Tag }
								</span>
								<span class="text-2xl font-bold font-mono text-content">
									{ feat.Metric }
								</span>
							</div>
							<h3 class="text-xl font-semibold text-content mb-2">
								{ feat.Title }
							</h3>
							<p class="text-sm text-content-muted leading-relaxed">
								{ feat.Description }
							</p>
						</div>

						<div class="mt-6 pt-4 border-t border-content/10">
							<a 
								href="#contact" 
								class="inline-flex items-center text-sm font-medium text-primary hover:underline gap-1.5 focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 rounded-sm"
							>
								Рассчитать стоимость
								<span aria-hidden="true">&rarr;</span>
							</a>
						</div>
					</div>
				}
			</div>
		</div>
	</section>
}
```

---

## 3. Рецепты командных модификаций

### Преобразование «робкого» блока в «смелый» (`bolder`):
1. Заголовок: `text-2xl font-medium` → `text-4xl lg:text-5xl font-black tracking-tight`.
2. Метрика: `text-xl` → `text-5xl font-mono font-bold text-primary`.
3. Разделитель: Добавить глубокую контрастную подложку `bg-surface-muted` или полноэкранный акцентный блок.

### Очистка блока от шума (`quieter`):
1. Бейджи: Убрать яркие разноцветные фоны `bg-amber-500`, `bg-blue-500` → единый нейтральный `bg-content/5 text-content-muted border border-content/10`.
2. Рамки: Убрать внутренние границы у мелких карточек, оставив разделение `gap-6`.
3. Анимации: Убрать бесконечные пульсации и вращения.
