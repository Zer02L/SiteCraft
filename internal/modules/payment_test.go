package modules

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sitecraft/internal/models"
)

func TestNewPaymentProvider(t *testing.T) {
	tests := []struct {
		name       string
		settings   models.SiteSettings
		expectMock bool
	}{
		{
			name: "payment disabled in settings",
			settings: models.SiteSettings{
				EnablePayment: false,
				PaymentShopID: "123456",
				PaymentSecret: "live_sec_xxx",
			},
			expectMock: true,
		},
		{
			name: "payment enabled but missing shop ID",
			settings: models.SiteSettings{
				EnablePayment: true,
				PaymentShopID: "",
				PaymentSecret: "live_sec_xxx",
			},
			expectMock: true,
		},
		{
			name: "payment enabled but missing secret",
			settings: models.SiteSettings{
				EnablePayment: true,
				PaymentShopID: "123456",
				PaymentSecret: "",
			},
			expectMock: true,
		},
		{
			name: "payment enabled with valid credentials",
			settings: models.SiteSettings{
				EnablePayment:  true,
				PaymentShopID:  "shop_789",
				PaymentSecret:  "sec_abc",
				PaymentSandbox: true,
			},
			expectMock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewPaymentProvider(tt.settings)
			if tt.expectMock {
				if _, ok := provider.(*MockPaymentProvider); !ok {
					t.Fatalf("expected *MockPaymentProvider, got %T", provider)
				}
			} else {
				yk, ok := provider.(*YooKassaClient)
				if !ok {
					t.Fatalf("expected *YooKassaClient, got %T", provider)
				}
				if yk.ShopID != tt.settings.PaymentShopID {
					t.Errorf("expected ShopID %q, got %q", tt.settings.PaymentShopID, yk.ShopID)
				}
				if yk.Secret != tt.settings.PaymentSecret {
					t.Errorf("expected Secret %q, got %q", tt.settings.PaymentSecret, yk.Secret)
				}
				if yk.Sandbox != tt.settings.PaymentSandbox {
					t.Errorf("expected Sandbox %v, got %v", tt.settings.PaymentSandbox, yk.Sandbox)
				}
			}
		})
	}
}

func TestMockPaymentProvider(t *testing.T) {
	provider := &MockPaymentProvider{}
	req := models.PaymentRequest{
		Amount:      1500.50,
		Currency:    "RUB",
		Description: "Оплата консультации",
		ReturnURL:   "/payment/success",
	}

	resp, err := provider.CreatePayment(req)
	if err != nil {
		t.Fatalf("unexpected error from mock provider: %v", err)
	}

	if !resp.IsMock {
		t.Errorf("expected IsMock to be true, got %v", resp.IsMock)
	}
	if resp.Status != "pending" {
		t.Errorf("expected status 'pending', got %q", resp.Status)
	}
	if !strings.HasPrefix(resp.PaymentID, "mock_pay_") {
		t.Errorf("expected PaymentID to start with 'mock_pay_', got %q", resp.PaymentID)
	}
	if !strings.Contains(resp.PaymentURL, "amount=1500.50") {
		t.Errorf("expected PaymentURL to contain amount, got %q", resp.PaymentURL)
	}
}

func TestYooKassaClient_CreatePayment_Success(t *testing.T) {
	expectedShopID := "shop_test_123"
	expectedSecret := "secret_key_456"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		user, pass, ok := r.BasicAuth()
		if !ok || user != expectedShopID || pass != expectedSecret {
			t.Errorf("invalid BasicAuth: got user=%q, pass=%q, ok=%v", user, pass, ok)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Idempotence-Key") == "" {
			t.Errorf("expected Idempotence-Key header to be present")
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("failed to parse json payload: %v", err)
		}

		amountMap, ok := payload["amount"].(map[string]interface{})
		if !ok || amountMap["value"] != "2990.00" || amountMap["currency"] != "RUB" {
			t.Errorf("unexpected amount payload: %#v", amountMap)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "2d8f99e0-000f-5000-8000-11bba8fb5e0d",
			"status": "pending",
			"confirmation": {
				"type": "redirect",
				"confirmation_url": "https://yoomoney.ru/checkout/payments/v2/contract?orderId=2d8f99e0"
			}
		}`))
	}))
	defer server.Close()

	client := &YooKassaClient{
		ShopID:     expectedShopID,
		Secret:     expectedSecret,
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}

	resp, err := client.CreatePayment(models.PaymentRequest{
		Amount:      2990.00,
		Currency:    "", // should default to RUB
		Description: "Подписка на сервис",
		ReturnURL:   "https://example.com/return",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.IsMock {
		t.Errorf("expected IsMock to be false")
	}
	if resp.PaymentID != "2d8f99e0-000f-5000-8000-11bba8fb5e0d" {
		t.Errorf("unexpected PaymentID: %s", resp.PaymentID)
	}
	if resp.Status != "pending" {
		t.Errorf("unexpected status: %s", resp.Status)
	}
	if resp.PaymentURL != "https://yoomoney.ru/checkout/payments/v2/contract?orderId=2d8f99e0" {
		t.Errorf("unexpected PaymentURL: %s", resp.PaymentURL)
	}
}

func TestYooKassaClient_CreatePayment_ApiError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"type":"error","code":"invalid_request","description":"Invalid amount value"}`))
	}))
	defer server.Close()

	client := &YooKassaClient{
		ShopID:     "shop_id",
		Secret:     "secret",
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}

	_, err := client.CreatePayment(models.PaymentRequest{
		Amount: -50,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected error to mention 400 status, got %v", err)
	}
}

func TestYooKassaClient_CreatePayment_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer server.Close()

	client := &YooKassaClient{
		ShopID:     "shop_id",
		Secret:     "secret",
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}

	_, err := client.CreatePayment(models.PaymentRequest{Amount: 100})
	if err == nil {
		t.Fatal("expected JSON parse error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestYooKassaClient_CreatePayment_NetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	serverURL := server.URL
	server.Close() // immediately close to simulate network failure

	client := &YooKassaClient{
		ShopID:     "shop_id",
		Secret:     "secret",
		BaseURL:    serverURL,
		HTTPClient: &http.Client{},
	}

	_, err := client.CreatePayment(models.PaymentRequest{Amount: 100})
	if err == nil {
		t.Fatal("expected network error, got nil")
	}
	if !strings.Contains(err.Error(), "yookassa request failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}
