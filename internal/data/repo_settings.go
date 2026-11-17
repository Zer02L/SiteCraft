package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"sitecraft/internal/db"
	"sitecraft/internal/models"

	"github.com/google/uuid"
)

// GetSiteSettings retrieves settings from cache or database.
func (r *Repository) GetSiteSettings(ctx context.Context) (models.SiteSettings, error) {
	c := r.Cache()
	if cached, ok := c.GetSettings(); ok {
		return cached, nil
	}

	row, err := r.queries.GetSiteSettings(ctx)
	if err == sql.ErrNoRows {
		// Return defaults
		def := models.SiteSettings{
			ID:                  uuid.New().String(),
			SiteName:            "SiteCraft",
			Tagline:             "Современный конструктор сайтов",
			Theme:               "main",
			EnableA11y:          true,
			EnableI18n:          true,
			DefaultLang:         "ru",
			EnableCareers:       true,
			EnableCookieConsent: true,
			HeaderVariant:       "classic",
			FooterVariant:       "classic",
		}
		return def, nil
	}
	if err != nil {
		return models.SiteSettings{}, err
	}

	settings := models.SiteSettings{
		ID:                      row.ID,
		SiteName:                row.SiteName,
		Tagline:                 row.Tagline,
		Theme:                   row.Theme,
		PrimaryColor:            row.PrimaryColor,
		AccentColor:             row.AccentColor,
		BackgroundColor:         row.BackgroundColor,
		FontFamily:              row.FontFamily,
		EnableA11y:              row.EnableA11y == 1,
		EnableAlpine:            row.EnableAlpine == 1,
		EnableTurnstile:         row.EnableTurnstile == 1,
		TurnstileSiteKey:        row.TurnstileSiteKey,
		TurnstileSecretKey:      row.TurnstileSecretKey,
		TelegramBotToken:        row.TelegramBotToken,
		TelegramChatID:          row.TelegramChatID,
		WebhookURL:              row.WebhookUrl,
		YandexMetrikaID:         row.YandexMetrikaID,
		MapMode:                 row.MapMode,
		BookingMode:             row.BookingMode,
		EnablePayment:           row.EnablePayment == 1,
		PaymentShopID:           row.PaymentShopID,
		PaymentSecret:           row.PaymentSecret,
		PaymentSandbox:          row.PaymentSandbox == 1,
		EnableBlog:              row.EnableBlog == 1,
		EnableTMA:               row.EnableTma == 1,
		EnableTracker:           row.EnableTracker == 1,
		EnableI18n:              row.EnableI18n == 1,
		DefaultLang:             row.DefaultLang,
		EnableFloatingContact:   row.EnableFloatingContact == 1,
		WhatsappPhone:           row.WhatsappPhone,
		TelegramContact:         row.TelegramContact,
		EnableCart:              row.EnableCart == 1,
		HeaderVariant:           row.HeaderVariant,
		FooterVariant:           row.FooterVariant,
		EnableLiveChat:          row.EnableLiveChat == 1,
		EnableFleet:             row.EnableFleet == 1,
		EnableRecentActivity:    row.EnableRecentActivity == 1,
		ClientAdminPassword:     row.ClientAdminPassword,
		ClientAdminPasswordHash: row.ClientAdminPasswordHash,
		LegalEntity:             row.LegalEntity,
		LegalAddress:            row.LegalAddress,
		INN:                     row.Inn,
		KPP:                     row.Kpp,
		OGRN:                    row.Ogrn,
		BankName:                row.BankName,
		BankBIK:                 row.BankBik,
		BankAccount:             row.BankAccount,
		BankCorr:                row.BankCorr,
		ContactEmail:            row.ContactEmail,
		EnableCareers:           row.EnableCareers == 1,
		EnableCookieConsent:     row.EnableCookieConsent == 1,
		AIProvider:              row.AiProvider,
		AIBaseURL:               row.AiBaseUrl,
		AIAPIKeyEncrypted:       row.AiApiKeyEncrypted,
		AIModelName:             row.AiModelName,
		AIProxyURL:              row.AiProxyUrl,
		AIFolderID:              row.AiFolderID,
	}

	c.SetSettings(settings)
	return settings, nil
}

