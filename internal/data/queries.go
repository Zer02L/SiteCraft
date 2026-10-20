package data

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sitecraft/internal/cache"
	"sitecraft/internal/models"

	"github.com/microcosm-cc/bluemonday"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"golang.org/x/crypto/bcrypt"
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
		EnableLiveChat: true,
		EnableFleet:         true,
		DefaultLang:         "ru",
		LegalEntity:         "ООО \"АвтоСервис Про\"",
		LegalAddress:        "г. Санкт-Петербург, Невский пр-т, д. 42",
		INN:                 "7801234567",
		KPP:                 "780101001",
		OGRN:                "1157847000123",
		BankName:            "ПАО Сбербанк",
		BankBIK:             "044030653",
		BankAccount:         "40702810900000012345",
		BankCorr:            "30101810500000000653",
		ContactEmail:        "info@sitecraft.local",
		EnableCareers:       true,
		EnableCookieConsent: true,
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
		if rec.Collection().Fields.GetByName("enable_live_chat") != nil {
			settings.EnableLiveChat = rec.GetBool("enable_live_chat")
		} else {
			settings.EnableLiveChat = true
		}
		if rec.Collection().Fields.GetByName("enable_fleet") != nil {
			settings.EnableFleet = rec.GetBool("enable_fleet")
		} else {
			settings.EnableFleet = true
		}
		if rec.Collection().Fields.GetByName("enable_recent_activity") != nil {
			settings.EnableRecentActivity = rec.GetBool("enable_recent_activity")
		} else {
			settings.EnableRecentActivity = true
		}
		settings.HeaderVariant = rec.GetString("header_variant")
		settings.FooterVariant = rec.GetString("footer_variant")
		if val := rec.GetString("legal_entity"); val != "" {
			settings.LegalEntity = val
		}
		if val := rec.GetString("legal_address"); val != "" {
			settings.LegalAddress = val
		}
		if val := rec.GetString("inn"); val != "" {
			settings.INN = val
		}
		if val := rec.GetString("kpp"); val != "" {
			settings.KPP = val
		}
		if val := rec.GetString("ogrn"); val != "" {
			settings.OGRN = val
		}
		if val := rec.GetString("bank_name"); val != "" {
			settings.BankName = val
		}
		if val := rec.GetString("bank_bik"); val != "" {
			settings.BankBIK = val
		}
		if val := rec.GetString("bank_account"); val != "" {
			settings.BankAccount = val
		}
		if val := rec.GetString("bank_corr"); val != "" {
			settings.BankCorr = val
		}
		if val := rec.GetString("contact_email"); val != "" {
			settings.ContactEmail = val
		}
		if rec.Collection().Fields.GetByName("enable_careers") != nil {
			settings.EnableCareers = rec.GetBool("enable_careers")
		}
		if rec.Collection().Fields.GetByName("enable_cookie_consent") != nil {
			settings.EnableCookieConsent = rec.GetBool("enable_cookie_consent")
		}
		settings.ClientAdminPassword = rec.GetString("client_admin_password")
		if rec.Collection().Fields.GetByName("client_admin_password_hash") != nil {
			settings.ClientAdminPasswordHash = rec.GetString("client_admin_password_hash")
		}
		// Авто-миграция legacy пароля в bcrypt хэш при необходимости
		if settings.ClientAdminPasswordHash == "" && settings.ClientAdminPassword != "" && settings.ClientAdminPassword != "admin" {
			if hashed, err := bcrypt.GenerateFromPassword([]byte(settings.ClientAdminPassword), bcrypt.DefaultCost); err == nil {
				settings.ClientAdminPasswordHash = string(hashed)
				rec.Set("client_admin_password_hash", string(hashed))
				rec.Set("client_admin_password", "")
				_ = app.Save(rec)
			}
		}
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
		content.HeroTitleB = rec.GetString("hero_title_b")
		content.HeroSubtitleB = rec.GetString("hero_subtitle_b")
		content.HeroCTAB = rec.GetString("hero_cta_b")
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
		stockCount := -1
		if r.Collection().Fields.GetByName("stock_count") != nil {
			stockCount = r.GetInt("stock_count")
		}
		isPreorder := false
		if r.Collection().Fields.GetByName("is_preorder") != nil {
			isPreorder = r.GetBool("is_preorder")
		}
		services = append(services, models.ServiceItem{
			ID:          r.Id,
			Title:       r.GetString("title"),
			Description: r.GetString("description"),
			Price:       r.GetString("price"),
			Order:       r.GetInt("order"),
			Active:      r.GetBool("active"),
			StockCount:  stockCount,
			IsPreorder:  isPreorder,
		})
	}

	c.SetServices(services)
	return services
}

