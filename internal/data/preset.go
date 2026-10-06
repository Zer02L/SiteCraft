package data

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"sitecraft/internal/cache"
	"sitecraft/internal/models"

	"github.com/pocketbase/pocketbase/core"
)

type PresetData struct {
	Settings models.SiteSettings  `json:"settings"`
	Content  models.SiteContent   `json:"content"`
	Services []models.ServiceItem `json:"services"`
}

// ApplyPreset loads a JSON preset from file and writes it to PocketBase collections
func ApplyPreset(app core.App, presetPath string) error {
	dataBytes, err := os.ReadFile(presetPath)
	if err != nil {
		return fmt.Errorf("failed to read preset file %s: %w", presetPath, err)
	}

	var preset PresetData
	if err := json.Unmarshal(dataBytes, &preset); err != nil {
		return fmt.Errorf("failed to parse preset JSON: %w", err)
	}

	// 1. Update site_settings
	settingsCol, err := app.FindCollectionByNameOrId("site_settings")
	if err == nil {
		rec, _ := app.FindFirstRecordByFilter("site_settings", "")
		if rec == nil {
			rec = core.NewRecord(settingsCol)
		}
		if preset.Settings.SiteName != "" {
			rec.Set("site_name", preset.Settings.SiteName)
		}
		if preset.Settings.Tagline != "" {
			rec.Set("tagline", preset.Settings.Tagline)
		}
		if preset.Settings.Theme != "" {
			rec.Set("theme", preset.Settings.Theme)
		}
		rec.Set("enable_a11y", preset.Settings.EnableA11y)
		rec.Set("enable_alpine", preset.Settings.EnableAlpine)
		if preset.Settings.MapMode != "" {
			rec.Set("map_mode", preset.Settings.MapMode)
		}
		if preset.Settings.BookingMode != "" {
			rec.Set("booking_mode", preset.Settings.BookingMode)
		}
		if err := app.Save(rec); err != nil {
			log.Printf("[Preset Error] Failed to save site_settings: %v\n", err)
		}
	}

	// 2. Update site_content
	contentCol, err := app.FindCollectionByNameOrId("site_content")
	if err == nil {
		rec, _ := app.FindFirstRecordByFilter("site_content", "")
		if rec == nil {
			rec = core.NewRecord(contentCol)
		}
		rec.Set("hero_badge", preset.Content.HeroBadge)
		rec.Set("hero_title", preset.Content.HeroTitle)
		rec.Set("hero_subtitle", preset.Content.HeroSubtitle)
		rec.Set("owner_name", preset.Content.OwnerName)
		rec.Set("owner_title", preset.Content.OwnerTitle)
		rec.Set("phone1", preset.Content.Phone1)
		rec.Set("phone2", preset.Content.Phone2)
		rec.Set("work_hours", preset.Content.WorkHours)
		rec.Set("lunch_break", preset.Content.LunchBreak)
		rec.Set("days_off", preset.Content.DaysOff)
		rec.Set("address", preset.Content.Address)
		rec.Set("address_sub", preset.Content.AddressSub)
		rec.Set("map_iframe", preset.Content.MapIframe)

		if err := app.Save(rec); err != nil {
			log.Printf("[Preset Error] Failed to save site_content: %v\n", err)
		}
	}

	// 3. Update services if provided
	if len(preset.Services) > 0 {
		servicesCol, err := app.FindCollectionByNameOrId("services")
		if err == nil {
			// Clear existing services
			existing, _ := app.FindRecordsByFilter("services", "", "", 0, 0)
			for _, r := range existing {
				_ = app.Delete(r)
			}
			for _, s := range preset.Services {
				rec := core.NewRecord(servicesCol)
				rec.Set("title", s.Title)
				rec.Set("description", s.Description)
				rec.Set("price", s.Price)
				rec.Set("order", s.Order)
				rec.Set("active", true)
				_ = app.Save(rec)
			}
		}
	}

	cache.GetStore().InvalidateAll()
	log.Printf("[Preset] Successfully applied preset from %s\n", presetPath)
	return nil
}