// UpsertSiteSettings persists site settings into database and invalidates cache.
func (r *Repository) UpsertSiteSettings(ctx context.Context, s models.SiteSettings) error {
	if s.ID == "" {
		s.ID = "default_settings"
	}
	if s.AIProvider == "" {
		s.AIProvider = "openai"
	}
	toInt := func(b bool) int64 {
		if b {
			return 1
		}
		return 0
	}

	_, err := r.queries.UpsertSiteSettings(ctx, db.UpsertSiteSettingsParams{
		ID:                      s.ID,
		SiteName:                s.SiteName,
		Tagline:                 s.Tagline,
		Theme:                   s.Theme,
		PrimaryColor:            s.PrimaryColor,
		AccentColor:             s.AccentColor,
		BackgroundColor:         s.BackgroundColor,
		FontFamily:              s.FontFamily,
		EnableA11y:              toInt(s.EnableA11y),
		EnableAlpine:            toInt(s.EnableAlpine),
		EnableTurnstile:         toInt(s.EnableTurnstile),
		TurnstileSiteKey:        s.TurnstileSiteKey,
		TurnstileSecretKey:      s.TurnstileSecretKey,
		TelegramBotToken:        s.TelegramBotToken,
		TelegramChatID:          s.TelegramChatID,
		WebhookUrl:              s.WebhookURL,
		YandexMetrikaID:         s.YandexMetrikaID,
		MapMode:                 s.MapMode,
		BookingMode:             s.BookingMode,
		EnablePayment:           toInt(s.EnablePayment),
		PaymentShopID:           s.PaymentShopID,
		PaymentSecret:           s.PaymentSecret,
		PaymentSandbox:          toInt(s.PaymentSandbox),
		EnableBlog:              toInt(s.EnableBlog),
		EnableTma:               toInt(s.EnableTMA),
		EnableTracker:           toInt(s.EnableTracker),
		EnableI18n:              toInt(s.EnableI18n),
		DefaultLang:             s.DefaultLang,
		EnableFloatingContact:   toInt(s.EnableFloatingContact),
		WhatsappPhone:           s.WhatsappPhone,
		TelegramContact:         s.TelegramContact,
		EnableCart:              toInt(s.EnableCart),
		HeaderVariant:           s.HeaderVariant,
		FooterVariant:           s.FooterVariant,
		EnableLiveChat:          toInt(s.EnableLiveChat),
		EnableFleet:             toInt(s.EnableFleet),
		EnableRecentActivity:    toInt(s.EnableRecentActivity),
		ClientAdminPassword:     s.ClientAdminPassword,
		ClientAdminPasswordHash: s.ClientAdminPasswordHash,
		LegalEntity:             s.LegalEntity,
		LegalAddress:            s.LegalAddress,
		Inn:                     s.INN,
		Kpp:                     s.KPP,
		Ogrn:                    s.OGRN,
		BankName:                s.BankName,
		BankBik:                 s.BankBIK,
		BankAccount:             s.BankAccount,
		BankCorr:                s.BankCorr,
		ContactEmail:            s.ContactEmail,
		EnableCareers:           toInt(s.EnableCareers),
		EnableCookieConsent:     toInt(s.EnableCookieConsent),
		AiProvider:              s.AIProvider,
		AiBaseUrl:               s.AIBaseURL,
		AiApiKeyEncrypted:       s.AIAPIKeyEncrypted,
		AiModelName:             s.AIModelName,
		AiProxyUrl:              s.AIProxyURL,
		AiFolderID:              s.AIFolderID,
	})
	if err != nil {
		return err
	}
	r.Cache().InvalidateSettings()
	return nil
}

// GetSiteContent retrieves content from cache or database.
func (r *Repository) GetSiteContent(ctx context.Context) (models.SiteContent, error) {
	c := r.Cache()
	if cached, ok := c.GetContent(); ok {
		return cached, nil
	}

	row, err := r.queries.GetSiteContent(ctx)
	if err == sql.ErrNoRows {
		return models.SiteContent{
			ID:           uuid.New().String(),
			HeroTitle:    "SiteCraft",
			HeroSubtitle: "Профессиональный сайт под ключ",
		}, nil
	}
	if err != nil {
		return models.SiteContent{}, err
	}

	content := models.SiteContent{
		ID:            row.ID,
		HeroBadge:     row.HeroBadge,
		HeroTitle:     row.HeroTitle,
		HeroSubtitle:  row.HeroSubtitle,
		HeroTitleB:    row.HeroTitleB,
		HeroSubtitleB: row.HeroSubtitleB,
		HeroCTAB:      row.HeroCtaB,
		OwnerName:     row.OwnerName,
		OwnerTitle:    row.OwnerTitle,
		NotaryName:    row.OwnerName,
		NotaryTitle:   row.OwnerTitle,
		Phone1:        row.Phone1,
		Phone2:        row.Phone2,
		WorkHours:     row.WorkHours,
		LunchBreak:    row.LunchBreak,
		DaysOff:       row.DaysOff,
		Address:       row.Address,
		AddressSub:    row.AddressSub,
		MapIframe:     row.MapIframe,
		HeroPhoto:     row.Photo,
	}

	c.SetContent(content)
	return content, nil
}

