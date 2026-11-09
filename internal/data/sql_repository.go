package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"sitecraft/internal/cache"
	"sitecraft/internal/db"
	"sitecraft/internal/models"

	"github.com/google/uuid"
)

// Repository manages database queries and in-memory cache synchronization.
type Repository struct {
	queries *db.Queries
	rawDB   *sql.DB
	cache   *cache.Store
}

var globalRepo *Repository

// SetGlobalRepository configures the singleton repository used by package functions.
func SetGlobalRepository(repo *Repository) {
	globalRepo = repo
}

// GetGlobalRepository returns the current configured repository.
func GetGlobalRepository() *Repository {
	return globalRepo
}

// NewRepository creates a new SQL repository instance.
func NewRepository(database *sql.DB, queries *db.Queries) *Repository {
	return &Repository{
		queries: queries,
		rawDB:   database,
		cache:   cache.GetStore(),
	}
}

// WithCache overrides the cache store for this repository instance.
func (r *Repository) WithCache(c *cache.Store) *Repository {
	r.cache = c
	return r
}

// Cache returns the configured cache store or falls back to default.
func (r *Repository) Cache() *cache.Store {
	if r.cache != nil {
		return r.cache
	}
	return cache.GetStore()
}

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
	}

	c.SetSettings(settings)
	return settings, nil
}

