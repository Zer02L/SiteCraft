package data

import (
	"encoding/json"
	"fmt"
	"strings"

	"sitecraft/internal/cache"
	"sitecraft/internal/models"

	"github.com/microcosm-cc/bluemonday"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

var htmlSanitizer = bluemonday.UGCPolicy()

// GetSiteSettings retrieves settings from In-Memory cache or PocketBase
func GetSiteSettings(app core.App) models.SiteSettings {
	c := cache.GetStore()
	if cached, ok := c.GetSettings(); ok {
		return cached
	}

	settings := models.SiteSettings{
		SiteName:       "Нотариальная контора",
		Tagline:        "Официальные нотариальные действия",
		Theme:          "main",
		EnableA11y:     true,
		EnableAlpine:   false,
		MapMode:        "lazy",
		BookingMode:    "slots",
		EnablePayment:  false,
		PaymentSandbox: true,
		EnableI18n:     true,
		DefaultLang:    "ru",
	}

	rec, err := app.FindFirstRecordByFilter("site_settings", "")
	if err == nil && rec != nil {
		settings.ID = rec.Id
		if val := rec.GetString("site_name"); val != "" {
			settings.SiteName = val
		}
		if val := rec.GetString("tagline"); val != "" {
			settings.Tagline = val
		}
		if val := rec.GetString("theme"); val != "" {
			settings.Theme = val
		}
		settings.PrimaryColor = rec.GetString("primary_color")
		settings.AccentColor = rec.GetString("accent_color")
		settings.BackgroundColor = rec.GetString("background_color")
		settings.FontFamily = rec.GetString("font_family")
		settings.EnableA11y = rec.GetBool("enable_a11y")
		settings.EnableAlpine = rec.GetBool("enable_alpine")
		settings.EnableTurnstile = rec.GetBool("enable_turnstile")
		settings.TurnstileSiteKey = rec.GetString("turnstile_site_key")
		settings.TurnstileSecretKey = rec.GetString("turnstile_secret_key")
		settings.TelegramBotToken = rec.GetString("telegram_bot_token")
		settings.TelegramChatID = rec.GetString("telegram_chat_id")
		settings.WebhookURL = rec.GetString("webhook_url")
		settings.YandexMetrikaID = rec.GetString("yandex_metrika_id")
		if val := rec.GetString("map_mode"); val != "" {
			settings.MapMode = val
		}
		if val := rec.GetString("booking_mode"); val != "" {
			settings.BookingMode = val
		}
		settings.EnablePayment = rec.GetBool("enable_payment")
		settings.PaymentShopID = rec.GetString("payment_shop_id")
		settings.PaymentSecret = rec.GetString("payment_secret")
		settings.PaymentSandbox = rec.GetBool("payment_sandbox")
		settings.EnableBlog = rec.GetBool("enable_blog")
		settings.EnableTMA = rec.GetBool("enable_tma")
		settings.EnableTracker = rec.GetBool("enable_tracker")
		if rec.Collection().Fields.GetByName("enable_i18n") != nil {
			settings.EnableI18n = rec.GetBool("enable_i18n")
		} else {
			settings.EnableI18n = true
		}
		if defL := rec.GetString("default_lang"); defL != "" {
			settings.DefaultLang = defL
		} else {
			settings.DefaultLang = "ru"
		}
		if rec.Collection().Fields.GetByName("enable_floating_contact") != nil {
			settings.EnableFloatingContact = rec.GetBool("enable_floating_contact")
		} else {
			settings.EnableFloatingContact = true
		}
		settings.WhatsappPhone = rec.GetString("whatsapp_phone")
		settings.TelegramContact = rec.GetString("telegram_contact")
		settings.EnableCart = rec.GetBool("enable_cart")
		settings.HeaderVariant = rec.GetString("header_variant")
		settings.FooterVariant = rec.GetString("footer_variant")
	}

	c.SetSettings(settings)
	return settings
}

// GetSiteContent retrieves content from In-Memory cache or PocketBase
func GetSiteContent(app core.App) models.SiteContent {
	c := cache.GetStore()
	if cached, ok := c.GetContent(); ok {
		return cached
	}

	content := models.SiteContent{
		HeroBadge:    "Нотариальная контора открыта",
		HeroTitle:    "Нотариус",
		HeroSubtitle: "Совершение всех видов нотариальных действий в строгом соответствии с законодательством РФ. Законность, оперативность и правовая защита ваших интересов.",
		OwnerName:    "Захаров Владислав Борисович",
		OwnerTitle:   "Нотариус города Краснодара",
		NotaryName:   "Захаров Владислав Борисович",
		NotaryTitle:  "Нотариус города Краснодара",
		Phone1:       "8-918-024-64-20",
		Phone2:       "8 (861) 224-89-55",
		WorkHours:    "Понедельник — Пятница: 9:00 - 17:30",
		LunchBreak:   "13:00 - 14:00",
		DaysOff:      "Выходные: суббота, воскресенье",
		Address:      "Нотариальная контора нотариуса Захарова В. Б.",
		AddressSub:   "Краснодарский край / г. Краснодар (код 861)",
		MapIframe:    "https://yandex.ru/map-widget/v1/?ll=38.975313%2C45.035470&z=15&pt=38.975313,45.035470,pm2rdm",
		HeroPhoto:    "/assets/images/notary_sign.png",
	}

	rec, err := app.FindFirstRecordByFilter("site_content", "")
	if err == nil && rec != nil {
		content.ID = rec.Id
		if val := rec.GetString("hero_badge"); val != "" {
			content.HeroBadge = val
		}
		if val := rec.GetString("hero_title"); val != "" {
			content.HeroTitle = val
		}
		if val := rec.GetString("hero_subtitle"); val != "" {
			content.HeroSubtitle = val
		}
		if val := rec.GetString("owner_name"); val != "" {
			content.OwnerName = val
			content.NotaryName = val
		} else if val := rec.GetString("notary_name"); val != "" {
			content.OwnerName = val
			content.NotaryName = val
		}
		if val := rec.GetString("owner_title"); val != "" {
			content.OwnerTitle = val
			content.NotaryTitle = val
		} else if val := rec.GetString("notary_title"); val != "" {
			content.OwnerTitle = val
			content.NotaryTitle = val
		}
		if val := rec.GetString("phone1"); val != "" {
			content.Phone1 = val
		}
		if val := rec.GetString("phone2"); val != "" {
			content.Phone2 = val
		}
		if val := rec.GetString("work_hours"); val != "" {
			content.WorkHours = val
		}
		if val := rec.GetString("lunch_break"); val != "" {
			content.LunchBreak = val
		}
		if val := rec.GetString("days_off"); val != "" {
			content.DaysOff = val
		}
		if val := rec.GetString("address"); val != "" {
			content.Address = val
		}
		if val := rec.GetString("address_sub"); val != "" {
			content.AddressSub = val
		}
		if val := rec.GetString("map_iframe"); val != "" {
			content.MapIframe = val
		}
		if photo := rec.GetString("photo"); photo != "" {
			content.HeroPhoto = "/api/files/site_content/" + rec.Id + "/" + photo
		}
	}

	c.SetContent(content)
	return content
}

// GetPageBlocks retrieves active page blocks sorted by order
func GetPageBlocks(app core.App) []models.PageBlock {
	c := cache.GetStore()
	if cached, ok := c.GetBlocks(); ok {
		return cached
	}

	records, err := app.FindRecordsByFilter("page_blocks", "active = true", "order", 0, 0)
	if err != nil || len(records) == 0 {
		return nil
	}

	var blocks []models.PageBlock
	for _, r := range records {
		rawContent := r.GetString("content")
		// Санитизируем произвольный HTML/контент от XSS через bluemonday
		sanitizedContent := htmlSanitizer.Sanitize(rawContent)

		var payload map[string]interface{}
		rawJSON := r.Get("payload")
		if rawJSON != nil {
			if m, ok := rawJSON.(map[string]interface{}); ok {
				payload = m
			} else {
				// В PocketBase JSONField может возвращать types.JSONRaw или байты
				b, err := json.Marshal(rawJSON)
				if err == nil && len(b) > 2 {
					_ = json.Unmarshal(b, &payload)
				}
			}
		}

		blocks = append(blocks, models.PageBlock{
			ID:      r.Id,
			Type:    r.GetString("type"),
			Title:   r.GetString("title"),
			Content: sanitizedContent,
			Order:   r.GetInt("order"),
			Active:  r.GetBool("active"),
			Payload: payload,
		})
	}

	c.SetBlocks(blocks)
	return blocks
}

// GetServices retrieves active services sorted by order from cache or PocketBase
func GetServices(app core.App) []models.ServiceItem {
	c := cache.GetStore()
	if cached, ok := c.GetServices(); ok {
		return cached
	}

	records, err := app.FindRecordsByFilter("services", "active = true", "order", 0, 0)
	if err != nil || len(records) == 0 {
		return nil
	}

	var services []models.ServiceItem
	for _, r := range records {
		services = append(services, models.ServiceItem{
			ID:          r.Id,
			Title:       r.GetString("title"),
			Description: r.GetString("description"),
			Price:       r.GetString("price"),
			Order:       r.GetInt("order"),
			Active:      r.GetBool("active"),
		})
	}

	c.SetServices(services)
	return services
}

// GetBookedSlots returns a map of date -> slice of booked times for active leads (status != 'отменена')
func GetBookedSlots(app core.App) map[string][]string {
	result := make(map[string][]string)
	records, err := app.FindRecordsByFilter("leads", "status != 'отменена'", "booking_date,booking_time", 0, 0)
	if err != nil {
		return result
	}

	for _, r := range records {
		date := r.GetString("booking_date")
		time := r.GetString("booking_time")
		if date != "" && time != "" {
			result[date] = append(result[date], time)
		}
	}
	return result
}

// GetArticles retrieves published articles from cache or PocketBase
func GetArticles(app core.App) []models.ArticleItem {
	c := cache.GetStore()
	if cached, ok := c.GetArticles(); ok {
		return cached
	}

	records, err := app.FindRecordsByFilter("articles", "is_published = true", "-published_at,-created", 0, 0)
	if err != nil || len(records) == 0 {
		return nil
	}

	var articles []models.ArticleItem
	for _, r := range records {
		var tags []string
		if rawTags := r.GetString("tags"); rawTags != "" {
			parts := strings.Split(rawTags, ",")
			for _, p := range parts {
				trimmed := strings.TrimSpace(p)
				if trimmed != "" {
					tags = append(tags, trimmed)
				}
			}
		}

		cover := ""
		if file := r.GetString("cover_image"); file != "" {
			cover = "/api/files/articles/" + r.Id + "/" + file
		}

		articles = append(articles, models.ArticleItem{
			ID:             r.Id,
			Title:          r.GetString("title"),
			Slug:           r.GetString("slug"),
			Excerpt:        r.GetString("excerpt"),
			Content:        htmlSanitizer.Sanitize(r.GetString("content")),
			CoverImage:     cover,
			AuthorName:     r.GetString("author_name"),
			PublishedAt:    r.GetString("published_at"),
			IsPublished:    r.GetBool("is_published"),
			ViewsCount:     r.GetInt("views_count"),
			SEOTitle:       r.GetString("seo_title"),
			SEODescription: r.GetString("seo_description"),
			Tags:           tags,
		})
	}

	c.SetArticles(articles)
	return articles
}

// GetArticleBySlug finds a published article by slug
func GetArticleBySlug(app core.App, slug string) (*models.ArticleItem, error) {
	articles := GetArticles(app)
	for _, a := range articles {
		if a.Slug == slug {
			return &a, nil
		}
	}

	rec, err := app.FindFirstRecordByFilter("articles", "slug = {:slug} && is_published = true", dbx.Params{"slug": slug})
	if err != nil || rec == nil {
		return nil, fmt.Errorf("article not found")
	}

	var tags []string
	if rawTags := rec.GetString("tags"); rawTags != "" {
		parts := strings.Split(rawTags, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	cover := ""
	if file := rec.GetString("cover_image"); file != "" {
		cover = "/api/files/articles/" + rec.Id + "/" + file
	}

	art := &models.ArticleItem{
		ID:             rec.Id,
		Title:          rec.GetString("title"),
		Slug:           rec.GetString("slug"),
		Excerpt:        rec.GetString("excerpt"),
		Content:        htmlSanitizer.Sanitize(rec.GetString("content")),
		CoverImage:     cover,
		AuthorName:     rec.GetString("author_name"),
		PublishedAt:    rec.GetString("published_at"),
		IsPublished:    rec.GetBool("is_published"),
		ViewsCount:     rec.GetInt("views_count"),
		SEOTitle:       rec.GetString("seo_title"),
		SEODescription: rec.GetString("seo_description"),
		Tags:           tags,
	}
	return art, nil
}

// IncrementArticleViews increments the views counter for an article
func IncrementArticleViews(app core.App, id string) {
	rec, err := app.FindRecordById("articles", id)
	if err == nil && rec != nil {
		rec.Set("views_count", rec.GetInt("views_count")+1)
		_ = app.Save(rec)
	}
}

// GetServiceOrderByCode retrieves a service order by its unique tracking code
func GetServiceOrderByCode(app core.App, code string) (*models.ServiceOrder, error) {
	cleanCode := strings.ToUpper(strings.TrimSpace(code))
	rec, err := app.FindFirstRecordByFilter("service_orders", "order_code = {:code}", dbx.Params{"code": cleanCode})
	if err != nil || rec == nil {
		return nil, fmt.Errorf("заказ с кодом %s не найден", code)
	}

	var history []models.OrderStageHistory
	rawJSON := rec.Get("stages_history")
	if rawJSON != nil {
		b, err := json.Marshal(rawJSON)
		if err == nil {
			_ = json.Unmarshal(b, &history)
		}
	}

	order := &models.ServiceOrder{
		ID:            rec.Id,
		OrderCode:     rec.GetString("order_code"),
		ClientName:    rec.GetString("client_name"),
		ClientPhone:   rec.GetString("client_phone"),
		ItemTitle:     rec.GetString("item_title"),
		CurrentStage:  rec.GetString("current_stage"),
		TotalAmount:   rec.GetFloat("total_amount"),
		IsPaid:        rec.GetBool("is_paid"),
		StagesHistory: history,
		Notes:         rec.GetString("notes"),
		CreatedAt:     rec.GetString("created"),
		UpdatedAt:     rec.GetString("updated"),
	}

	return order, nil
}
