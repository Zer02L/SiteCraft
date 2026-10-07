package handlers

import (
	"os"

	"sitecraft/internal/events"
	"sitecraft/internal/i18n"
	"sitecraft/internal/notifications"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
)

// RegisterRoutes registers all website and API routes onto the PocketBase router.
func RegisterRoutes(app *pocketbase.PocketBase, r *router.Router[*core.RequestEvent]) {
	// Middleware: локализация (ru/en определение из cookie, query, Accept-Language)
	r.BindFunc(i18n.Middleware())

	bus := events.Default()
	notifications.RegisterSubscribers(bus)

	webHandler := NewWebHandler(app)
	leadHandler := NewLeadHandler(app, bus)
	paymentHandler := NewPaymentHandler(app, WithPaymentEventBus(bus))
	tmaHandler := NewTMAHandler(app)
	trackerHandler := NewTrackerHandler(app)

	// Статические файлы (скомпилированные CSS, ассеты)
	r.GET("/assets/{path...}", apis.Static(os.DirFS("./ui/assets"), false))
	r.GET("/public/{path...}", apis.Static(os.DirFS("./public"), false))

	// Главная страница (SSR с помощью Templ + In-Memory Cache)
	r.GET("/", webHandler.HandleHome)

	// Этап 1: SEO-Блог и База знаний
	r.GET("/blog", webHandler.HandleBlogList)
	r.GET("/blog/{slug}", webHandler.HandleBlogPost)
	r.GET("/sitemap.xml", webHandler.HandleSitemap)

	// Этап 2: Ниша Г - B2B-трекер заказов в реальном времени
	r.GET("/track", trackerHandler.HandleTrackerPage)
	r.GET("/track/{code}", trackerHandler.HandleTrackOrder)
	r.GET("/api/track/status/{code}", trackerHandler.HandleTrackStatusPartial)
	r.GET("/api/track/sse/{code}", trackerHandler.HandleTrackSSE)

	// TMA (Telegram Mini App): Ниша А (E-commerce / Shop) и Ниша В (Paywall / VIP Club)
	r.GET("/shop", tmaHandler.HandleShop)
	r.GET("/vip", tmaHandler.HandleVIPClub)
	r.POST("/api/tma/auth", tmaHandler.HandleTMAAuth)
	r.POST("/api/shop/checkout", tmaHandler.HandleShopCheckout)

	// API HTMX компоненты
	r.GET("/api/map-frame", webHandler.HandleMapFrame)
	r.GET("/api/slots", webHandler.HandleSlots)
	r.GET("/api/payment/modal", webHandler.HandlePaymentModal)

	// API Бизнес-действия
	r.POST("/api/leads", leadHandler.HandleCreateLead)
	r.POST("/api/payment/create", paymentHandler.HandleCreatePayment)
	r.GET("/api/lang/{code}", i18n.HandleSetLang)
}
