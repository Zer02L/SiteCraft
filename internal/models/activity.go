package models

// RecentActivityItem represents a depersonalized social proof event.
type RecentActivityItem struct {
	Name     string `json:"name"`      // e.g. "Константин" or "Елена"
	City     string `json:"city"`      // e.g. "СПб", "Москва", "Краснодар"
	Action   string `json:"action"`    // e.g. "заказал аудит", "записался на консультацию"
	TimeAgo  string `json:"time_ago"`  // e.g. "12 мин назад", "35 мин назад"
	Icon     string `json:"icon"`      // "lead", "order", "booking"
}
