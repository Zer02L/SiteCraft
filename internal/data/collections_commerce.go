package data

import (
	"log"

	"sitecraft/internal/models"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

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
