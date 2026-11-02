package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"

	"sitecraft/internal/data"
	"sitecraft/internal/events"
	"sitecraft/internal/i18n"
	"sitecraft/internal/models"

	"ganigeorgiev.com/pocketbase/dbx"
	"ganigeorgiev.com/pocketbase/pocketbase"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

// LeadHandler processes incoming leads and inquiries.
type LeadHandler struct {
	app      *pocketbase.PocketBase
	eventBus events.Bus
}

// NewLeadHandler creates a new instance of LeadHandler.
func NewLeadHandler(app *pocketbase.PocketBase, bus ...events.Bus) *LeadHandler {
	var eb events.Bus = events.Default()
	if len(bus) > 0 && bus[0] != nil {
		eb = bus[0]
	}
	return &LeadHandler{
		app:      app,
		eventBus: eb,
	}
}

// HandleCreateLead handles form submissions from the landing page.
func (h *LeadHandler) HandleCreateLead(e *core.RequestEvent) error {
	ctx := e.Request.Context()
	lang := i18n.GetLocale(ctx)
	settings := data.GetSiteSettings(h.app)

	name := strings.TrimSpace(e.Request.FormValue("name"))
	phone := strings.TrimSpace(e.Request.FormValue("phone"))
	bookingDate := strings.TrimSpace(e.Request.FormValue("booking_date"))
	bookingTime := strings.TrimSpace(e.Request.FormValue("booking_time"))
	actionType := strings.TrimSpace(e.Request.FormValue("action_type"))
	comment := strings.TrimSpace(e.Request.FormValue("comment"))
	consent := e.Request.FormValue("consent")
	honeypot := e.Request.FormValue("company_website_url")

	input := LeadFormInput{
		Name:        name,
		Phone:       phone,
		BookingDate: bookingDate,
		BookingTime: bookingTime,
		ActionType:  actionType,
		Comment:     comment,
		Consent:     consent,
		Honeypot:    honeypot,
		Mode:        settings.BookingMode,
	}

	validationErrors := ValidateLeadFormLocalized(input, lang)

	// Если есть ошибки валидации — возвращаем структурированное сообщение
	if len(validationErrors) > 0 {
		var errList strings.Builder
		for _, msg := range validationErrors {
			errList.WriteString(fmt.Sprintf(
				`<li class="flex items-start gap-2"><svg class="w-4 h-4 text-rose-500 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg><span>%s</span></li>`,
				html.EscapeString(msg),
			))
		}

		errorTitle := i18n.T(ctx, "error.form_check")
		errorHTML := fmt.Sprintf(`
			<div class="p-4 bg-rose-50 border border-rose-300 rounded-xl text-rose-900 text-sm shadow-sm transition-all duration-200">
				<div class="font-bold flex items-center gap-2 mb-2 text-rose-800">
					<svg class="w-5 h-5 text-rose-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
					%s
				</div>
				<ul class="text-xs text-rose-700 space-y-1.5 ml-1">
					%s
				</ul>
			</div>
		`, html.EscapeString(errorTitle), errList.String())
		return e.HTML(http.StatusUnprocessableEntity, errorHTML)
	}

	// Проверка на повторное бронирование слота (только если указаны дата и время)
	if bookingDate != "" && bookingTime != "" {
		existingRecords, _ := h.app.FindRecordsByFilter(
			"leads",
			"booking_date = {:date} && booking_time = {:time} && status != 'отменена'",
			"",
			1,
			0,
			dbx.Params{"date": bookingDate, "time": bookingTime},
		)
		if len(existingRecords) > 0 {
			slotTitle := i18n.T(ctx, "booking.slot_taken_title")
			slotDesc := i18n.T(ctx, "booking.slot_taken_desc")
			return e.HTML(http.StatusConflict, fmt.Sprintf(`
				<div class="p-4 bg-amber-50 border border-amber-300 rounded-xl text-amber-900 text-sm shadow-sm">
					<div class="font-bold flex items-center gap-2 mb-1 text-amber-800">
						<svg class="w-5 h-5 text-amber-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
						%s
					</div>
					<p class="text-xs text-amber-700">%s</p>
				</div>
			`, html.EscapeString(slotTitle), html.EscapeString(slotDesc)))
		}
	}

	source := strings.TrimSpace(e.Request.FormValue("source"))
	if source == "" {
		source = "website"
	}

	// Подготовка JSON payload для гибких полей
	payloadMap := map[string]interface{}{}
	if actionType != "" {
		payloadMap["action_type"] = actionType
	}
	if comment != "" {
		payloadMap["comment"] = comment
	}

	// Обработка результатов онлайн-квиза/калькулятора
	if quizRaw := strings.TrimSpace(e.Request.FormValue("quiz_payload")); quizRaw != "" {
		var quizMap map[string]interface{}
		if err := json.Unmarshal([]byte(quizRaw), &quizMap); err == nil {
			payloadMap["quiz_summary"] = quizMap
		} else {
			payloadMap["quiz_summary"] = quizRaw
		}
	}

	// Сквозная UTM-атрибуция (из контекста middleware или формы)
	utm := GetUTM(ctx)
	if utm.IsEmpty() {
		utm.Source = strings.TrimSpace(e.Request.FormValue("utm_source"))
		utm.Medium = strings.TrimSpace(e.Request.FormValue("utm_medium"))
		utm.Campaign = strings.TrimSpace(e.Request.FormValue("utm_campaign"))
		utm.Content = strings.TrimSpace(e.Request.FormValue("utm_content"))
		utm.Term = strings.TrimSpace(e.Request.FormValue("utm_term"))
	}
	if !utm.IsEmpty() {
		payloadMap["utm"] = utm
		payloadMap["utm_source"] = utm.Source
		payloadMap["utm_campaign"] = utm.Campaign
	}

	// Фиксация варианта A/B тестирования (A или B)
	abVariant := GetABVariant(ctx)
	if abParam := strings.ToUpper(strings.TrimSpace(e.Request.FormValue("ab_variant"))); abParam == "A" || abParam == "B" {
		abVariant = abParam
	}
	payloadMap["ab_variant"] = abVariant

	// Фиксация выбранного специалиста/поста (Resource Booking)
	resourceID := strings.TrimSpace(e.Request.FormValue("resource_id"))
	if resourceID != "" {
		payloadMap["resource_id"] = resourceID
	}

	// Сохраняем в коллекцию leads в PocketBase
	leadModel := models.Lead{
		Name:        name,
		Phone:       phone,
		Status:      "новая",
		Source:      source,
		BookingDate: bookingDate,
		BookingTime: bookingTime,
		Payload:     payloadMap,
	}

	if leadsCol, err := h.app.FindCollectionByNameOrId("leads"); err == nil {
		leadRecord := core.NewRecord(leadsCol)
		leadRecord.Set("name", name)
		leadRecord.Set("phone", phone)
		leadRecord.Set("booking_date", bookingDate)
		leadRecord.Set("booking_time", bookingTime)
		leadRecord.Set("resource_id", resourceID)
		leadRecord.Set("status", "новая")
		leadRecord.Set("source", source)
		leadRecord.Set("payload", payloadMap)

		if err := h.app.Save(leadRecord); err != nil {
			log.Printf("[Leads Error] Не удалось сохранить заявку в базу: %v\n", err)
			serverTitle := i18n.T(ctx, "error.server_title")
			serverDesc := i18n.T(ctx, "error.server_desc")
			return e.HTML(http.StatusInternalServerError, fmt.Sprintf(`
				<div class="p-4 bg-rose-50 border border-rose-300 rounded-xl text-rose-800 text-sm">
					<div class="font-bold">%s</div>
					<p class="mt-1 text-xs">%s</p>
				</div>
			`, html.EscapeString(serverTitle), html.EscapeString(serverDesc)))
		}
		leadModel.ID = leadRecord.Id
	}

	// Публикуем событие через Event Bus (подписчики Telegram, CRM и др. обработают его асинхронно)
	h.eventBus.Publish(context.Background(), events.LeadCreatedEvent{
		Settings: settings,
		Lead:     leadModel,
	})

	log.Printf("Новая заявка: ФИО=%s, Тел=%s, Дата=%s %s, Ресурс=%s, Инфо=%+v\n",
		name, phone, bookingDate, bookingTime, resourceID, payloadMap)

	// Возвращаем сообщение об успехе
	var successMsg string
	var calendarLinkHTML string
	if bookingDate != "" && bookingTime != "" {
		successMsg = fmt.Sprintf(i18n.T(ctx, "booking.success_slots"),
			"<strong>"+html.EscapeString(bookingDate)+"</strong>",
			"<strong>"+html.EscapeString(bookingTime)+"</strong>",
			"<strong>"+html.EscapeString(phone)+"</strong>")

		icsURL := fmt.Sprintf("/api/booking/calendar.ics?date=%s&time=%s&name=%s",
			bookingDate, bookingTime, html.EscapeString(name))
		calendarLinkHTML = fmt.Sprintf(`
			<div class="mt-3 pt-3 border-t border-emerald-300 dark:border-emerald-800 flex items-center justify-between gap-2 flex-wrap">
				<span class="text-xs text-emerald-800 dark:text-emerald-300 font-medium">Событие визита:</span>
				<a href="%s" download="visit.ics" class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white font-medium text-xs transition-colors shadow-xs">
					<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path></svg>
					<span>Добавить в календарь (.ics)</span>
				</a>
			</div>
		`, icsURL)
	} else {
		successMsg = fmt.Sprintf(i18n.T(ctx, "booking.success_simple"),
			"<strong>"+html.EscapeString(phone)+"</strong>")
	}

	successTitle := i18n.T(ctx, "booking.success_title")
	successHTML := fmt.Sprintf(`
		<div class="p-4 bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-300 dark:border-emerald-800 rounded-xl text-emerald-800 dark:text-emerald-200 text-sm shadow-sm transition-all duration-200">
			<div class="font-bold flex items-center gap-2">
				<svg class="w-5 h-5 text-emerald-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
				%s
			</div>
			<p class="mt-2 text-xs text-emerald-700 dark:text-emerald-300">%s</p>
			%s
		</div>
	`, html.EscapeString(successTitle), successMsg, calendarLinkHTML)
	return e.HTML(http.StatusOK, successHTML)
}

