package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"sitecraft/internal/models"
)

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
}

// SendLeadNotification sends notification to Telegram and/or external CRM webhook
func SendLeadNotification(settings models.SiteSettings, lead models.Lead) {
	if settings.TelegramBotToken != "" && settings.TelegramChatID != "" {
		go sendTelegram(settings.TelegramBotToken, settings.TelegramChatID, lead)
	}

	if settings.WebhookURL != "" {
		go sendWebhook(settings.WebhookURL, lead)
	}
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

	if len(lead.Payload) > 0 {
		for k, v := range lead.Payload {
			if strVal := fmt.Sprintf("%v", v); strVal != "" {
				text += fmt.Sprintf("📝 <b>%s:</b> %s\n", k, strVal)
			}
		}
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	bodyData := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

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
