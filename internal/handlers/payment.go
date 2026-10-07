package handlers

import (
	"fmt"
	"html"
	"net/http"
	"strings"

	"sitecraft/internal/data"
	"sitecraft/internal/events"
	"sitecraft/internal/models"
	"sitecraft/internal/modules"
	"sitecraft/ui/components"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// PaymentHandler handles payment creation and processing.
type PaymentHandler struct {
	app             *pocketbase.PocketBase
	providerFactory func(settings models.SiteSettings) modules.PaymentProvider
	eventBus        events.Bus
}

// NewPaymentHandler creates a new instance of PaymentHandler.
func NewPaymentHandler(app *pocketbase.PocketBase, opts ...func(*PaymentHandler)) *PaymentHandler {
	h := &PaymentHandler{
		app:             app,
		providerFactory: modules.NewPaymentProvider,
		eventBus:        events.Default(),
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// WithPaymentProviderFactory overrides default payment provider factory.
func WithPaymentProviderFactory(factory func(models.SiteSettings) modules.PaymentProvider) func(*PaymentHandler) {
	return func(h *PaymentHandler) {
		h.providerFactory = factory
	}
}

// WithPaymentEventBus injects custom event bus.
func WithPaymentEventBus(bus events.Bus) func(*PaymentHandler) {
	return func(h *PaymentHandler) {
		h.eventBus = bus
	}
}

// HandleCreatePayment initiates payment via configured PaymentProvider.
func (h *PaymentHandler) HandleCreatePayment(e *core.RequestEvent) error {
	settings := data.GetSiteSettings(h.app)
	orderCode := strings.TrimSpace(e.Request.FormValue("order_code"))
	description := e.Request.FormValue("description")
	email := e.Request.FormValue("email")
	phone := e.Request.FormValue("phone")
	consent := e.Request.FormValue("consent")
	honeypot := e.Request.FormValue("company_website_url")
	var amount float64 = 1000.0

	// ВАЖНО: Если передан order_code (оплата по трекеру заказов),
	// сумма СТРОГО берется из базы данных PocketBase!
	// Пользователь НЕ может изменить сумму через HTML/POST форму.
	if orderCode != "" {
		order, err := data.GetServiceOrderByCode(h.app, orderCode)
		if err != nil || order == nil {
			return e.HTML(http.StatusBadRequest, `<div class="p-3 bg-rose-50 text-rose-700 text-xs rounded-lg border border-rose-200">Заказ-наряд не найден в системе.</div>`)
		}
		if order.IsPaid {
			return e.HTML(http.StatusBadRequest, `<div class="p-3 bg-amber-50 text-amber-800 text-xs rounded-lg border border-amber-200">Данный заказ-наряд уже был успешно оплачен ранее.</div>`)
		}
		// Перезаписываем сумму значением из базы данных!
		amount = order.TotalAmount
		description = fmt.Sprintf("Оплата заказ-наряда № %s (%s)", order.OrderCode, order.ItemTitle)
		if phone == "" {
			phone = order.ClientPhone
		}
	} else {
		_, _ = fmt.Sscanf(e.Request.FormValue("amount"), "%f", &amount)
	}

	// Валидация формы оплаты
	input := PaymentFormInput{
		OrderCode: orderCode,
		Amount:    amount,
		Email:     email,
		Phone:     phone,
		Consent:   consent,
		Honeypot:  honeypot,
	}

	if valErrs := ValidatePaymentForm(input, orderCode != ""); len(valErrs) > 0 {
		var sb strings.Builder
		sb.WriteString(`<div class="p-3 bg-rose-50 text-rose-700 text-xs rounded-lg border border-rose-200 space-y-1"><div class="font-bold">Пожалуйста, исправьте ошибки:</div><ul class="list-disc list-inside">`)
		for _, errText := range valErrs {
			sb.WriteString(fmt.Sprintf("<li>%s</li>", html.EscapeString(errText)))
		}
		sb.WriteString(`</ul></div>`)
		return e.HTML(http.StatusUnprocessableEntity, sb.String())
	}

	provider := h.providerFactory(settings)
	res, err := provider.CreatePayment(models.PaymentRequest{
		Amount:      amount,
		Currency:    "RUB",
		Description: description,
		ReturnURL:   "/payment/success",
		Email:       email,
		Phone:       phone,
	})
	if err != nil {
		return e.HTML(http.StatusInternalServerError, fmt.Sprintf(`<div class="p-3 bg-rose-50 text-rose-700 text-xs rounded-lg">Ошибка создания платежа: %s</div>`, html.EscapeString(err.Error())))
	}

	// Если оплата сформирована для заказа, привязываем ее к заказу
	if orderCode != "" {
		if rec, err := h.app.FindFirstRecordByFilter("service_orders", "order_code = {:code}", dbx.Params{"code": orderCode}); err == nil && rec != nil {
			rec.Set("is_paid", true)
			_ = h.app.Save(rec)

			h.eventBus.Publish(e.Request.Context(), events.OrderPaidEvent{
				OrderCode: orderCode,
				Amount:    amount,
				PaymentID: res.PaymentID,
			})
		}
	}

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return components.PaymentSuccessMessage(res).Render(e.Request.Context(), e.Response)
}
