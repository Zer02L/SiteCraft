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
}
