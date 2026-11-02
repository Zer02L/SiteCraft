package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sitecraft/internal/data"
	"sitecraft/internal/events"
	"sitecraft/internal/i18n"
	"sitecraft/internal/models"
	"sitecraft/internal/modules"

	"ganigeorgiev.com/pocketbase/dbx"
	"ganigeorgiev.com/pocketbase/pocketbase"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

// CartHandler handles web e-commerce cart checkouts, promo validation, and delivery.
type CartHandler struct {
	app      *pocketbase.PocketBase
	eventBus events.Bus
}

// NewCartHandler creates a new CartHandler instance.
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

// HandleValidateCoupon validates promo coupons and calculates dynamic discount.
func (h *CartHandler) HandleValidateCoupon(e *core.RequestEvent) error {
	code := strings.TrimSpace(strings.ToUpper(e.Request.FormValue("code")))
	subtotalStr := strings.TrimSpace(e.Request.FormValue("subtotal"))
	subtotal, _ := strconv.ParseFloat(subtotalStr, 64)

	if code == "" {
		return e.HTML(http.StatusOK, `<span class="text-xs text-rose-500">Введите промокод</span>`)
	}

	rec, err := h.app.FindFirstRecordByFilter("coupons", "code = {:code} && active = true", dbx.Params{"code": code})
	if err != nil || rec == nil {
		return e.HTML(http.StatusOK, `<span class="text-xs text-rose-500 font-medium">Промокод не найден или неактивен</span>`)
	}

	minOrder := rec.GetFloat("min_order_amount")
	if minOrder > 0 && subtotal < minOrder {
		return e.HTML(http.StatusOK, fmt.Sprintf(`<span class="text-xs text-amber-500">Минимальная сумма заказа для этого купона: <strong>%.0f ₽</strong></span>`, minOrder))
	}

	usageLimit := rec.GetInt("usage_limit")
	usedCount := rec.GetInt("used_count")
	if usageLimit > 0 && usedCount >= usageLimit {
		return e.HTML(http.StatusOK, `<span class="text-xs text-rose-500">Лимит активаций промокода исчерпан</span>`)
	}

	discType := rec.GetString("discount_type")
	discVal := rec.GetFloat("discount_value")
	var discount float64
	if discType == "percent" {
		discount = subtotal * (discVal / 100.0)
	} else {
		discount = discVal
	}
	if discount > subtotal {
		discount = subtotal
	}

	return e.HTML(http.StatusOK, fmt.Sprintf(`
		<div class="p-2.5 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs flex items-center justify-between">
			<div class="flex items-center gap-1.5 font-bold">
				<span>✓ Промокод %s:</span>
				<span class="text-accent">-%.0f ₽</span>
			</div>
			<input type="hidden" name="applied_coupon" value="%s"/>
			<input type="hidden" name="discount_amount" value="%.2f"/>
		</div>
	`, html.EscapeString(code), discount, html.EscapeString(code), discount))
}

