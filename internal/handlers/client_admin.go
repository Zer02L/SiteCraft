package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"sitecraft/internal/cache"
	"sitecraft/internal/data"
	"sitecraft/internal/models"
	"sitecraft/ui/pages"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"golang.org/x/crypto/bcrypt"
)

const ClientAuthCookieName = "sitecraft_client_auth"

type sessionInfo struct {
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionStore provides thread-safe session token management for client admin.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]sessionInfo
}

// NewSessionStore initializes a new SessionStore.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]sessionInfo),
	}
}

// CreateSession generates a secure 32-byte session token.
func (s *SessionStore) CreateSession() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[token] = sessionInfo{
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	return token, nil
}

// IsValid checks if the token is valid and unexpired.
func (s *SessionStore) IsValid(token string) bool {
	if token == "" {
		return false
	}
	s.mu.RLock()
	info, ok := s.sessions[token]
	s.mu.RUnlock()
	if !ok {
		return false
	}
	if time.Now().After(info.ExpiresAt) {
		s.RevokeSession(token)
		return false
	}
	return true
}

// RevokeSession deletes a session token.
func (s *SessionStore) RevokeSession(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

// RateLimiter manages failed login attempts to prevent brute-force attacks.
type RateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
}

type loginAttempt struct {
	count     int
	firstFail time.Time
	blockedTo time.Time
}

const (
	maxLoginAttempts = 5
	attemptWindow    = 15 * time.Minute
	blockDuration    = 15 * time.Minute
)

// NewRateLimiter initializes a new RateLimiter.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		attempts: make(map[string]*loginAttempt),
	}
}

// IsBlocked returns true if IP is currently blocked from logging in.
func (r *RateLimiter) IsBlocked(ip string) (bool, time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	att, ok := r.attempts[ip]
	if !ok {
		return false, 0
	}
	now := time.Now()
	if now.Before(att.blockedTo) {
		return true, att.blockedTo.Sub(now)
	}
	if att.count >= maxLoginAttempts && now.After(att.blockedTo) {
		delete(r.attempts, ip)
		return false, 0
	}
	if now.Sub(att.firstFail) > attemptWindow {
		delete(r.attempts, ip)
		return false, 0
	}
	return false, 0
}

// RecordFailure records a failed login attempt and blocks IP if limit exceeded.
func (r *RateLimiter) RecordFailure(ip string) (isBlocked bool, waitDuration time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	att, ok := r.attempts[ip]
	if !ok || now.Sub(att.firstFail) > attemptWindow {
		att = &loginAttempt{
			count:     1,
			firstFail: now,
		}
		r.attempts[ip] = att
		return false, 0
	}

	att.count++
	if att.count >= maxLoginAttempts {
		att.blockedTo = now.Add(blockDuration)
		return true, blockDuration
	}
	return false, 0
}

// Reset clears failure tracking for an IP upon successful login.
func (r *RateLimiter) Reset(ip string) {
	r.mu.Lock()
	delete(r.attempts, ip)
	r.mu.Unlock()
}

// ClientAdminHandler manages client administration routes.
type ClientAdminHandler struct {
	app          *pocketbase.PocketBase
	sessionStore *SessionStore
	rateLimiter  *RateLimiter
}

// NewClientAdminHandler creates a new ClientAdminHandler instance.
func NewClientAdminHandler(app *pocketbase.PocketBase) *ClientAdminHandler {
	return &ClientAdminHandler{
		app:          app,
		sessionStore: NewSessionStore(),
		rateLimiter:  NewRateLimiter(),
	}
}

// SessionStore returns the active session store (useful for testing).
func (h *ClientAdminHandler) SessionStore() *SessionStore {
	return h.sessionStore
}

// RateLimiter returns the active rate limiter (useful for testing).
func (h *ClientAdminHandler) RateLimiter() *RateLimiter {
	return h.rateLimiter
}

// isPasswordConfigured returns true if a master password has been configured.
func (h *ClientAdminHandler) isPasswordConfigured() bool {
	settings := data.GetSiteSettings(h.app)
	if settings.ClientAdminPasswordHash != "" {
		return true
	}
	if settings.ClientAdminPassword != "" && settings.ClientAdminPassword != "admin" {
		return true
	}
	return false
}

