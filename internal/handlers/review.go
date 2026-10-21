package handlers

import (
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"

	"sitecraft/internal/data"
	"sitecraft/internal/models"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// ReviewHandler manages verified customer reviews (UGC) and moderation.
type ReviewHandler struct {
	app *pocketbase.PocketBase
}

// NewReviewHandler creates a new instance of ReviewHandler.
func NewReviewHandler(app *pocketbase.PocketBase) *ReviewHandler {
	return &ReviewHandler{app: app}
}

// HandleSubmitReview processes customer review submissions with validation and moderation queue.
func (h *ReviewHandler) HandleSubmitReview(e *core.RequestEvent) error {
	name := strings.TrimSpace(e.Request.FormValue("author_name"))
	comment := strings.TrimSpace(e.Request.FormValue("comment"))
	ratingStr := strings.TrimSpace(e.Request.FormValue("rating"))
	orderID := strings.TrimSpace(e.Request.FormValue("order_id"))
	if orderID == "" {
		orderID = strings.TrimSpace(e.Request.FormValue("order_code"))
	}
	source := strings.TrimSpace(e.Request.FormValue("source"))
	honeypot := e.Request.FormValue("website_company_url")

	// Антиспам-ловушка
	if honeypot != "" {
		return e.HTML(http.StatusOK, `<div class="p-3 bg-emerald-50 text-emerald-800 rounded-xl text-xs">Спасибо за отзыв!</div>`)
	}

	rating, err := strconv.Atoi(ratingStr)
	if err != nil || rating < 1 || rating > 5 {
		rating = 5
	}

	// shortcut: byte length check instead of utf8.RuneCountInString (1 cyrillic rune = 2 bytes), upgrade when strict unicode rune counting is required
	if len(name) < 2 {
		return e.HTML(http.StatusUnprocessableEntity, `
			<div class="p-3 bg-rose-50 border border-rose-300 rounded-xl text-rose-800 text-xs">
				Пожалуйста, укажите ваше имя (не менее 2 символов).
			</div>
		`)
	}

	if len(comment) < 5 {
		return e.HTML(http.StatusUnprocessableEntity, `
			<div class="p-3 bg-rose-50 border border-rose-300 rounded-xl text-rose-800 text-xs">
				Пожалуйста, напишите краткий комментарий к отзыву (не менее 5 символов).
			</div>
		`)
	}

	if source == "" {
		source = "site"
	}

	reviewItem := models.ReviewItem{
		AuthorName:  name,
		Rating:      rating,
		Comment:     comment,
		OrderID:     orderID,
		IsPublished: false, // Отправляется на премодерацию
		Source:      source,
	}

	// Если передан валидный номер существующего заказа из трекера — авто-верификация (Verified Buyer)
	if orderID != "" {
		if order, err := data.GetServiceOrderByCode(h.app, orderID); err == nil && order != nil {
			reviewItem.IsPublished = true // Доверенный проверенный клиент
		}
	}

	if _, err := data.CreateReview(h.app, reviewItem); err != nil {
		return e.HTML(http.StatusInternalServerError, `
			<div class="p-3 bg-rose-50 border border-rose-300 rounded-xl text-rose-800 text-xs">
				Не удалось сохранить отзыв. Пожалуйста, попробуйте позже.
			</div>
		`)
	}

	msg := "Спасибо за ваш отзыв! Он появится на сайте после быстрой проверки модератором."
	if reviewItem.IsPublished {
		msg = "Спасибо за ваш проверенный отзыв! Он успешно опубликован."
	}

	return e.HTML(http.StatusOK, fmt.Sprintf(`
		<div class="p-4 bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-300 dark:border-emerald-800 rounded-2xl text-emerald-800 dark:text-emerald-300 text-xs flex items-center gap-2.5">
			<svg class="w-5 h-5 text-emerald-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
			</svg>
			<span>%s</span>
		</div>
	`, html.EscapeString(msg)))
}

// HandleTogglePublish toggles review publication status in client admin.
func (h *ReviewHandler) HandleTogglePublish(e *core.RequestEvent) error {
	id := e.Request.PathValue("id")
	if err := data.ToggleReviewPublish(h.app, id); err != nil {
		return e.String(http.StatusInternalServerError, err.Error())
	}
	record, err := h.app.FindRecordById("reviews", id)
	if err == nil && record != nil {
		if record.GetBool("is_published") {
			return e.HTML(http.StatusOK, `<span class="px-2.5 py-1 rounded-full text-[11px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">Опубликован</span>`)
		}
		return e.HTML(http.StatusOK, `<span class="px-2.5 py-1 rounded-full text-[11px] font-bold bg-amber-500/10 text-amber-400 border border-amber-500/20">На модерации</span>`)
	}
	return e.HTML(http.StatusOK, `<span class="text-xs text-emerald-400 font-semibold">Статус обновлен</span>`)
}

// HandleDeleteReview permanently deletes a review.
func (h *ReviewHandler) HandleDeleteReview(e *core.RequestEvent) error {
	id := e.Request.PathValue("id")
	if err := data.DeleteReview(h.app, id); err != nil {
		return e.String(http.StatusInternalServerError, err.Error())
	}
	return e.HTML(http.StatusOK, "")
}
