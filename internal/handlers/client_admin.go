package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"sitecraft/internal/data"
	"sitecraft/internal/models"
	"sitecraft/ui/pages"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

const ClientAuthCookieName = "sitecraft_client_auth"

// ClientAdminHandler manages client administration routes.
type ClientAdminHandler struct {
	app *pocketbase.PocketBase
}

// NewClientAdminHandler creates a new ClientAdminHandler instance.
func NewClientAdminHandler(app *pocketbase.PocketBase) *ClientAdminHandler {
	return &ClientAdminHandler{app: app}
}

// isAuthorized checks if the client request contains a valid auth session cookie.
func (h *ClientAdminHandler) isAuthorized(e *core.RequestEvent) bool {
	cookie, err := e.Request.Cookie(ClientAuthCookieName)
	if err != nil || cookie == nil || cookie.Value == "" {
		return false
	}
	settings := data.GetSiteSettings(h.app)
	expectedPassword := settings.ClientAdminPassword
	if expectedPassword == "" {
		expectedPassword = "admin"
	}
	return cookie.Value == expectedPassword
}

// HandleLoginPage renders the login page or redirects to dashboard if already authenticated.
func (h *ClientAdminHandler) HandleLoginPage(e *core.RequestEvent) error {
	if h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client")
	}
	settings := data.GetSiteSettings(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientLogin(settings, "").Render(e.Request.Context(), e.Response)
}

// HandleLogin processes password submission.
func (h *ClientAdminHandler) HandleLogin(e *core.RequestEvent) error {
	settings := data.GetSiteSettings(h.app)
	password := strings.TrimSpace(e.Request.FormValue("password"))

	expectedPassword := settings.ClientAdminPassword
	if expectedPassword == "" {
		expectedPassword = "admin"
	}

	if password != expectedPassword {
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return pages.ClientLogin(settings, "Неверный пароль доступа").Render(e.Request.Context(), e.Response)
	}

	http.SetCookie(e.Response, &http.Cookie{
		Name:     ClientAuthCookieName,
		Value:    password,
		Path:     "/client",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return e.Redirect(http.StatusFound, "/client")
}

// HandleLogout clears auth cookie.
func (h *ClientAdminHandler) HandleLogout(e *core.RequestEvent) error {
	http.SetCookie(e.Response, &http.Cookie{
		Name:     ClientAuthCookieName,
		Value:    "",
		Path:     "/client",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	})
	return e.Redirect(http.StatusFound, "/client/login")
}

// HandleDashboard renders the main client cabinet.
func (h *ClientAdminHandler) HandleDashboard(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}

	settings := data.GetSiteSettings(h.app)
	content := data.GetSiteContent(h.app)
	leads := data.GetAllLeads(h.app)
	articles := data.GetAllArticlesAdmin(h.app)
	trashArticles := data.GetTrashArticlesAdmin(h.app)

	tab := e.Request.URL.Query().Get("tab")
	if tab == "" {
		tab = "leads"
	}

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientDashboard(settings, content, leads, articles, trashArticles, tab).Render(e.Request.Context(), e.Response)
}

// HandleTabLeads renders leads tab fragment for HTMX.
func (h *ClientAdminHandler) HandleTabLeads(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}
	leads := data.GetAllLeads(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientLeadsTab(leads).Render(e.Request.Context(), e.Response)
}

// HandleTabContacts renders contacts tab fragment for HTMX.
func (h *ClientAdminHandler) HandleTabContacts(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}
	settings := data.GetSiteSettings(h.app)
	content := data.GetSiteContent(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientContactsTab(content, settings).Render(e.Request.Context(), e.Response)
}

// HandleTabArticles renders articles tab fragment for HTMX.
func (h *ClientAdminHandler) HandleTabArticles(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}
	articles := data.GetAllArticlesAdmin(h.app)
	trashArticles := data.GetTrashArticlesAdmin(h.app)
	showTrash := e.Request.URL.Query().Get("view") == "trash"

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticlesTab(articles, trashArticles, showTrash).Render(e.Request.Context(), e.Response)
}

