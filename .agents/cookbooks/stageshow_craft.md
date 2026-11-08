# Cookbook: Stageshow-Craft (Faux Product UI in SiteCraft)

Каноническое руководство по созданию убедительных, кинематографичных мокапов продуктов (Hero-шоты SaaS, мобильные экраны, Bento-сетки фичей, анимированный курсор) в SiteCraft на базе Go, Templ, Alpine.js и Tailwind CSS v4 без зависимостей от Node.js или React.

---

## 1. Главные принципы
- **3–7 осмысленных строк:** реальный текст пишется только на фокусном действии; второстепенный заменяется на `PlaceholderBar`.
- **Срез краев (Edge Bleed):** окно уходит за границу карточки или контейнера (`-mr-16` / `-mb-20`), создавая ощущение масштаба и глубины.
- **Сателлиты (Satellites):** нависающие карточки результатов (квитанция, push, тост, плашка метрики), перекрывающие край основного окна.
- **ГОСТ a11y & Reduced Motion:** при `prefers-reduced-motion` виртуальный курсор автоматически скрывается, а макет показывает статичный финальный кадр.

---

## 2. Пример 1: Hero-демо SaaS с окном приложения, курсором и сателлитом

```templ
package sections

import "sitecraft/ui/components/cinematic"

templ SaaSHeroDemo() {
	@cinematic.StageshowStage(cinematic.StageshowStageProps{
		ID:         "hero-demo-stage",
		Background: "bg-surface/50 border border-content/10",
		Class:      "min-h-[440px] p-8 flex items-center justify-center relative",
	}) {
		<!-- Главное окно приложения со срезом по нижнему правому краю -->
		@cinematic.MockWindow(cinematic.MockWindowProps{
			Title: "SiteCraft Cloud / Deployments",
			URL:   "app.sitecraft.dev/live",
			Class: "w-[560px] h-[340px] -mr-8 -mb-12",
		}) {
			<div class="p-5 flex flex-col gap-3">
				<!-- Заголовок с реальным текстом -->
				<div class="flex items-center justify-between pb-3 border-b border-content/10">
					<div class="flex flex-col gap-1">
						<span class="text-xs font-semibold text-content">Production Cluster</span>
						@cinematic.PlaceholderBar(cinematic.PlaceholderBarProps{Width: "w-32", Height: "h-2"})
					</div>
					<span class="text-[10px] font-mono px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-500 border border-emerald-500/20">
						Active
					</span>
				</div>

				<!-- Список элементов: фокусный элемент подсвечен -->
				<div class="flex flex-col gap-2">
					@cinematic.MockRow(cinematic.MockRowProps{
						Title:      "edge-gateway-eu-central",
						Subtitle:   "Deployed 2m ago · SHA 9f2a01",
						BadgeText:  "Healthy",
						BadgeColor: "text-emerald-400 bg-emerald-500/10",
						Active:     true,
					})
					@cinematic.MockRow(cinematic.MockRowProps{
						Title:      "api-worker-node-01",
						Subtitle:   "Uptime 99.98% · 24 req/s",
						BadgeText:  "200 OK",
						BadgeColor: "text-zinc-400 bg-content/5",
						Active:     false,
					})
					@cinematic.MockRow(cinematic.MockRowProps{
						Title:      "pocketbase-sqlite-replica",
						Subtitle:   "Sync complete · WAL mode",
						BadgeText:  "Standby",
						BadgeColor: "text-zinc-400 bg-content/5",
						Active:     false,
					})
				</div>
			</div>
		}

		<!-- Сателлит: нависающая карточка подтверждения -->
		@cinematic.MockSatellite(cinematic.MockSatelliteProps{
			Class: "-bottom-2 -left-2 sm:left-6 max-w-xs",
		}) {
			<div class="flex items-center gap-3">
				<div class="w-8 h-8 rounded-full bg-emerald-500/15 flex items-center justify-center text-emerald-500">
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
					</svg>
				</div>
				<div class="flex flex-col">
					<span class="text-xs font-semibold text-content">1-Click Deploy</span>
					<span class="text-[10px] text-content/60">SSL Certificate auto-renewed</span>
				</div>
			</div>
		}

		<!-- Виртуальный анимированный курсор -->
		@cinematic.MockCursor(cinematic.MockCursorProps{
			ID: "hero-cursor",
			Waypoints: []cinematic.MockCursorWaypoint{
				{X: 180, Y: 120, Click: false, DurationMs: 1200},
				{X: 280, Y: 195, Click: true, DurationMs: 2000},
				{X: 380, Y: 260, Click: false, DurationMs: 1500},
			},
		})
	}
}
```

---

## 3. Пример 2: Мобильный Faux UI (Смартфон с Dynamic Island)

