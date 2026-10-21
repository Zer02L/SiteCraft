package registry

import (
	"encoding/json"
	"fmt"
	"sitecraft/internal/models"
)

// ParsePayload converts a raw block payload map into a typed struct T pointer.
// Returns nil if payload is nil or JSON unmarshaling fails.
func ParsePayload[T any](payload map[string]interface{}) *T {
	if payload == nil {
		return nil
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	var res T
	if err := json.Unmarshal(b, &res); err != nil {
		return nil
	}
	return &res
}

// ParseFeaturesPayload converts raw block payload map into typed FeaturesPayload
func ParseFeaturesPayload(payload map[string]interface{}) *models.FeaturesPayload {
	return ParsePayload[models.FeaturesPayload](payload)
}

// ParseFAQPayload converts raw block payload map into typed FAQPayload
func ParseFAQPayload(payload map[string]interface{}) *models.FAQPayload {
	return ParsePayload[models.FAQPayload](payload)
}

// ParseStepsPayload converts raw block payload map into typed StepsPayload
func ParseStepsPayload(payload map[string]interface{}) *models.StepsPayload {
	return ParsePayload[models.StepsPayload](payload)
}

// ParseQuizPayload converts raw block payload map into typed QuizPayload with intelligent niche defaults
func ParseQuizPayload(payload map[string]interface{}, theme string) models.QuizPayload {
	if parsed := ParsePayload[models.QuizPayload](payload); parsed != nil && len(parsed.Steps) > 0 {
		if parsed.Currency == "" {
			parsed.Currency = "₽"
		}
		return *parsed
	}

	// Дефолтные конфигурации по нишам
	switch theme {
	case "auto", "repair":
		return models.QuizPayload{
			Badge:     "Калькулятор ТО и ремонта",
			Title:     "Рассчитайте точную стоимость работ за 1 минуту",
			Subtitle:  "Ответьте на 3 вопроса о вашем автомобиле и получите персональную скидку 10% на первое посещение",
			BasePrice: 1500,
			Currency:  "₽",
			BonusText: "Зафиксируем скидку 10% на запчасти и работу за вашим номером телефона.",
			Steps: []models.QuizStep{
				{
					ID:       "service_type",
					Question: "Какая услуга требуется автомобилю?",
					Subtitle: "Выберите основной вид работ",
					Options: QuizOptionDefault("Комплексное ТО и замена масла", "Замена моторного масла и всех фильтров", 1500,
						"Диагностика подвески и тормозов", "Компьютерное сканирование + подъемник", 1200,
						"Ремонт ходовой / двигателя", "Дефектовка узлов и устранение стуков", 3500,
						"Автоэлектрика и кондиционер", "Поиск утечек, заправка фреоном, датчики", 2000),
				},
				{
					ID:       "car_class",
					Question: "Класс вашего автомобиля",
					Subtitle: "Влияет на объем расходных материалов",
					Options: QuizOptionDefault("Легковой седан / хэтчбек", "Стандартный городской класс", 0,
						"Кроссовер / Внедорожник (SUV)", "Повышенный объем масла и фильтров", 800,
						"Коммерческий микроавтобус / пикап", "Тяжелая техника", 1500),
				},
				{
					ID:       "urgency",
					Question: "Когда планируете визит на СТО?",
					Subtitle: "Поможем забронировать свободный подъемник",
					Options: QuizOptionDefault("Срочно сегодня или завтра", "Приоритетный заезд вне очереди", 500,
						"В течение 3-5 дней", "В удобное согласованное время", 0,
						"В ближайшие выходные", "Суббота или воскресенье", 0),
				},
			},
		}

	case "beauty":
		return models.QuizPayload{
			Badge:     "Бьюти-калькулятор",
			Title:     "Подберите идеальный уход и рассчитайте стоимость",
			Subtitle:  "Пройдите быстрый тест и получите подарок: SPA-уход для рук при первой записи",
			BasePrice: 2000,
			Currency:  "₽",
			BonusText: "Промокод на SPA-уход и скидка 15% будут зафиксированы за вашим номером.",
			Steps: []models.QuizStep{
				{
					ID:       "procedure",
					Question: "Какая процедура вас интересует?",
					Options: QuizOptionDefault("Сложное окрашивание / Airtouch", "Премиальные красители и тонирование", 4500,
						"Стрижка и укладка волос", "Авторская форма и термозащита", 1500,
						"Маникюр и педикюр с покрытием", "Аппаратная обработка, гель-лак", 2200,
						"SPA-уход и восстановление", "Глубокое питание и массаж", 2000),
				},
				{
					ID:       "master_level",
					Question: "Категория мастера",
					Options: QuizOptionDefault("Топ-стилист / Арт-директор", "Опыт от 8 лет, победитель конкурсов", 1200,
						"Ведущий мастер", "Опыт от 4 лет, постоянное повышение квалификации", 500,
						"Мастер салона", "Качественное выполнение по регламенту", 0),
				},
				{
					ID:       "time_slot",
					Question: "Желаемое время визита",
					Options: QuizOptionDefault("Утренние часы (10:00 - 13:00)", "Скидка «Счастливые часы»", 0,
						"Дневное / вечернее время", "После рабочего дня", 0,
						"Выходные дни", "Суббота / Воскресенье", 300),
				},
			},
		}

	default: // legal, medical, tech, general
		return models.QuizPayload{
			Badge:     "Экспресс-оценка задачи",
			Title:     "Рассчитайте предварительную стоимость услуг",
			Subtitle:  "Ответьте на 3 уточняющих вопроса для точного определения объема работ",
			BasePrice: 3000,
			Currency:  "₽",
			BonusText: "Бесплатный первичный экспресс-аудит документов при подтверждении заявки.",
			Steps: []models.QuizStep{
				{
					ID:       "task_scope",
					Question: "Какая задача стоит перед вами?",
					Options: QuizOptionDefault("Сделка с недвижимостью / землей", "Проверка чистоты, сопровождение регистрации", 3000,
						"Защита бизнеса и споры", "Претензии, досудебное урегулирование, суды", 6000,
						"Оформление доверенностей / документов", "Удостоверение, подготовка проектов", 1500,
						"Комплексная консультация и аудит", "Детальный правовой разбор ситуации", 2500),
				},
				{
					ID:       "client_type",
					Question: "Форма взаимодействия",
					Options: QuizOptionDefault("Физическое лицо", "Для личных вопросов и семьи", 0,
						"Индивидуальный предприниматель", "Малый бизнес, самозанятые", 1500,
						"Юридическое лицо (ООО/АО)", "Корпоративное обслуживание", 3500),
				},
				{
					ID:       "urgency",
					Question: "Срочность подготовки документов",
					Options: QuizOptionDefault("Срочно (в течение 24 часов)", "Экстренное включение в работу", 2000,
						"В стандартные сроки (3-5 дней)", "Плановое выполнение регламента", 0),
				},
			},
		}
	}
}

func QuizOptionDefault(args ...interface{}) []models.QuizOption {
	var opts []models.QuizOption
	for i := 0; i < len(args); i += 3 {
		title, _ := args[i].(string)
		desc, _ := args[i+1].(string)
		var price float64
		switch v := args[i+2].(type) {
		case int:
			price = float64(v)
		case float64:
			price = v
		}
		opts = append(opts, models.QuizOption{
			ID:            fmt.Sprintf("opt_%d", (i/3)+1),
			Title:         title,
			Description:   desc,
			PriceModifier: price,
		})
	}
	return opts
}

// ParseBeforeAfterPayload converts raw block payload map into typed BeforeAfterPayload
func ParseBeforeAfterPayload(payload map[string]interface{}, theme string) models.BeforeAfterPayload {
	if parsed := ParsePayload[models.BeforeAfterPayload](payload); parsed != nil && len(parsed.Items) > 0 {
		return *parsed
	}

	// Дефолтные кейсы с качественными оптимизированными фото
	switch theme {
	case "auto", "repair":
		return models.BeforeAfterPayload{
			Badge:    "Реальные кейсы СТО",
			Title:    "Примеры наших работ: До и После",
			Subtitle: "Оцените результат восстановительного ремонта кузова и детейлинга",
			Items: []models.BeforeAfterItem{
				{
					Title:         "Восстановительная полировка и керамическая защита кузова",
					Description:   "Устранили паутинку мелких царапин, восстановили зеркальный блеск лака и нанесли 2 слоя нанокерамики.",
					BeforeImage:   "https://images.unsplash.com/photo-1617814076367-b759c7d7e738?auto=format&fit=crop&w=1200&q=80",
					AfterImage:    "https://images.unsplash.com/photo-1603584173870-7f23fdae1b7a?auto=format&fit=crop&w=1200&q=80",
					BeforeLabel:   "До полировки",
					AfterLabel:    "После керамики",
					Duration:      "2 рабочих дня",
					Cost:          "38 000 ₽",
					MaterialsCost: "16 500 ₽ (керамический состав Gyeon)",
					WorksCost:     "21 500 ₽",
					Stages: []string{
						"Глубокая двухфазная мойка и очистка кузова автоскрабом",
						"Трехэтапная абразивная полировка пастами различной зернистости",
						"Обезжиривание и нанесение подложки Ceramic Primer",
						"Нанесение 2 слоев гидрофобного керамического покрытия 9H",
					},
				},
				{
					Title:         "Глубокая химчистка салона и реставрация кожи",
					Description:   "Полная дезинфекция паром, удаление застарелых пятен и кондиционирование кожаной обивки сидений.",
					BeforeImage:   "https://images.unsplash.com/photo-1503376780353-7e6692767b70?auto=format&fit=crop&w=1200&q=80",
					AfterImage:    "https://images.unsplash.com/photo-1552519507-da3b142c6e3d?auto=format&fit=crop&w=1200&q=80",
					BeforeLabel:   "До очистки",
					AfterLabel:    "Чистый салон",
					Duration:      "1 рабочий день",
					Cost:          "19 500 ₽",
					MaterialsCost: "4 500 ₽ (гипоаллергенная биохимия Koch)",
					WorksCost:     "15 000 ₽",
					Stages: []string{
						"Сухая уборка мощным пылеводососом и продувка торнадором",
						"Экстракторная чистка ковролина и потолочного покрытия",
						"Деликатная ручная очистка и нанесение бальзама на кожу сидений",
						"Озонирование и устранение неприятных запахов кондиционера",
					},
				},
			},
		}

	case "beauty":
		return models.BeforeAfterPayload{
			Badge:    "Портфолио мастеров",
			Title:    "Преображение клиентов: До и После",
			Subtitle: "Индивидуальный подход к подбору оттенка и формы стрижки",
			Items: []models.BeforeAfterItem{
				{
					Title:       "Сложное рельефное окрашивание Airtouch",
					Description: "Плавная растяжка цвета от натуральных корней к чистому холодному блонду без повреждения структуры волос.",
					BeforeImage: "https://images.unsplash.com/photo-1522337360788-8b13dee7a37e?auto=format&fit=crop&w=1200&q=80",
					AfterImage:  "https://images.unsplash.com/photo-1560066984-138dadb4c035?auto=format&fit=crop&w=1200&q=80",
					BeforeLabel: "До окрашивания",
					AfterLabel:  "Результат Airtouch",
				},
			},
		}

	default: // legal, medical, default
		return models.BeforeAfterPayload{
			Badge:    "Эффективность решений",
			Title:    "Результаты до и после правового сопровождения",
			Subtitle: "Как профессиональная юридическая поддержка защищает активы доверителей",
			Items: []models.BeforeAfterItem{
				{
					Title:       "Урегулирование задолженности и снятие ареста с активов",
					Description: "До: риск субсидиарной ответственности на 18 млн ₽. После: мировое соглашение и полное сохранение бизнеса.",
					BeforeImage: "https://images.unsplash.com/photo-1450133064473-71024230f91b?auto=format&fit=crop&w=1200&q=80",
					AfterImage:  "https://images.unsplash.com/photo-1589829545856-d10d557cf95f?auto=format&fit=crop&w=1200&q=80",
					BeforeLabel: "Спор в суде",
					AfterLabel:  "Активы защищены",
				},
			},
		}
	}
}
