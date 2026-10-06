package data

import (
	"log"
	"os"

	"sitecraft/internal/cache"
	"sitecraft/internal/models"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

// InitSchema ensures site_settings, site_content, services, and leads collections exist with default data.
func InitSchema(app core.App) error {
	if err := initSiteSettingsCollection(app); err != nil {
		return err
	}
	if err := initSiteContentCollection(app); err != nil {
		return err
	}
	if err := initServicesCollection(app); err != nil {
		return err
	}
	if err := initLeadsCollection(app); err != nil {
		return err
	}
	RegisterCacheHooks(app)
	return nil
}

// RegisterCacheHooks sets up event listeners to invalidate in-memory cache upon admin updates
func RegisterCacheHooks(app core.App) {
	app.OnRecordAfterCreateSuccess("site_settings", "site_content", "services").BindFunc(func(e *core.RecordEvent) error {
		invalidateByCollection(e.Record.Collection().Name)
		return e.Next()
	})
	app.OnRecordAfterUpdateSuccess("site_settings", "site_content", "services").BindFunc(func(e *core.RecordEvent) error {
		invalidateByCollection(e.Record.Collection().Name)
		return e.Next()
	})
	app.OnRecordAfterDeleteSuccess("site_settings", "site_content", "services").BindFunc(func(e *core.RecordEvent) error {
		invalidateByCollection(e.Record.Collection().Name)
		return e.Next()
	})
}

func invalidateByCollection(colName string) {
	c := cache.GetStore()
	switch colName {
	case "site_settings":
		c.InvalidateSettings()
		log.Println("[Cache] site_settings invalidated")
	case "site_content":
		c.InvalidateContent()
		log.Println("[Cache] site_content invalidated")
	case "services":
		c.InvalidateServices()
		log.Println("[Cache] services invalidated")
	}
}

func initSiteSettingsCollection(app core.App) error {
	_, err := app.FindCollectionByNameOrId("site_settings")
	if err != nil {
		col := core.NewBaseCollection("site_settings")
		col.Fields.Add(
			&core.TextField{Name: "site_name", Required: true},
			&core.TextField{Name: "tagline"},
			&core.TextField{Name: "theme"}, // legal, medical, auto, tech, default
			&core.BoolField{Name: "enable_a11y"},
			&core.BoolField{Name: "enable_alpine"},
			&core.BoolField{Name: "enable_turnstile"},
			&core.TextField{Name: "turnstile_site_key"},
			&core.TextField{Name: "turnstile_secret_key"},
			&core.TextField{Name: "telegram_bot_token"},
			&core.TextField{Name: "telegram_chat_id"},
			&core.TextField{Name: "webhook_url"},
			&core.TextField{Name: "yandex_metrika_id"},
			&core.TextField{Name: "map_mode"},     // iframe, lazy, static
			&core.TextField{Name: "booking_mode"}, // slots, simple
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция site_settings успешно создана")
	}

	count, _ := app.CountRecords("site_settings")
	if count == 0 {
		col, err := app.FindCollectionByNameOrId("site_settings")
		if err != nil {
			return err
		}
		record := core.NewRecord(col)
		record.Set("site_name", "Нотариальная контора")
		record.Set("tagline", "Официальные нотариальные действия в соответствии с законодательством РФ")
		record.Set("theme", "legal")
		record.Set("enable_a11y", true)
		record.Set("enable_alpine", false) // Zero-JS по умолчанию!
		record.Set("enable_turnstile", false)
		record.Set("map_mode", "lazy")
		record.Set("booking_mode", "slots")

		if err := app.Save(record); err != nil {
			log.Printf("[Migration Error] Ошибка создания записи site_settings: %v", err)
		} else {
			log.Println("[Migration] Базовые настройки site_settings созданы")
		}
	}
	return nil
}

func initSiteContentCollection(app core.App) error {
	col, err := app.FindCollectionByNameOrId("site_content")
	if err != nil {
		col = core.NewBaseCollection("site_content")
		col.Fields.Add(
			&core.TextField{Name: "hero_badge"},
			&core.TextField{Name: "hero_title"},
			&core.TextField{Name: "hero_subtitle"},
			&core.TextField{Name: "owner_name"},
			&core.TextField{Name: "owner_title"},
			&core.TextField{Name: "phone1"},
			&core.TextField{Name: "phone2"},
			&core.TextField{Name: "work_hours"},
			&core.TextField{Name: "lunch_break"},
			&core.TextField{Name: "days_off"},
			&core.TextField{Name: "address"},
			&core.TextField{Name: "address_sub"},
			&core.TextField{Name: "map_iframe"},
			&core.FileField{
				Name:      "photo",
				MaxSelect: 1,
				MimeTypes: []string{"image/png", "image/jpeg", "image/webp", "image/svg+xml", "image/gif"},
			},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция site_content успешно создана")
	} else {
		// Ensure owner_name and owner_title fields exist (or notary_name if legacy)
		hasOwnerName := col.Fields.GetByName("owner_name") != nil
		if !hasOwnerName {
			col.Fields.Add(&core.TextField{Name: "owner_name"})
			col.Fields.Add(&core.TextField{Name: "owner_title"})
			_ = app.Save(col)
		}
	}

	count, _ := app.CountRecords("site_content")
	if count == 0 {
		col, err := app.FindCollectionByNameOrId("site_content")
		if err != nil {
			return err
		}
		record := core.NewRecord(col)
		record.Set("hero_badge", "Нотариальная контора открыта")
		record.Set("hero_title", "Нотариус")
		record.Set("hero_subtitle", "Совершение всех видов нотариальных действий в строгом соответствии с законодательством РФ. Законность, оперативность и правовая защита ваших интересов.")
		record.Set("owner_name", "Захаров Владислав Борисович")
		record.Set("owner_title", "Нотариус города Краснодара")
		record.Set("phone1", "8-918-024-64-20")
		record.Set("phone2", "8 (861) 224-89-55")
		record.Set("work_hours", "Понедельник — Пятница: 9:00 - 17:30")
		record.Set("lunch_break", "13:00 - 14:00")
		record.Set("days_off", "Выходные: суббота, воскресенье")
		record.Set("address", "Нотариальная контора нотариуса Захарова В. Б.")
		record.Set("address_sub", "Краснодарский край / г. Краснодар (код 861)")
		record.Set("map_iframe", "https://yandex.ru/map-widget/v1/?ll=38.975313%2C45.035470&z=15&pt=38.975313,45.035470,pm2rdm")

		if _, err := os.Stat("ui/assets/images/notary_sign.png"); err == nil {
			if f, err := filesystem.NewFileFromPath("ui/assets/images/notary_sign.png"); err == nil {
				record.Set("photo", f)
			}
		}

		if err := app.Save(record); err != nil {
			log.Printf("[Migration Error] Ошибка создания записи site_content: %v", err)
		} else {
			log.Println("[Migration] Начальные тексты и фото сайта созданы в site_content")
		}
	}

	return nil
}

func initServicesCollection(app core.App) error {
	_, err := app.FindCollectionByNameOrId("services")
	if err != nil {
		col := core.NewBaseCollection("services")
		col.Fields.Add(
			&core.TextField{Name: "title", Required: true},
			&core.TextField{Name: "description"},
			&core.TextField{Name: "price"},
			&core.NumberField{Name: "order", OnlyInt: true},
			&core.BoolField{Name: "active"},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция services успешно создана")
	}

	count, _ := app.CountRecords("services")
	if count == 0 {
		col, err := app.FindCollectionByNameOrId("services")
		if err != nil {
			return err
		}

		defaultServices := []struct {
			Title       string
			Description string
			Price       string
			Order       int
		}{
			{
				Title:       "Удостоверение сделок и договоров",
				Description: "Купля-продажа, дарение недвижимости, долей в уставном капитале ООО, брачные договоры.",
				Price:       "по тарифу НК РФ",
				Order:       1,
			},
			{
				Title:       "Оформление доверенностей",
				Description: "Генеральные доверенности, судебные, на управление автомобилем, распоряжение имуществом.",
				Price:       "от 1 500 ₽",
				Order:       2,
			},
			{
				Title:       "Ведение наследственных дел",
				Description: "Открытие наследственного дела, розыск завещаний, выдача свидетельств о праве на наследство.",
				Price:       "по тарифу НК РФ",
				Order:       3,
			},
			{
				Title:       "Нотариальные согласия и отказы",
				Description: "Согласия супругов на сделки, выезд несовершеннолетних за границу, отказ от преимущественного права покупки.",
				Price:       "от 1 800 ₽",
				Order:       4,
			},
			{
				Title:       "Свидетельствование верности копий",
				Description: "Заверение копий учредительных документов, паспортов, дипломов, подлинности подписи переводчиков.",
				Price:       "от 100 ₽ / стр",
				Order:       5,
			},
			{
				Title:       "Обеспечение доказательств",
				Description: "Нотариальный протокол осмотра интернет-страниц, мессенджеров, переписки для суда.",
				Price:       "по регламенту ФНП",
				Order:       6,
			},
		}

		for _, s := range defaultServices {
			rec := core.NewRecord(col)
			rec.Set("title", s.Title)
			rec.Set("description", s.Description)
			rec.Set("price", s.Price)
			rec.Set("order", s.Order)
			rec.Set("active", true)
			if err := app.Save(rec); err != nil {
				log.Printf("[Migration Error] Ошибка добавления услуги %s: %v", s.Title, err)
			}
		}
		log.Println("[Migration] Базовые услуги нотариуса успешно добавлены в services")
	}

	return nil
}

func initLeadsCollection(app core.App) error {
	col, err := app.FindCollectionByNameOrId("leads")
	if err != nil {
		col = core.NewBaseCollection("leads")
		col.Fields.Add(
			&core.TextField{Name: "name", Required: true},
			&core.TextField{Name: "phone", Required: true},
			&core.TextField{Name: "email"},
			&core.TextField{Name: "source"},
			&core.TextField{Name: "booking_date"},
			&core.TextField{Name: "booking_time"},
			&core.TextField{Name: "status"}, // например, "новая", "обработана", "отменена"
			&core.JSONField{Name: "payload"}, // динамические поля: action_type, comment, vin, etc.
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция leads для онлайн-заявок успешно создана")
	} else {
		// Ensure payload and email/source fields exist
		changed := false
		if col.Fields.GetByName("payload") == nil {
			col.Fields.Add(&core.JSONField{Name: "payload"})
			changed = true
		}
		if col.Fields.GetByName("email") == nil {
			col.Fields.Add(&core.TextField{Name: "email"})
			changed = true
		}
		if col.Fields.GetByName("source") == nil {
			col.Fields.Add(&core.TextField{Name: "source"})
			changed = true
		}
		if changed {
			_ = app.Save(col)
		}
	}
	return nil
}

// GetSiteSettings retrieves settings from In-Memory cache or PocketBase
func GetSiteSettings(app core.App) models.SiteSettings {
	c := cache.GetStore()
	if cached, ok := c.GetSettings(); ok {
		return cached
	}

	settings := models.SiteSettings{
		SiteName:     "Нотариальная контора",
		Tagline:      "Официальные нотариальные действия",
		Theme:        "legal",
		EnableA11y:   true,
		EnableAlpine: false,
		MapMode:      "lazy",
		BookingMode:  "slots",
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
		// Support both owner_name and legacy notary_name
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
