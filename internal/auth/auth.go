package auth

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
	"golang.org/x/crypto/bcrypt"
)

const (
	SessionKeyAdminAuth  = "admin_authenticated"
	SessionKeyAdminEmail = "admin_email"
	SessionKeyFlashError = "flash_error"
	SessionKeyFlashInfo  = "flash_info"
)

// NewSessionManager creates a configured SCS session manager.
func NewSessionManager(db *sql.DB, isPostgres bool) *scs.SessionManager {
	sm := scs.New()
	sm.Store = NewSQLStore(db, isPostgres)
	sm.Lifetime = 24 * time.Hour
	sm.IdleTimeout = 4 * time.Hour
	sm.Cookie.Name = "sitecraft_session"
	sm.Cookie.HttpOnly = true
	sm.Cookie.Path = "/"
	sm.Cookie.SameSite = http.SameSiteLaxMode

	// Secure cookie only in production (HTTPS) or when explicitly configured
	if os.Getenv("APP_ENV") == "production" || os.Getenv("COOKIE_SECURE") == "true" {
		sm.Cookie.Secure = true
	} else {
		sm.Cookie.Secure = false
	}

	return sm
}

// HashPassword hashes plain text password with bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword verifies plain password against bcrypt hash.
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// RequireAdminAuth returns a middleware protecting routes from unauthenticated users.
func RequireAdminAuth(sm *scs.SessionManager, redirectURL string) func(http.Handler) http.Handler {
	if redirectURL == "" {
		redirectURL = "/admin/login"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isAuth := sm.GetBool(r.Context(), SessionKeyAdminAuth)
			if !isAuth {
				// If HTMX request, ask client to redirect via HX-Redirect header
				if r.Header.Get("HX-Request") == "true" {
					w.Header().Set("HX-Redirect", redirectURL)
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				http.Redirect(w, r, redirectURL, http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CSRFProtection returns a nosurf CSRF middleware with exemptions for webhook and API paths.
func CSRFProtection(exemptPrefixes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		csrfHandler := nosurf.New(next)
		csrfHandler.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "400 Bad Request: Invalid or missing CSRF token", http.StatusBadRequest)
		}))
		csrfHandler.ExemptFunc(func(r *http.Request) bool {
			// Exempt specific API and webhook endpoints
			path := r.URL.Path
			for _, prefix := range exemptPrefixes {
				if strings.HasPrefix(path, prefix) {
					return true
				}
			}
			// Default exempt paths for external integrations / webhooks
			if strings.HasPrefix(path, "/api/webhook") || strings.HasPrefix(path, "/api/telegram") {
				return true
			}
			return false
		})
		return csrfHandler
	}
}

// CSRFToken returns the current request CSRF token string.
func CSRFToken(r *http.Request) string {
	return nosurf.Token(r)
}

// SetAdminSession marks the current session as authenticated for the given email.
func SetAdminSession(ctx context.Context, sm *scs.SessionManager, email string) error {
	if err := sm.RenewToken(ctx); err != nil {
		return err
	}
	sm.Put(ctx, SessionKeyAdminAuth, true)
	sm.Put(ctx, SessionKeyAdminEmail, email)
	return nil
}

// ClearAdminSession logs out administrator by clearing session values and renewing token.
func ClearAdminSession(ctx context.Context, sm *scs.SessionManager) error {
	sm.Remove(ctx, SessionKeyAdminAuth)
	sm.Remove(ctx, SessionKeyAdminEmail)
	return sm.RenewToken(ctx)
}
