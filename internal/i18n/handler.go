package i18n

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// HandleSetLang handles language switching via /api/lang/{code}.
// It sets the cookie and redirects the visitor back to the Referer or root.
func HandleSetLang(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		code = r.PathValue("code")
	}
	normalized := Normalize(code)

	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    normalized,
		Path:     "/",
		MaxAge:   365 * 24 * 3600,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: false,
	})

	targetURL := r.Header.Get("Referer")
	if targetURL == "" {
		targetURL = "/"
	} else {
		if strings.Contains(targetURL, "lang=") {
			targetURL = strings.ReplaceAll(targetURL, "?lang=ru", "")
			targetURL = strings.ReplaceAll(targetURL, "&lang=ru", "")
			targetURL = strings.ReplaceAll(targetURL, "?lang=en", "")
			targetURL = strings.ReplaceAll(targetURL, "&lang=en", "")
		}
	}

	http.Redirect(w, r, targetURL, http.StatusSeeOther)
}
