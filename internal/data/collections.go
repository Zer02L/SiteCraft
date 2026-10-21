package data

import (
	"log"
	"os"

	"sitecraft/internal/models"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

func initSiteSettingsCollection(app core.App) error {
	col, err := app.FindCollectionByNameOrId("site_settings")
	if err != nil {
		col = core.NewBaseCollection("site_settings")
		col.Fields.Add(
			&core.TextField{Name: "site_name", Required: true},
			&core.TextField{Name: "tagline"},
			&core.TextField{Name: "theme"}, // legal, medical, auto, tech, default, custom
			&core.TextField{Name: "primary_color"},
			&core.TextField{Name: "accent_color"},
			&core.TextField{Name: "background_color"},
			&core.TextField{Name: "font_family"},
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
			&core.BoolField{Name: "enable_payment"},
			&core.TextField{Name: "payment_shop_id"},
			&core.TextField{Name: "payment_secret"},
			&core.BoolField{Name: "payment_sandbox"},
			&core.BoolField{Name: "enable_blog"},
			&core.BoolField{Name: "enable_tma"},
			&core.BoolField{Name: "enable_tracker"},
			&core.BoolField{Name: "enable_i18n"},
			&core.BoolField{Name: "enable_floating_contact"},
			&core.TextField{Name: "whatsapp_phone"},
			&core.TextField{Name: "telegram_contact"},
			&core.BoolField{Name: "enable_cart"},
			&core.TextField{Name: "header_variant"},
			&core.TextField{Name: "footer_variant"},
			&core.BoolField{Name: "enable_live_chat"},
			&core.BoolField{Name: "enable_fleet"},
			&core.BoolField{Name: "enable_recent_activity"},
			&core.TextField{Name: "client_admin_password"},
			&core.TextField{Name: "client_admin_password_hash"},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция site_settings успешно создана")
	} else {
		// Ensure payment, theming, blog, TMA, i18n, floating contact, cart, chat, and fleet fields exist
		changed := false
		if col.Fields.GetByName("primary_color") == nil {
			col.Fields.Add(&core.TextField{Name: "primary_color"})
			col.Fields.Add(&core.TextField{Name: "accent_color"})
			col.Fields.Add(&core.TextField{Name: "background_color"})
			col.Fields.Add(&core.TextField{Name: "font_family"})
			changed = true
		}
		if col.Fields.GetByName("enable_payment") == nil {
			col.Fields.Add(&core.BoolField{Name: "enable_payment"})
			col.Fields.Add(&core.TextField{Name: "payment_shop_id"})
			col.Fields.Add(&core.TextField{Name: "payment_secret"})
			col.Fields.Add(&core.BoolField{Name: "payment_sandbox"})
			changed = true
		}
		if col.Fields.GetByName("enable_blog") == nil {
			col.Fields.Add(&core.BoolField{Name: "enable_blog"})
			changed = true
		}
		if col.Fields.GetByName("enable_tma") == nil {
			col.Fields.Add(&core.BoolField{Name: "enable_tma"})
			changed = true
		}
		if col.Fields.GetByName("enable_tracker") == nil {
			col.Fields.Add(&core.BoolField{Name: "enable_tracker"})
			changed = true
		}
		if col.Fields.GetByName("enable_i18n") == nil {
			col.Fields.Add(&core.BoolField{Name: "enable_i18n"})
			col.Fields.Add(&core.TextField{Name: "default_lang"})
			changed = true
		}
		if col.Fields.GetByName("enable_floating_contact") == nil {
			col.Fields.Add(&core.BoolField{Name: "enable_floating_contact"})
			col.Fields.Add(&core.TextField{Name: "whatsapp_phone"})
			col.Fields.Add(&core.TextField{Name: "telegram_contact"})
			col.Fields.Add(&core.BoolField{Name: "enable_cart"})
			changed = true
		}
		if col.Fields.GetByName("enable_live_chat") == nil {
			col.Fields.Add(&core.BoolField{Name: "enable_live_chat"})
			changed = true
		}
		if col.Fields.GetByName("enable_fleet") == nil {
			col.Fields.Add(&core.BoolField{Name: "enable_fleet"})
			changed = true
		}
		if col.Fields.GetByName("enable_recent_activity") == nil {
			col.Fields.Add(&core.BoolField{Name: "enable_recent_activity"})
			changed = true
		}
		if col.Fields.GetByName("header_variant") == nil {
			col.Fields.Add(&core.TextField{Name: "header_variant"})
			col.Fields.Add(&core.TextField{Name: "footer_variant"})
			changed = true
		}
		if col.Fields.GetByName("client_admin_password") == nil {
			col.Fields.Add(&core.TextField{Name: "client_admin_password"})
			changed = true
		}
		if col.Fields.GetByName("client_admin_password_hash") == nil {
			col.Fields.Add(&core.TextField{Name: "client_admin_password_hash"})
			changed = true
		}
		if col.Fields.GetByName("legal_entity") == nil {
			col.Fields.Add(&core.TextField{Name: "legal_entity"})
			col.Fields.Add(&core.TextField{Name: "legal_address"})
			col.Fields.Add(&core.TextField{Name: "inn"})
			col.Fields.Add(&core.TextField{Name: "kpp"})
			col.Fields.Add(&core.TextField{Name: "ogrn"})
			col.Fields.Add(&core.TextField{Name: "bank_name"})
			col.Fields.Add(&core.TextField{Name: "bank_bik"})
			col.Fields.Add(&core.TextField{Name: "bank_account"})
			col.Fields.Add(&core.TextField{Name: "bank_corr"})
			col.Fields.Add(&core.TextField{Name: "contact_email"})
			col.Fields.Add(&core.BoolField{Name: "enable_careers"})
			col.Fields.Add(&core.BoolField{Name: "enable_cookie_consent"})
			changed = true
		}
		if changed {
			_ = app.Save(col)
		}
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
		record.Set("theme", "main")
		record.Set("enable_a11y", true)
		record.Set("enable_alpine", false) // Zero-JS по умолчанию!
		record.Set("enable_turnstile", false)
		record.Set("map_mode", "lazy")
		record.Set("booking_mode", "slots")
		record.Set("enable_payment", false)
		record.Set("payment_sandbox", true)

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
			&core.TextField{Name: "hero_title_b"},
			&core.TextField{Name: "hero_subtitle_b"},
			&core.TextField{Name: "hero_cta_b"},
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
		changed := false
		if col.Fields.GetByName("hero_title_b") == nil {
			col.Fields.Add(&core.TextField{Name: "hero_title_b"})
			col.Fields.Add(&core.TextField{Name: "hero_subtitle_b"})
			col.Fields.Add(&core.TextField{Name: "hero_cta_b"})
			changed = true
		}
		if changed {
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

func initPageBlocksCollection(app core.App) error {
	col, err := app.FindCollectionByNameOrId("page_blocks")
	if err != nil {
		col = core.NewBaseCollection("page_blocks")
		col.Fields.Add(
			&core.TextField{Name: "type", Required: true}, // hero, schedule, services, booking, map, rich_text, html
			&core.TextField{Name: "title"},
			&core.TextField{Name: "content"},
			&core.NumberField{Name: "order", OnlyInt: true},
			&core.BoolField{Name: "active"},
			&core.JSONField{Name: "payload"},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция page_blocks успешно создана")
	} else {
		// Убеждаемся, что поле payload существует у существующей коллекции
		if col.Fields.GetByName("payload") == nil {
			col.Fields.Add(&core.JSONField{Name: "payload"})
			if err := app.Save(col); err != nil {
				log.Printf("[Migration Error] Не удалось добавить поле payload в page_blocks: %v\n", err)
			} else {
				log.Println("[Migration] Поле payload успешно добавлено в схему page_blocks")
			}
		}
	}

	count, _ := app.CountRecords("page_blocks")
	if count == 0 {
		col, err := app.FindCollectionByNameOrId("page_blocks")
		if err != nil {
			return err
		}

		// Дефолтные секции для модульного конструктора
		defaultBlocks := []struct {
			Type    string
			Title   string
			Content string
			Order   int
		}{
			{Type: "hero", Title: "Главный экран", Order: 1},
			{Type: "schedule", Title: "График работы и контакты", Order: 2},
			{Type: "services", Title: "Услуги и цены", Order: 3},
			{Type: "booking", Title: "Онлайн-запись / Лидген", Order: 4},
			{Type: "map", Title: "Интерактивная карта", Order: 5},
		}

		for _, b := range defaultBlocks {
			rec := core.NewRecord(col)
			rec.Set("type", b.Type)
			rec.Set("title", b.Title)
			rec.Set("content", b.Content)
			rec.Set("order", b.Order)
			rec.Set("active", true)
			_ = app.Save(rec)
		}
		log.Println("[Migration] Базовые блоки страницы созданы в page_blocks")
	}
	return nil
}

func initServicesCollection(app core.App) error {
	col, err := app.FindCollectionByNameOrId("services")
	if err != nil {
		col = core.NewBaseCollection("services")
		col.Fields.Add(
			&core.TextField{Name: "title", Required: true},
			&core.TextField{Name: "description"},
			&core.TextField{Name: "price"},
			&core.NumberField{Name: "order", OnlyInt: true},
			&core.BoolField{Name: "active"},
			&core.NumberField{Name: "stock_count", OnlyInt: true},
			&core.BoolField{Name: "is_preorder"},
			&core.BoolField{Name: "is_deleted"},
			&core.TextField{Name: "deleted_at"},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция services успешно создана")
	} else {
		changed := false
		if col.Fields.GetByName("stock_count") == nil {
			col.Fields.Add(&core.NumberField{Name: "stock_count", OnlyInt: true})
			changed = true
		}
		if col.Fields.GetByName("is_preorder") == nil {
			col.Fields.Add(&core.BoolField{Name: "is_preorder"})
			changed = true
		}
		if col.Fields.GetByName("is_deleted") == nil {
			col.Fields.Add(&core.BoolField{Name: "is_deleted"})
			changed = true
		}
		if col.Fields.GetByName("deleted_at") == nil {
			col.Fields.Add(&core.TextField{Name: "deleted_at"})
			changed = true
		}
		if changed {
			_ = app.Save(col)
		}
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
		}

		for _, s := range defaultServices {
			rec := core.NewRecord(col)
			rec.Set("title", s.Title)
			rec.Set("description", s.Description)
			rec.Set("price", s.Price)
			rec.Set("order", s.Order)
			rec.Set("active", true)
			_ = app.Save(rec)
		}
		log.Println("[Migration] Базовые услуги успешно добавлены в services")
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
			&core.TextField{Name: "resource_id"},
			&core.TextField{Name: "status"}, // например, "новая", "обработана", "отменена"
			&core.JSONField{Name: "payload"}, // динамические поля: action_type, comment, vin, etc.
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция leads для онлайн-заявок успешно создана")
	} else {
		changed := false
		if col.Fields.GetByName("resource_id") == nil {
			col.Fields.Add(&core.TextField{Name: "resource_id"})
			changed = true
		}
		if col.Fields.GetByName("created") == nil {
			col.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
			changed = true
		}
		if col.Fields.GetByName("updated") == nil {
			col.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
			changed = true
		}
		if changed {
			_ = app.Save(col)
		}
	}
	return nil
}

func initArticlesCollection(app core.App) error {
	col, err := app.FindCollectionByNameOrId("articles")
	if err != nil {
		col = core.NewBaseCollection("articles")
		col.Fields.Add(
			&core.TextField{Name: "title", Required: true},
			&core.TextField{Name: "slug", Required: true},
			&core.TextField{Name: "excerpt"},
			&core.EditorField{Name: "content"},
			&core.FileField{Name: "cover_image"},
			&core.TextField{Name: "author_name"},
			&core.TextField{Name: "published_at"},
			&core.BoolField{Name: "is_published"},
			&core.BoolField{Name: "is_deleted"},
			&core.TextField{Name: "deleted_at"},
			&core.NumberField{Name: "views_count", OnlyInt: true},
			&core.TextField{Name: "seo_title"},
			&core.TextField{Name: "seo_description"},
			&core.TextField{Name: "tags"},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция articles для SEO-блога успешно создана")
	} else {
		changed := false
		if col.Fields.GetByName("is_deleted") == nil {
			col.Fields.Add(&core.BoolField{Name: "is_deleted"})
			changed = true
		}
		if col.Fields.GetByName("deleted_at") == nil {
			col.Fields.Add(&core.TextField{Name: "deleted_at"})
			changed = true
		}
		if col.Fields.GetByName("created") == nil {
			col.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
			changed = true
		}
		if col.Fields.GetByName("updated") == nil {
			col.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
			changed = true
		}
		if changed {
			_ = app.Save(col)
		}
	}

	count, _ := app.CountRecords("articles")
	if count == 0 {
		col, err := app.FindCollectionByNameOrId("articles")
		if err != nil {
			return err
		}

		defaultArticles := []struct {
			Title   string
			Slug    string
			Excerpt string
			Content string
			Author  string
			Date    string
			Tags    string
			Views   int
		}{
			{
				Title:   "Как правильно подготовить документы для сделки с недвижимостью",
				Slug:    "kak-podgotovit-dokumenty-dlya-sdelki",
				Excerpt: "Полное пошаговое руководство от экспертов: перечень обязательных справок, типичные юридические ошибки и сроки действия документов.",
				Content: `<h2>1. Базовый пакет документов</h2><p>При совершении любых юридически значимых действий с недвижимым имуществом критически важно заблаговременно проверить актуальность правоустанавливающих документов.</p><h3>Обязательный перечень:</h3><ul><li>Выписка из ЕГРН об основных характеристиках и зарегистрированных правах (срок действия — 30 календарных дней).</li><li>Договор-основание (купли-продажи, дарения, приватизации или свидетельство о праве на наследство).</li><li>Нотариально удостоверенное согласие супруга (при наличии зарегистрированного брака).</li></ul><h2>2. Риски и подводные камни</h2><p>Особое внимание следует уделить наличию скрытых обременений, несовершеннолетних собственников и задолженностей по коммунальным платежам. Своевременное нотариальное удостоверение договора гарантирует чистоту сделки и защищает стороны от судебных претензий третьих лиц.</p>`,
				Author:  "Владислав Захаров",
				Date:    "2026-10-01",
				Tags:    "Недвижимость, Советы, Право",
				Views:   142,
			},
			{
				Title:   "Электронная регистрация и цифровые нотариальные действия в 2026 году",
				Slug:    "elektronnaya-registratsiya-notariat-2026",
				Excerpt: "Как современные технологии экономят ваше время: удаленное удостоверение сделок, передача заявлений в Росреестр и цифровая подпись.",
				Content: `<h2>Цифровизация юридических процедур</h2><p>В 2026 году нотариат полностью интегрирован с государственными цифровыми сервисами. Граждане и представители бизнеса могут совершать большинство действий без очередей в ведомствах.</p><h3>Преимущества электронной подачи:</h3><p>Нотариус самостоятельно направляет пакет документов в Росреестр и налоговые органы в электронном виде с использованием квалифицированной ЭЦП. Регистрация прав собственности происходит всего за 1 рабочий день.</p>`,
				Author:  "Юридический отдел",
				Date:    "2026-10-05",
				Tags:    "Цифровизация, Технологии, Нотариат",
				Views:   89,
			},
		}

		for _, a := range defaultArticles {
			rec := core.NewRecord(col)
			rec.Set("title", a.Title)
			rec.Set("slug", a.Slug)
			rec.Set("excerpt", a.Excerpt)
			rec.Set("content", a.Content)
			rec.Set("author_name", a.Author)
			rec.Set("published_at", a.Date)
			rec.Set("is_published", true)
			rec.Set("views_count", a.Views)
			rec.Set("seo_title", a.Title)
			rec.Set("seo_description", a.Excerpt)
			rec.Set("tags", a.Tags)
			_ = app.Save(rec)
		}
		log.Println("[Migration] Демонстрационные экспертные статьи успешно добавлены в articles")
	}

	return nil
}

func initServiceOrdersCollection(app core.App) error {
	col, err := app.FindCollectionByNameOrId("service_orders")
	if err != nil {
		col = core.NewBaseCollection("service_orders")
		col.Fields.Add(
			&core.TextField{Name: "order_code", Required: true},
			&core.TextField{Name: "client_name"},
			&core.TextField{Name: "client_phone"},
			&core.TextField{Name: "item_title", Required: true},
			&core.TextField{Name: "current_stage"}, // received, diagnostics, in_progress, ready, completed
			&core.NumberField{Name: "total_amount"},
			&core.BoolField{Name: "is_paid"},
			&core.JSONField{Name: "stages_history"},
			&core.TextField{Name: "notes"},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция service_orders для B2B-трекера успешно создана")
	}

	count, _ := app.CountRecords("service_orders")
	if count == 0 {
		col, err := app.FindCollectionByNameOrId("service_orders")
		if err != nil {
			return err
		}

		demoHistory := []models.OrderStageHistory{
			{Stage: "received", Title: "Автомобиль принят мастером", Timestamp: "06.10 09:15", Comment: "Пробег 84 200 км, визуальный осмотр кузова", Done: true},
			{Stage: "diagnostics", Title: "Компьютерная диагностика и дефектовка", Timestamp: "06.10 11:30", Comment: "Выявлен износ тормозных колодок и замена масла", Done: true},
			{Stage: "in_progress", Title: "Выполнение сервисных работ", Timestamp: "07.10 10:00", Comment: "Замена масла в двигателе, установка оригинальных колодок", Done: false},
			{Stage: "ready", Title: "Готов к выдаче", Timestamp: "", Comment: "Автомобиль на чистке и готов к передаче владельцу", Done: false},
		}

		rec := core.NewRecord(col)
		rec.Set("order_code", "SRV-4182")
		rec.Set("client_name", "Константин")
		rec.Set("client_phone", "+7 (999) 123-45-67")
		rec.Set("item_title", "BMW X5 3.0d (х819ок123)")
		rec.Set("current_stage", "in_progress")
		rec.Set("total_amount", 14500.0)
		rec.Set("is_paid", false)
		rec.Set("stages_history", demoHistory)
		rec.Set("notes", "Комплексное ТО и замена расходников")
		_ = app.Save(rec)

		// Второй тестовый заказ со статусом ready для проверки оплаты
		demoReadyHistory := []models.OrderStageHistory{
			{Stage: "received", Title: "Устройство принято в СЦ", Timestamp: "06.10 14:00", Comment: "iPhone 14 Pro, трещина стекла дисплея", Done: true},
			{Stage: "diagnostics", Title: "Инженерная диагностика", Timestamp: "06.10 15:20", Comment: "Матрица цела, требуется замена верхнего стекла", Done: true},
			{Stage: "in_progress", Title: "Замена стекла на OCA-пленку", Timestamp: "07.10 11:00", Comment: "Процедура выполнена на заводском оборудовании", Done: true},
			{Stage: "ready", Title: "Готов к выдаче клиенту", Timestamp: "07.10 12:45", Comment: "Устройство собрано, проверено, готово к выдаче", Done: true},
		}
		recReady := core.NewRecord(col)
		recReady.Set("order_code", "FIX-9021")
		recReady.Set("client_name", "Елена")
		recReady.Set("client_phone", "+7 (918) 765-43-21")
		recReady.Set("item_title", "iPhone 14 Pro (Deep Purple)")
		recReady.Set("current_stage", "ready")
		recReady.Set("total_amount", 6800.0)
		recReady.Set("is_paid", false)
		recReady.Set("stages_history", demoReadyHistory)
		recReady.Set("notes", "Замена стекла экрана по заводской технологии")
		_ = app.Save(recReady)

		log.Println("[Migration] Демонстрационные заказы SRV-4182 и FIX-9021 успешно созданы")
	}

	return nil
}

func initChatMessagesCollection(app core.App) error {
	col, err := app.FindCollectionByNameOrId("chat_messages")
	if err != nil {
		col = core.NewBaseCollection("chat_messages")
		col.Fields.Add(
			&core.TextField{Name: "session_id", Required: true},
			&core.TextField{Name: "sender", Required: true}, // "visitor" or "operator"
			&core.TextField{Name: "message", Required: true},
			&core.NumberField{Name: "tg_msg_id"},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция chat_messages для Live Web-to-TG чата успешно создана")
	}
	return nil
}

func initCouponsCollection(app core.App) error {
	col, err := app.FindCollectionByNameOrId("coupons")
	if err != nil {
		col = core.NewBaseCollection("coupons")
		col.Fields.Add(
			&core.TextField{Name: "code", Required: true},
			&core.TextField{Name: "discount_type", Required: true}, // "percent" or "fixed"
			&core.NumberField{Name: "discount_value", Required: true},
			&core.NumberField{Name: "min_order_amount"},
			&core.TextField{Name: "valid_until"},
			&core.NumberField{Name: "usage_limit", OnlyInt: true},
			&core.NumberField{Name: "used_count", OnlyInt: true},
			&core.BoolField{Name: "active"},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция coupons успешно создана")
	}

	count, _ := app.CountRecords("coupons")
	if count == 0 {
		col, err := app.FindCollectionByNameOrId("coupons")
		if err != nil {
			return err
		}

		c1 := core.NewRecord(col)
		c1.Set("code", "PROMO10")
		c1.Set("discount_type", "percent")
		c1.Set("discount_value", 10.0)
		c1.Set("min_order_amount", 1000.0)
		c1.Set("usage_limit", 100)
		c1.Set("used_count", 0)
		c1.Set("active", true)
		_ = app.Save(c1)

		c2 := core.NewRecord(col)
		c2.Set("code", "WELCOME500")
		c2.Set("discount_type", "fixed")
		c2.Set("discount_value", 500.0)
		c2.Set("min_order_amount", 2500.0)
		c2.Set("usage_limit", 50)
		c2.Set("used_count", 0)
		c2.Set("active", true)
		_ = app.Save(c2)

		log.Println("[Migration] Демонстрационные промокоды PROMO10 и WELCOME500 созданы")
	}
	return nil
}

func initResourcesCollection(app core.App) error {
	col, err := app.FindCollectionByNameOrId("resources")
	if err != nil {
		col = core.NewBaseCollection("resources")
		col.Fields.Add(
			&core.TextField{Name: "name", Required: true},
			&core.TextField{Name: "role", Required: true},
			&core.TextField{Name: "avatar"},
			&core.BoolField{Name: "active"},
			&core.NumberField{Name: "order", OnlyInt: true},
			&core.TextField{Name: "service_ids"},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция resources для записи к специалистам успешно создана")
	}

	count, _ := app.CountRecords("resources")
	if count == 0 {
		col, err := app.FindCollectionByNameOrId("resources")
		if err != nil {
			return err
		}

		r1 := core.NewRecord(col)
		r1.Set("name", "Захаров Владислав Борисович")
		r1.Set("role", "Нотариус / Ведущий специалист")
		r1.Set("avatar", "👨‍⚖️")
		r1.Set("active", true)
		r1.Set("order", 1)
		_ = app.Save(r1)

		r2 := core.NewRecord(col)
		r2.Set("name", "Васильева Елена Сергеевна")
		r2.Set("role", "Помощник нотариуса / Сделки и доверенности")
		r2.Set("avatar", "👩‍💼")
		r2.Set("active", true)
		r2.Set("order", 2)
		_ = app.Save(r2)

		log.Println("[Migration] Демонстрационные специалисты добавлены в resources")
	}
	return nil
}

func initReviewsCollection(app core.App) error {
	col, err := app.FindCollectionByNameOrId("reviews")
	if err != nil {
		col = core.NewBaseCollection("reviews")
		col.Fields.Add(
			&core.TextField{Name: "author_name", Required: true},
			&core.NumberField{Name: "rating", OnlyInt: true, Required: true},
			&core.TextField{Name: "comment", Required: true},
			&core.TextField{Name: "order_id"},
			&core.TextField{Name: "photo"},
			&core.BoolField{Name: "is_published"},
			&core.TextField{Name: "source"},
			&core.AutodateField{Name: "created", OnCreate: true},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция reviews для проверенных отзывов успешно создана")
	}

	count, _ := app.CountRecords("reviews")
	if count == 0 {
		col, err := app.FindCollectionByNameOrId("reviews")
		if err != nil {
			return err
		}

		demoReviews := []struct {
			Author  string
			Rating  int
			Comment string
			OrderID string
			Source  string
		}{
			{
				Author:  "Константин Мельников",
				Rating:  5,
				Comment: "Все оформили быстро и профессионально. Сделка по недвижимости прошла идеально, без задержек.",
				OrderID: "SRV-4182",
				Source:  "track",
			},
			{
				Author:  "Анна Сергеева",
				Rating:  5,
				Comment: "Записалась через сайт на конкретное время к специалисту, приняли ровно по записи без очереди. Спасибо за уважение к времени!",
				OrderID: "",
				Source:  "site",
			},
			{
				Author:  "Дмитрий В.",
				Rating:  5,
				Comment: "Грамотные консультации, помогли с оформлением сложного наследственного дела. Рекомендую.",
				OrderID: "FIX-9021",
				Source:  "track",
			},
		}

		for _, rev := range demoReviews {
			r := core.NewRecord(col)
			r.Set("author_name", rev.Author)
			r.Set("rating", rev.Rating)
			r.Set("comment", rev.Comment)
			r.Set("order_id", rev.OrderID)
			r.Set("is_published", true)
			r.Set("source", rev.Source)
			_ = app.Save(r)
		}
		log.Println("[Migration] Начальные проверенные отзывы успешно созданы в reviews")
	}
	return nil
}

func initBranchesCollection(app core.App) error {
	col, err := app.FindCollectionByNameOrId("branches")
	if err != nil {
		col = core.NewBaseCollection("branches")
		col.Fields.Add(
			&core.TextField{Name: "name", Required: true},
			&core.TextField{Name: "city", Required: true},
			&core.TextField{Name: "address", Required: true},
			&core.TextField{Name: "phone", Required: true},
			&core.TextField{Name: "work_hours"},
			&core.NumberField{Name: "lat"},
			&core.NumberField{Name: "lng"},
			&core.TextField{Name: "metro"},
			&core.TextField{Name: "services"},
			&core.BoolField{Name: "is_main"},
			&core.BoolField{Name: "active"},
			&core.NumberField{Name: "order", OnlyInt: true},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция branches для сети филиалов успешно создана")
	}

	count, _ := app.CountRecords("branches")
	if count == 0 {
		col, err := app.FindCollectionByNameOrId("branches")
		if err != nil {
			return err
		}

		b1 := core.NewRecord(col)
		b1.Set("name", "Филиал «Центральный»")
		b1.Set("city", "Санкт-Петербург")
		b1.Set("address", "Невский пр-т, д. 42")
		b1.Set("phone", "+7 (812) 309-88-21")
		b1.Set("work_hours", "Пн-Пт: 09:00 — 21:00, Сб-Вс: 10:00 — 19:00")
		b1.Set("lat", 59.934280)
		b1.Set("lng", 30.335099)
		b1.Set("metro", "Гостиный двор")
		b1.Set("services", "Комплексное ТО, Экспресс-диагностика, Оформление сделок")
		b1.Set("is_main", true)
		b1.Set("active", true)
		b1.Set("order", 1)
		_ = app.Save(b1)

		b2 := core.NewRecord(col)
		b2.Set("name", "Филиал «Северный»")
		b2.Set("city", "Санкт-Петербург")
		b2.Set("address", "пр. Просвещения, д. 80, корп. 1")
		b2.Set("phone", "+7 (812) 309-88-22")
		b2.Set("work_hours", "Ежедневно: 09:00 — 21:00")
		b2.Set("lat", 60.038560)
		b2.Set("lng", 30.395420)
		b2.Set("metro", "Гражданский проспект")
		b2.Set("services", "Кузовной цех, Компьютерная диагностика, Замена масел")
		b2.Set("is_main", false)
		b2.Set("active", true)
		b2.Set("order", 2)
		_ = app.Save(b2)

		b3 := core.NewRecord(col)
		b3.Set("name", "Филиал «Московский»")
		b3.Set("city", "Москва")
		b3.Set("address", "Ленинградский пр-т, д. 36, стр. 2")
		b3.Set("phone", "+7 (495) 710-99-44")
		b3.Set("work_hours", "Пн-Сб: 08:30 — 20:30, Вс: выходной")
		b3.Set("lat", 55.791550)
		b3.Set("lng", 37.558310)
		b3.Set("metro", "Динамо")
		b3.Set("services", "Полный цикл сервиса, Корпоративное обслуживание B2B, Нотариат")
		b3.Set("is_main", false)
		b3.Set("active", true)
		b3.Set("order", 3)
		_ = app.Save(b3)

		log.Println("[Migration] Демонстрационные филиалы успешно созданы в branches")
	}
	return nil
}

func initVacanciesCollection(app core.App) error {
	col, err := app.FindCollectionByNameOrId("vacancies")
	if err != nil {
		col = core.NewBaseCollection("vacancies")
		col.Fields.Add(
			&core.TextField{Name: "title", Required: true},
			&core.TextField{Name: "department"},
			&core.TextField{Name: "salary"},
			&core.TextField{Name: "experience"},
			&core.TextField{Name: "schedule"},
			&core.JSONField{Name: "responsibilities"},
			&core.JSONField{Name: "requirements"},
			&core.JSONField{Name: "benefits"},
			&core.BoolField{Name: "active"},
			&core.NumberField{Name: "order", OnlyInt: true},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция vacancies для HR-модуля успешно создана")
	}

	count, _ := app.CountRecords("vacancies")
	if count == 0 {
		col, err := app.FindCollectionByNameOrId("vacancies")
		if err != nil {
			return err
		}

		v1 := core.NewRecord(col)
		v1.Set("title", "Мастер-приемщик / Сервисный консультант")
		v1.Set("department", "Сервисный отдел")
		v1.Set("salary", "от 110 000 до 160 000 ₽")
		v1.Set("experience", "от 2 лет в аналогичной сфере")
		v1.Set("schedule", "Сменный график 2/2, 09:00 — 21:00")
		v1.Set("responsibilities", []string{
			"Прием клиентов, первичная дефектовка и согласование перечня работ",
			"Контроль соблюдения регламентных сроков выполнения заказов",
			"Оформление заказ-нарядов, счетов и актов выполненных работ",
			"Выдача готовой техники и консультирование по рекомендациям",
		})
		v1.Set("requirements", []string{
			"Грамотная речь, клиентоориентированность, вежливость",
			"Знание устройства современных систем и регламентов ТО",
			"Опыт работы в профильных CRM / 1С / учетных системах",
		})
		v1.Set("benefits", []string{
			"Официальное оформление по ТК РФ с первого дня",
			"Выплаты 2 раза в месяц без задержек + KPI за объем",
			"Теплый современный офис, корпоративная форма, скидки на услуги",
		})
		v1.Set("active", true)
		v1.Set("order", 1)
		_ = app.Save(v1)

		v2 := core.NewRecord(col)
		v2.Set("title", "Специалист по ремонту / Автоэлектрик-диагност")
		v2.Set("department", "Технический цех")
		v2.Set("salary", "от 130 000 до 220 000 ₽")
		v2.Set("experience", "от 3 лет практического стажа")
		v2.Set("schedule", "Сменный график 3/3, 09:00 — 21:00")
		v2.Set("responsibilities", []string{
			"Компьютерная диагностика электронных систем и датчиков",
			"Поиск обрывов проводки, ремонт блоков управления и CAN-шин",
			"Установка дополнительного оборудования и систем мониторинга",
		})
		v2.Set("requirements", []string{
			"Профильное техническое образование",
			"Уверенные навыки чтения электрических схем",
			"Ответственность, чистоплотность и аккуратность в работе",
		})
		v2.Set("benefits", []string{
			"Индивидуальный рабочий пост с передовым сканером и инструментарием",
			"Обучение за счет компании и повышение квалификации",
			"Спецодежда с еженедельной химчисткой, душевая, комната отдыха",
		})
		v2.Set("active", true)
		v2.Set("order", 2)
		_ = app.Save(v2)

		log.Println("[Migration] Демонстрационные вакансии успешно созданы в vacancies")
	}
	return nil
}

