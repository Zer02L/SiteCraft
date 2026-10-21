package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"sitecraft/internal/testutil"

	"github.com/pocketbase/pocketbase/core"
)

func TestHTMXLeadFormSubmission(t *testing.T) {
	app := testutil.NewTestApp(t)

	handler := NewLeadHandler(app)

	// 1. Ошибки валидации: пустые обязательные поля -> HTTP 422 и HTML-фрагмент со списком ошибок
	t.Run("Validation failure returns 422 and error fragment", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", "")
		form.Set("phone", "invalid-phone")
		form.Set("consent", "") // 152-ФЗ не отмечен

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/leads", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request", "true")

		re := &core.RequestEvent{App: app, Request: req, Response: rec}
		if err := handler.HandleCreateLead(re); err != nil {
			t.Fatalf("Unexpected error from handler: %v", err)
		}

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("Expected status 422, got %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "bg-rose-50") {
			t.Errorf("Expected rose error alert in response, got: %s", body)
		}
		if !strings.Contains(body, "flex items-start gap-2") {
			t.Errorf("Expected structured error list items in response, got: %s", body)
		}
	})

	// 2. Антиспам ловушка (Honeypot): скрытое поле заполнено ботом -> HTTP 422
	t.Run("Honeypot detects spam bot and rejects", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", "Робот Спамер")
		form.Set("phone", "+79991234567")
		form.Set("consent", "on")
		form.Set("company_website_url", "https://spam-link.org")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/leads", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request", "true")

		re := &core.RequestEvent{App: app, Request: req, Response: rec}
		if err := handler.HandleCreateLead(re); err != nil {
			t.Fatalf("Unexpected error from handler: %v", err)
		}

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("Expected status 422 for bot submission, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "Спам-фильтр") {
			t.Errorf("Expected honeypot error message, got: %s", rec.Body.String())
		}
	})

	// 3. Успешная запись на конкретный слот -> HTTP 200, плашка успеха и .ics ссылка
	t.Run("Successful slot booking returns 200 with ics download link", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", "Алексей Смирнов")
		form.Set("phone", "+7 (999) 555-44-33")
		form.Set("booking_date", "2026-12-05")
		form.Set("booking_time", "14:00")
		form.Set("consent", "on")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/leads", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request", "true")

		re := &core.RequestEvent{App: app, Request: req, Response: rec}
		if err := handler.HandleCreateLead(re); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "bg-emerald-50") {
			t.Errorf("Expected emerald success alert, got: %s", body)
		}
		if !strings.Contains(body, "visit.ics") {
			t.Errorf("Expected visit.ics calendar download link, got: %s", body)
		}
		if !strings.Contains(body, "+7 (999) 555-44-33") {
			t.Errorf("Expected customer phone in confirmation, got: %s", body)
		}

		// Проверяем, что заявка физически создалась в базе
		recCount, err := app.CountRecords("leads")
		if err != nil || recCount == 0 {
			t.Errorf("Expected lead record saved in DB, count=%d, err=%v", recCount, err)
		}
	})

	// 4. Коллизия слота (Double Booking): повторная запись на занятое время -> HTTP 409 Conflict
	t.Run("Duplicate slot booking returns 409 Conflict", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", "Другой Клиент")
		form.Set("phone", "+7 (999) 000-11-22")
		form.Set("booking_date", "2026-12-05")
		form.Set("booking_time", "14:00") // Тот же слот!
		form.Set("consent", "on")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/leads", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request", "true")

		re := &core.RequestEvent{App: app, Request: req, Response: rec}
		if err := handler.HandleCreateLead(re); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if rec.Code != http.StatusConflict {
			t.Errorf("Expected status 409 Conflict for double booking, got %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "bg-amber-50") {
			t.Errorf("Expected amber warning alert for taken slot, got: %s", body)
		}
	})
}

