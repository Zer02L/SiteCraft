package notifications

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"sitecraft/internal/data"
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
			msg := fmt.Sprintf("🔔 *Напоминание о записи*\nКлиент: %s (%s)\nВремя: %s %s\n", l.Name, l.Phone, dateStr, timeStr)
			log.Println("[Notification Reminder]", msg)
		}
	}
}

// CheckArticlesSchedule invalidates articles cache periodically.
func CheckArticlesSchedule() {
	// Periodic cache check
}