// GetBookedSlots returns a map of date -> slice of booked times for active leads (status != 'отменена')
// Optionally filters by resource_id (specialist or workstation).
func GetBookedSlots(app core.App, resourceID ...string) map[string][]string {
	result := make(map[string][]string)

	filter := "status != 'отменена'"
	params := dbx.Params{}
	if len(resourceID) > 0 && resourceID[0] != "" {
		filter += " && resource_id = {:resID}"
		params["resID"] = resourceID[0]
	}

	records, err := app.FindRecordsByFilter("leads", filter, "booking_date,booking_time", 0, 0, params)
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

// RecordToArticleItem maps PocketBase record to models.ArticleItem
func RecordToArticleItem(r *core.Record) models.ArticleItem {
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

	return models.ArticleItem{
		ID:             r.Id,
		Title:          r.GetString("title"),
		Slug:           r.GetString("slug"),
		Excerpt:        r.GetString("excerpt"),
		Content:        htmlSanitizer.Sanitize(r.GetString("content")),
		CoverImage:     cover,
		AuthorName:     r.GetString("author_name"),
		PublishedAt:    r.GetString("published_at"),
		IsPublished:    r.GetBool("is_published"),
		IsDeleted:      r.GetBool("is_deleted"),
		DeletedAt:      r.GetString("deleted_at"),
		ViewsCount:     r.GetInt("views_count"),
		SEOTitle:       r.GetString("seo_title"),
		SEODescription: r.GetString("seo_description"),
		Tags:           tags,
	}
}

// GetArticles retrieves published and live articles (not scheduled in future, not in trash) from cache or PocketBase
func GetArticles(app core.App) []models.ArticleItem {
	c := cache.GetStore()
	if cached, ok := c.GetArticles(); ok {
		return cached
	}

	records, err := app.FindRecordsByFilter("articles", "is_published = true && (is_deleted = false || is_deleted = null)", "-published_at,-created", 0, 0)
	if err != nil || len(records) == 0 {
		return nil
	}

	var articles []models.ArticleItem
	for _, r := range records {
		art := RecordToArticleItem(r)
		if art.IsLive() {
			articles = append(articles, art)
		}
	}

	c.SetArticles(articles)
	return articles
}

// GetArticleBySlug finds a published live article by slug
func GetArticleBySlug(app core.App, slug string) (*models.ArticleItem, error) {
	articles := GetArticles(app)
	for _, a := range articles {
		if a.Slug == slug && a.IsLive() {
			return &a, nil
		}
	}

	rec, err := app.FindFirstRecordByFilter("articles", "slug = {:slug} && is_published = true && (is_deleted = false || is_deleted = null)", dbx.Params{"slug": slug})
	if err != nil || rec == nil {
		return nil, fmt.Errorf("article not found")
	}

	art := RecordToArticleItem(rec)
	if !art.IsLive() {
		return nil, fmt.Errorf("article is scheduled or not published")
	}
	return &art, nil
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

// GetAllLeads retrieves all leads sorted by created descending (or rowid/id if created is not present)
func GetAllLeads(app core.App) []models.Lead {
	records, err := app.FindRecordsByFilter("leads", "", "-created", 0, 0)
	if err != nil {
		records, err = app.FindRecordsByFilter("leads", "", "-rowid", 0, 0)
		if err != nil {
			records, _ = app.FindRecordsByFilter("leads", "", "", 0, 0)
		}
	}
	if len(records) == 0 {
		return nil
	}

	var leads []models.Lead
	for _, r := range records {
		var payload map[string]interface{}
		rawJSON := r.Get("payload")
		if rawJSON != nil {
			if m, ok := rawJSON.(map[string]interface{}); ok {
				payload = m
			} else {
				b, err := json.Marshal(rawJSON)
				if err == nil {
					_ = json.Unmarshal(b, &payload)
				}
			}
		}

		leads = append(leads, models.Lead{
			ID:          r.Id,
			Name:        r.GetString("name"),
			Phone:       r.GetString("phone"),
			Email:       r.GetString("email"),
			Status:      r.GetString("status"),
			Source:      r.GetString("source"),
			BookingDate: r.GetString("booking_date"),
			BookingTime: r.GetString("booking_time"),
			Payload:     payload,
		})
	}
	return leads
}

// GetAllArticlesAdmin retrieves all active articles (not in trash) for client admin
func GetAllArticlesAdmin(app core.App) []models.ArticleItem {
	records, err := app.FindRecordsByFilter("articles", "is_deleted = false || is_deleted = null", "-created,-published_at", 0, 0)
	if err != nil {
		records, err = app.FindRecordsByFilter("articles", "", "-created", 0, 0)
		if err != nil {
			return nil
		}
	}

	var articles []models.ArticleItem
	for _, r := range records {
		if r.GetBool("is_deleted") {
			continue
		}
		articles = append(articles, RecordToArticleItem(r))
	}
	return articles
}

// GetTrashArticlesAdmin retrieves deleted articles (in trash) for client admin
func GetTrashArticlesAdmin(app core.App) []models.ArticleItem {
	records, err := app.FindRecordsByFilter("articles", "is_deleted = true", "-deleted_at,-updated", 0, 0)
	if err != nil || len(records) == 0 {
		return nil
	}

	var articles []models.ArticleItem
	for _, r := range records {
		articles = append(articles, RecordToArticleItem(r))
	}
	return articles
}

// ToggleArticlePublish toggles is_published status
func ToggleArticlePublish(app core.App, id string) (*models.ArticleItem, error) {
	rec, err := app.FindRecordById("articles", id)
	if err != nil || rec == nil {
		return nil, fmt.Errorf("статья не найдена")
	}
	current := rec.GetBool("is_published")
	rec.Set("is_published", !current)
	if !current && rec.GetString("published_at") == "" {
		rec.Set("published_at", time.Now().Format("2006-01-02 15:04"))
	}
	if err := app.Save(rec); err != nil {
		return nil, err
	}
	cache.GetStore().InvalidateArticles()
	art := RecordToArticleItem(rec)
	return &art, nil
}

// MoveArticleToTrash moves article to trash (soft delete)
func MoveArticleToTrash(app core.App, id string) error {
	rec, err := app.FindRecordById("articles", id)
	if err != nil || rec == nil {
		return fmt.Errorf("статья не найдена")
	}
	rec.Set("is_deleted", true)
	rec.Set("deleted_at", time.Now().Format("2006-01-02 15:04:05"))
	if err := app.Save(rec); err != nil {
		return err
	}
	cache.GetStore().InvalidateArticles()
	return nil
}

// RestoreArticleFromTrash restores article from trash
func RestoreArticleFromTrash(app core.App, id string) error {
	rec, err := app.FindRecordById("articles", id)
	if err != nil || rec == nil {
		return fmt.Errorf("статья не найдена")
	}
	rec.Set("is_deleted", false)
	rec.Set("deleted_at", "")
	if err := app.Save(rec); err != nil {
		return err
	}
	cache.GetStore().InvalidateArticles()
	return nil
}

// PermanentDeleteArticle permanently removes article from database
func PermanentDeleteArticle(app core.App, id string) error {
	rec, err := app.FindRecordById("articles", id)
	if err != nil || rec == nil {
		return fmt.Errorf("статья не найдена")
	}
	if err := app.Delete(rec); err != nil {
		return err
	}
	cache.GetStore().InvalidateArticles()
	return nil
}

// EmptyTrashArticles permanently deletes all articles in trash
func EmptyTrashArticles(app core.App) error {
	records, err := app.FindRecordsByFilter("articles", "is_deleted = true", "", 0, 0)
	if err != nil {
		return err
	}
	for _, rec := range records {
		_ = app.Delete(rec)
	}
	cache.GetStore().InvalidateArticles()
	return nil
}

func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "недавно"
	}
	diff := time.Since(t)
	if diff < time.Minute {
		return "только что"
	}
	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins <= 0 {
			mins = 1
		}
		return fmt.Sprintf("%d мин назад", mins)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%d ч назад", hours)
	}
	days := int(diff.Hours() / 24)
	return fmt.Sprintf("%d дн назад", days)
}

