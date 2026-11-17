package admin_blocks

import (
	"strings"
)

// HumanBlockType returns a friendly Russian name and an emoji/icon for a block type.
func HumanBlockType(bType string) (name string, icon string) {
	lower := strings.ToLower(strings.TrimSpace(bType))
	switch lower {
	case "hero":
		return "Главный экран (Hero)", "⚡"
	case "cinematic_odometer", "cinematic_sticky_scroll_odometer", "odometer":
		return "Интерактивный счётчик", "⏱"
	case "cinematic_spotlight", "spotlight":
		return "Световой прожектор", "💡"
	case "comparison", "before_after", "cinematic_comparison":
		return "Сравнение до / после", "⚖️"
	case "cinematic_sticky_stack", "cinematic_sticky_cards":
		return "Каскадные карточки", "📑"
	case "cinematic_accordion":
		return "Аккордеон с подсветкой", "☰"
	case "cinematic_marquee":
		return "Бегущая строка", "⇄"
	case "cinematic_text_mask":
		return "Кинетическая типографика", "🔤"
	case "features":
		return "Преимущества", "✦"
	case "services":
		return "Услуги и цены", "💼"
	case "booking":
		return "Онлайн-запись", "📅"
	case "quiz", "calculator":
		return "Калькулятор / Квиз", "🧮"
	case "steps":
		return "Этапы работы", "🪜"
	case "reviews":
		return "Отзывы клиентов", "★"
	case "team":
		return "Команда специалистов", "👥"
	case "faq":
		return "Вопросы и ответы", "💬"
	case "gallery":
		return "Фотогалерея", "🖼"
	case "map", "contacts", "branches":
		return "Контакты и карта", "📍"
	case "showcase":
		return "Витрина проектов", "🖥"
	default:
		clean := strings.TrimPrefix(lower, "cinematic_")
		clean = strings.ReplaceAll(clean, "_", " ")
		if len(clean) > 0 {
			clean = strings.ToUpper(clean[:1]) + clean[1:]
		}
		return clean, "❖"
	}
}
