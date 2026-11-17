package events

import "sitecraft/internal/models"

const (
	EventLeadCreated = "lead.created"
	EventOrderPaid   = "order.paid"
	EventSiteMutated = "site.mutated"
)

// LeadCreatedEvent is fired when a visitor submits a lead/booking form.
type LeadCreatedEvent struct {
	Settings models.SiteSettings
	Lead     models.Lead
}

func (e LeadCreatedEvent) EventName() string {
	return EventLeadCreated
}

// OrderPaidEvent is fired when an order/service invoice is successfully paid.
type OrderPaidEvent struct {
	OrderCode string
	Amount    float64
	PaymentID string
}

func (e OrderPaidEvent) EventName() string {
	return EventOrderPaid
}

// SiteMutatedEvent is fired when administrative configuration or content changes.
type SiteMutatedEvent struct {
	Scope string // "settings", "content", "services", "articles", "blocks", "reviews", "theme"
}

func (e SiteMutatedEvent) EventName() string {
	return EventSiteMutated
}

