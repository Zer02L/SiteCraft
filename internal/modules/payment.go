package modules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"sitecraft/internal/models"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// PaymentProvider abstracts online payment systems (ЮKassa, Tinkoff, Stripe, Mock)
type PaymentProvider interface {
	CreatePayment(req models.PaymentRequest) (models.PaymentResponse, error)
}

// YooKassaClient implements payments via ЮKassa REST API
type YooKassaClient struct {
	ShopID     string
	Secret     string
	Sandbox    bool
	BaseURL    string       // Optional override for testing (default: https://api.yookassa.ru/v3/payments)
	HTTPClient *http.Client // Optional HTTP client override for testing
}

// NewPaymentProvider creates a payment provider based on site settings
func NewPaymentProvider(settings models.SiteSettings) PaymentProvider {
	if !settings.EnablePayment || settings.PaymentShopID == "" || settings.PaymentSecret == "" {
		// Mock-провайдер для тестирования и песочницы без боевых ключей
		return &MockPaymentProvider{}
	}
	return &YooKassaClient{
		ShopID:   settings.PaymentShopID,
		Secret:   settings.PaymentSecret,
		Sandbox:  settings.PaymentSandbox,
	}
}

// CreatePayment creates a payment order in ЮKassa
func (y *YooKassaClient) CreatePayment(req models.PaymentRequest) (models.PaymentResponse, error) {
	currency := req.Currency
	if currency == "" {
		currency = "RUB"
	}

	payload := map[string]interface{}{
		"amount": map[string]interface{}{
			"value":    fmt.Sprintf("%.2f", req.Amount),
			"currency": currency,
		},
		"capture": true,
		"confirmation": map[string]interface{}{
			"type":       "redirect",
			"return_url": req.ReturnURL,
		},
		"description": req.Description,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return models.PaymentResponse{}, err
	}

	apiURL := y.BaseURL
	if apiURL == "" {
		apiURL = "https://api.yookassa.ru/v3/payments"
	}
	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return models.PaymentResponse{}, err
	}

	httpReq.SetBasicAuth(y.ShopID, y.Secret)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", fmt.Sprintf("%d", time.Now().UnixNano()))

	client := y.HTTPClient
	if client == nil {
		client = httpClient
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return models.PaymentResponse{}, fmt.Errorf("yookassa request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return models.PaymentResponse{}, fmt.Errorf("yookassa error status %d: %s", resp.StatusCode, string(respBody))
	}

	var resData struct {
		ID           string `json:"id"`
		Status       string `json:"status"`
		Confirmation struct {
			ConfirmationURL string `json:"confirmation_url"`
		} `json:"confirmation"`
	}

	if err := json.Unmarshal(respBody, &resData); err != nil {
		return models.PaymentResponse{}, fmt.Errorf("failed to parse yookassa response: %w", err)
	}

	return models.PaymentResponse{
		PaymentID:  resData.ID,
		Status:     resData.Status,
		PaymentURL: resData.Confirmation.ConfirmationURL,
		IsMock:     false,
	}, nil
}

// MockPaymentProvider handles test payments
type MockPaymentProvider struct{}

func (m *MockPaymentProvider) CreatePayment(req models.PaymentRequest) (models.PaymentResponse, error) {
	mockID := fmt.Sprintf("mock_pay_%d", time.Now().Unix())
	log.Printf("[Payment Mock] Created test payment %s for amount: %.2f %s\n", mockID, req.Amount, req.Currency)
	return models.PaymentResponse{
		PaymentID:  mockID,
		Status:     "pending",
		PaymentURL: fmt.Sprintf("/payment/success?payment_id=%s&amount=%.2f", mockID, req.Amount),
		IsMock:     true,
	}, nil
}
