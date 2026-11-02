package data

import (
	"log"
	"os"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/filesystem"
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
			&core.SelectField{
				Name:      "header_variant",
				Values:    []string{"classic", "minimal", "centered"},
				MaxSelect: 1,
			},
			&core.SelectField{
				Name:      "footer_variant",
				Values:    []string{"classic", "minimal", "columns"},
				MaxSelect: 1,
			},
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
		if f := col.Fields.GetByName("header_variant"); f == nil {
			col.Fields.Add(&core.SelectField{
				Name:      "header_variant",
				Values:    []string{"classic", "minimal", "centered"},
				MaxSelect: 1,
			})
			changed = true
		} else if _, ok := f.(*core.SelectField); !ok {
			col.Fields.RemoveByName("header_variant")
			col.Fields.Add(&core.SelectField{
				Name:      "header_variant",
				Values:    []string{"classic", "minimal", "centered"},
				MaxSelect: 1,
			})
			changed = true
		}
		if f := col.Fields.GetByName("footer_variant"); f == nil {
			col.Fields.Add(&core.SelectField{
				Name:      "footer_variant",
				Values:    []string{"classic", "minimal", "columns"},
				MaxSelect: 1,
			})
			changed = true
		} else if _, ok := f.(*core.SelectField); !ok {
			col.Fields.RemoveByName("footer_variant")
			col.Fields.Add(&core.SelectField{
				Name:      "footer_variant",
				Values:    []string{"classic", "minimal", "columns"},
				MaxSelect: 1,
			})
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
