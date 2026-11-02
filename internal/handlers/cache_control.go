package handlers

import (
	"net/http"
	"strings"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

// CacheControlMiddleware applies granular Cache-Control HTTP headers
// based on request paths, methods and content volatility.
func CacheControlMiddleware() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		path := e.Request.URL.Path
		method := e.Request.Method

		switch {
		// 1. Статические ассеты (CSS, JS, шрифты, иконки) — максимальное иммутабельное кэширование браузером
		case strings.HasPrefix(path, "/assets/"):
			e.Response.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

		// 2. Панель управления клиентом и Fleet — категорический запрет кэширования для безопасности
		case strings.HasPrefix(path, "/client"), strings.HasPrefix(path, "/fleet"):
			e.Response.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate")
			e.Response.Header().Set("Pragma", "no-cache")
			e.Response.Header().Set("Expires", "0")

		// 3. SSE-стриминг (живой чат /api/chat/sse и трекер заказов /api/track/sse)
		case strings.HasSuffix(path, "/sse") || strings.Contains(path, "/sse/"):
			e.Response.Header().Set("Cache-Control", "no-cache, no-transform")
			e.Response.Header().Set("Connection", "keep-alive")

		// 4. Публичные статьи блога — редкие обновления, допустим stale-while-revalidate
		case strings.HasPrefix(path, "/blog/") && method == http.MethodGet:
			if e.Response.Header().Get("Cache-Control") == "" {
				e.Response.Header().Set("Cache-Control", "public, max-age=300, stale-while-revalidate=3600")
			}

		// 5. Динамические публичные страницы SSR (Главная, каталог, трекер, TMA-витрина)
		case method == http.MethodGet:
			// private, no-cache защищает от рассинхрона A/B тестов, i18n и броней, сохраняя BFCache
			if e.Response.Header().Get("Cache-Control") == "" {
				e.Response.Header().Set("Cache-Control", "private, no-cache")
			}

		// 6. Мутации и отправка форм (POST/DELETE)
		default:
			if e.Response.Header().Get("Cache-Control") == "" {
				e.Response.Header().Set("Cache-Control", "no-store")
			}
		}

		return e.Next()
	}
}
