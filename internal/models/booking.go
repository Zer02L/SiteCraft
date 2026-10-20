package models

// ResourceItem represents a specialist, master, workstation or cabinet for resource booking.
type ResourceItem struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`               // e.g. "Захаров Владислав Борисович" or "Мастер Алексей"
	Role     string   `json:"role"`               // e.g. "Ведущий юрист", "Старший специалист", "Бокс №1"
	Avatar   string   `json:"avatar,omitempty"`     // Image URL or emoji/icon
	Active   bool     `json:"active"`
	Order    int      `json:"order"`
	Services []string `json:"services,omitempty"` // IDs of services this resource can perform
}
