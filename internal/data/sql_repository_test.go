package data_test

import (
	"context"
	"testing"

	"sitecraft/internal/data"
	"sitecraft/internal/db"
	"sitecraft/internal/models"
)

func TestSQLRepositoryOperations(t *testing.T) {
	conn, queries, err := db.Open("file:repotest?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer conn.Close()

	repo := data.NewRepository(conn, queries)
	ctx := context.Background()

	// 1. Settings
	settings, err := repo.GetSiteSettings(ctx)
	if err != nil {
		t.Fatalf("GetSiteSettings failed: %v", err)
	}
	settings.SiteName = "Тестовая компания"
	if err := repo.UpsertSiteSettings(ctx, settings); err != nil {
		t.Fatalf("UpsertSiteSettings failed: %v", err)
	}
	updatedSettings, err := repo.GetSiteSettings(ctx)
	if err != nil {
		t.Fatalf("GetSiteSettings after update failed: %v", err)
	}
	if updatedSettings.SiteName != "Тестовая компания" {
		t.Errorf("expected SiteName 'Тестовая компания', got '%s'", updatedSettings.SiteName)
	}

	// 2. Content
	content, err := repo.GetSiteContent(ctx)
	if err != nil {
		t.Fatalf("GetSiteContent failed: %v", err)
	}
	content.HeroTitle = "Главный заголовок"
	if err := repo.UpsertSiteContent(ctx, content); err != nil {
		t.Fatalf("UpsertSiteContent failed: %v", err)
	}
	updatedContent, err := repo.GetSiteContent(ctx)
	if err != nil {
		t.Fatalf("GetSiteContent after update failed: %v", err)
	}
	if updatedContent.HeroTitle != "Главный заголовок" {
		t.Errorf("expected HeroTitle 'Главный заголовок', got '%s'", updatedContent.HeroTitle)
	}

	// 3. Lead Creation & Retrieval
	lead, err := repo.CreateLead(ctx, models.Lead{
		Name:   "Мария",
		Phone:  "+7 999 111-22-33",
		Source: "сайт",
	})
	if err != nil {
		t.Fatalf("CreateLead failed: %v", err)
	}
	if lead.ID == "" {
		t.Errorf("expected lead ID to be populated")
	}

	leads, err := repo.GetLeads(ctx)
	if err != nil {
		t.Fatalf("GetLeads failed: %v", err)
	}
	if len(leads) != 1 {
		t.Errorf("expected 1 lead, got %d", len(leads))
	}

	// 4. Preset Application
	err = repo.ApplyPreset(ctx, "presets/b2b_legal.json")
	if err != nil {
		t.Fatalf("ApplyPreset failed: %v", err)
	}

	services, err := repo.GetServices(ctx)
	if err != nil {
		t.Fatalf("GetServices failed: %v", err)
	}
	if len(services) == 0 {
		t.Errorf("expected services to be loaded from preset")
	}

	blocks, err := repo.GetPageBlocks(ctx)
	if err != nil {
		t.Fatalf("GetPageBlocks failed: %v", err)
	}
	if len(blocks) == 0 {
		t.Errorf("expected page blocks to be loaded from preset")
	}

	// 5. Test Service Mutations & Cache Invalidation
	testSvc := models.ServiceItem{
		Title:       "Консультация юриста",
		Description: "Первичная консультация",
		Price:       "5000",
		Order:       10,
		Active:      true,
	}
	if err := repo.SaveService(ctx, testSvc); err != nil {
		t.Fatalf("SaveService failed: %v", err)
	}
	svcsAfterSave, err := repo.GetServices(ctx)
	if err != nil {
		t.Fatalf("GetServices failed: %v", err)
	}
	foundSvc := false
	var savedSvcID string
	for _, s := range svcsAfterSave {
		if s.Title == "Консультация юриста" {
			foundSvc = true
			savedSvcID = s.ID
			break
		}
	}
	if !foundSvc {
		t.Errorf("expected to find newly saved service in cache/db")
	}

	// Toggle Service
	if err := repo.ToggleService(ctx, savedSvcID); err != nil {
		t.Fatalf("ToggleService failed: %v", err)
	}
	toggledSvc, err := repo.GetServiceByID(ctx, savedSvcID)
	if err != nil {
		t.Fatalf("GetServiceByID failed: %v", err)
	}
	if toggledSvc.Active {
		t.Errorf("expected service to be inactive after toggle")
	}

	// 6. Test Block Swap & Transactions
	b1, err := repo.CreatePageBlock(ctx, models.PageBlock{Type: "hero", Title: "B1", Order: 1, Active: true})
	if err != nil {
		t.Fatalf("CreatePageBlock 1 failed: %v", err)
	}
	b2, err := repo.CreatePageBlock(ctx, models.PageBlock{Type: "faq", Title: "B2", Order: 2, Active: true})
	if err != nil {
		t.Fatalf("CreatePageBlock 2 failed: %v", err)
	}
	if err := repo.SwapPageBlockOrder(ctx, b1.ID, 1, b2.ID, 2); err != nil {
		t.Fatalf("SwapPageBlockOrder failed: %v", err)
	}
	b1Updated, _ := repo.GetPageBlockByID(ctx, b1.ID)
	b2Updated, _ := repo.GetPageBlockByID(ctx, b2.ID)
	if b1Updated.Order != 2 || b2Updated.Order != 1 {
		t.Errorf("expected swapped orders (2 and 1), got b1=%d, b2=%d", b1Updated.Order, b2Updated.Order)
	}

	// Test UpdatePageBlock
	b1Updated.Title = "Updated Title"
	b1Updated.Content = "Updated Content"
	b1Updated.Payload = map[string]interface{}{"variant": "split"}
	if err := repo.UpdatePageBlock(ctx, b1Updated); err != nil {
		t.Fatalf("UpdatePageBlock failed: %v", err)
	}
	b1Fetched, err := repo.GetPageBlockByID(ctx, b1.ID)
	if err != nil {
		t.Fatalf("GetPageBlockByID failed: %v", err)
	}
	if b1Fetched.Title != "Updated Title" || b1Fetched.Content != "Updated Content" {
		t.Errorf("expected updated title and content, got title=%q, content=%q", b1Fetched.Title, b1Fetched.Content)
	}
	if v, ok := b1Fetched.Payload["variant"].(string); !ok || v != "split" {
		t.Errorf("expected payload variant 'split', got %v", b1Fetched.Payload["variant"])
	}
}