// isAuthorized checks if the client request contains a valid auth session cookie.
func (h *ClientAdminHandler) isAuthorized(e *core.RequestEvent) bool {
	cookie, err := e.Request.Cookie(ClientAuthCookieName)
	if err != nil || cookie == nil || cookie.Value == "" {
		return false
	}
	return h.sessionStore.IsValid(cookie.Value)
}

// HandleLoginPage renders the login page or setup wizard if password is not configured.
func (h *ClientAdminHandler) HandleLoginPage(e *core.RequestEvent) error {
	if h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client")
	}
	settings := data.GetSiteSettings(h.app)
	if !h.isPasswordConfigured() {
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return pages.ClientInitialSetup(settings, "").Render(e.Request.Context(), e.Response)
	}
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientLogin(settings, "").Render(e.Request.Context(), e.Response)
}

// HandleInitialSetup processes first-time master password configuration.
func (h *ClientAdminHandler) HandleInitialSetup(e *core.RequestEvent) error {
	if h.isPasswordConfigured() {
		return e.Redirect(http.StatusFound, "/client/login")
	}
	settings := data.GetSiteSettings(h.app)
	password := strings.TrimSpace(e.Request.FormValue("password"))
	passwordConfirm := strings.TrimSpace(e.Request.FormValue("password_confirm"))

	if utf8.RuneCountInString(password) < 8 {
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return pages.ClientInitialSetup(settings, "Пароль должен содержать не менее 8 символов").Render(e.Request.Context(), e.Response)
	}
	if password != passwordConfirm {
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return pages.ClientInitialSetup(settings, "Пароли не совпадают").Render(e.Request.Context(), e.Response)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return pages.ClientInitialSetup(settings, "Ошибка хэширования пароля").Render(e.Request.Context(), e.Response)
	}

	if setRec, err := h.app.FindFirstRecordByFilter("site_settings", ""); err == nil && setRec != nil {
		setRec.Set("client_admin_password_hash", string(hash))
		setRec.Set("client_admin_password", "")
		_ = h.app.Save(setRec)
	}
	cache.GetStore().InvalidateSettings()

	token, _ := h.sessionStore.CreateSession()
	http.SetCookie(e.Response, &http.Cookie{
		Name:     ClientAuthCookieName,
		Value:    token,
		Path:     "/client",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return e.Redirect(http.StatusFound, "/client")
}

// HandleLogin processes password submission with rate-limiting and bcrypt verification.
func (h *ClientAdminHandler) HandleLogin(e *core.RequestEvent) error {
	settings := data.GetSiteSettings(h.app)
	if !h.isPasswordConfigured() {
		return e.Redirect(http.StatusFound, "/client/login")
	}

	ip := e.RealIP()
	if blocked, waitDur := h.rateLimiter.IsBlocked(ip); blocked {
		minutes := int(waitDur.Minutes()) + 1
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		e.Response.WriteHeader(http.StatusTooManyRequests)
		return pages.ClientLogin(settings, fmt.Sprintf("Слишком много неудачных попыток входа. Попробуйте через %d мин.", minutes)).Render(e.Request.Context(), e.Response)
	}

	password := strings.TrimSpace(e.Request.FormValue("password"))
	hash := settings.ClientAdminPasswordHash
	var valid bool

	if hash != "" {
		valid = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
	} else if settings.ClientAdminPassword != "" && settings.ClientAdminPassword != "admin" {
		valid = password == settings.ClientAdminPassword
		if valid {
			// Автомиграция при входе
			if newHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost); err == nil {
				if setRec, err := h.app.FindFirstRecordByFilter("site_settings", ""); err == nil && setRec != nil {
					setRec.Set("client_admin_password_hash", string(newHash))
					setRec.Set("client_admin_password", "")
					_ = h.app.Save(setRec)
					cache.GetStore().InvalidateSettings()
				}
			}
		}
	}

	if !valid {
		isBlocked, _ := h.rateLimiter.RecordFailure(ip)
		errMsg := "Неверный пароль доступа"
		if isBlocked {
			errMsg = "Слишком много неудачных попыток входа. Доступ заблокирован на 15 минут."
			e.Response.WriteHeader(http.StatusTooManyRequests)
		}
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return pages.ClientLogin(settings, errMsg).Render(e.Request.Context(), e.Response)
	}

	// Сброс счетчика неудач при успехе
	h.rateLimiter.Reset(ip)

	token, _ := h.sessionStore.CreateSession()
	http.SetCookie(e.Response, &http.Cookie{
		Name:     ClientAuthCookieName,
		Value:    token,
		Path:     "/client",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return e.Redirect(http.StatusFound, "/client")
}

// HandleLogout clears auth session and cookie.
func (h *ClientAdminHandler) HandleLogout(e *core.RequestEvent) error {
	if cookie, err := e.Request.Cookie(ClientAuthCookieName); err == nil && cookie != nil {
		h.sessionStore.RevokeSession(cookie.Value)
	}
	http.SetCookie(e.Response, &http.Cookie{
		Name:     ClientAuthCookieName,
		Value:    "",
		Path:     "/client",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
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
	reviews := data.GetAllReviewsAdmin(h.app)

	tab := e.Request.URL.Query().Get("tab")
	if tab == "" {
		tab = "leads"
	}

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientDashboard(settings, content, leads, articles, trashArticles, reviews, tab).Render(e.Request.Context(), e.Response)
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

// HandleTabReviews renders reviews moderation tab fragment for HTMX.
func (h *ClientAdminHandler) HandleTabReviews(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}
	reviews := data.GetAllReviewsAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientReviewsTab(reviews).Render(e.Request.Context(), e.Response)
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

// HandleTabSecurity renders security tab fragment for HTMX.
func (h *ClientAdminHandler) HandleTabSecurity(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientSecurityTab("", "").Render(e.Request.Context(), e.Response)
}

// HandleChangePassword processes password change in client cabinet.
func (h *ClientAdminHandler) HandleChangePassword(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.String(http.StatusUnauthorized, "Неавторизован")
	}

	currentPassword := strings.TrimSpace(e.Request.FormValue("current_password"))
	newPassword := strings.TrimSpace(e.Request.FormValue("new_password"))
	confirmPassword := strings.TrimSpace(e.Request.FormValue("confirm_password"))

	settings := data.GetSiteSettings(h.app)

	var currentValid bool
	if settings.ClientAdminPasswordHash != "" {
		currentValid = bcrypt.CompareHashAndPassword([]byte(settings.ClientAdminPasswordHash), []byte(currentPassword)) == nil
	} else if settings.ClientAdminPassword != "" {
		currentValid = currentPassword == settings.ClientAdminPassword
	}

	if !currentValid {
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return pages.ClientSecurityTab("", "Неверный текущий пароль").Render(e.Request.Context(), e.Response)
	}

	if utf8.RuneCountInString(newPassword) < 8 {
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return pages.ClientSecurityTab("", "Новый пароль должен содержать не менее 8 символов").Render(e.Request.Context(), e.Response)
	}

	if newPassword != confirmPassword {
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return pages.ClientSecurityTab("", "Новый пароль и подтверждение не совпадают").Render(e.Request.Context(), e.Response)
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return pages.ClientSecurityTab("", "Ошибка шифрования нового пароля").Render(e.Request.Context(), e.Response)
	}

	if setRec, err := h.app.FindFirstRecordByFilter("site_settings", ""); err == nil && setRec != nil {
		setRec.Set("client_admin_password_hash", string(newHash))
		setRec.Set("client_admin_password", "")
		_ = h.app.Save(setRec)
	}
	cache.GetStore().InvalidateSettings()

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientSecurityTab("Пароль доступа успешно изменен!", "").Render(e.Request.Context(), e.Response)
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
			l.ID,
			l.Name,
			l.Phone,
			l.Email,
			l.BookingDate,
			l.BookingTime,
			l.Status,
			l.Source,
			utmSource,
			utmCampaign,
			quizInfo,
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
