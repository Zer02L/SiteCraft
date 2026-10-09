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