// UpsertSiteSettings persists site settings into database and invalidates cache.
func (r *Repository) UpsertSiteSettings(ctx context.Context, s models.SiteSettings) error {
	if s.ID == "" {
		s.ID = "default_settings"
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

// GetServices returns active services.
func (r *Repository) GetServices(ctx context.Context) ([]models.ServiceItem, error) {
	c := r.Cache()
	if cached, ok := c.GetServices(); ok {
		return cached, nil
	}

	rows, err := r.queries.ListActiveServices(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]models.ServiceItem, len(rows))
	for i, row := range rows {
		items[i] = models.ServiceItem{
			ID:          row.ID,
			Title:       row.Title,
			Description: row.Description,
			Price:       row.Price,
			Order:       int(row.OrderIndex),
			Active:      row.Active == 1,
			StockCount:  int(row.StockCount),
			IsPreorder:  row.IsPreorder == 1,
			IsDeleted:   row.IsDeleted == 1,
		}
	}
	c.SetServices(items)
	return items, nil
}

// GetPageBlocks returns active page blocks ordered by index.
func (r *Repository) GetPageBlocks(ctx context.Context) ([]models.PageBlock, error) {
	c := r.Cache()
	if cached, ok := c.GetBlocks(); ok {
		return cached, nil
	}

	rows, err := r.queries.ListPageBlocks(ctx, "")
	if err != nil {
		return nil, err
	}

	blocks := make([]models.PageBlock, len(rows))
	for i, row := range rows {
		var payload map[string]interface{}
		if row.Payload != "" {
			_ = json.Unmarshal([]byte(row.Payload), &payload)
		}
		blocks[i] = models.PageBlock{
			ID:      row.ID,
			Type:    row.Type,
			Title:   row.Title,
			Content: row.Content,
			Order:   int(row.OrderIndex),
			Active:  row.Active == 1,
			Payload: payload,
		}
	}
	c.SetBlocks(blocks)
	return blocks, nil
}

// GetLeads returns recent leads.
func (r *Repository) GetLeads(ctx context.Context) ([]models.Lead, error) {
	rows, err := r.queries.ListLeads(ctx)
	if err != nil {
		return nil, err
	}

	leads := make([]models.Lead, len(rows))
	for i, row := range rows {
		leads[i] = models.Lead{
			ID:          row.ID,
			Name:        row.Name,
			Phone:       row.Phone,
			Email:       row.Email,
			Source:      row.Source,
			BookingDate: row.BookingDate,
			BookingTime: row.BookingTime,
			ResourceID:  row.ResourceID,
			Status:      row.Status,
			Created:     row.CreatedAt,
		}
	}
	return leads, nil
}

// CreateLead inserts a new lead into database.
func (r *Repository) CreateLead(ctx context.Context, lead models.Lead) (models.Lead, error) {
	if lead.ID == "" {
		lead.ID = uuid.New().String()
	}
	if lead.Status == "" {
		lead.Status = "новая"
	}
	row, err := r.queries.CreateLead(ctx, db.CreateLeadParams{
		ID:          lead.ID,
		Name:        lead.Name,
		Phone:       lead.Phone,
		Email:       lead.Email,
		Source:      lead.Source,
		BookingDate: lead.BookingDate,
		BookingTime: lead.BookingTime,
		ResourceID:  lead.ResourceID,
		Status:      lead.Status,
		Payload:     "{}",
	})
	if err != nil {
		return lead, err
	}
	lead.Created = row.CreatedAt
	return lead, nil
}

// GetArticles returns published articles.
func (r *Repository) GetArticles(ctx context.Context) ([]models.ArticleItem, error) {
	c := r.Cache()
	if cached, ok := c.GetArticles(); ok {
		return cached, nil
	}

	rows, err := r.queries.ListPublishedArticles(ctx)
	if err != nil {
		return nil, err
	}

	articles := make([]models.ArticleItem, len(rows))
	for i, row := range rows {
		articles[i] = models.ArticleItem{
			ID:             row.ID,
			Title:          row.Title,
			Slug:           row.Slug,
			Excerpt:        row.Excerpt,
			Content:        row.Content,
			CoverImage:     row.CoverImage,
			AuthorName:     row.AuthorName,
			PublishedAt:    row.PublishedAt,
			IsPublished:    row.IsPublished == 1,
			ViewsCount:     int(row.ViewsCount),
			SEOTitle:       row.SeoTitle,
			SEODescription: row.SeoDescription,
			Tags:           strings.Split(row.Tags, ","),
		}
	}
	c.SetArticles(articles)
	return articles, nil
}

// GetReviews returns published reviews.
func (r *Repository) GetReviews(ctx context.Context) ([]models.ReviewItem, error) {
	rows, err := r.queries.ListPublishedReviews(ctx)
	if err != nil {
		return nil, err
	}

	reviews := make([]models.ReviewItem, len(rows))
	for i, row := range rows {
		var photos []string
		if row.Photo != "" {
			photos = []string{row.Photo}
		}
		reviews[i] = models.ReviewItem{
			ID:          row.ID,
			AuthorName:  row.AuthorName,
			Rating:      int(row.Rating),
			Comment:     row.Comment,
			OrderID:     row.OrderID,
			Photos:      photos,
			IsPublished: row.IsPublished == 1,
			Source:      row.Source,
			CreatedAt:   row.CreatedAt,
		}
	}
	return reviews, nil
}

// GetResources returns active resources / specialists.
func (r *Repository) GetResources(ctx context.Context) ([]models.ResourceItem, error) {
	rows, err := r.queries.ListActiveResources(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]models.ResourceItem, len(rows))
	for i, row := range rows {
		items[i] = models.ResourceItem{
			ID:     row.ID,
			Name:   row.Name,
			Role:   row.Role,
			Avatar: row.Avatar,
			Active: row.Active == 1,
			Order:  int(row.OrderIndex),
		}
	}
	return items, nil
}

// GetBranches returns active company branches.
func (r *Repository) GetBranches(ctx context.Context) ([]models.BranchItem, error) {
	rows, err := r.queries.ListActiveBranches(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]models.BranchItem, len(rows))
	for i, row := range rows {
		items[i] = models.BranchItem{
			ID:        row.ID,
			Name:      row.Name,
			City:      row.City,
			Address:   row.Address,
			Phone:     row.Phone,
			WorkHours: row.WorkHours,
			Lat:       row.Lat,
			Lng:       row.Lng,
			Metro:     row.Metro,
			IsMain:    row.IsMain == 1,
			Active:    row.Active == 1,
		}
	}
	return items, nil
}

