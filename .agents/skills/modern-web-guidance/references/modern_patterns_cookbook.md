# Рецепты современных веб-стандартов (Modern Web Cookbook)

Канонические примеры компонентов для архитектуры **Go 1.27 + Templ + Tailwind CSS v4 + Alpine.js + HTMX**.

---

## 1. Нативное модальное окно (`<dialog>` + `@starting-style`)

Заменяет устаревший оверлей `fixed inset-0 z-50` и ручной фокус-трап на JS.

```templ
package components

// ModalDialog демонстрирует нативный тег <dialog> с плавной анимацией через @starting-style
templ ModalDialog(id string, title string) {
	<dialog
		id={ id }
		class="m-auto rounded-2xl bg-surface p-6 shadow-2xl backdrop:bg-black/60 backdrop:backdrop-blur-sm max-w-lg w-full transition-[opacity,transform,overlay,display] duration-200 ease-out [transition-behavior:allow-discrete] opacity-0 scale-95 open:opacity-100 open:scale-100 starting:open:opacity-0 starting:open:scale-95"
		data-component="modal-dialog"
		data-source="ui/components/modal_dialog.templ"
	>
		<div class="flex items-center justify-between pb-4 border-b border-content/10">
			<h3 class="text-xl font-bold text-content">{ title }</h3>
			<button
				type="button"
				onclick={ templ.SafeScript("document.getElementById('" + id + "').close()") }
				class="p-2 text-content-muted hover:text-content rounded-lg focus-visible:ring-2 focus-visible:ring-primary focus:outline-none"
				aria-label="Закрыть диалоговое окно"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
				</svg>
			</button>
		</div>

		<div class="py-4 text-content-muted">
			{ children... }
		</div>
	</dialog>
}
```

**Открытие из любого места:**
```html
<button onclick="document.getElementById('lead-modal').showModal()" class="btn-primary">
    Оставить заявку
</button>
```

---

## 2. Всплывающее меню на Popover API (Zero-JS Dropdown)

Браузер автоматически обеспечивает закрытие по клику вне (Light Dismiss) и по клавише Esc.

```templ
package components

templ PopoverDropdown(id string, buttonText string) {
	<div class="relative inline-block">
		<!-- Кнопка-триггер связывается через popovertarget -->
		<button
			type="button"
			popovertarget={ id }
			class="px-4 py-2 bg-surface border border-content/10 rounded-xl text-content font-medium hover:bg-surface-muted focus-visible:ring-2 focus-visible:ring-primary focus:outline-none flex items-center gap-2"
		>
			<span>{ buttonText }</span>
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
			</svg>
		</button>

		<!-- Popover элемент: открывается в Top Layer -->
		<div
			id={ id }
			popover="auto"
			class="m-0 mt-2 p-2 w-56 rounded-xl bg-surface border border-content/10 shadow-xl transition-[opacity,transform,overlay,display] duration-150 ease-out [transition-behavior:allow-discrete] opacity-0 translate-y-1 open:opacity-100 open:translate-y-0 starting:open:opacity-0 starting:open:translate-y-1"
		>
			<ul class="space-y-1 text-sm text-content">
				<li>
					<a href="/profile" class="block px-3 py-2 rounded-lg hover:bg-surface-muted">Профиль</a>
				</li>
				<li>
					<a href="/settings" class="block px-3 py-2 rounded-lg hover:bg-surface-muted">Настройки</a>
				</li>
				<li class="border-t border-content/10 my-1"></li>
				<li>
					<button class="w-full text-left px-3 py-2 rounded-lg text-red-500 hover:bg-red-500/10">Выход</button>
				</li>
			</ul>
		</div>
	</div>
}
```

---

## 3. View Transitions в HTMX связке

Бесшовный переход между представлениями без SPA-фреймворков.

### Разметка в Templ:
```templ
<button
	hx-get="/api/services/filter?category=b2b"
	hx-target="#services-grid"
	hx-swap="innerHTML transition:true"
	class="px-5 py-2.5 rounded-xl font-medium bg-primary text-white"
>
	Корпоративные услуги
</button>

<div id="services-grid" class="grid grid-cols-1 md:grid-cols-3 gap-6">
	<!-- Содержимое плавно морфится при свопе HTMX -->
</div>
```

