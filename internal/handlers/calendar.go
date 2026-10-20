package handlers

import (
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"sitecraft/internal/data"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// GenerateICS builds a RFC 5545 compliant iCalendar (.ics) string using pure Go stdlib.
func GenerateICS(title, description, location string, startTime, endTime time.Time, uid string) []byte {
	if uid == "" {
		uid = fmt.Sprintf("%d@sitecraft", time.Now().UnixNano())
	}

	const icsTimeFormat = "20060102T150405Z"
	nowUTC := time.Now().UTC().Format(icsTimeFormat)
	startUTC := startTime.UTC().Format(icsTimeFormat)
	endUTC := endTime.UTC().Format(icsTimeFormat)

	var sb strings.Builder
	sb.WriteString("BEGIN:VCALENDAR\r\n")
	sb.WriteString("VERSION:2.0\r\n")
	sb.WriteString("PRODID:-//SiteCraft Engine//Booking Calendar 1.0//RU\r\n")
	sb.WriteString("CALSCALE:GREGORIAN\r\n")
	sb.WriteString("METHOD:PUBLISH\r\n")
	sb.WriteString("BEGIN:VEVENT\r\n")
	sb.WriteString(fmt.Sprintf("UID:%s\r\n", uid))
	sb.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", nowUTC))
	sb.WriteString(fmt.Sprintf("DTSTART:%s\r\n", startUTC))
	sb.WriteString(fmt.Sprintf("DTEND:%s\r\n", endUTC))
	sb.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", cleanICSText(title)))
	sb.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", cleanICSText(description)))
	if location != "" {
		sb.WriteString(fmt.Sprintf("LOCATION:%s\r\n", cleanICSText(location)))
	}
	sb.WriteString("STATUS:CONFIRMED\r\n")
	sb.WriteString("END:VEVENT\r\n")
	sb.WriteString("END:VCALENDAR\r\n")

	return []byte(sb.String())
}

func cleanICSText(text string) string {
	r := strings.NewReplacer(
		"\\", "\\\\",
		";", "\\;",
		",", "\\,",
		"\n", "\\n",
		"\r", "",
	)
	return r.Replace(text)
}

// BookingHandler handles .ics calendar downloads and 1-click confirm/cancel actions.
type BookingHandler struct {
	app *pocketbase.PocketBase
}

// NewBookingHandler creates a new instance of BookingHandler.
func NewBookingHandler(app *pocketbase.PocketBase) *BookingHandler {
	return &BookingHandler{app: app}
}

// HandleDownloadICS serves an .ics calendar file for the booked visit.
func (h *BookingHandler) HandleDownloadICS(e *core.RequestEvent) error {
	dateStr := strings.TrimSpace(e.Request.URL.Query().Get("date")) // YYYY-MM-DD
	timeStr := strings.TrimSpace(e.Request.URL.Query().Get("time")) // HH:MM
	serviceName := strings.TrimSpace(e.Request.URL.Query().Get("service"))
	clientName := strings.TrimSpace(e.Request.URL.Query().Get("name"))

	content := data.GetSiteContent(h.app)
	settings := data.GetSiteSettings(h.app)

	if serviceName == "" {
		serviceName = "Прием в " + settings.SiteName
	}

	var startTime time.Time
	if dateStr != "" && timeStr != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+timeStr, time.Local); err == nil {
			startTime = t
		}
	}

	if startTime.IsZero() {
		startTime = time.Now().Add(24 * time.Hour)
	}
	// Длительность визита по умолчанию — 45 минут
	endTime := startTime.Add(45 * time.Minute)

	desc := fmt.Sprintf("Запись на прием: %s. Клиент: %s. Контакт: %s", serviceName, clientName, content.Phone1)
	location := content.Address
	if content.AddressSub != "" {
		location += ", " + content.AddressSub
	}

	icsBytes := GenerateICS(serviceName, desc, location, startTime, endTime, "")

	e.Response.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	e.Response.Header().Set("Content-Disposition", `attachment; filename="visit.ics"`)
	_, err := e.Response.Write(icsBytes)
	return err
}

// HandleBookingAction handles 1-click confirmation or cancellation from Telegram/Email reminder.
func (h *BookingHandler) HandleBookingAction(e *core.RequestEvent) error {
	id := strings.TrimSpace(e.Request.URL.Query().Get("id"))
	action := strings.ToLower(strings.TrimSpace(e.Request.URL.Query().Get("action")))

	if id == "" || (action != "confirm" && action != "cancel") {
		return e.HTML(http.StatusBadRequest, renderActionPage("Неверные параметры", "Не указан идентификатор брони или действие.", "warning"))
	}

	rec, err := h.app.FindRecordById("leads", id)
	if err != nil || rec == nil {
		return e.HTML(http.StatusNotFound, renderActionPage("Запись не найдена", "Бронирование с указанным номером не найдено или было удалено.", "error"))
	}

	if action == "confirm" {
		rec.Set("status", "подтверждена")
		_ = h.app.Save(rec)
		return e.HTML(http.StatusOK, renderActionPage(
			"Визит подтвержден!",
			fmt.Sprintf("Ваша запись на <strong>%s %s</strong> успешно подтверждена. Ждем вас по адресу: %s.",
				html.EscapeString(rec.GetString("booking_date")),
				html.EscapeString(rec.GetString("booking_time")),
				html.EscapeString(data.GetSiteContent(h.app).Address)),
			"success",
		))
	}

	// Cancel action
	rec.Set("status", "отменена")
	_ = h.app.Save(rec)
	return e.HTML(http.StatusOK, renderActionPage(
		"Запись отменена",
		fmt.Sprintf("Ваша бронь на <strong>%s %s</strong> была отменена. Слот времени освобожден.",
			html.EscapeString(rec.GetString("booking_date")),
			html.EscapeString(rec.GetString("booking_time"))),
		"info",
	))
}

func renderActionPage(title, message, statusType string) string {
	badgeColor := "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20"
	icon := "✓"
	if statusType == "error" || statusType == "warning" {
		badgeColor = "bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20"
		icon = "!"
	} else if statusType == "info" {
		badgeColor = "bg-slate-500/10 text-slate-600 dark:text-slate-400 border-slate-500/20"
		icon = "ℹ"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="utf-8"/>
	<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
	<title>%s</title>
	<link rel="stylesheet" href="/assets/styles.css"/>
</head>
<body class="bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 flex items-center justify-center min-h-screen p-4 font-sans antialiased">
	<div class="max-w-md w-full bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-3xl p-8 shadow-2xl text-center">
		<div class="w-16 h-16 rounded-2xl %s border flex items-center justify-center mx-auto mb-4 text-2xl font-bold">
			%s
		</div>
		<h1 class="text-xl font-bold mb-2">%s</h1>
		<p class="text-sm text-slate-600 dark:text-slate-400 leading-relaxed mb-6">%s</p>
		<a href="/" class="inline-flex items-center justify-center px-6 py-2.5 rounded-xl bg-slate-900 hover:bg-slate-800 text-white dark:bg-white dark:hover:bg-slate-100 dark:text-slate-900 text-xs font-semibold shadow transition-all">
			Вернуться на сайт
		</a>
	</div>
</body>
</html>`, html.EscapeString(title), badgeColor, icon, html.EscapeString(title), message)
}
