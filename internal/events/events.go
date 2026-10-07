package events

import "sitecraft/internal/models"

const (
	EventLeadCreated = "lead.created"
	EventOrderPaid   = "order.paid"
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
