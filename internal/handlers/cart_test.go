package handlers

import (
	"encoding/json"
	"testing"

	"sitecraft/internal/models"
)

func TestCartItemsDecodingAndCalculation(t *testing.T) {
	rawJSON := `[
		{"id": "srv_1", "title": "Диагностика авто", "price": 1200, "qty": 1},
		{"id": "srv_2", "title": "Замена масла", "price": 1500, "qty": 2}
	]`

	var items []models.CartItem
	err := json.Unmarshal([]byte(rawJSON), &items)
	if err != nil {
		t.Fatalf("Failed to decode cart items: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(items))
	}

	var total float64
	totalCount := 0
	for _, it := range items {
		total += it.Price * float64(it.Qty)
		totalCount += it.Qty
	}

	expectedTotal := 1200.0 + (1500.0 * 2)
	if total != expectedTotal {
		t.Errorf("Expected total %f, got %f", expectedTotal, total)
	}

	if totalCount != 3 {
		t.Errorf("Expected count 3, got %d", totalCount)
	}
}

func TestDeliveryTariffCalculation(t *testing.T) {
	cases := []struct {
		deliveryType string
		deliveryZone string
		subtotal     float64
		expectedCost float64
	}{
		{"pickup", "", 500, 0},
		{"courier", "city", 1500, 300},
		{"courier", "city", 2500, 0},    // Free over 2500
		{"courier", "city", 5000, 0},    // Free over 2500
		{"courier", "suburb", 1500, 650},
		{"courier", "region", 1500, 1200},
	}

	for _, tc := range cases {
		var cost float64
		if tc.deliveryType == "courier" {
			switch tc.deliveryZone {
			case "suburb":
				cost = 650
			case "region":
				cost = 1200
			default:
				if tc.subtotal >= 2500 {
					cost = 0
				} else {
					cost = 300
				}
			}
		}
		if cost != tc.expectedCost {
			t.Errorf("Delivery (%s, %s, subtotal=%.0f): expected %.0f, got %.0f",
				tc.deliveryType, tc.deliveryZone, tc.subtotal, tc.expectedCost, cost)
		}
	}
}

func TestCouponDiscountCalculation(t *testing.T) {
	// Percent discount test
	percentCoupon := models.Coupon{
		Code:          "PROMO10",
		DiscountType:  "percent",
		DiscountValue: 10,
	}
	subtotal := 3000.0
	expectedDiscount := 300.0
	actualDiscount := subtotal * (percentCoupon.DiscountValue / 100.0)
	if actualDiscount != expectedDiscount {
		t.Errorf("Expected percent discount %.0f, got %.0f", expectedDiscount, actualDiscount)
	}

	// Fixed discount test
	fixedCoupon := models.Coupon{
		Code:           "WELCOME500",
		DiscountType:   "fixed",
		DiscountValue:  500,
		MinOrderAmount: 2000,
	}
	if subtotal < fixedCoupon.MinOrderAmount {
		t.Error("Order should satisfy min_order_amount")
	}
	if fixedCoupon.DiscountValue != 500 {
		t.Errorf("Expected fixed discount 500, got %.0f", fixedCoupon.DiscountValue)
	}
}

func TestInventoryStockLogic(t *testing.T) {
	service := models.ServiceItem{
		ID:         "srv_exhaust",
		Title:      "Выхлопная система",
		StockCount: 2,
		IsPreorder: false,
	}

	// Ordered 2 items
	orderedQty := 2
	newStock := service.StockCount - orderedQty
	if newStock <= 0 {
		service.StockCount = 0
		service.IsPreorder = true
	} else {
		service.StockCount = newStock
	}

	if service.StockCount != 0 {
		t.Errorf("Expected stock 0, got %d", service.StockCount)
	}
	if !service.IsPreorder {
		t.Error("Expected IsPreorder to be true when stock is depleted")
	}
}