// HandleCheckout processes the cart checkout form submission.
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

	// Новые параметры Фазы 2: доставка и промокоды
	deliveryType := strings.TrimSpace(e.Request.FormValue("delivery_type"))
	deliveryZone := strings.TrimSpace(e.Request.FormValue("delivery_zone"))
	appliedCoupon := strings.TrimSpace(strings.ToUpper(e.Request.FormValue("applied_coupon")))

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

	// 1. Подсчет промежуточной суммы товаров (subtotal)
	var itemsSubtotal float64
	var itemTitles []string
	for _, item := range cartItems {
		itemsSubtotal += (item.Price * float64(item.Qty))
		itemTitles = append(itemTitles, fmt.Sprintf("%s (x%d)", item.Title, item.Qty))
	}
	itemsSummary := strings.Join(itemTitles, ", ")

	// 2. Расчет гео-зоны доставки (Модуль #1)
	var deliveryCost float64 = 0
	deliveryTitle := "Самовывоз"
	if deliveryType == "courier" {
		switch deliveryZone {
		case "suburb":
			deliveryCost = 650
			deliveryTitle = "Курьер: Пригород (до 15 км)"
		case "region":
			deliveryCost = 1200
			deliveryTitle = "Курьер: Область (до 30 км)"
		default: // city
			if itemsSubtotal >= 2500 {
				deliveryCost = 0
				deliveryTitle = "Курьер: По городу (Бесплатно от 2500 ₽)"
			} else {
				deliveryCost = 300
				deliveryTitle = "Курьер: По городу (300 ₽)"
			}
		}
	}

	// 3. Валидация и списание промокода (Модуль #4)
	var discountAmount float64 = 0
	if appliedCoupon != "" {
		if cRec, err := h.app.FindFirstRecordByFilter("coupons", "code = {:code} && active = true", dbx.Params{"code": appliedCoupon}); err == nil && cRec != nil {
			minOrder := cRec.GetFloat("min_order_amount")
			usageLimit := cRec.GetInt("usage_limit")
			usedCount := cRec.GetInt("used_count")
			if (minOrder == 0 || itemsSubtotal >= minOrder) && (usageLimit == 0 || usedCount < usageLimit) {
				discType := cRec.GetString("discount_type")
				discVal := cRec.GetFloat("discount_value")
				if discType == "percent" {
					discountAmount = itemsSubtotal * (discVal / 100.0)
				} else {
					discountAmount = discVal
				}
				if discountAmount > itemsSubtotal {
					discountAmount = itemsSubtotal
				}
				// Фиксируем использование купона
				cRec.Set("used_count", usedCount+1)
				_ = h.app.Save(cRec)
			}
		}
	}

	// 4. Итоговая сумма заказа
	totalAmount := (itemsSubtotal - discountAmount) + deliveryCost
	if totalAmount < 0 {
		totalAmount = 0
	}

	// 5. Авто-списание остатков со склада Simple Inventory (Модуль #3)
	if servicesCol, err := h.app.FindCollectionByNameOrId("services"); err == nil {
		for _, item := range cartItems {
			if item.ID != "" && !item.IsUpsell {
				if sRec, err := h.app.FindRecordById(servicesCol, item.ID); err == nil && sRec != nil {
					if sRec.Collection().Fields.GetByName("stock_count") != nil {
						currStock := sRec.GetInt("stock_count")
						if currStock > 0 {
							newStock := currStock - item.Qty
							if newStock <= 0 {
								newStock = 0
								sRec.Set("is_preorder", true)
							}
							sRec.Set("stock_count", newStock)
							_ = h.app.Save(sRec)
						}
					}
				}
			}
		}
	}

	// Генерируем уникальный номер заказа
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	orderCode := fmt.Sprintf("ORD-%04d", r.Intn(10000))

	// 6. Сохраняем в коллекцию service_orders для трекинга (/track/ORD-XXXX)
	demoHistory := []models.OrderStageHistory{
		{
			Stage:     "received",
			Title:     "Заказ принят и передан в обработку",
			Timestamp: time.Now().Format("02.01 15:04"),
			Comment:   fmt.Sprintf("Оформлен на сайте. Позиций: %d шт. Доставка: %s", len(cartItems), deliveryTitle),
			Done:      true,
		},
	}

	if ordersCol, err := h.app.FindCollectionByNameOrId("service_orders"); err == nil {
		orderRec := core.NewRecord(ordersCol)
		orderRec.Set("order_code", orderCode)
		orderRec.Set("client_name", name)
		orderRec.Set("client_phone", phone)
		orderRec.Set("item_title", fmt.Sprintf("Заказ: %s", itemsSummary))
		orderRec.Set("current_stage", "received")
		orderRec.Set("total_amount", totalAmount)
		orderRec.Set("is_paid", false)
		orderRec.Set("stages_history", demoHistory)
		orderRec.Set("notes", comment)
		if err := h.app.Save(orderRec); err != nil {
			log.Printf("[Cart Error] Не удалось сохранить service_order: %v\n", err)
		}
	}

	// 7. Сохраняем в коллекцию leads для CRM
	payloadMap := map[string]interface{}{
		"source":          "web_cart",
		"order_code":      orderCode,
		"items_count":     len(cartItems),
		"items":           cartItems,
		"items_subtotal":  itemsSubtotal,
		"delivery_type":   deliveryType,
		"delivery_title":  deliveryTitle,
		"delivery_cost":   deliveryCost,
		"coupon_code":     appliedCoupon,
		"discount_amount": discountAmount,
		"total_amount":    totalAmount,
		"comment":         comment,
		"pay_online":      payOnlineStr == "true",
		"ab_variant":      GetABVariant(ctx),
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
						<div class="flex justify-between"><span>Доставка:</span><span>%s</span></div>
						<div class="flex justify-between"><span>Позиций:</span><span class="font-bold">%d шт.</span></div>
						<div class="flex justify-between font-bold text-accent pt-1 border-t border-content/10"><span>Итого:</span><span>%.0f ₽</span></div>
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
			`, orderCode, totalAmount, html.EscapeString(name), html.EscapeString(deliveryTitle), len(cartItems), totalAmount, payResp.PaymentURL, orderCode))
		}
	}

	// Обычный сценарий (оплата при получении)
	discountHTML := ""
	if discountAmount > 0 {
		discountHTML = fmt.Sprintf(`<div class="flex justify-between text-accent"><span>Скидка (%s):</span><span>-%.0f ₽</span></div>`, html.EscapeString(appliedCoupon), discountAmount)
	}

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
				<div class="flex justify-between"><span>Товары (сумма):</span><span>%.0f ₽</span></div>
				%s
				<div class="flex justify-between"><span>Доставка:</span><span>%s</span></div>
				<div class="flex justify-between pt-1 border-t border-content/10 font-bold"><span>Итого к оплате:</span><span class="text-accent text-sm">%.0f ₽</span></div>
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
	`, orderCode, orderCode, html.EscapeString(name), itemsSubtotal, discountHTML, html.EscapeString(deliveryTitle), totalAmount, orderCode))
}
