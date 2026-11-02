package data

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sitecraft/internal/cache"
	"sitecraft/internal/models"

	"ganigeorgiev.com/pocketbase/dbx"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

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
