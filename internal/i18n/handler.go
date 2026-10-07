package i18n

import (
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// HandleSetLang handles language switching via /api/lang/{code}.
// It sets the cookie and redirects the visitor back to the Referer or root.
func HandleSetLang(e *core.RequestEvent) error {
	code := e.Request.PathValue("code")
	normalized := Normalize(code)

	http.SetCookie(e.Response, &http.Cookie{
		Name:     CookieName,
		Value:    normalized,
		Path:     "/",
		MaxAge:   365 * 24 * 3600,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: false,
	})

	targetURL := e.Request.Header.Get("Referer")
	if targetURL == "" {
		targetURL = "/"
	} else {
		// Clean up any ?lang=... in Referer to avoid overriding the newly set cookie
		if strings.Contains(targetURL, "lang=") {
			// If referer has query param lang=, replace or strip it
			targetURL = strings.ReplaceAll(targetURL, "?lang=ru", "")
			targetURL = strings.ReplaceAll(targetURL, "&lang=ru", "")
			targetURL = strings.ReplaceAll(targetURL, "?lang=en", "")
			targetURL = strings.ReplaceAll(targetURL, "&lang=en", "")
		}
	}

	e.Response.Header().Set("Location", targetURL)
	e.Response.WriteHeader(http.StatusSeeOther)
	return nil
}
