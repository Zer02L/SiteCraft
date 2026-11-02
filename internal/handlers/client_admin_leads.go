package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"sitecraft/internal/data"
	"sitecraft/internal/models"
	"sitecraft/ui/pages"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

// HandleTabLeads renders leads tab fragment for HTMX.
func (h *ClientAdminHandler) HandleTabLeads(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}
	leads := data.GetAllLeads(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientLeadsTab(leads).Render(e.Request.Context(), e.Response)
}

// HandleUpdateLeadStatus updates status of a specific lead via HTMX.
func (h *ClientAdminHandler) HandleUpdateLeadStatus(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}

	leadID := e.Request.PathValue("id")
	newStatus := strings.TrimSpace(e.Request.FormValue("status"))

	rec, err := h.app.FindRecordById("leads", leadID)
	if err != nil || rec == nil {
		return e.String(http.StatusNotFound, "Заявка не найдена")
	}

	rec.Set("status", newStatus)
	if err := h.app.Save(rec); err != nil {
		return e.String(http.StatusInternalServerError, "Ошибка сохранения статуса")
	}

	lead := models.Lead{
		ID:          rec.Id,
		Name:        rec.GetString("name"),
		Phone:       rec.GetString("phone"),
		Email:       rec.GetString("email"),
		Status:      rec.GetString("status"),
		Source:      rec.GetString("source"),
		BookingDate: rec.GetString("booking_date"),
		BookingTime: rec.GetString("booking_time"),
	}

	allLeads := data.GetAllLeads(h.app)
	newCount, inProgressCount := countLeadsByStatus(allLeads)

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.LeadRowWithOOBBadges(lead, newCount, inProgressCount, len(allLeads)).Render(e.Request.Context(), e.Response)
}

func countLeadsByStatus(leads []models.Lead) (newCount int, inProgressCount int) {
	for _, l := range leads {
		if l.Status == "новая" || l.Status == "" {
			newCount++
		} else if l.Status == "в обработке" {
			inProgressCount++
		}
	}
	return newCount, inProgressCount
}

// sanitizeCSVField neutralizes CSV Formula Injection (OWASP guidance for spreadsheet export).
// Values starting with =, +, -, @, \t, or \r are prepended with a single quote.
func sanitizeCSVField(val string) string {
	trimmed := strings.TrimSpace(val)
	if len(trimmed) > 0 {
		switch trimmed[0] {
		case '=', '+', '-', '@', '\t', '\r':
			return "'" + val
		}
	}
	return val
}

// HandleExportLeadsCSV exports all leads as a CSV file download.
func (h *ClientAdminHandler) HandleExportLeadsCSV(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}

	leads := data.GetAllLeads(h.app)

	b := &bytes.Buffer{}
	// BOM для корректного открытия русских символов в Excel
	b.WriteString("\xEF\xBB\xBF")

	w := csv.NewWriter(b)
	w.Comma = ';'

	_ = w.Write([]string{"ID", "Имя", "Телефон", "Email", "Дата записи", "Время", "Статус", "Источник", "UTM Источник", "UTM Кампания", "Квиз / Смета", "Комментарий"})

	for _, l := range leads {
		comment := ""
		utmSource := ""
		utmCampaign := ""
		quizInfo := ""

		if l.Payload != nil {
			if c, ok := l.Payload["comment"].(string); ok {
				comment = c
			}
			if s, ok := l.Payload["utm_source"].(string); ok {
				utmSource = s
			}
			if cmp, ok := l.Payload["utm_campaign"].(string); ok {
				utmCampaign = cmp
			}
			if qMap, ok := l.Payload["quiz_summary"].(map[string]interface{}); ok {
				if est, ok := qMap["estimated_price"]; ok {
					quizInfo = fmt.Sprintf("%v руб.", est)
				}
			}
		}
		_ = w.Write([]string{
			sanitizeCSVField(l.ID),
			sanitizeCSVField(l.Name),
			sanitizeCSVField(l.Phone),
			sanitizeCSVField(l.Email),
			sanitizeCSVField(l.BookingDate),
			sanitizeCSVField(l.BookingTime),
			sanitizeCSVField(l.Status),
			sanitizeCSVField(l.Source),
			sanitizeCSVField(utmSource),
			sanitizeCSVField(utmCampaign),
			sanitizeCSVField(quizInfo),
			sanitizeCSVField(comment),
		})
	}
	w.Flush()

	filename := fmt.Sprintf("leads_%s.csv", time.Now().Format("2006-01-02"))
	e.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	e.Response.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	_, err := e.Response.Write(b.Bytes())
	return err
}