```templ
@cinematic.StageshowStage(cinematic.StageshowStageProps{
	Background: "bg-surface/30",
	Class:      "p-6 flex items-center justify-center",
}) {
	@cinematic.MockPhone(cinematic.MockPhoneProps{
		DynamicIsland: true,
		Class:         "-mb-16",
	}) {
		<div class="p-4 flex flex-col gap-4">
			<div class="flex justify-between items-center text-xs font-medium text-content">
				<span>SiteCraft Pay</span>
				<span class="font-mono text-[10px] text-content/50">12:45</span>
			</div>
			
			<div class="p-3 bg-content/5 rounded-xl border border-content/10 flex flex-col gap-1">
				<span class="text-[11px] text-content/60">Баланс счета</span>
				<span class="text-xl font-bold font-mono text-content">1,420,500 ₽</span>
			</div>

			<div class="space-y-2">
				@cinematic.MockRow(cinematic.MockRowProps{
					Title:      "Перевод по СБП",
					Subtitle:   "Сегодня, 11:20",
					BadgeText:  "+ 45,000 ₽",
					BadgeColor: "text-emerald-400 bg-emerald-500/10",
					Active:     true,
				})
				@cinematic.MockRow(cinematic.MockRowProps{
					Title:      "Подписка Cloud Pro",
					Subtitle:   "Вчера, 19:40",
					BadgeText:  "- 1,290 ₽",
					BadgeColor: "text-content/60 bg-content/5",
				})
			</div>
		</div>
	}
}
```

---

## 4. Пример 3: Продуктовая Bento-сетка (4 Tiles)

```templ
<div class="grid grid-cols-1 md:grid-cols-12 gap-4 sm:gap-6">
	<!-- Tile 1: 8 cols — Hero Faux UI с обрезанным окном -->
	<div class="md:col-span-8 bg-surface rounded-2xl border border-content/10 p-6 overflow-hidden relative min-h-[300px]">
		<div class="max-w-xs mb-4">
			<h3 class="text-base font-bold text-content">Управление кластером</h3>
			<p class="text-xs text-content/60 mt-1">Автоматическая репликация SQLite с нулевой задержкой.</p>
		</div>
		<div class="absolute right-0 bottom-0 translate-x-12 translate-y-8 w-[420px]">
			@cinematic.MockWindow(cinematic.MockWindowProps{Title: "Cluster Dashboard"}) {
				<div class="p-4 space-y-2">
					@cinematic.MockRow(cinematic.MockRowProps{Title: "Node Alpha", Active: true, BadgeText: "Primary"})
					@cinematic.MockRow(cinematic.MockRowProps{Title: "Node Beta", BadgeText: "Replica"})
				</div>
			}
		</div>
	</div>

	<!-- Tile 2: 4 cols — Ключевая метрика -->
	<div class="md:col-span-4 bg-surface rounded-2xl border border-content/10 p-6 flex flex-col justify-between">
		<span class="text-xs font-mono uppercase tracking-wider text-content/50">Performance</span>
		<div class="my-4">
			<div class="text-4xl font-extrabold font-mono text-content">0.12 ms</div>
			<div class="text-xs text-emerald-500 mt-1 flex items-center gap-1 font-medium">
				<span>↑ 94%</span> быстрее внешних СУБД
			</div>
		</div>
		<p class="text-xs text-content/60 leading-relaxed">
			Встроенный PocketBase исключает сетевые задержки при запросах к базе данных.
		</p>
	</div>

	<!-- Tile 3: 4 cols — Смартфон с пушем -->
	<div class="md:col-span-4 bg-surface rounded-2xl border border-content/10 p-6 overflow-hidden relative min-h-[280px]">
		<h3 class="text-sm font-bold text-content">Мобильный мониторинг</h3>
		<div class="absolute -bottom-16 left-1/2 -translate-x-1/2">
			@cinematic.MockPhone(cinematic.MockPhoneProps{DynamicIsland: true}) {
				<div class="p-3 text-[11px] text-content/80">Оповещения в реальном времени</div>
			}
		</div>
	</div>

	<!-- Tile 4: 8 cols — Терминал логов -->
	<div class="md:col-span-8 bg-surface rounded-2xl border border-content/10 p-6 flex flex-col justify-between">
		<div class="mb-4">
			<h3 class="text-base font-bold text-content">Автономный деплой</h3>
			<p class="text-xs text-content/60 mt-1">Один бинарный файл без внешних демонов и npm.</p>
		</div>
		@cinematic.MockTerminal(cinematic.MockTerminalProps{Title: "sitecraft-build"}) {
			<div class="text-zinc-400">$ make build</div>
			<div class="text-emerald-400">==> Binary compiled: bin/server (24 MB)</div>
			<div class="text-zinc-500">==> Single binary ready for production deployment.</div>
		}
	</div>
</div>
```
