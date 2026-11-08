# Motion: Хореография движений и виртуальный курсор (SiteCraft Edition)

Анимация в Faux UI существует не ради движения, а ради **демонстрации ценности продукта** («one thing happens»). Каждое действие имеет видимую причину и понятный результат.

---

## 1. Анатомия сценария (Story & Beats)

Хорошая демонстрация длится **8–14 секунд** и состоит из 3–4 шагов:
1. **Beat 1: Наведение (Intent).** Виртуальный курсор плавно перемещается к целевому элементу (кнопке, строке, переключателю).
2. **Beat 2: Действие (Action).** Клик с микро-сжатием курсора (`scale-90`) и подсветкой элемента (hover/active state).
3. **Beat 3: Реакция системы (Reaction).** Появление попапа, плавная смена текста, переключение таба или появление строки лога.
4. **Beat 4: Награда / Фиксация (Payoff).** Конечный результат удерживается на экране 2–3 секунды для комфортного чтения.
5. **Beat 5: Зацикливание (Loop Reset).** Мягкий фейд или возврат в исходное состояние с логической паузой.

---

## 2. Реализация виртуального курсора на Alpine.js

В SiteCraft мы не используем тяжелые JS-библиотеки. Виртуальный курсор управляется легковесным контроллером Alpine.js `stageshowCursorController` через CSS-транзишены:

```javascript
// Паттерн контроллера виртуального курсора
function stageshowCursorController(steps) {
    return {
        step: 0,
        x: steps[0].x,
        y: steps[0].y,
        clicking: false,
        timer: null,
        reducedMotion: window.matchMedia('(prefers-reduced-motion: reduce)').matches,

        init() {
            if (this.reducedMotion) return;
            this.runLoop();
        },
        runLoop() {
            const current = steps[this.step];
            this.x = current.x;
            this.y = current.y;
            this.clicking = current.click || false;

            this.timer = setTimeout(() => {
                this.step = (this.step + 1) % steps.length;
                this.runLoop();
            }, current.duration || 1500);
        },
        destroy() {
            if (this.timer) clearTimeout(this.timer);
        }
    }
}
```

* **Поддержка `prefers-reduced-motion`:**
  * Если у пользователя включен режим пониженной подвижности (или активен ГОСТ a11y режим), курсор скрывается (`hidden`), а сцена отображает итоговое статическое состояние (Poster frame).

---

## 3. Кривые плавности (Easing) и 60 FPS

* **Никаких `linear` движений:**
  * Перемещение курсора: `cubic-bezier(0.2, 0.8, 0.2, 1)` (плавный разгон и замедление перед кликом).
  * Появление плашек и карточек: `cubic-bezier(0.16, 1, 0.3, 1)` (пружинящий вход без резкости).
* **Аппаратные свойства:**
  * Анимируются исключительно `transform` (`translate3d`, `scale`) и `opacity`.
  * Никогда не анимировать `top`, `left`, `width`, `height` напрямую, чтобы избежать reflow/repaint.