// HandleApplyVacancy processes HR job applications submitted from careers section/page (Блок 4: HR).
func (h *LeadHandler) HandleApplyVacancy(e *core.RequestEvent) error {
	name := strings.TrimSpace(e.Request.FormValue("name"))
	phone := strings.TrimSpace(e.Request.FormValue("phone"))
	vacancyTitle := strings.TrimSpace(e.Request.FormValue("vacancy_title"))
	resumeLink := strings.TrimSpace(e.Request.FormValue("resume_link"))
	honeypot := e.Request.FormValue("company_website_url")

	if honeypot != "" {
		return e.HTML(http.StatusOK, `<div class="p-4 bg-emerald-50 rounded-xl text-xs text-emerald-800">Спасибо! Заявка принята.</div>`)
	}

	if name == "" || phone == "" {
		return e.HTML(http.StatusBadRequest, `<div class="p-3 bg-rose-50 border border-rose-200 rounded-xl text-rose-800 text-xs font-semibold">Пожалуйста, укажите имя и телефон для связи.</div>`)
	}

	if vacancyTitle == "" {
		vacancyTitle = "Специалист"
	}

	err := data.CreateVacancyLead(h.app, name, phone, "", vacancyTitle, resumeLink, "")
	if err != nil {
		return e.HTML(http.StatusInternalServerError, `<div class="p-3 bg-rose-50 border border-rose-200 rounded-xl text-rose-800 text-xs font-semibold">Не удалось сохранить отклик. Пожалуйста, позвоните нам напрямую.</div>`)
	}

	settings := data.GetSiteSettings(h.app)
	leadModel := models.Lead{
		Name:    name,
		Phone:   phone,
		Source:  "vacancy",
		Status:  "новая",
		Payload: map[string]interface{}{
			"action_type": "hr_application",
			"vacancy":     vacancyTitle,
			"resume":      resumeLink,
		},
	}
	h.eventBus.Publish(context.Background(), events.LeadCreatedEvent{
		Settings: settings,
		Lead:     leadModel,
	})

	successHTML := fmt.Sprintf(`
		<div class="p-4 bg-emerald-50 border border-emerald-300 rounded-xl text-emerald-900 text-xs shadow-xs">
			<div class="font-bold text-sm mb-1">✓ Отклик успешно отправлен!</div>
			<p class="text-emerald-700">Спасибо, %s! Наш HR-специалист свяжется с вами по указанному телефону для обсуждения вакансии «%s».</p>
		</div>
	`, html.EscapeString(name), html.EscapeString(vacancyTitle))

	return e.HTML(http.StatusOK, successHTML)
}