func TestHTMXReviewFormSubmission(t *testing.T) {
	app := testutil.NewTestApp(t)

	handler := NewReviewHandler(app)

	// 1. Ошибка валидации: пустое имя или короткий комментарий
	t.Run("Review short inputs return 422", func(t *testing.T) {
		form := url.Values{}
		form.Set("author_name", "")
		form.Set("comment", "Ок")
		form.Set("rating", "5")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/reviews", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request", "true")

		re := &core.RequestEvent{App: app, Request: req, Response: rec}
		if err := handler.HandleSubmitReview(re); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("Expected 422 for short input, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "не менее 2 символов") {
			t.Errorf("Expected min name length message, got: %s", rec.Body.String())
		}
	})

	// 2. Антиспам ловушка: honeypot возвращает 200 без создания отзыва
	t.Run("Honeypot returns 200 dummy response", func(t *testing.T) {
		form := url.Values{}
		form.Set("author_name", "Бот")
		form.Set("comment", "Отличный сервис, заходите на наш сайт")
		form.Set("website_company_url", "http://spam.ru")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/reviews", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request", "true")

		re := &core.RequestEvent{App: app, Request: req, Response: rec}
		if err := handler.HandleSubmitReview(re); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Expected 200 for bot trap, got %d", rec.Code)
		}
	})

	// 3. Обычный отзыв без номера заказа -> премодерация
	t.Run("Standard review sends to moderation queue", func(t *testing.T) {
		form := url.Values{}
		form.Set("author_name", "Дмитрий")
		form.Set("comment", "Все сделали быстро и без лишней бюрократии!")
		form.Set("rating", "5")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/reviews", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request", "true")

		re := &core.RequestEvent{App: app, Request: req, Response: rec}
		if err := handler.HandleSubmitReview(re); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "после быстрой проверки модератором") {
			t.Errorf("Expected moderation notice, got: %s", rec.Body.String())
		}
	})

	// 4. Проверенный отзыв с валидным кодом заказа (SRV-4182 из начальных данных) -> авто-публикация
	t.Run("Verified buyer review auto publishes", func(t *testing.T) {
		form := url.Values{}
		form.Set("author_name", "Сергей С.")
		form.Set("comment", "Заказ выполнен отлично, трекинг статуса очень удобный.")
		form.Set("rating", "5")
		form.Set("order_code", "SRV-4182")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/reviews", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request", "true")

		re := &core.RequestEvent{App: app, Request: req, Response: rec}
		if err := handler.HandleSubmitReview(re); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "проверенный отзыв! Он успешно опубликован") {
			t.Errorf("Expected verified buyer success notice, got: %s", rec.Body.String())
		}
	})
}

func TestHTMXCouponValidation(t *testing.T) {
	app := testutil.NewTestApp(t)

	handler := NewCartHandler(app)

	// 1. Пустой купон
	t.Run("Empty coupon returns prompt", func(t *testing.T) {
		form := url.Values{}
		form.Set("code", "")
		form.Set("subtotal", "2000")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/cart/coupon", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		re := &core.RequestEvent{App: app, Request: req, Response: rec}
		if err := handler.HandleValidateCoupon(re); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !strings.Contains(rec.Body.String(), "Введите промокод") {
			t.Errorf("Expected prompt to enter coupon, got: %s", rec.Body.String())
		}
	})

	// 2. Несуществующий купон
	t.Run("Invalid coupon returns not found", func(t *testing.T) {
		form := url.Values{}
		form.Set("code", "FAKE_COUPON_123")
		form.Set("subtotal", "2000")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/cart/coupon", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		re := &core.RequestEvent{App: app, Request: req, Response: rec}
		if err := handler.HandleValidateCoupon(re); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !strings.Contains(rec.Body.String(), "не найден или неактивен") {
			t.Errorf("Expected not found error, got: %s", rec.Body.String())
		}
	})

	// 3. Валидный купон PROMO10 (10% скидка при сумме >= 1000)
	t.Run("Valid coupon PROMO10 calculates discount fragment", func(t *testing.T) {
		form := url.Values{}
		form.Set("code", "PROMO10")
		form.Set("subtotal", "3000")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/cart/coupon", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		re := &core.RequestEvent{App: app, Request: req, Response: rec}
		if err := handler.HandleValidateCoupon(re); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "Промокод PROMO10") {
			t.Errorf("Expected applied coupon notice, got: %s", body)
		}
		if !strings.Contains(body, "-300 ₽") { // 10% от 3000 = 300 ₽
			t.Errorf("Expected -300 ₽ discount in body, got: %s", body)
		}
	})
}