// GetRecentActivity retrieves depersonalized marketing events (leads + orders) cached in memory for 5 minutes
func GetRecentActivity(app core.App) []models.RecentActivityItem {
	c := cache.GetStore()
	if cached, ok := c.GetRecentActivity(); ok {
		return cached
	}

	content := GetSiteContent(app)
	defaultCity := "СПб"
	if strings.Contains(content.AddressSub, "Краснодар") || strings.Contains(content.Address, "Краснодар") {
		defaultCity = "Краснодар"
	} else if strings.Contains(content.AddressSub, "Москва") || strings.Contains(content.Address, "Москва") {
		defaultCity = "Москва"
	} else if strings.Contains(content.AddressSub, "Петербург") || strings.Contains(content.Address, "СПб") {
		defaultCity = "СПб"
	}

	var items []models.RecentActivityItem

	// 1. Недавние лиды (leads)
	if leadRecords, err := app.FindRecordsByFilter("leads", "status != 'отменена'", "-created", 6, 0); err == nil {
		for _, r := range leadRecords {
			name := strings.TrimSpace(r.GetString("name"))
			if fields := strings.Fields(name); len(fields) > 0 {
				name = fields[0]
			}
			if name == "" {
				name = "Посетитель"
			}

			action := "оформил заявку"
			icon := "lead"
			if r.GetString("booking_date") != "" {
				action = "записался на консультацию"
				icon = "booking"
			}

			var payload map[string]interface{}
			if rawJSON := r.Get("payload"); rawJSON != nil {
				if m, ok := rawJSON.(map[string]interface{}); ok {
					payload = m
				}
			}
			if payload != nil {
				if act, ok := payload["action_type"].(string); ok && act != "" {
					action = "оформил заявку на " + act
				} else if _, ok := payload["quiz_summary"]; ok {
					action = "рассчитал стоимость сметы"
				}
			}

			createdTime := r.GetDateTime("created").Time()
			items = append(items, models.RecentActivityItem{
				Name:    name,
				City:    defaultCity,
				Action:  action,
				TimeAgo: formatTimeAgo(createdTime),
				Icon:    icon,
			})
		}
	}

	// 2. Недавние сервисные заказы (service_orders)
	if orderRecords, err := app.FindRecordsByFilter("service_orders", "", "-created", 4, 0); err == nil {
		for _, r := range orderRecords {
			name := strings.TrimSpace(r.GetString("client_name"))
			if fields := strings.Fields(name); len(fields) > 0 {
				name = fields[0]
			}
			if name == "" {
				name = "Клиент"
			}

			action := "оформил заказ"
			if itemTitle := strings.TrimSpace(r.GetString("item_title")); itemTitle != "" {
				// берем краткое название
				if len(itemTitle) > 28 {
					itemTitle = itemTitle[:25] + "..."
				}
				action = "заказал " + itemTitle
			}

			createdTime := r.GetDateTime("created").Time()
			items = append(items, models.RecentActivityItem{
				Name:    name,
				City:    defaultCity,
				Action:  action,
				TimeAgo: formatTimeAgo(createdTime),
				Icon:    "order",
			})
		}
	}

	// 3. Если реальных записей пока мало, подмешиваем органичные события для начального Social Proof
	if len(items) < 2 {
		items = append(items,
			models.RecentActivityItem{
				Name:    "Константин",
				City:    defaultCity,
				Action:  "заказал консультацию",
				TimeAgo: "12 мин назад",
				Icon:    "lead",
			},
			models.RecentActivityItem{
				Name:    "Елена",
				City:    defaultCity,
				Action:  "записалась на прием",
				TimeAgo: "28 мин назад",
				Icon:    "booking",
			},
			models.RecentActivityItem{
				Name:    "Михаил",
				City:    defaultCity,
				Action:  "оформил заказ онлайн",
				TimeAgo: "45 мин назад",
				Icon:    "order",
			},
		)
	}

	// Кэшируем в памяти ровно на 5 минут
	c.SetRecentActivity(items, 5*time.Minute)
	return items
}

