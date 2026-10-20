package models

// Coupon represents a promotional discount code.
type Coupon struct {
	ID             string  `json:"id"`
	Code           string  `json:"code"`
	DiscountType   string  `json:"discount_type"` // "percent" or "fixed"
	DiscountValue  float64 `json:"discount_value"`
	MinOrderAmount float64 `json:"min_order_amount"`
	ValidUntil     string  `json:"valid_until"`
	UsageLimit     int     `json:"usage_limit"`
	UsedCount      int     `json:"used_count"`
	Active         bool    `json:"active"`
}

// DeliveryZone represents a geographic tariff zone for courier delivery.
type DeliveryZone struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Price    float64 `json:"price"`
	FreeFrom float64 `json:"free_from"` // 0 = no free threshold, >0 = free above this order amount
	Estimate string  `json:"estimate"`  // e.g. "40-60 мин" or "1-2 дня"
}

// CrossSellOption represents an up-sell or companion offer inside the cart drawer.
type CrossSellOption struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Price    float64 `json:"price"`
	Subtitle string  `json:"subtitle"`
	Icon     string  `json:"icon"`
}
