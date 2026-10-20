package notifications

import (
	"fmt"
	"log"
	"strings"
	"time"

	"sitecraft/internal/data"

	"github.com/pocketbase/pocketbase/core"
)

// CheckBookingReminders scans upcoming bookings at NOW() + 2 hours horizon and dispatches 1-click confirm/cancel links.
func CheckBookingReminders(app core.App) {
	settings := data.GetSiteSettings(app)
	if settings.TelegramBotToken == "" || settings.TelegramChatID == "" {
		return
	}

	records, err := app.FindRecordsByFilter("leads", "status = 'новая' || status = 'подтверждена'", "booking_date,booking_time", 50, 0)
	if err != nil || len(records) == 0 {
		return
	}

	now := time.Now()
	for _, r := range records {
		dateStr := strings.TrimSpace(r.GetString("booking_date"))
		timeStr := strings.TrimSpace(r.GetString("booking_time"))
		if dateStr == "" || timeStr == "" {
			continue
		}

		visitTime, err := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+timeStr, time.Local)
		if err != nil {
			continue
		}

		diff := visitTime.Sub(now)
		// Горизонт NOW() + 2 hours: проверяем записи от 1 часа 20 мин до 2 часов 20 мин
		if diff >= 80*time.Minute && diff <= 140*time.Minute {
			var payload map[string]interface{}
			if rawJSON := r.Get("payload"); rawJSON != nil {
				if m, ok := rawJSON.(map[string]interface{}); ok {
					payload = m
				}
			}
			if payload == nil {
				payload = map[string]interface{}{}
			}

			// Если напоминание уже отправлялось — пропускаем
			if sent, ok := payload["reminder_sent"].(bool); ok && sent {
				continue
			}

			name := r.GetString("name")
			phone := r.GetString("phone")
			id := r.Id

			// Формируем текст напоминания со ссылками действия в 1 клик
			msg := fmt.Sprintf(
				"⏰ <b>Напоминание о визите через 2 часа!</b>\n\n"+
					"👤 <b>Клиент:</b> %s\n"+
					"📞 <b>Телефон:</b> %s\n"+
					"🗓 <b>Время визита:</b> %s в %s\n\n"+
					"Действия в 1 клик:\n"+
					"👉 <a href=\"/api/booking/action?id=%s&action=confirm\">✅ Подтвердить визит</a>\n"+
					"👉 <a href=\"/api/booking/action?id=%s&action=cancel\">❌ Отменить запись</a>",
				name, phone, dateStr, timeStr, id, id,
			)

			bodyData := map[string]interface{}{
				"chat_id":    settings.TelegramChatID,
				"text":       msg,
				"parse_mode": "HTML",
			}

			SendTelegramMessage(settings.TelegramBotToken, bodyData)

			payload["reminder_sent"] = true
			r.Set("payload", payload)
			if err := app.Save(r); err != nil {
				log.Printf("[Reminder Error] Не удалось обновить флаг reminder_sent для лида %s: %v\n", id, err)
			} else {
				log.Printf("[Reminder] Уведомление о визите за 2 часа отправлено для лида %s (%s %s)\n", id, dateStr, timeStr)
			}
		}
	}
}