// GetResources retrieves active specialists, workstations or cabinets sorted by order
func GetResources(app core.App) []models.ResourceItem {
	records, err := app.FindRecordsByFilter("resources", "active = true", "order", 0, 0)
	if err != nil || len(records) == 0 {
		return nil
	}

	var items []models.ResourceItem
	for _, r := range records {
		var services []string
		if sStr := r.GetString("service_ids"); sStr != "" {
			parts := strings.Split(sStr, ",")
			for _, p := range parts {
				if trimmed := strings.TrimSpace(p); trimmed != "" {
					services = append(services, trimmed)
				}
			}
		}

		items = append(items, models.ResourceItem{
			ID:       r.Id,
			Name:     r.GetString("name"),
			Role:     r.GetString("role"),
			Avatar:   r.GetString("avatar"),
			Active:   r.GetBool("active"),
			Order:    r.GetInt("order"),
			Services: services,
		})
	}
	return items
}

// GetReviews retrieves published customer reviews for public display
func GetReviews(app core.App) []models.ReviewItem {
	records, err := app.FindRecordsByFilter("reviews", "is_published = true", "-created", 0, 0)
	if err != nil || len(records) == 0 {
		return nil
	}

	var items []models.ReviewItem
	for _, r := range records {
		items = append(items, models.ReviewItem{
			ID:          r.Id,
			AuthorName:  r.GetString("author_name"),
			Rating:      r.GetInt("rating"),
			Comment:     r.GetString("comment"),
			OrderID:     r.GetString("order_id"),
			IsPublished: r.GetBool("is_published"),
			Source:      r.GetString("source"),
			CreatedAt:   r.GetString("created"),
		})
	}
	return items
}

