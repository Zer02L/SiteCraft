package data

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sitecraft/internal/cache"
	"sitecraft/internal/models"

	"github.com/microcosm-cc/bluemonday"
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
		SiteName:            "Нотариальная контора",
		Tagline:             "Официальные нотариальные действия",
		Theme:               "main",
		EnableA11y:          true,
		EnableAlpine:        false,
		MapMode:             "lazy",
		BookingMode:         "slots",
		EnablePayment:       false,
		PaymentSandbox:      true,
		EnableI18n:          true,
		EnableLiveChat:      true,
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