// GetVacancies returns active vacancies.
func (r *Repository) GetVacancies(ctx context.Context) ([]models.VacancyItem, error) {
	rows, err := r.queries.ListActiveVacancies(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]models.VacancyItem, len(rows))
	for i, row := range rows {
		var resp, req, ben []string
		_ = json.Unmarshal([]byte(row.Responsibilities), &resp)
		_ = json.Unmarshal([]byte(row.Requirements), &req)
		_ = json.Unmarshal([]byte(row.Benefits), &ben)

		items[i] = models.VacancyItem{
			ID:               row.ID,
			Title:            row.Title,
			Department:       row.Department,
			Salary:           row.Salary,
			Experience:       row.Experience,
			Schedule:         row.Schedule,
			Responsibilities: resp,
			Requirements:     req,
			Benefits:         ben,
			Active:           row.Active == 1,
			Order:            int(row.OrderIndex),
		}
	}
	return items, nil
}

// GetBookedSlots returns list of booked time slots today.
func (r *Repository) GetBookedSlots(ctx context.Context, resourceID ...string) ([]string, error) {
	today := time.Now().Format("2006-01-02")
	leads, err := r.queries.ListLeads(ctx)
	if err != nil {
		return nil, err
	}

	var resID string
	if len(resourceID) > 0 {
		resID = resourceID[0]
	}

	var slots []string
	for _, l := range leads {
		if l.BookingDate == today && l.BookingTime != "" && l.Status != "отменена" {
			if resID == "" || l.ResourceID == resID {
				slots = append(slots, l.BookingTime)
			}
		}
	}
	return slots, nil
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

// --- Services Operations ---

// GetServiceByID returns a single service by ID.
func (r *Repository) GetServiceByID(ctx context.Context, id string) (models.ServiceItem, error) {
	row, err := r.queries.GetServiceByID(ctx, id)
	if err != nil {
		return models.ServiceItem{}, err
	}
	return models.ServiceItem{
		ID:          row.ID,
		Title:       row.Title,
		Description: row.Description,
		Price:       row.Price,
		Order:       int(row.OrderIndex),
		Active:      row.Active == 1,
		StockCount:  int(row.StockCount),
		IsPreorder:  row.IsPreorder == 1,
		IsDeleted:   row.IsDeleted == 1,
	}, nil
}

// SaveService creates or updates a service and invalidates services cache.
func (r *Repository) SaveService(ctx context.Context, item models.ServiceItem) error {
	actInt := int64(0)
	if item.Active {
		actInt = 1
	}
	stock := int64(item.StockCount)
	if stock == 0 {
		stock = -1
	}
	preorder := int64(0)
	if item.IsPreorder {
		preorder = 1
	}

	var err error
	if item.ID != "" {
		_, err = r.queries.UpdateService(ctx, db.UpdateServiceParams{
			ID:          item.ID,
			Title:       item.Title,
			Description: item.Description,
			Price:       item.Price,
			OrderIndex:  int64(item.Order),
			Active:      actInt,
			StockCount:  stock,
			IsPreorder:  preorder,
		})
	} else {
		_, err = r.queries.CreateService(ctx, db.CreateServiceParams{
			ID:          uuid.New().String(),
			Title:       item.Title,
			Description: item.Description,
			Price:       item.Price,
			OrderIndex:  int64(item.Order),
			Active:      actInt,
			StockCount:  stock,
			IsPreorder:  preorder,
		})
	}
	if err != nil {
		return err
	}
	r.Cache().InvalidateServices()
	return nil
}

// ToggleService toggles the active state of a service and invalidates services cache.
func (r *Repository) ToggleService(ctx context.Context, id string) error {
	svc, err := r.queries.GetServiceByID(ctx, id)
	if err != nil {
		return err
	}
	newActive := int64(1)
	if svc.Active == 1 {
		newActive = 0
	}
	_, err = r.queries.UpdateService(ctx, db.UpdateServiceParams{
		ID:          svc.ID,
		Title:       svc.Title,
		Description: svc.Description,
		Price:       svc.Price,
		OrderIndex:  svc.OrderIndex,
		Active:      newActive,
		StockCount:  svc.StockCount,
		IsPreorder:  svc.IsPreorder,
	})
	if err != nil {
		return err
	}
	r.Cache().InvalidateServices()
	return nil
}

// SoftDeleteService marks a service as deleted and invalidates services cache.
func (r *Repository) SoftDeleteService(ctx context.Context, id string) error {
	if err := r.queries.SoftDeleteService(ctx, id); err != nil {
		return err
	}
	r.Cache().InvalidateServices()
	return nil
}

// --- Page Block Operations ---

// GetPageBlockByID retrieves a page block by ID.
func (r *Repository) GetPageBlockByID(ctx context.Context, id string) (models.PageBlock, error) {
	row, err := r.queries.GetPageBlockByID(ctx, id)
	if err != nil {
		return models.PageBlock{}, err
	}
	var payload map[string]interface{}
	if row.Payload != "" {
		_ = json.Unmarshal([]byte(row.Payload), &payload)
	}
	return models.PageBlock{
		ID:      row.ID,
		Type:    row.Type,
		Title:   row.Title,
		Content: row.Content,
		Order:   int(row.OrderIndex),
		Active:  row.Active == 1,
		Payload: payload,
	}, nil
}

// CreatePageBlock inserts a page block and invalidates blocks cache.
func (r *Repository) CreatePageBlock(ctx context.Context, block models.PageBlock) (models.PageBlock, error) {
	if block.ID == "" {
		block.ID = uuid.New().String()
	}
	payloadBytes, _ := json.Marshal(block.Payload)
	act := int64(1)
	if !block.Active {
		act = 0
	}
	_, err := r.queries.CreatePageBlock(ctx, db.CreatePageBlockParams{
		ID:         block.ID,
		PageID:     "",
		Type:       block.Type,
		Title:      block.Title,
		Content:    block.Content,
		OrderIndex: int64(block.Order),
		Active:     act,
		Payload:    string(payloadBytes),
	})
	if err != nil {
		return block, err
	}
	r.Cache().InvalidateBlocks()
	return block, nil
}

// TogglePageBlock toggles block visibility and invalidates blocks cache.
func (r *Repository) TogglePageBlock(ctx context.Context, id string) error {
	block, err := r.queries.GetPageBlockByID(ctx, id)
	if err != nil {
		return err
	}
	newActive := int64(1)
	if block.Active == 1 {
		newActive = 0
	}
	_, err = r.queries.UpdatePageBlock(ctx, db.UpdatePageBlockParams{
		ID:         block.ID,
		Type:       block.Type,
		Title:      block.Title,
		Content:    block.Content,
		OrderIndex: block.OrderIndex,
		Active:     newActive,
		Payload:    block.Payload,
	})
	if err != nil {
		return err
	}
	r.Cache().InvalidateBlocks()
	return nil
}

// SwapPageBlockOrder swaps order of two blocks atomically inside a transaction and invalidates cache.
func (r *Repository) SwapPageBlockOrder(ctx context.Context, id1 string, order1 int64, id2 string, order2 int64) error {
	tx, err := r.rawDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	if err := qtx.UpdatePageBlockOrder(ctx, db.UpdatePageBlockOrderParams{
		ID:         id1,
		OrderIndex: order2,
	}); err != nil {
		return fmt.Errorf("failed to update block order %s: %w", id1, err)
	}
	if err := qtx.UpdatePageBlockOrder(ctx, db.UpdatePageBlockOrderParams{
		ID:         id2,
		OrderIndex: order1,
	}); err != nil {
		return fmt.Errorf("failed to update block order %s: %w", id2, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.Cache().InvalidateBlocks()
	return nil
}

// DeletePageBlock deletes a block by ID and invalidates cache.
func (r *Repository) DeletePageBlock(ctx context.Context, id string) error {
	if err := r.queries.DeletePageBlock(ctx, id); err != nil {
		return err
	}
	r.Cache().InvalidateBlocks()
	return nil
}

// --- Article Operations ---

// GetArticleByID returns an article by ID.
func (r *Repository) GetArticleByID(ctx context.Context, id string) (models.ArticleItem, error) {
	art, err := r.queries.GetArticleByID(ctx, id)
	if err != nil {
		return models.ArticleItem{}, err
	}
	return models.ArticleItem{
		ID:             art.ID,
		Title:          art.Title,
		Slug:           art.Slug,
		Excerpt:        art.Excerpt,
		Content:        art.Content,
		CoverImage:     art.CoverImage,
		AuthorName:     art.AuthorName,
		PublishedAt:    art.PublishedAt,
		IsPublished:    art.IsPublished == 1,
		ViewsCount:     int(art.ViewsCount),
		SEOTitle:       art.SeoTitle,
		SEODescription: art.SeoDescription,
		Tags:           strings.Split(art.Tags, ","),
	}, nil
}

// SaveArticle creates or updates an article and invalidates articles cache.
func (r *Repository) SaveArticle(ctx context.Context, item models.ArticleItem) error {
	pubInt := int64(0)
	if item.IsPublished {
		pubInt = 1
	}
	tags := strings.Join(item.Tags, ",")

	var err error
	if item.ID != "" {
		_, err = r.queries.UpdateArticle(ctx, db.UpdateArticleParams{
			ID:             item.ID,
			Title:          item.Title,
			Slug:           item.Slug,
			Excerpt:        item.Excerpt,
			Content:        item.Content,
			CoverImage:     item.CoverImage,
			AuthorName:     item.AuthorName,
			PublishedAt:    item.PublishedAt,
			IsPublished:    pubInt,
			SeoTitle:       item.SEOTitle,
			SeoDescription: item.SEODescription,
			Tags:           tags,
		})
	} else {
		_, err = r.queries.CreateArticle(ctx, db.CreateArticleParams{
			ID:             uuid.New().String(),
			Title:          item.Title,
			Slug:           item.Slug,
			Excerpt:        item.Excerpt,
			Content:        item.Content,
			CoverImage:     item.CoverImage,
			AuthorName:     item.AuthorName,
			PublishedAt:    item.PublishedAt,
			IsPublished:    pubInt,
			SeoTitle:       item.SEOTitle,
			SeoDescription: item.SEODescription,
			Tags:           tags,
		})
	}
	if err != nil {
		return err
	}
	r.Cache().InvalidateArticles()
	return nil
}

// ToggleArticle toggles publication status of an article and invalidates articles cache.
func (r *Repository) ToggleArticle(ctx context.Context, id string) error {
	art, err := r.queries.GetArticleByID(ctx, id)
	if err != nil {
		return err
	}
	newPub := int64(1)
	if art.IsPublished == 1 {
		newPub = 0
	}
	_, err = r.queries.UpdateArticle(ctx, db.UpdateArticleParams{
		ID:             art.ID,
		Title:          art.Title,
		Slug:           art.Slug,
		Excerpt:        art.Excerpt,
		Content:        art.Content,
		CoverImage:     art.CoverImage,
		AuthorName:     art.AuthorName,
		PublishedAt:    art.PublishedAt,
		IsPublished:    newPub,
		SeoTitle:       art.SeoTitle,
		SeoDescription: art.SeoDescription,
		Tags:           art.Tags,
	})
	if err != nil {
		return err
	}
	r.Cache().InvalidateArticles()
	return nil
}

// SoftDeleteArticle soft deletes an article and invalidates articles cache.
func (r *Repository) SoftDeleteArticle(ctx context.Context, id string) error {
	if err := r.queries.SoftDeleteArticle(ctx, id); err != nil {
		return err
	}
	r.Cache().InvalidateArticles()
	return nil
}

// --- Leads and Reviews ---

// UpdateLeadStatus updates the status of a lead.
func (r *Repository) UpdateLeadStatus(ctx context.Context, id, status string) error {
	_, err := r.queries.UpdateLeadStatus(ctx, db.UpdateLeadStatusParams{
		ID:     id,
		Status: status,
	})
	return err
}

// ApproveReview marks a review as approved.
func (r *Repository) ApproveReview(ctx context.Context, id string) error {
	_, err := r.queries.ApproveReview(ctx, id)
	return err
}

// DeleteReview deletes a review by ID.
func (r *Repository) DeleteReview(ctx context.Context, id string) error {
	_, err := r.rawDB.ExecContext(ctx, "DELETE FROM reviews WHERE id = ?", id)
	return err
}

