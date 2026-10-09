package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"sitecraft/internal/data"
	"sitecraft/internal/events"
	"sitecraft/internal/i18n"
	"sitecraft/internal/models"
	"sitecraft/internal/modules"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// CartHandler handles web e-commerce cart checkouts
type CartHandler struct {
	app      *pocketbase.PocketBase
	eventBus events.Bus
}

// NewCartHandler creates a new CartHandler instance
func NewCartHandler(app *pocketbase.PocketBase, bus ...events.Bus) *CartHandler {
	var eb events.Bus = events.Default()
	if len(bus) > 0 && bus[0] != nil {
		eb = bus[0]
	}
	return &CartHandler{
		app:      app,
		eventBus: eb,
	}
}

// HandleCheckout processes the cart checkout form submission
func (h *CartHandler) HandleCheckout(e *core.RequestEvent) error {
	ctx := e.Request.Context()
	lang := i18n.GetLocale(ctx)
	settings := data.GetSiteSettings(h.app)

	name := strings.TrimSpace(e.Request.FormValue("name"))
	phone := strings.TrimSpace(e.Request.FormValue("phone"))
	comment := strings.TrimSpace(e.Request.FormValue("comment"))
	payOnlineStr := e.Request.FormValue("pay_online")
	honeypot := e.Request.FormValue("company_website_url")
	cartPayloadRaw := e.Request.FormValue("cart_payload")

	input := LeadFormInput{
		Name:     name,
		Phone:    phone,
		Consent:  "on",
		Honeypot: honeypot,
		Mode:     "simple",
	}

	validationErrors := ValidateLeadFormLocalized(input, lang)
	if len(validationErrors) > 0 {
		var errList strings.Builder
		for _, msg := range validationErrors {
			errList.WriteString(fmt.Sprintf(
				`<li class="flex items-start gap-2"><span class="text-rose-500">•</span><span>%s</span></li>`,
				html.EscapeString(msg),
			))
		}
		errorHTML := fmt.Sprintf(`
			<div class="p-4 bg-rose-50 border border-rose-300 rounded-2xl text-rose-900 text-xs shadow-sm space-y-2">
				<div class="font-bold flex items-center gap-2 text-rose-800">
					<span>Ошибка валидации формы</span>
				</div>
				<ul class="space-y-1">
					%s
				</ul>
			</div>
		`, errList.String())
		return e.HTML(http.StatusUnprocessableEntity, errorHTML)
	}

	// Десериализуем товары корзины
	var cartItems []models.CartItem
	if cartPayloadRaw != "" {
		_ = json.Unmarshal([]byte(cartPayloadRaw), &cartItems)
	}

	if len(cartItems) == 0 {
		return e.HTML(http.StatusBadRequest, `
			<div class="p-4 bg-amber-50 border border-amber-300 rounded-2xl text-amber-900 text-xs">
				Ваша корзина пуста. Добавьте товары перед оформлением.
			</div>
		`)
	}

	// Подсчет общей суммы и сводки
	var totalAmount float64
	var itemTitles []string
	for _, item := range cartItems {
		totalAmount += (item.Price * float64(item.Qty))
		itemTitles = append(itemTitles, fmt.Sprintf("%s (x%d)", item.Title, item.Qty))
	}
	itemsSummary := strings.Join(itemTitles, ", ")

	// Генерируем уникальный номер заказа
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	orderCode := fmt.Sprintf("ORD-%04d", r.Intn(10000))

	// 1. Сохраняем в коллекцию service_orders для трекинга (/track/ORD-XXXX)
	demoHistory := []models.OrderStageHistory{
		{
			Stage:     "received",
			Title:     "Заказ принят и передан в обработку",
			Timestamp: time.Now().Format("02.01 15:04"),
			Comment:   fmt.Sprintf("Оформлен на сайте. Позиций: %d шт.", len(cartItems)),
			Done:      true,
		},
	}

	if ordersCol, err := h.app.FindCollectionByNameOrId("service_orders"); err == nil {
		orderRec := core.NewRecord(ordersCol)
		orderRec.Set("order_code", orderCode)
		orderRec.Set("client_name", name)
		orderRec.Set("client_phone", phone)
		orderRec.Set("item_title", fmt.Sprintf("Заказ товаров/услуг: %s", itemsSummary))
		orderRec.Set("current_stage", "received")
		orderRec.Set("total_amount", totalAmount)
		orderRec.Set("is_paid", false)
		orderRec.Set("stages_history", demoHistory)
		orderRec.Set("notes", comment)
		if err := h.app.Save(orderRec); err != nil {
			log.Printf("[Cart Error] Не удалось сохранить service_order: %v\n", err)
		}
	}

	// 2. Сохраняем в коллекцию leads для CRM
	payloadMap := map[string]interface{}{
		"source":       "web_cart",
		"order_code":   orderCode,
		"items_count":  len(cartItems),
		"items":        cartItems,
		"total_amount": totalAmount,
		"comment":      comment,
		"pay_online":   payOnlineStr == "true",
	}

	leadModel := models.Lead{
		Name:    name,
		Phone:   phone,
		Status:  "новый заказ",
		Source:  "web_cart",
		Payload: payloadMap,
	}

	if leadsCol, err := h.app.FindCollectionByNameOrId("leads"); err == nil {
		leadRec := core.NewRecord(leadsCol)
		leadRec.Set("name", name)
		leadRec.Set("phone", phone)
		leadRec.Set("status", "новый заказ")
		leadRec.Set("source", "web_cart")
		leadRec.Set("payload", payloadMap)
		_ = h.app.Save(leadRec)
		leadModel.ID = leadRec.Id
	}

	// Публикуем событие через EventBus для Telegram и CRM
	h.eventBus.Publish(ctx, events.LeadCreatedEvent{
		Settings: settings,
		Lead:     leadModel,
	})

	// Сигнализируем клиенту очистить корзину в LocalStorage
	e.Response.Header().Set("HX-Trigger", "cart-cleared")

	// Если выбрана онлайн-оплата и эквайринг включен
	if payOnlineStr == "true" && settings.EnablePayment && totalAmount > 0 {
		payReq := models.PaymentRequest{
			Amount:      totalAmount,
			Currency:    "RUB",
			Description: fmt.Sprintf("Заказ %s (%s)", orderCode, name),
			Phone:       phone,
			ReturnURL:   fmt.Sprintf("/track/%s", orderCode),
		}

		provider := modules.NewPaymentProvider(settings)
		payResp, err := provider.CreatePayment(payReq)
		if err == nil && payResp.PaymentURL != "" {
			return e.HTML(http.StatusOK, fmt.Sprintf(`
				<div class="py-8 text-center space-y-5">
					<div class="w-16 h-16 rounded-full bg-emerald-500/10 text-emerald-500 mx-auto flex items-center justify-center">
						<svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
					</div>
					<div>
						<span class="inline-block px-3 py-1 rounded-full bg-accent/15 text-accent text-xs font-bold mb-2">Заказ № %s</span>
						<h3 class="text-xl font-bold text-content">Заказ успешно сформирован!</h3>
						<p class="text-xs text-content-muted mt-1">Сумма к оплате: <strong>%.0f ₽</strong></p>
					</div>

					<div class="p-4 rounded-2xl bg-surface-muted border border-content/10 text-left text-xs space-y-1.5">
						<div class="flex justify-between"><span>Клиент:</span><span class="font-bold">%s</span></div>
						<div class="flex justify-between"><span>Телефон:</span><span class="font-bold">%s</span></div>
						<div class="flex justify-between"><span>Позиций:</span><span class="font-bold">%d шт.</span></div>
					</div>

					<div class="space-y-2 pt-2">
						<a
							href="%s"
							class="w-full inline-flex items-center justify-center gap-2 py-3.5 px-6 rounded-2xl bg-accent text-primary-dark font-extrabold text-sm shadow-xl hover:brightness-105 transition"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z"></path></svg>
							<span>Перейти к безопасной оплате</span>
						</a>
						<a
							href="/track/%s"
							class="w-full inline-block text-center py-2 text-xs text-content-muted hover:text-accent transition font-semibold"
						>
							Отслеживать статус заказа в трекере →
						</a>
					</div>
				</div>
			`, orderCode, totalAmount, html.EscapeString(name), html.EscapeString(phone), len(cartItems), payResp.PaymentURL, orderCode))
		}
	}

	// Обычный сценарий (оплата при получении)
	return e.HTML(http.StatusOK, fmt.Sprintf(`
		<div class="py-8 text-center space-y-5">
			<div class="w-16 h-16 rounded-full bg-emerald-500/10 text-emerald-500 mx-auto flex items-center justify-center">
				<svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
			</div>
			<div>
				<span class="inline-block px-3 py-1 rounded-full bg-accent/15 text-accent text-xs font-bold mb-2">Заказ № %s</span>
				<h3 class="text-xl font-bold text-content">Спасибо за заказ!</h3>
				<p class="text-xs text-content-muted mt-1">Менеджер уже получил спецификацию и свяжется с вами для подтверждения.</p>
			</div>

			<div class="p-4 rounded-2xl bg-surface-muted border border-content/10 text-left text-xs space-y-1.5">
				<div class="flex justify-between"><span>Номер для отслеживания:</span><span class="font-bold text-accent">%s</span></div>
				<div class="flex justify-between"><span>Клиент:</span><span class="font-bold">%s</span></div>
				<div class="flex justify-between"><span>Сумма заказа:</span><span class="font-bold">%.0f ₽</span></div>
				<div class="flex justify-between"><span>Оплата:</span><span class="font-bold">При получении</span></div>
			</div>

			<div class="space-y-2 pt-2">
				<a
					href="/track/%s"
					class="w-full inline-flex items-center justify-center gap-2 py-3 px-6 rounded-2xl bg-surface border border-content/15 text-content hover:border-accent text-xs font-bold transition shadow-sm"
				>
					<span>Открыть заказ в онлайн-трекере</span>
					<svg class="w-4 h-4 text-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3"></path></svg>
				</a>
			</div>
		</div>
	`, orderCode, orderCode, html.EscapeString(name), totalAmount, orderCode))
}