// HandleSaveContacts updates site_settings and site_content with validation.
func (h *ClientAdminHandler) HandleSaveContacts(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}

	siteName := strings.TrimSpace(e.Request.FormValue("site_name"))
	tagline := strings.TrimSpace(e.Request.FormValue("tagline"))
	whatsapp := strings.TrimSpace(e.Request.FormValue("whatsapp_phone"))
	tg := strings.TrimSpace(e.Request.FormValue("telegram_contact"))

	phone1 := strings.TrimSpace(e.Request.FormValue("phone1"))
	phone2 := strings.TrimSpace(e.Request.FormValue("phone2"))
	address := strings.TrimSpace(e.Request.FormValue("address"))
	addressSub := strings.TrimSpace(e.Request.FormValue("address_sub"))
	workHours := strings.TrimSpace(e.Request.FormValue("work_hours"))
	daysOff := strings.TrimSpace(e.Request.FormValue("days_off"))

	// Валидация полей формы
	var validationErrors []string

	if utf8.RuneCountInString(siteName) < 2 {
		validationErrors = append(validationErrors, "Название компании / сайта должно содержать от 2 до 100 символов")
	} else if utf8.RuneCountInString(siteName) > 100 {
		validationErrors = append(validationErrors, "Название компании / сайта не должно превышать 100 символов")
	}

	if utf8.RuneCountInString(address) < 5 {
		validationErrors = append(validationErrors, "Адрес офиса должен содержать не менее 5 символов")
	} else if utf8.RuneCountInString(address) > 200 {
		validationErrors = append(validationErrors, "Адрес офиса не должен превышать 200 символов")
	}

	if !isValidPhone(phone1) {
		validationErrors = append(validationErrors, "Основной телефон должен быть в формате +7 (999) 999-99-99")
	}

	if phone2 != "" && !isValidPhone(phone2) {
		validationErrors = append(validationErrors, "Дополнительный телефон должен быть в формате +7 (999) 999-99-99")
	}

	if whatsapp != "" && !isValidPhone(whatsapp) {
		validationErrors = append(validationErrors, "WhatsApp должен быть в формате +7 (999) 999-99-99")
	}

	if tg != "" && !isValidTelegram(tg) {
		validationErrors = append(validationErrors, "Telegram должен быть в формате @username или t.me/username")
	}

	if len(validationErrors) > 0 {
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		var errItems strings.Builder
		for _, errText := range validationErrors {
			errItems.WriteString(fmt.Sprintf("<li>%s</li>", errText))
		}
		return e.String(http.StatusOK, fmt.Sprintf(`<div class="text-xs text-rose-400 font-medium p-3 rounded-xl bg-rose-500/10 border border-rose-500/20"><div class="flex items-center gap-1.5 font-bold mb-1 text-rose-300"><svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg> <span>Ошибка валидации формы:</span></div><ul class="list-disc list-inside space-y-0.5">%s</ul></div>`, errItems.String()))
	}

	// Форматирование телефонов и телеграм
	formattedPhone1 := formatPhoneMask(phone1)
	formattedPhone2 := ""
	if phone2 != "" {
		formattedPhone2 = formatPhoneMask(phone2)
	}
	formattedWA := ""
	if whatsapp != "" {
		formattedWA = formatPhoneMask(whatsapp)
	}
	formattedTG := tg
	if formattedTG != "" && !strings.HasPrefix(formattedTG, "@") && !strings.HasPrefix(formattedTG, "http") && !strings.HasPrefix(formattedTG, "t.me") {
		formattedTG = "@" + formattedTG
	}

	// Обновление site_settings
	if setRec, err := h.app.FindFirstRecordByFilter("site_settings", ""); err == nil && setRec != nil {
		if siteName != "" {
			setRec.Set("site_name", siteName)
		}
		setRec.Set("tagline", tagline)
		setRec.Set("whatsapp_phone", formattedWA)
		setRec.Set("telegram_contact", formattedTG)
		_ = h.app.Save(setRec)
	}

	// Обновление site_content
	if cntRec, err := h.app.FindFirstRecordByFilter("site_content", ""); err == nil && cntRec != nil {
		cntRec.Set("phone1", formattedPhone1)
		cntRec.Set("phone2", formattedPhone2)
		cntRec.Set("address", address)
		cntRec.Set("address_sub", addressSub)
		cntRec.Set("work_hours", workHours)
		cntRec.Set("days_off", daysOff)
		_ = h.app.Save(cntRec)
	}

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return e.String(http.StatusOK, `<div class="text-xs text-emerald-400 font-medium flex items-center gap-1.5 p-2.5 rounded-xl bg-emerald-500/10 border border-emerald-500/20"><svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg> <span>Контакты и реквизиты успешно сохранены!</span></div>`)
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

