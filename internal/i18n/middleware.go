package i18n

import (
	"net/http"
	"strings"
)

// HTTPMiddleware returns a standard net/http middleware compatible with Chi.
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := DetectLang(r)

		qLang := strings.TrimSpace(r.URL.Query().Get("lang"))
		if qLang != "" {
			normalized := Normalize(qLang)
			http.SetCookie(w, &http.Cookie{
				Name:     CookieName,
				Value:    normalized,
				Path:     "/",
				MaxAge:   365 * 24 * 3600,
				SameSite: http.SameSiteLaxMode,
				HttpOnly: false,
			})
			lang = normalized
		}

		ctx := WithLocale(r.Context(), lang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Middleware is an alias for HTTPMiddleware.
func Middleware(next http.Handler) http.Handler {
	return HTTPMiddleware(next)
}
