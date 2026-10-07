package handlers

import (
	"context"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"

	"sitecraft/internal/data"
	"sitecraft/internal/events"
	"sitecraft/internal/models"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
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

	validationErrors := ValidateLeadForm(input)

	// Если есть ошибки валидации — возвращаем структурированное сообщение
	if len(validationErrors) > 0 {
		var errList strings.Builder
		for _, msg := range validationErrors {
			errList.WriteString(fmt.Sprintf(
				`<li class="flex items-start gap-2"><svg class="w-4 h-4 text-rose-500 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg><span>%s</span></li>`,
				html.EscapeString(msg),
			))
		}

		errorHTML := fmt.Sprintf(`
			<div class="p-4 bg-rose-50 border border-rose-300 rounded-xl text-rose-900 text-sm shadow-sm transition-all duration-200">
				<div class="font-bold flex items-center gap-2 mb-2 text-rose-800">
					<svg class="w-5 h-5 text-rose-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
					Пожалуйста, проверьте данные формы:
				</div>
				<ul class="text-xs text-rose-700 space-y-1.5 ml-1">
					%s
				</ul>
			</div>
		`, errList.String())
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
			return e.HTML(http.StatusConflict, `
				<div class="p-4 bg-amber-50 border border-amber-300 rounded-xl text-amber-900 text-sm shadow-sm">
					<div class="font-bold flex items-center gap-2 mb-1 text-amber-800">
						<svg class="w-5 h-5 text-amber-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
						Выбранное время уже занято
					</div>
					<p class="text-xs text-amber-700">К сожалению, этот слот только что забронировали. Пожалуйста, выберите другое свободное время.</p>
				</div>
			`)
		}
	}

	// Подготовка JSON payload для гибких полей
	payloadMap := map[string]interface{}{}
	if actionType != "" {
		payloadMap["action_type"] = actionType
	}
	if comment != "" {
		payloadMap["comment"] = comment
	}

	// Сохраняем в коллекцию leads в PocketBase
	leadModel := models.Lead{
		Name:        name,
		Phone:       phone,
		Status:      "новая",
		Source:      "website",
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
		leadRecord.Set("status", "новая")
		leadRecord.Set("source", "website")
		leadRecord.Set("payload", payloadMap)

		if err := h.app.Save(leadRecord); err != nil {
			log.Printf("[Leads Error] Не удалось сохранить заявку в базу: %v\n", err)
			return e.HTML(http.StatusInternalServerError, `
				<div class="p-4 bg-rose-50 border border-rose-300 rounded-xl text-rose-800 text-sm">
					<div class="font-bold">Ошибка сервера</div>
					<p class="mt-1 text-xs">Не удалось сохранить запись. Пожалуйста, повторите попытку или позвоните нам по телефону.</p>
				</div>
			`)
		}
		leadModel.ID = leadRecord.Id
	}

	// Публикуем событие через Event Bus (подписчики Telegram, CRM и др. обработают его асинхронно)
	h.eventBus.Publish(context.Background(), events.LeadCreatedEvent{
		Settings: settings,
		Lead:     leadModel,
	})

	log.Printf("Новая заявка: ФИО=%s, Тел=%s, Дата=%s %s, Инфо=%+v\n",
		name, phone, bookingDate, bookingTime, payloadMap)

	// Возвращаем сообщение об успехе
	var successMsg string
	if bookingDate != "" && bookingTime != "" {
		successMsg = fmt.Sprintf(`Мы зарезервировали для вас <strong>%s в %s</strong>. Наш специалист свяжется с вами по номеру <strong>%s</strong> для подтверждения.`,
			html.EscapeString(bookingDate), html.EscapeString(bookingTime), html.EscapeString(phone))
	} else {
		successMsg = fmt.Sprintf(`Спасибо за обращение! Наш менеджер свяжется с вами по номеру <strong>%s</strong> в ближайшее рабочее время.`,
			html.EscapeString(phone))
	}

	successHTML := fmt.Sprintf(`
		<div class="p-4 bg-emerald-50 border border-emerald-300 rounded-xl text-emerald-800 text-sm shadow-sm transition-all duration-200">
			<div class="font-bold flex items-center gap-2">
				<svg class="w-5 h-5 text-emerald-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
				Заявка успешно принята!
			</div>
			<p class="mt-2 text-xs text-emerald-700">%s</p>
		</div>
	`, successMsg)
	return e.HTML(http.StatusOK, successHTML)
}
