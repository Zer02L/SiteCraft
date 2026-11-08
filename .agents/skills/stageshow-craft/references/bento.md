# Bento: Построение продуктовых Bento-сеток (SiteCraft Edition)

Bento-сетка — это единая картина из плиток, а не 4–6 разрозненных скриншотов окон. Если в каждой плитке поместить по мини-окну, макет превращается в шумный каталог.

---

## 1. Золотое правило Bento: Баланс контента

В гармоничной Bento-сетке:
* **~35–45% плиток** — это Faux UI (главное окно, телефон или демо).
* **~30% плиток** — это метрики / статистика (крупные цифры `text-4xl font-bold`, процент роста, граф).
* **~25% плиток** — это типографика или тезис (яркий инсайт, цитата, архитектурная схема).

---

## 2. Анатомия Bento-сетки из 4 плиток

```
+------------------------------------------+--------------------+
|                                          |                    |
|  TILE 1 (Hero Faux UI - 8 cols)          |  TILE 2 (Stat)     |
|  Окно приложения со срезом по правому     |  +99.8%            |
|  краю. Виртуальный курсор кликает экшен. |  Uptime SLA        |
|                                          |                    |
+--------------------+---------------------+--------------------+
|                    |                                          |
|  TILE 3 (Mobile)   |  TILE 4 (Workflow / Terminal - 8 cols)   |
|  Смартфон с пушем  |  Строка развертывания и интерактивный    |
|  авторизации       |  чек-лист                                |
|                    |                                          |
+--------------------+------------------------------------------+
```

---

## 3. Компоновка классов в Tailwind CSS v4

```html
<div class="grid grid-cols-1 md:grid-cols-12 gap-4 sm:gap-6">
    <!-- Tile 1: Hero UI (8 cols) -->
    <div class="md:col-span-8 bg-surface rounded-2xl border border-content/10 p-6 overflow-hidden relative">
        ...
    </div>

    <!-- Tile 2: Metric (4 cols) -->
    <div class="md:col-span-4 bg-surface rounded-2xl border border-content/10 p-6 flex flex-col justify-between">
        ...
    </div>

    <!-- Tile 3: Satellite Device (4 cols) -->
    <div class="md:col-span-4 bg-surface rounded-2xl border border-content/10 p-6 overflow-hidden">
        ...
    </div>

    <!-- Tile 4: Command / Action (8 cols) -->
    <div class="md:col-span-8 bg-surface rounded-2xl border border-content/10 p-6">
        ...
    </div>
</div>
```
