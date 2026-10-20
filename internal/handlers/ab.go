package handlers

import (
	"context"
	"math/rand/v2"
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

const ABCookieName = "sitecraft_ab"

type abContextKey struct{}

// WithABVariant injects the A/B test variant ("A" or "B") into context.
func WithABVariant(ctx context.Context, variant string) context.Context {
	return context.WithValue(ctx, abContextKey{}, variant)
}

// GetABVariant returns current A/B test variant from context ("A" or "B", defaults to "A").
func GetABVariant(ctx context.Context) string {
	if val, ok := ctx.Value(abContextKey{}).(string); ok && (val == "A" || val == "B") {
		return val
	}
	return "A"
}

// ABTestingMiddleware handles zero-flicker server-side split testing.
// Assigns 50/50 A or B variant on first visit, persists in cookie for 30 days,
// and injects the assigned variant into request context.
func ABTestingMiddleware() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var variant string

		// Поддержка ручного переключения через query-параметр для тестирования (?ab=A или ?ab=B)
		if qParam := strings.ToUpper(strings.TrimSpace(e.Request.URL.Query().Get("ab"))); qParam == "A" || qParam == "B" {
			variant = qParam
			http.SetCookie(e.Response, &http.Cookie{
				Name:     ABCookieName,
				Value:    variant,
				Path:     "/",
				MaxAge:   30 * 24 * 3600, // 30 дней
				SameSite: http.SameSiteLaxMode,
				HttpOnly: false,
			})
		} else if cookie, err := e.Request.Cookie(ABCookieName); err == nil && cookie != nil && (cookie.Value == "A" || cookie.Value == "B") {
			variant = cookie.Value
		} else {
			// Случайное 50/50 распределение при первом визите
			if rand.IntN(2) == 0 {
				variant = "A"
			} else {
				variant = "B"
			}

			http.SetCookie(e.Response, &http.Cookie{
				Name:     ABCookieName,
				Value:    variant,
				Path:     "/",
				MaxAge:   30 * 24 * 3600, // 30 дней
				SameSite: http.SameSiteLaxMode,
				HttpOnly: false,
			})
		}

		ctx := WithABVariant(e.Request.Context(), variant)
		e.Request = e.Request.WithContext(ctx)

		return e.Next()
	}
}
