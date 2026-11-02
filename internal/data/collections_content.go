package data

import (
	"log"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

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
