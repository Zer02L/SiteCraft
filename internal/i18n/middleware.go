package i18n

import (
	"net/http"
	"strings"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

// Middleware returns a PocketBase router middleware function that detects
// language preference, sets persistent cookie on ?lang=... parameter,
// and attaches the locale into the request context.
func Middleware() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Detect language from request (query, cookie, header)
		lang := DetectLang(e.Request)

		// If query param is explicitly passed and valid, update cookie
		qLang := strings.TrimSpace(e.Request.URL.Query().Get("lang"))
		if qLang != "" {
			normalized := Normalize(qLang)
			http.SetCookie(e.Response, &http.Cookie{
				Name:     CookieName,
				Value:    normalized,
				Path:     "/",
				MaxAge:   365 * 24 * 3600, // 1 year
				SameSite: http.SameSiteLaxMode,
				HttpOnly: false, // accessible to client JS if needed
			})
			lang = normalized
		}

		// Inject into context
		ctx := WithLocale(e.Request.Context(), lang)
		e.Request = e.Request.WithContext(ctx)

		return e.Next()
	}
}
