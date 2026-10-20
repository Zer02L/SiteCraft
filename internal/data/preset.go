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
	Blocks   []models.PageBlock   `json:"blocks"`
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
		if preset.Settings.PrimaryColor != "" {
			rec.Set("primary_color", preset.Settings.PrimaryColor)
		}
		if preset.Settings.AccentColor != "" {
			rec.Set("accent_color", preset.Settings.AccentColor)
		}
		if preset.Settings.BackgroundColor != "" {
			rec.Set("background_color", preset.Settings.BackgroundColor)
		}
		rec.Set("enable_a11y", preset.Settings.EnableA11y)
		rec.Set("enable_alpine", preset.Settings.EnableAlpine)
		rec.Set("enable_tracker", preset.Settings.EnableTracker)
		rec.Set("enable_blog", preset.Settings.EnableBlog)
		rec.Set("enable_tma", preset.Settings.EnableTMA)
		rec.Set("enable_payment", preset.Settings.EnablePayment)
		rec.Set("payment_sandbox", preset.Settings.PaymentSandbox)
		rec.Set("enable_i18n", preset.Settings.EnableI18n)
		rec.Set("enable_cart", preset.Settings.EnableCart)
		rec.Set("enable_floating_contact", preset.Settings.EnableFloatingContact)
		rec.Set("enable_live_chat", preset.Settings.EnableLiveChat)
		rec.Set("enable_fleet", preset.Settings.EnableFleet)
		rec.Set("enable_recent_activity", preset.Settings.EnableRecentActivity)
		rec.Set("enable_cookie_consent", preset.Settings.EnableCookieConsent)
		rec.Set("enable_careers", preset.Settings.EnableCareers)

		if preset.Settings.WhatsappPhone != "" {
			rec.Set("whatsapp_phone", preset.Settings.WhatsappPhone)
		}
		if preset.Settings.TelegramContact != "" {
			rec.Set("telegram_contact", preset.Settings.TelegramContact)
		}
		if preset.Settings.DefaultLang != "" {
			rec.Set("default_lang", preset.Settings.DefaultLang)
		}
		if preset.Settings.MapMode != "" {
			rec.Set("map_mode", preset.Settings.MapMode)
		}
		if preset.Settings.BookingMode != "" {
			rec.Set("booking_mode", preset.Settings.BookingMode)
		}
		if preset.Settings.HeaderVariant != "" {
			rec.Set("header_variant", preset.Settings.HeaderVariant)
		}
		if preset.Settings.FooterVariant != "" {
			rec.Set("footer_variant", preset.Settings.FooterVariant)
		}
		if preset.Settings.LegalEntity != "" {
			rec.Set("legal_entity", preset.Settings.LegalEntity)
		}
		if preset.Settings.LegalAddress != "" {
			rec.Set("legal_address", preset.Settings.LegalAddress)
		}
		if preset.Settings.INN != "" {
			rec.Set("inn", preset.Settings.INN)
		}
		if preset.Settings.KPP != "" {
			rec.Set("kpp", preset.Settings.KPP)
		}
		if preset.Settings.OGRN != "" {
			rec.Set("ogrn", preset.Settings.OGRN)
		}
		if preset.Settings.BankName != "" {
			rec.Set("bank_name", preset.Settings.BankName)
		}
		if preset.Settings.BankBIK != "" {
			rec.Set("bank_bik", preset.Settings.BankBIK)
		}
		if preset.Settings.BankAccount != "" {
			rec.Set("bank_account", preset.Settings.BankAccount)
		}
		if preset.Settings.BankCorr != "" {
			rec.Set("bank_corr", preset.Settings.BankCorr)
		}
		if preset.Settings.ContactEmail != "" {
			rec.Set("contact_email", preset.Settings.ContactEmail)
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
		if preset.Content.HeroTitleB != "" {
			rec.Set("hero_title_b", preset.Content.HeroTitleB)
		}
		if preset.Content.HeroSubtitleB != "" {
			rec.Set("hero_subtitle_b", preset.Content.HeroSubtitleB)
		}
		if preset.Content.HeroCTAB != "" {
			rec.Set("hero_cta_b", preset.Content.HeroCTAB)
		}
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
				rec.Set("stock_count", s.StockCount)
				rec.Set("is_preorder", s.IsPreorder)
				_ = app.Save(rec)
			}
		}
	}

	// 4. Update page blocks if provided (Structural variations between presets)
	if len(preset.Blocks) > 0 {
		blocksCol, err := app.FindCollectionByNameOrId("page_blocks")
		if err == nil {
			existing, _ := app.FindRecordsByFilter("page_blocks", "", "", 0, 0)
			for _, r := range existing {
				_ = app.Delete(r)
			}
			for _, b := range preset.Blocks {
				rec := core.NewRecord(blocksCol)
				rec.Set("type", b.Type)
				rec.Set("title", b.Title)
				rec.Set("content", b.Content)
				rec.Set("order", b.Order)
				rec.Set("active", true)
				if b.Payload != nil {
					rec.Set("payload", b.Payload)
				}
				if err := app.Save(rec); err != nil {
					log.Printf("[Preset Error] Failed to save page_block (%s): %v\n", b.Type, err)
				}
			}
		}
	}

	cache.GetStore().InvalidateAll()
	log.Printf("[Preset] Successfully applied preset from %s\n", presetPath)
	return nil
}