func isValidPhone(phone string) bool {
	digits := extractDigits(phone)
	if len(digits) == 11 && (digits[0] == '7' || digits[0] == '8') {
		return true
	}
	if len(digits) == 10 {
		return true
	}
	return false
}

func isValidTelegram(tg string) bool {
	tg = strings.TrimSpace(tg)
	tg = strings.TrimPrefix(tg, "https://")
	tg = strings.TrimPrefix(tg, "http://")
	tg = strings.TrimPrefix(tg, "t.me/")
	tg = strings.TrimPrefix(tg, "@")
	if len(tg) < 3 || len(tg) > 32 {
		return false
	}
	for _, r := range tg {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}

func extractDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func formatPhoneMask(phone string) string {
	d := extractDigits(phone)
	if len(d) == 11 && (d[0] == '7' || d[0] == '8') {
		return fmt.Sprintf("+7 (%s) %s-%s-%s", d[1:4], d[4:7], d[7:9], d[9:11])
	}
	if len(d) == 10 {
		return fmt.Sprintf("+7 (%s) %s-%s-%s", d[0:3], d[3:6], d[6:8], d[8:10])
	}
	return phone
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

	_ = w.Write([]string{"ID", "Имя", "Телефон", "Email", "Дата записи", "Время", "Статус", "Источник", "Комментарий"})

	for _, l := range leads {
		comment := ""
		if l.Payload != nil {
			if c, ok := l.Payload["comment"].(string); ok {
				comment = c
			}
		}
		_ = w.Write([]string{
			l.ID,
			l.Name,
			l.Phone,
			l.Email,
			l.BookingDate,
			l.BookingTime,
			l.Status,
			l.Source,
			comment,
		})
	}
	w.Flush()

	filename := fmt.Sprintf("leads_%s.csv", time.Now().Format("2006-01-02"))
	e.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	e.Response.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	_, err := e.Response.Write(b.Bytes())
	return err
}

// HandleNewArticleForm renders form for creating a new article.
func (h *ClientAdminHandler) HandleNewArticleForm(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticleForm(models.ArticleItem{
		IsPublished: true,
		PublishedAt: time.Now().Format("2006-01-02 15:04"),
	}, true).Render(e.Request.Context(), e.Response)
}

// HandleEditArticleForm renders form for editing an existing article.
func (h *ClientAdminHandler) HandleEditArticleForm(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}

	id := e.Request.PathValue("id")
	rec, err := h.app.FindRecordById("articles", id)
	if err != nil || rec == nil {
		return e.String(http.StatusNotFound, "Статья не найдена")
	}

	article := data.RecordToArticleItem(rec)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticleForm(article, false).Render(e.Request.Context(), e.Response)
}

