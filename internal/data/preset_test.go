package data

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"sitecraft/internal/models"
	"sitecraft/ui/pages"
	"sitecraft/ui/registry"
)

func TestAllModulesPresetStructure(t *testing.T) {
	presetPaths := []string{
		"../../presets/test_all_modules.json",
		"../../presets/test.json",
	}

	for _, path := range presetPaths {
		t.Run(path, func(t *testing.T) {
			dataBytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read preset file %s: %v", path, err)
			}

			var preset PresetData
			if err := json.Unmarshal(dataBytes, &preset); err != nil {
				t.Fatalf("Failed to unmarshal %s: %v", path, err)
			}

			// 1. Verify E-Commerce flags
			if !preset.Settings.EnableCart {
				t.Errorf("Expected EnableCart=true")
			}
			if !preset.Settings.EnablePayment || !preset.Settings.PaymentSandbox {
				t.Errorf("Expected EnablePayment=true and PaymentSandbox=true")
			}

			// 2. Verify CRO & Marketing flags
			if !preset.Settings.EnableLiveChat {
				t.Errorf("Expected EnableLiveChat=true")
			}
			if !preset.Settings.EnableRecentActivity {
				t.Errorf("Expected EnableRecentActivity=true")
			}
			if !preset.Settings.EnableFloatingContact {
				t.Errorf("Expected EnableFloatingContact=true")
			}
			if preset.Content.HeroTitleB == "" {
				t.Errorf("Expected HeroTitleB for A/B testing")
			}

			// 3. Verify Service & Booking flags
			if preset.Settings.BookingMode != "slots" {
				t.Errorf("Expected BookingMode='slots', got %s", preset.Settings.BookingMode)
			}
			if !preset.Settings.EnableTracker {
				t.Errorf("Expected EnableTracker=true")
			}

			// 4. Verify B2B & Compliance flags
			if !preset.Settings.EnableCareers {
				t.Errorf("Expected EnableCareers=true")
			}
			if !preset.Settings.EnableCookieConsent {
				t.Errorf("Expected EnableCookieConsent=true")
			}
			if preset.Settings.LegalEntity == "" || preset.Settings.INN == "" {
				t.Errorf("Expected LegalEntity and INN for B2B Invoicing & Compliance")
			}

			// 5. Verify Agency & Core flags
			if !preset.Settings.EnableFleet {
				t.Errorf("Expected EnableFleet=true")
			}
			if !preset.Settings.EnableBlog {
				t.Errorf("Expected EnableBlog=true")
			}
			if !preset.Settings.EnableTMA {
				t.Errorf("Expected EnableTMA=true")
			}
			if !preset.Settings.EnableI18n {
				t.Errorf("Expected EnableI18n=true")
			}
			if !preset.Settings.EnableA11y {
				t.Errorf("Expected EnableA11y=true")
			}

			// 6. Verify Services Inventory
			if len(preset.Services) < 4 {
				t.Fatalf("Expected at least 4 services, got %d", len(preset.Services))
			}
			hasLimitedStock := false
			hasPreorder := false
			for _, s := range preset.Services {
				if s.StockCount > 0 && s.StockCount <= 5 {
					hasLimitedStock = true
				}
				if s.StockCount == 0 && s.IsPreorder {
					hasPreorder = true
				}
			}
			if !hasLimitedStock {
				t.Errorf("Expected at least one service with limited stock (<=5) for FOMO badge")
			}
			if !hasPreorder {
				t.Errorf("Expected at least one service with preorder enabled")
			}

			// 7. Verify Page Blocks
			expectedBlocks := []string{
				"hero",
				"cinematic_odometer",
				"cinematic_spotlight",
				"cinematic_marquee",
				"services",
				"quiz",
				"before_after",
				"features",
				"steps",
				"team",
				"reviews",
				"booking",
				"careers",
				"faq",
				"map",
				"showcase",
			}

			foundBlocks := make(map[string]bool)
			for _, b := range preset.Blocks {
				foundBlocks[b.Type] = true
			}

			for _, expected := range expectedBlocks {
				if !foundBlocks[expected] {
					t.Errorf("Expected block type '%s' in preset blocks", expected)
				}
			}
		})
	}
}

