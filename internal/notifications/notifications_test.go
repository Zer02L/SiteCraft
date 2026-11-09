package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"sitecraft/internal/events"
	"sitecraft/internal/models"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestTelegramNotifier_EmptyConfig(t *testing.T) {
	notifier := &TelegramNotifier{}
	ctx := context.Background()

	// Should safely return nil without making any network call
	err := notifier.NotifyLead(ctx, models.SiteSettings{}, models.Lead{Name: "Иван"})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestTelegramNotifier_SendLead(t *testing.T) {
	var capturedReq *http.Request
	var capturedBody []byte
	var mu sync.Mutex

	reset := SetHTTPClient(&http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			mu.Lock()
			defer mu.Unlock()
			capturedReq = r
			capturedBody, _ = io.ReadAll(r.Body)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"ok":true}`))),
				Header:     make(http.Header),
			}, nil
		}),
	})
	defer reset()

	notifier := &TelegramNotifier{}
	settings := models.SiteSettings{
		TelegramBotToken: "test_bot_token_123",
		TelegramChatID:   "chat_999",
	}

	lead := models.Lead{
		Name:        "Алексей Смирнов",
		Phone:       "+7 (999) 111-22-33",
		BookingDate: "2026-12-01",
		BookingTime: "14:30",
		Source:      "yandex_direct",
		Payload: map[string]interface{}{
			"utm_source":   "yandex",
			"utm_campaign": "promo_winter",
			"quiz_summary": map[string]interface{}{
				"estimated_price":   "15 000",
				"preferred_contact": "Telegram",
			},
			"ab_variant": "B",
			"comment":    "Срочный заказ",
		},
	}

	err := notifier.NotifyLead(context.Background(), settings, lead)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if capturedReq == nil {
		t.Fatal("expected HTTP request to have been sent to Telegram")
	}

	expectedURL := "https://api.telegram.org/bottest_bot_token_123/sendMessage"
	if capturedReq.URL.String() != expectedURL {
		t.Errorf("expected URL %s, got %s", expectedURL, capturedReq.URL.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(capturedBody, &payload); err != nil {
		t.Fatalf("failed to decode JSON body: %v", err)
	}

	if payload["chat_id"] != "chat_999" {
		t.Errorf("expected chat_id 'chat_999', got %v", payload["chat_id"])
	}
	if payload["parse_mode"] != "HTML" {
		t.Errorf("expected parse_mode 'HTML', got %v", payload["parse_mode"])
	}

	text, _ := payload["text"].(string)
	checks := []string{
		"Алексей Смирнов",
		"+7 (999) 111-22-33",
		"2026-12-01 в 14:30",
		"yandex_direct",
		"yandex (кампания: promo_winter)",
		"15 000 ₽ (контакт: Telegram)",
		"Вариант B",
		"Срочный заказ",
	}

	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Errorf("expected notification text to contain %q, but got:\n%s", check, text)
		}
	}
}

func TestWebhookNotifier(t *testing.T) {
	t.Run("empty webhook url", func(t *testing.T) {
		w := &WebhookNotifier{}
		err := w.NotifyLead(context.Background(), models.SiteSettings{WebhookURL: ""}, models.Lead{})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("valid webhook dispatch", func(t *testing.T) {
		var capturedReq *http.Request
		var capturedBody []byte

		reset := SetHTTPClient(&http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				capturedReq = r
				capturedBody, _ = io.ReadAll(r.Body)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"status":"ok"}`))),
					Header:     make(http.Header),
				}, nil
			}),
		})
		defer reset()

		w := &WebhookNotifier{}
		settings := models.SiteSettings{
			WebhookURL: "https://crm.example.com/api/leads",
		}
		lead := models.Lead{
			Name:  "Екатерина",
			Phone: "+7 (911) 222-33-44",
			Email: "ekaterina@example.com",
		}

		err := w.NotifyLead(context.Background(), settings, lead)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if capturedReq == nil {
			t.Fatal("expected webhook HTTP request")
		}
		if capturedReq.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json Content-Type, got %s", capturedReq.Header.Get("Content-Type"))
		}
		if capturedReq.Header.Get("User-Agent") != "SiteCraft-Notifier/1.0" {
			t.Errorf("expected SiteCraft user agent, got %s", capturedReq.Header.Get("User-Agent"))
		}

		var receivedLead models.Lead
		if err := json.Unmarshal(capturedBody, &receivedLead); err != nil {
			t.Fatalf("failed to decode webhook body: %v", err)
		}
		if receivedLead.Name != lead.Name || receivedLead.Phone != lead.Phone {
			t.Errorf("expected %+v, got %+v", lead, receivedLead)
		}
	})
}

func TestCompositeNotifier(t *testing.T) {
	var callCount1, callCount2 int

	m1 := &mockNotifier{fn: func() { callCount1++ }}
	m2 := &mockNotifier{fn: func() { callCount2++ }}

	comp := NewCompositeNotifier(m1, m2)
	err := comp.NotifyLead(context.Background(), models.SiteSettings{}, models.Lead{Name: "Test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount1 != 1 || callCount2 != 1 {
		t.Errorf("expected both notifiers called once, got m1=%d, m2=%d", callCount1, callCount2)
	}
}

type mockNotifier struct {
	fn func()
}

func (m *mockNotifier) NotifyLead(ctx context.Context, settings models.SiteSettings, lead models.Lead) error {
	m.fn()
	return nil
}

func TestRegisterSubscribers(t *testing.T) {
	bus := events.NewMemoryBus()

	done := make(chan bool, 1)
	reset := SetHTTPClient(&http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			done <- true
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"ok":true}`))),
				Header:     make(http.Header),
			}, nil
		}),
	})
	defer reset()

	RegisterSubscribers(bus)

	bus.Publish(context.Background(), events.LeadCreatedEvent{
		Settings: models.SiteSettings{
			TelegramBotToken: "tok",
			TelegramChatID:   "chat",
		},
		Lead: models.Lead{Name: "EventBus User"},
	})

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for RegisterSubscribers listener to be triggered")
	}
}