// GetReviewAggregate computes aggregate metrics for Schema.org JSON-LD
func GetReviewAggregate(reviews []models.ReviewItem) models.ReviewAggregate {
	if len(reviews) == 0 {
		return models.ReviewAggregate{
			TotalReviews: 0,
			AverageScore: 5.0,
			BestRating:   5,
			WorstRating:  1,
		}
	}

	total := len(reviews)
	sum := 0
	for _, r := range reviews {
		sum += r.Rating
	}
	avg := float64(sum) / float64(total)

	return models.ReviewAggregate{
		TotalReviews: total,
		AverageScore: avg,
		BestRating:   5,
		WorstRating:  1,
	}
}

// CreateReview saves a customer review with is_published = false (pending moderation)
func CreateReview(app core.App, review models.ReviewItem) (*core.Record, error) {
	col, err := app.FindCollectionByNameOrId("reviews")
	if err != nil {
		return nil, err
	}

	rec := core.NewRecord(col)
	rec.Set("author_name", review.AuthorName)
	rec.Set("rating", review.Rating)
	rec.Set("comment", htmlSanitizer.Sanitize(review.Comment))
	rec.Set("order_id", review.OrderID)
	rec.Set("is_published", review.IsPublished)
	rec.Set("source", review.Source)

	if err := app.Save(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// GetAllReviewsAdmin retrieves all reviews for client admin moderation
func GetAllReviewsAdmin(app core.App) []models.ReviewItem {
	records, err := app.FindRecordsByFilter("reviews", "", "-created", 0, 0)
	if err != nil || len(records) == 0 {
		return nil
	}

	var items []models.ReviewItem
	for _, r := range records {
		items = append(items, models.ReviewItem{
			ID:          r.Id,
			AuthorName:  r.GetString("author_name"),
			Rating:      r.GetInt("rating"),
			Comment:     r.GetString("comment"),
			OrderID:     r.GetString("order_id"),
			IsPublished: r.GetBool("is_published"),
			Source:      r.GetString("source"),
			CreatedAt:   r.GetString("created"),
		})
	}
	return items
}

// ToggleReviewPublish toggles the publication status of a review
func ToggleReviewPublish(app core.App, id string) error {
	rec, err := app.FindRecordById("reviews", id)
	if err != nil || rec == nil {
		return fmt.Errorf("отзыв не найден")
	}
	current := rec.GetBool("is_published")
	rec.Set("is_published", !current)
	return app.Save(rec)
}

// DeleteReview permanently deletes a review
func DeleteReview(app core.App, id string) error {
	rec, err := app.FindRecordById("reviews", id)
	if err != nil || rec == nil {
		return fmt.Errorf("отзыв не найден")
	}
	return app.Delete(rec)
}

// GetBranches returns all active branches ordered by display order (Блок 4: Сеть филиалов).
func GetBranches(app core.App) []models.BranchItem {
	records, err := app.FindRecordsByFilter("branches", "active = true", "order", 50, 0)
	if err != nil {
		return nil
	}

	var items []models.BranchItem
	for _, r := range records {
		var services []string
		srvStr := r.GetString("services")
		if srvStr != "" {
			for _, s := range strings.Split(srvStr, ",") {
				if trimmed := strings.TrimSpace(s); trimmed != "" {
					services = append(services, trimmed)
				}
			}
		}

		items = append(items, models.BranchItem{
			ID:        r.Id,
			Name:      r.GetString("name"),
			City:      r.GetString("city"),
			Address:   r.GetString("address"),
			Phone:     r.GetString("phone"),
			WorkHours: r.GetString("work_hours"),
			Lat:       r.GetFloat("lat"),
			Lng:       r.GetFloat("lng"),
			Metro:     r.GetString("metro"),
			Services:  services,
			IsMain:    r.GetBool("is_main"),
			Active:    r.GetBool("active"),
		})
	}
	return items
}

// GetVacancies returns all active vacancies for Careers/HR module (Блок 4: Вакансии).
func GetVacancies(app core.App) []models.VacancyItem {
	records, err := app.FindRecordsByFilter("vacancies", "active = true", "order", 50, 0)
	if err != nil {
		return nil
	}

	var items []models.VacancyItem
	for _, r := range records {
		var resp, req, ben []string
		_ = json.Unmarshal([]byte(r.GetString("responsibilities")), &resp)
		_ = json.Unmarshal([]byte(r.GetString("requirements")), &req)
		_ = json.Unmarshal([]byte(r.GetString("benefits")), &ben)

		items = append(items, models.VacancyItem{
			ID:               r.Id,
			Title:            r.GetString("title"),
			Department:       r.GetString("department"),
			Salary:           r.GetString("salary"),
			Experience:       r.GetString("experience"),
			Schedule:         r.GetString("schedule"),
			Responsibilities: resp,
			Requirements:     req,
			Benefits:         ben,
			Active:           r.GetBool("active"),
			Order:            int(r.GetInt("order")),
		})
	}
	return items
}

// CreateVacancyLead saves an HR job application lead into database.
func CreateVacancyLead(app core.App, name, phone, email, vacancyTitle, resumeLink, coverLetter string) error {
	col, err := app.FindCollectionByNameOrId("leads")
	if err != nil {
		return err
	}

	rec := core.NewRecord(col)
	rec.Set("name", name)
	rec.Set("phone", phone)
	rec.Set("email", email)
	rec.Set("source", "vacancy")
	rec.Set("status", "новая")
	rec.Set("payload", map[string]interface{}{
		"action_type":  "hr_application",
		"vacancy":      vacancyTitle,
		"resume":       resumeLink,
		"cover_letter": coverLetter,
	})
	return app.Save(rec)
}

// GetInvoiceData retrieves or generates official B2B invoice data for print (Блок 4: Invoicing).
func GetInvoiceData(app core.App, orderCode string) (*models.InvoiceData, error) {
	settings := GetSiteSettings(app)
	orderCode = strings.TrimSpace(orderCode)

	var order *models.ServiceOrder
	if orderCode != "" {
		o, err := GetServiceOrderByCode(app, orderCode)
		if err == nil {
			order = o
		}
	}

	invoiceNum := "СЧ-" + time.Now().Format("060102") + "-01"
	totalAmount := 14500.0
	clientName := "ООО «ТехноПром Логистик»"
	clientPhone := "+7 (999) 123-45-67"
	itemTitle := "Комплексное техническое обслуживание и сервис оборудования"

	if order != nil {
		invoiceNum = "СЧ-" + order.OrderCode
		totalAmount = order.TotalAmount
		if order.ClientName != "" {
			clientName = order.ClientName
		}
		if order.ClientPhone != "" {
			clientPhone = order.ClientPhone
		}
		if order.ItemTitle != "" {
			itemTitle = order.ItemTitle
		}
	}

	now := time.Now()
	monthsRu := []string{
		"", "января", "февраля", "марта", "апреля", "мая", "июня",
		"июля", "августа", "сентября", "октября", "ноября", "декабря",
	}
	dateStr := fmt.Sprintf("%d %s %d г.", now.Day(), monthsRu[now.Month()], now.Year())

	supplier := models.InvoiceEntity{
		Name:        settings.LegalEntity,
		INN:         settings.INN,
		KPP:         settings.KPP,
		OGRN:        settings.OGRN,
		Address:     settings.LegalAddress,
		Phone:       settings.WhatsappPhone,
		BankName:    settings.BankName,
		BankBIK:     settings.BankBIK,
		BankAccount: settings.BankAccount,
		BankCorr:    settings.BankCorr,
	}
	if supplier.Name == "" {
		supplier.Name = "ООО «АвтоСервис Про»"
		supplier.INN = "7801234567"
		supplier.KPP = "780101001"
		supplier.OGRN = "1157847000123"
		supplier.Address = "191025, г. Санкт-Петербург, Невский пр-т, д. 42"
		supplier.BankName = "ПАО Сбербанк"
		supplier.BankBIK = "044030653"
		supplier.BankAccount = "40702810900000012345"
		supplier.BankCorr = "30101810500000000653"
	}

	customer := models.InvoiceEntity{
		Name:    clientName,
		Phone:   clientPhone,
		Address: "г. Санкт-Петербург",
	}

	lines := []models.InvoiceLineItem{
		{
			Index:    1,
			Title:    itemTitle,
			Unit:     "усл.",
			Quantity: 1,
			Price:    totalAmount,
			Amount:   totalAmount,
		},
	}

	// QR-код по ГОСТ Р 56042-2014 для моментальной оплаты в мобильных банках
	paymentCode := fmt.Sprintf("ST00012|Name=%s|PersonalAcc=%s|BankName=%s|BIC=%s|CorrespAcc=%s|PayeeINN=%s|Purpose=Оплата по счету %s|Sum=%d",
		supplier.Name, supplier.BankAccount, supplier.BankName, supplier.BankBIK, supplier.BankCorr, supplier.INN, invoiceNum, int64(totalAmount*100))

	return &models.InvoiceData{
		InvoiceNumber: invoiceNum,
		InvoiceDate:   dateStr,
		OrderCode:     orderCode,
		Supplier:      supplier,
		Customer:      customer,
		Items:         lines,
		TotalAmount:   totalAmount,
		VATText:       "Без НДС (в связи с применением УСН, ст. 346.11 НК РФ)",
		TotalInWords:  fmt.Sprintf("%.0f рублей 00 копеек", totalAmount),
		PaymentCode:   paymentCode,
		Notes:         "Оплата настоящего счета означает согласие с условиями поставки и выполнения работ. Срок оплаты — 3 банковских дня.",
	}, nil
}


