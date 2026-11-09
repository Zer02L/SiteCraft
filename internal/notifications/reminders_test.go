package notifications

import (
	"context"
	"testing"
	"time"

	"sitecraft/internal/models"
	"sitecraft/internal/testutil"
)

func TestCheckBookingReminders(t *testing.T) {
	repo, sqlDB, _ := testutil.NewTestRepo(t)
	ctx := context.Background()

	// Update settings with telegram credentials
	_, err := sqlDB.ExecContext(ctx, "UPDATE site_settings SET telegram_bot_token = 'test_token', telegram_chat_id = 'test_chat'")
	if err != nil {
		t.Fatalf("failed to update site settings: %v", err)
	}

	now := time.Now()

	// 1. Far away lead
	_, _ = repo.CreateLead(ctx, models.Lead{
		Name:        "Клиент Далеко",
		Phone:       "+79991110001",
		Status:      "новая",
		BookingDate: now.Add(24 * time.Hour).Format("2006-01-02"),
		BookingTime: now.Add(24 * time.Hour).Format("15:04"),
	})

	// 2. Target lead in 100 min window
	targetDate := now.Add(100 * time.Minute)
	_, _ = repo.CreateLead(ctx, models.Lead{
		Name:        "Клиент 2Часа",
		Phone:       "+79992220002",
		Status:      "подтверждена",
		BookingDate: targetDate.Format("2006-01-02"),
		BookingTime: targetDate.Format("15:04"),
	})

	// 3. Cancelled lead in window
	_, _ = repo.CreateLead(ctx, models.Lead{
		Name:        "Клиент Отменен",
		Phone:       "+79993330003",
		Status:      "отменена",
		BookingDate: targetDate.Format("2006-01-02"),
		BookingTime: targetDate.Format("15:04"),
	})

	// Execute check without panics
	CheckBookingReminders(ctx, repo)
}