func TestRenderAllModulesHomePage(t *testing.T) {
	dataBytes, err := os.ReadFile("../../presets/test_all_modules.json")
	if err != nil {
		t.Fatalf("Failed to read test_all_modules.json: %v", err)
	}

	var preset PresetData
	if err := json.Unmarshal(dataBytes, &preset); err != nil {
		t.Fatalf("Failed to unmarshal preset: %v", err)
	}

	// Mock entities for dependencies
	resources := []models.ResourceItem{
		{ID: "res_1", Name: "Иванов Иван", Role: "Главный инженер", Active: true, Order: 1},
		{ID: "res_2", Name: "Петров Петр", Role: "Специалист техподдержки", Active: true, Order: 2},
	}
	reviews := []models.ReviewItem{
		{ID: "rev_1", AuthorName: "Алексей С.", Rating: 5, Comment: "Отличный тестовый стенд!", IsPublished: true},
	}
	branches := []models.BranchItem{
		{ID: "br_1", Name: "Центральный офис", City: "Москва", Address: "Пресненская наб., 12", Phone: "+7 (495) 123-45-67", Active: true},
	}
	vacancies := []models.VacancyItem{
		{ID: "vac_1", Title: "Senior Go Developer", Department: "Инжиниринг", Salary: "от 300 000 ₽", Active: true},
	}
	bookedSlots := map[string][]string{
		"2026-10-20": {"10:00", "12:00"},
	}

	// Render Home Page with all blocks and components
	var buf bytes.Buffer
	component := pages.Home(pages.HomePageData{
		Settings:    preset.Settings,
		Content:     preset.Content,
		Services:    preset.Services,
		BookedSlots: bookedSlots,
		Blocks:      preset.Blocks,
		Resources:   resources,
		Reviews:     reviews,
		Branches:    branches,
		Vacancies:   vacancies,
	})

	err = component.Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render Home page with all modules: %v", err)
	}

	html := buf.String()
	if len(html) < 5000 {
		t.Fatalf("Rendered HTML is suspiciously small (%d bytes)", len(html))
	}

	// Verify key components rendered into HTML
	checks := []struct {
		desc    string
		snippet string
	}{
		{"A11y Bar", "a11y-toolbar"},
		{"Header", "<header"},
		{"Language Switcher", "?lang=en"},
		{"Services Section", "$store.cart.add"},
		{"Quiz Section", "quiz"},
		{"Before/After Section", "before-after"},
		{"Booking Section", "booking"},
		{"Reviews Section", "reviews"},
		{"Careers Section", "careers"},
		{"FAQ Section", "faq"},
		{"Footer", "<footer"},
		{"Cart Drawer", "cart-drawer"},
		{"Chat Widget", "sitecraft_chat_sid"},
		{"Recent Activity Toast", "recent-activity"},
		{"Exit Intent Modal", "exit-intent"},
		{"Cookie Consent Banner", "sitecraft_cookie_consent"},
		{"Floating Contact", "https://t.me/"},
	}

	for _, check := range checks {
		if !strings.Contains(html, check.snippet) {
			t.Errorf("Missing %s (expected snippet '%s' in rendered HTML)", check.desc, check.snippet)
		}
	}
}

func TestParsersWithTestPresetPayloads(t *testing.T) {
	dataBytes, err := os.ReadFile("../../presets/test_all_modules.json")
	if err != nil {
		t.Fatalf("Failed to read test_all_modules.json: %v", err)
	}

	var preset PresetData
	if err := json.Unmarshal(dataBytes, &preset); err != nil {
		t.Fatalf("Failed to unmarshal preset: %v", err)
	}

	for _, block := range preset.Blocks {
		switch block.Type {
		case "quiz":
			payload := registry.ParseQuizPayload(block.Payload, preset.Settings.Theme)
			if len(payload.Steps) != 2 {
				t.Errorf("Expected 2 quiz steps, got %d", len(payload.Steps))
			}
			if payload.BasePrice != 10000 {
				t.Errorf("Expected BasePrice 10000, got %f", payload.BasePrice)
			}
		case "before_after":
			payload := registry.ParseBeforeAfterPayload(block.Payload, preset.Settings.Theme)
			if len(payload.Items) != 1 {
				t.Errorf("Expected 1 before/after item, got %d", len(payload.Items))
			}
			if payload.Items[0].Duration != "3 рабочих дня" {
				t.Errorf("Expected Duration '3 рабочих дня', got %s", payload.Items[0].Duration)
			}
		case "features":
			payload := registry.ParseFeaturesPayload(block.Payload)
			if payload == nil || len(payload.Items) != 4 {
				t.Errorf("Expected 4 feature items, got %v", payload)
			}
		case "faq":
			payload := registry.ParseFAQPayload(block.Payload)
			if payload == nil || len(payload.Items) != 4 {
				t.Errorf("Expected 4 FAQ items, got %v", payload)
			}
		}
	}
}
