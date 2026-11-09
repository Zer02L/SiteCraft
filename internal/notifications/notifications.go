package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"sitecraft/internal/events"
	"sitecraft/internal/models"
)

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
}

// SetHTTPClient overrides the internal http client (primarily for testing) and returns a cleanup function.
func SetHTTPClient(client *http.Client) func() {
	prev := httpClient
	httpClient = client
	return func() {
		httpClient = prev
	}
}

// Notifier defines the interface for outbound lead notifications.
type Notifier interface {
	NotifyLead(ctx context.Context, settings models.SiteSettings, lead models.Lead) error
}

// TelegramNotifier sends lead messages to Telegram.
type TelegramNotifier struct{}

func (t *TelegramNotifier) NotifyLead(ctx context.Context, settings models.SiteSettings, lead models.Lead) error {
	if settings.TelegramBotToken == "" || settings.TelegramChatID == "" {
		return nil
	}
	sendTelegram(settings.TelegramBotToken, settings.TelegramChatID, lead)
	return nil
}

// WebhookNotifier dispatches leads to external CRM systems.
type WebhookNotifier struct{}

func (w *WebhookNotifier) NotifyLead(ctx context.Context, settings models.SiteSettings, lead models.Lead) error {
	if settings.WebhookURL == "" {
		return nil
	}
	sendWebhook(settings.WebhookURL, lead)
	return nil
}

// CompositeNotifier dispatches notifications to multiple notifiers.
type CompositeNotifier struct {
	notifiers []Notifier
}

func NewCompositeNotifier(notifiers ...Notifier) *CompositeNotifier {
	return &CompositeNotifier{notifiers: notifiers}
}

func (c *CompositeNotifier) NotifyLead(ctx context.Context, settings models.SiteSettings, lead models.Lead) error {
	for _, n := range c.notifiers {
		_ = n.NotifyLead(ctx, settings, lead)
	}
	return nil
}

// RegisterSubscribers attaches notification handlers to the event bus.
func RegisterSubscribers(bus events.Bus) {
	composite := NewCompositeNotifier(
		&TelegramNotifier{},
		&WebhookNotifier{},
	)

	bus.Subscribe(events.EventLeadCreated, func(ctx context.Context, event events.Event) error {
		leadEvent, ok := event.(events.LeadCreatedEvent)
		if !ok {
			return nil
		}
		return composite.NotifyLead(ctx, leadEvent.Settings, leadEvent.Lead)
	})
}

// SendLeadNotification maintains backward compatibility while delegating.
func SendLeadNotification(settings models.SiteSettings, lead models.Lead) {
	events.Default().Publish(context.Background(), events.LeadCreatedEvent{
		Settings: settings,
		Lead:     lead,
	})
}

func sendTelegram(token, chatID string, lead models.Lead) {
	text := fmt.Sprintf(
		"🔔 <b>Новая заявка с сайта!</b>\n\n"+
			"👤 <b>Имя:</b> %s\n"+
			"📞 <b>Телефон:</b> %s\n",
		lead.Name, lead.Phone,
	)

	if lead.BookingDate != "" && lead.BookingTime != "" {
		text += fmt.Sprintf("🗓 <b>Дата и время:</b> %s в %s\n", lead.BookingDate, lead.BookingTime)
	}

	if lead.Source != "" && lead.Source != "website" {
		text += fmt.Sprintf("📍 <b>Источник:</b> %s\n", lead.Source)
	}

	if len(lead.Payload) > 0 {
		// Отображаем рекламную кампанию (UTM)
		if src, ok := lead.Payload["utm_source"].(string); ok && src != "" {
			cmp, _ := lead.Payload["utm_campaign"].(string)
			if cmp != "" {
				text += fmt.Sprintf("🎯 <b>Реклама:</b> %s (кампания: %s)\n", src, cmp)
			} else {
				text += fmt.Sprintf("🎯 <b>Реклама:</b> %s\n", src)
			}
		}

		// Отображаем расчет квиза, если есть
		if qMap, ok := lead.Payload["quiz_summary"].(map[string]interface{}); ok {
			if est, ok := qMap["estimated_price"]; ok {
				pref, _ := qMap["preferred_contact"].(string)
				text += fmt.Sprintf("💡 <b>Расчет квиза:</b> %v ₽ (контакт: %s)\n", est, pref)
			}
		}

		// Отображаем вариант серверного A/B сплит-теста
		if ab, ok := lead.Payload["ab_variant"].(string); ok && ab != "" {
			text += fmt.Sprintf("🧪 <b>A/B тест:</b> Вариант %s\n", ab)
		}

		for k, v := range lead.Payload {
			if k == "utm" || k == "utm_source" || k == "utm_campaign" || k == "quiz_summary" || k == "ab_variant" {
				continue
			}
			if strVal := fmt.Sprintf("%v", v); strVal != "" {
				text += fmt.Sprintf("📝 <b>%s:</b> %s\n", k, strVal)
			}
		}
	}

	bodyData := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	SendTelegramMessage(token, bodyData)
}

// SendTelegramMessage dispatches generic telegram bot payloads.
func SendTelegramMessage(token string, bodyData map[string]interface{}) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payloadBytes, err := json.Marshal(bodyData)
	if err != nil {
		log.Printf("[Notification Error] Failed to encode TG payload: %v\n", err)
		return
	}

	resp, err := httpClient.Post(apiURL, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		log.Printf("[Notification Error] Failed to send Telegram alert: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("[Notification Error] Telegram API responded with status: %d\n", resp.StatusCode)
	}
}

func sendWebhook(webhookURL string, lead models.Lead) {
	payloadBytes, err := json.Marshal(lead)
	if err != nil {
		log.Printf("[Notification Error] Failed to encode webhook payload: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(payloadBytes))
	if err != nil {
		log.Printf("[Notification Error] Failed to create webhook request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "SiteCraft-Notifier/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[Notification Error] Failed to dispatch webhook: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("[Notification Error] Webhook responded with status: %d\n", resp.StatusCode)
	}
}