### Настройка в `ui/assets/input.css`:
```css
/* Включение нативных переходов между страницами */
@view-transition {
  navigation: auto;
}

/* Настройка морфинга конкретных элементов */
#services-grid {
  view-transition-name: services-catalog;
}

::view-transition-old(services-catalog),
::view-transition-new(services-catalog) {
  animation-duration: 220ms;
  animation-timing-function: cubic-bezier(0.16, 1, 0.3, 1);
}
```

---

## 4. Контейнерные запросы в Tailwind CSS v4 (`@container`)

Карточка сама меняет компоновку с вертикальной на горизонтальную в зависимости от ширины родителя:

```templ
package components

templ AdaptiveBentoCard(title, desc, tag, imgURL string) {
	<!-- Определяем контекст контейнера -->
	<div class="@container w-full">
		<article class="flex flex-col @md:flex-row items-center gap-6 p-6 rounded-2xl bg-surface border border-content/10">
			<div class="w-full @md:w-1/3 aspect-video @md:aspect-square overflow-hidden rounded-xl">
				<img
					src={ imgURL }
					alt={ title }
					loading="lazy"
					decoding="async"
					class="w-full h-full object-cover"
				/>
			</div>
			<div class="w-full @md:w-2/3 space-y-2">
				<span class="text-xs font-mono font-medium text-accent uppercase tracking-wider">{ tag }</span>
				<h3 class="text-xl @lg:text-2xl font-bold text-content">{ title }</h3>
				<p class="text-content-muted text-sm @lg:text-base leading-relaxed">{ desc }</p>
			</div>
		</article>
	</div>
}
```

---

## 5. Оптимизация LCP Hero-изображения (Core Web Vitals)

Гарантирует мгновенную предзагрузку ключевого баннера без блокировки потока:

```templ
package sections

templ HeroSection(title, subtitle, heroImgURL string) {
	<section class="relative min-h-[70vh] flex items-center bg-surface py-16">
		<div class="container mx-auto px-4 grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
			<div class="space-y-6">
				<h1 class="text-4xl sm:text-6xl font-black text-content tracking-tight">{ title }</h1>
				<p class="text-lg text-content-muted leading-relaxed">{ subtitle }</p>
			</div>
			<div class="relative aspect-video rounded-3xl overflow-hidden shadow-2xl">
				<!-- LCP Image: fetchpriority="high", loading="eager", decoding="async" -->
				<img
					src={ heroImgURL }
					alt="Обложка сервиса"
					fetchpriority="high"
					loading="eager"
					decoding="async"
					class="w-full h-full object-cover"
				/>
			</div>
		</div>
	</section>
}
```

---

## 6. Валидация форм без JS через `:user-invalid`

Ошибки подсвечиваются только после того, как пользователь потрогал поле и убрал фокус.

```templ
package components

templ ModernFormInput(id, name, label, inputType, pattern, errorMsg string, required bool) {
	<div class="space-y-1">
		<label for={ id } class="block text-sm font-medium text-content">
			{ label }
			if required {
				<span class="text-accent">*</span>
			}
		</label>
		<input
			id={ id }
			name={ name }
			type={ inputType }
			if required {
				required
			}
			if pattern != "" {
				pattern={ pattern }
			}
			class="w-full px-4 py-2.5 rounded-xl border border-content/20 bg-surface text-content focus-visible:ring-2 focus-visible:ring-primary focus:outline-none invalid:[&:not(:placeholder-shown):not(:focus)]:border-red-500 invalid:[&:not(:placeholder-shown):not(:focus)]:text-red-500 transition-colors"
			placeholder=" "
		/>
		<p class="hidden text-xs text-red-500 mt-1 [input:invalid:not(:placeholder-shown):not(:focus)~&]:block">
			{ errorMsg }
		</p>
	</div>
}
```
