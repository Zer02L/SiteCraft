package checker_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"sitecraft/internal/checker"
	"sitecraft/internal/events"
	"sitecraft/internal/models"

	_ "modernc.org/sqlite"
)

type mockRegistry struct {
	disabled map[string]bool
}

func (m *mockRegistry) IsActive(id string, settings models.SiteSettings) bool {
	if m.disabled[id] {
		return false
	}
	return settings.IsModuleActive(id)
}

func TestEvaluateSiteHealth_InitialSettings(t *testing.T) {
	settings := models.SiteSettings{
		SiteName: "Новый сайт",
	}

	res := checker.EvaluateSiteHealth(settings, nil)

	if res.TotalCount == 0 {
		t.Fatal("expected non-zero total items count")
	}

	if res.IsReady {
		t.Error("expected site not to be ready initially")
	}

	if res.Percentage > 50 {
		t.Errorf("expected low completion percentage, got %d%%", res.Percentage)
	}

	pending := res.PendingItems()
	if len(pending) == 0 {
		t.Error("expected pending items")
	}

	critCount := res.CriticalPendingCount()
	if critCount == 0 {
		t.Error("expected pending critical items (contacts, legal, blocks)")
	}
}

func TestEvaluateSiteHealth_FullyConfigured(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE services (id TEXT PRIMARY KEY, is_deleted INTEGER, active INTEGER);
		CREATE TABLE articles (id TEXT PRIMARY KEY, is_deleted INTEGER, is_published INTEGER);
		CREATE TABLE reviews (id TEXT PRIMARY KEY, is_published INTEGER);
		CREATE TABLE page_blocks (id TEXT PRIMARY KEY, active INTEGER);
		CREATE TABLE site_content (phone1 TEXT, phone2 TEXT, address TEXT, address_sub TEXT);

		INSERT INTO services VALUES ('s1', 0, 1);
		INSERT INTO articles VALUES ('a1', 0, 1);
		INSERT INTO reviews VALUES ('r1', 1);
		INSERT INTO page_blocks VALUES ('b1', 1);
		INSERT INTO site_content VALUES ('+7 999 111-22-33', '', 'г. Москва, ул. Арбат, 1', '');
	`)
	if err != nil {
		t.Fatalf("failed to init schema: %v", err)
	}

	settings := models.SiteSettings{
		SiteName:            "Готовая Компания",
		LegalEntity:         "ООО 'Вектор'",
		INN:                 "7701234567",
		EnableCookieConsent: true,
		ContactEmail:        "info@vector.ru",
		TelegramBotToken:    "bot12345:TOKEN",
		TelegramChatID:      "-10012345",
		YandexMetrikaID:     "12345678",
		EnablePayment:       true,
		PaymentShopID:       "shop_123",
		PaymentSecret:       "sec_xyz",
		EnableBlog:          true,
	}

	res := checker.EvaluateSiteHealth(settings, db)

	if res.Percentage != 100 {
		t.Errorf("expected 100%% completion, got %d%% (pending: %d)", res.Percentage, len(res.PendingItems()))
		for _, it := range res.PendingItems() {
			t.Logf("pending item: %s (%s)", it.Title, it.ID)
		}
	}

	if !res.IsReady {
		t.Error("expected IsReady = true for 100% completed checklist")
	}

	if len(res.PendingItems()) != 0 {
		t.Errorf("expected 0 pending items, got %d", len(res.PendingItems()))
	}

	if len(res.CompletedItems()) != res.TotalCount {
		t.Errorf("expected all %d items to be completed, got %d", res.TotalCount, len(res.CompletedItems()))
	}
}

func TestEvaluateSiteHealth_ModuleFiltering(t *testing.T) {
	settings := models.SiteSettings{
		SiteName:      "Базовый тариф",
		EnablePayment: false,
		EnableBlog:    false,
	}

	reg := &mockRegistry{
		disabled: map[string]bool{
			"services": true,
			"reviews":  true,
			"articles": true,
		},
	}

	res := checker.EvaluateSiteHealth(settings, nil, checker.WithRegistry(reg.IsActive))

	for _, item := range res.Items {
		if item.ID == "services" {
			t.Errorf("services module was disabled, but item appeared in checklist")
		}
		if item.ID == "articles" || item.ID == "blog" {
			t.Errorf("blog/articles was disabled, but item appeared in checklist")
		}
		if item.ID == "reviews" {
			t.Errorf("reviews module was disabled, but item appeared in checklist")
		}
		if item.ID == "payment" {
			t.Errorf("payment was disabled, but item appeared in checklist")
		}
	}
}

func TestService_CachingAndEventBus(t *testing.T) {
	bus := events.NewMemoryBus()
	reg := &mockRegistry{disabled: map[string]bool{}}

	settings := models.SiteSettings{
		SiteName: "Тестовый инстанс",
	}

	svc := checker.NewService(nil, nil, bus, reg.IsActive)

	res1 := svc.GetChecklist(context.Background(), settings)
	if res1.TotalCount == 0 {
		t.Fatal("expected non-zero total items")
	}

	// Immediate subsequent call should return cached result
	res2 := svc.GetChecklist(context.Background(), settings)
	if res1.Percentage != res2.Percentage {
		t.Errorf("expected cached percentage to match: %d vs %d", res1.Percentage, res2.Percentage)
	}

	// Trigger mutation event via EventBus
	bus.Publish(context.Background(), events.SiteMutatedEvent{Scope: "settings"})

	// Give goroutine a moment to invalidate
	time.Sleep(50 * time.Millisecond)

	// Update settings
	settings.YandexMetrikaID = "998877"
	res3 := svc.GetChecklist(context.Background(), settings)

	foundYandex := false
	for _, it := range res3.Items {
		if it.ID == "analytics" && it.IsCompleted {
			foundYandex = true
			break
		}
	}
	if !foundYandex {
		t.Error("expected analytics to be marked completed after cache invalidation and updated settings")
	}
}