// UpsertSiteContent persists site content into database and invalidates cache.
func (r *Repository) UpsertSiteContent(ctx context.Context, c models.SiteContent) error {
	if c.ID == "" {
		c.ID = "default_content"
	}
	_, err := r.queries.UpsertSiteContent(ctx, db.UpsertSiteContentParams{
		ID:            c.ID,
		HeroBadge:     c.HeroBadge,
		HeroTitle:     c.HeroTitle,
		HeroSubtitle:  c.HeroSubtitle,
		HeroTitleB:    c.HeroTitleB,
		HeroSubtitleB: c.HeroSubtitleB,
		HeroCtaB:      c.HeroCTAB,
		OwnerName:     c.OwnerName,
		OwnerTitle:    c.OwnerTitle,
		Phone1:        c.Phone1,
		Phone2:        c.Phone2,
		WorkHours:     c.WorkHours,
		LunchBreak:    c.LunchBreak,
		DaysOff:       c.DaysOff,
		Address:       c.Address,
		AddressSub:    c.AddressSub,
		MapIframe:     c.MapIframe,
		Photo:         c.HeroPhoto,
	})
	if err != nil {
		return err
	}
	r.Cache().InvalidateContent()
	return nil
}

// ApplyPreset loads a JSON preset and applies it to SQL tables.
func (r *Repository) ApplyPreset(ctx context.Context, presetPath string) error {
	resolvedPath := resolvePresetPath(presetPath)
	dataBytes, err := os.ReadFile(resolvedPath)
	if err != nil {
		return fmt.Errorf("failed to read preset file %s: %w", presetPath, err)
	}

	var preset PresetData
	if err := json.Unmarshal(dataBytes, &preset); err != nil {
		return fmt.Errorf("failed to parse preset JSON: %w", err)
	}

	// 1. Settings
	if err := r.UpsertSiteSettings(ctx, preset.Settings); err != nil {
		slog.Error("failed to apply preset settings", "error", err)
	}

	// 2. Content
	if err := r.UpsertSiteContent(ctx, preset.Content); err != nil {
		slog.Error("failed to apply preset content", "error", err)
	}

	// 3. Services
	if len(preset.Services) > 0 {
		_ = r.queries.DeleteAllServices(ctx)
		for _, s := range preset.Services {
			isPre := int64(0)
			if s.IsPreorder {
				isPre = 1
			}
			_, _ = r.queries.CreateService(ctx, db.CreateServiceParams{
				ID:          uuid.New().String(),
				Title:       s.Title,
				Description: s.Description,
				Price:       s.Price,
				OrderIndex:  int64(s.Order),
				Active:      1,
				StockCount:  int64(s.StockCount),
				IsPreorder:  isPre,
			})
		}
		r.Cache().InvalidateServices()
	}

	// 4. Page Blocks
	if len(preset.Blocks) > 0 {
		_ = r.queries.DeleteAllPageBlocks(ctx)
		for _, b := range preset.Blocks {
			payloadBytes, _ := json.Marshal(b.Payload)
			act := int64(1)
			if !b.Active {
				act = 0
			}
			_, _ = r.queries.CreatePageBlock(ctx, db.CreatePageBlockParams{
				ID:         uuid.New().String(),
				PageID:     "",
				Type:       b.Type,
				Title:      b.Title,
				Content:    b.Content,
				OrderIndex: int64(b.Order),
				Active:     act,
				Payload:    string(payloadBytes),
			})
		}
		r.Cache().InvalidateBlocks()
	}

	slog.Info("Successfully applied preset to SQL database", "preset", presetPath)
	return nil
}
