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
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция site_settings успешно создана")
	} else {
		// Ensure payment, theming, blog, TMA, i18n, floating contact, and cart fields exist
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
		if col.Fields.GetByName("header_variant") == nil {
			col.Fields.Add(&core.TextField{Name: "header_variant"})
			col.Fields.Add(&core.TextField{Name: "footer_variant"})
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
			&core.TextField{Name: "status"}, // например, "новая", "обработана", "отменена"
			&core.JSONField{Name: "payload"}, // динамические поля: action_type, comment, vin, etc.
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция leads для онлайн-заявок успешно создана")
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
			&core.NumberField{Name: "views_count", OnlyInt: true},
			&core.TextField{Name: "seo_title"},
			&core.TextField{Name: "seo_description"},
			&core.TextField{Name: "tags"},
		)

		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("[Migration] Коллекция articles для SEO-блога успешно создана")
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