// HandleSaveArticle creates or updates an article.
func (h *ClientAdminHandler) HandleSaveArticle(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}

	id := strings.TrimSpace(e.Request.FormValue("id"))
	title := strings.TrimSpace(e.Request.FormValue("title"))
	slug := strings.TrimSpace(e.Request.FormValue("slug"))
	excerpt := strings.TrimSpace(e.Request.FormValue("excerpt"))
	content := strings.TrimSpace(e.Request.FormValue("content"))
	author := strings.TrimSpace(e.Request.FormValue("author_name"))
	tags := strings.TrimSpace(e.Request.FormValue("tags"))
	publishedAt := strings.TrimSpace(e.Request.FormValue("published_at"))
	removeCover := e.Request.FormValue("remove_cover") == "true"
	isPublished := e.Request.FormValue("is_published") == "true"

	articlesCol, err := h.app.FindCollectionByNameOrId("articles")
	if err != nil {
		return e.String(http.StatusInternalServerError, "Коллекция articles не найдена")
	}

	var rec *core.Record
	if id != "" {
		rec, _ = h.app.FindRecordById("articles", id)
	}
	if rec == nil {
		rec = core.NewRecord(articlesCol)
	}

	if publishedAt == "" {
		if rec.GetString("published_at") != "" {
			publishedAt = rec.GetString("published_at")
		} else {
			publishedAt = time.Now().Format("2006-01-02 15:04")
		}
	} else {
		publishedAt = strings.Replace(publishedAt, "T", " ", 1)
	}

	rec.Set("title", title)
	rec.Set("slug", slug)
	rec.Set("excerpt", excerpt)
	rec.Set("content", content)
	rec.Set("author_name", author)
	rec.Set("tags", tags)
	rec.Set("published_at", publishedAt)
	rec.Set("is_published", isPublished)
	rec.Set("is_deleted", false)

	if removeCover {
		rec.Set("cover_image", nil)
	}

	if mf, fileHeader, err := e.Request.FormFile("cover_image"); err == nil && fileHeader != nil && fileHeader.Filename != "" {
		_ = mf.Close()
		if file, err := filesystem.NewFileFromMultipart(fileHeader); err == nil {
			rec.Set("cover_image", file)
		}
	}

	if err := h.app.Save(rec); err != nil {
		return e.String(http.StatusInternalServerError, fmt.Sprintf("Ошибка сохранения: %v", err))
	}

	articles := data.GetAllArticlesAdmin(h.app)
	trashArticles := data.GetTrashArticlesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticlesTab(articles, trashArticles, false).Render(e.Request.Context(), e.Response)
}

// HandleToggleArticlePublish toggles article published status.
func (h *ClientAdminHandler) HandleToggleArticlePublish(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}
	id := e.Request.PathValue("id")
	if _, err := data.ToggleArticlePublish(h.app, id); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	articles := data.GetAllArticlesAdmin(h.app)
	trashArticles := data.GetTrashArticlesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticlesTab(articles, trashArticles, false).Render(e.Request.Context(), e.Response)
}

// HandleTrashArticle moves article to trash.
func (h *ClientAdminHandler) HandleTrashArticle(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}
	id := e.Request.PathValue("id")
	if err := data.MoveArticleToTrash(h.app, id); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	articles := data.GetAllArticlesAdmin(h.app)
	trashArticles := data.GetTrashArticlesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticlesTab(articles, trashArticles, false).Render(e.Request.Context(), e.Response)
}

// HandleRestoreArticle restores article from trash.
func (h *ClientAdminHandler) HandleRestoreArticle(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}
	id := e.Request.PathValue("id")
	if err := data.RestoreArticleFromTrash(h.app, id); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	articles := data.GetAllArticlesAdmin(h.app)
	trashArticles := data.GetTrashArticlesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticlesTab(articles, trashArticles, true).Render(e.Request.Context(), e.Response)
}

// HandlePermanentDeleteArticle permanently removes an article from database.
func (h *ClientAdminHandler) HandlePermanentDeleteArticle(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}
	id := e.Request.PathValue("id")
	if err := data.PermanentDeleteArticle(h.app, id); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	articles := data.GetAllArticlesAdmin(h.app)
	trashArticles := data.GetTrashArticlesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticlesTab(articles, trashArticles, true).Render(e.Request.Context(), e.Response)
}

// HandleEmptyTrashArticles permanently removes all articles in trash.
func (h *ClientAdminHandler) HandleEmptyTrashArticles(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}
	if err := data.EmptyTrashArticles(h.app); err != nil {
		return e.String(http.StatusBadRequest, err.Error())
	}
	articles := data.GetAllArticlesAdmin(h.app)
	trashArticles := data.GetTrashArticlesAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientArticlesTab(articles, trashArticles, true).Render(e.Request.Context(), e.Response)
}
