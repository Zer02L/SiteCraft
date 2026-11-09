package db_test

import (
	"context"
	"testing"

	"sitecraft/internal/db"
)

func TestOpenAndMigrations(t *testing.T) {
	conn, queries, err := db.Open("file:memdb1?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("db.Open failed: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()

	// Verify Querier interface works
	var q db.Querier = queries

	// Test Insert and Get Page
	page, err := q.CreatePage(ctx, db.CreatePageParams{
		ID:             "test-page-1",
		Slug:           "home",
		Title:          "Home Page",
		SeoTitle:       "SEO Home",
		SeoDescription: "Desc",
		IsPublished:    1,
	})
	if err != nil {
		t.Fatalf("CreatePage failed: %v", err)
	}

	fetched, err := q.GetPageBySlug(ctx, "home")
	if err != nil {
		t.Fatalf("GetPageBySlug failed: %v", err)
	}
	if fetched.ID != page.ID {
		t.Errorf("expected ID %s, got %s", page.ID, fetched.ID)
	}

	// Test Create Lead
	lead, err := q.CreateLead(ctx, db.CreateLeadParams{
		ID:          "lead-1",
		Name:        "Иван Иванов",
		Phone:       "+7 (999) 000-00-00",
		Email:       "ivan@test.local",
		Source:      "form",
		BookingDate: "2026-11-10",
		BookingTime: "14:00",
		ResourceID:  "res-1",
		Status:      "новая",
		Payload:     "{}",
	})
	if err != nil {
		t.Fatalf("CreateLead failed: %v", err)
	}
	if lead.Name != "Иван Иванов" {
		t.Errorf("unexpected lead name: %s", lead.Name)
	}

	count, err := q.CountLeads(ctx)
	if err != nil {
		t.Fatalf("CountLeads failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
}
