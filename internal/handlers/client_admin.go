package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"sitecraft/internal/cache"
	"sitecraft/internal/data"
	"sitecraft/ui/pages"

	"ganigeorgiev.com/pocketbase/pocketbase"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
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

// RequireAuth returns a route middleware that protects client admin endpoints.
// If the session is invalid or expired, it redirects to /client/login.
// For HTMX requests, it sets the HX-Redirect header to trigger a full window navigation.
func (h *ClientAdminHandler) RequireAuth() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if !h.isAuthorized(e) {
			if e.Request.Header.Get("HX-Request") == "true" {
				e.Response.Header().Set("HX-Redirect", "/client/login")
				return e.String(http.StatusOK, "")
			}
			return e.Redirect(http.StatusFound, "/client/login")
		}
		return e.Next()
	}
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
	services := data.GetAllServicesAdmin(h.app)
	trashServices := data.GetTrashServicesAdmin(h.app)

	tab := e.Request.URL.Query().Get("tab")
	if tab == "" {
		tab = "leads"
	}

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientDashboard(settings, content, leads, articles, trashArticles, reviews, services, trashServices, tab).Render(e.Request.Context(), e.Response)
}
