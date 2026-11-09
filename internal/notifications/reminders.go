package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"sitecraft/internal/data"
)

var (
	sentReminders   = make(map[string]time.Time)
	sentRemindersMu sync.Mutex
)

// CheckBookingReminders scans upcoming bookings at NOW() + 2 hours horizon and dispatches notifications.
func CheckBookingReminders(ctx context.Context, repo *data.Repository) {
	settings, err := repo.GetSiteSettings(ctx)
	if err != nil || settings.TelegramBotToken == "" || settings.TelegramChatID == "" {
		return
	}

	leads, err := repo.GetLeads(ctx)
	if err != nil || len(leads) == 0 {
		return
	}

	now := time.Now()

	// Clean up old sent reminders older than 24 hours
	sentRemindersMu.Lock()
	for id, t := range sentReminders {
		if now.Sub(t) > 24*time.Hour {
			delete(sentReminders, id)
		}
	}
	sentRemindersMu.Unlock()

	for _, l := range leads {
		if l.Status != "новая" && l.Status != "подтверждена" {
			continue
		}
		dateStr := strings.TrimSpace(l.BookingDate)
		timeStr := strings.TrimSpace(l.BookingTime)
		if dateStr == "" || timeStr == "" {
			continue
		}

		visitTime, err := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+timeStr, time.Local)
		if err != nil {
			continue
		}

		diff := visitTime.Sub(now)
		// Horizon NOW() + 2 hours: check bookings between 80 min and 140 min
		if diff >= 80*time.Minute && diff <= 140*time.Minute {
			sentRemindersMu.Lock()
			_, alreadySent := sentReminders[l.ID]
			if !alreadySent {
				sentReminders[l.ID] = now
			}
			sentRemindersMu.Unlock()

			if alreadySent {
				continue
			}

			msg := fmt.Sprintf("🔔 <b>Напоминание о записи</b>\nКлиент: %s (%s)\nВремя: %s %s", l.Name, l.Phone, dateStr, timeStr)
			slog.Info("Sending booking reminder", "lead_id", l.ID, "client", l.Name)

			bodyData := map[string]interface{}{
				"chat_id":    settings.TelegramChatID,
				"text":       msg,
				"parse_mode": "HTML",
			}
			SendTelegramMessage(settings.TelegramBotToken, bodyData)
		}
	}
}

// CheckArticlesSchedule invalidates articles cache periodically so scheduled posts appear.
func CheckArticlesSchedule(repo *data.Repository) {
	if repo != nil {
		repo.Cache().InvalidateArticles()
	}
}
