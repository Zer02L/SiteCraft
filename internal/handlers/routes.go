package handlers

import (
	"context"

	"sitecraft/internal/events"
	"sitecraft/internal/i18n"
	"sitecraft/internal/notifications"
	"sitecraft/ui"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
)

// RegisterRoutes registers all website and API routes onto the PocketBase router.
func RegisterRoutes(app *pocketbase.PocketBase, r *router.Router[*core.RequestEvent]) {
	// Middleware: структурированное логирование запросов (slog)
	r.BindFunc(RequestLoggerMiddleware())
	// Middleware: гранулярные HTTP заголовки Cache-Control (статика, BFCache, SSE, админка)
	r.BindFunc(CacheControlMiddleware())
	// Middleware: локализация (ru/en определение из cookie, query, Accept-Language)
	r.BindFunc(i18n.Middleware())
	// Middleware: сквозная UTM-атрибуция (захват utm_* и yclid в 30-дневную cookie)
	r.BindFunc(UTMMiddleware())
	// Middleware: серверное A/B сплит-тестирование (50/50 cookie sitecraft_ab)
	r.BindFunc(ABTestingMiddleware())

	bus := events.Default()
	notifications.RegisterSubscribers(bus)

	webHandler := NewWebHandler(app)
	leadHandler := NewLeadHandler(app, bus)
	paymentHandler := NewPaymentHandler(app, WithPaymentEventBus(bus))
	tmaHandler := NewTMAHandler(app)
	trackerHandler := NewTrackerHandler(app)
	cartHandler := NewCartHandler(app, bus)
	clientAdminHandler := NewClientAdminHandler(app)
	chatHandler := NewChatHandler(app)
	chatHandler.StartBackgroundPoller(context.Background())
	fleetHandler := NewFleetHandler(app, clientAdminHandler.SessionStore())
	bookingHandler := NewBookingHandler(app)
	reviewHandler := NewReviewHandler(app)

	// Статические файлы (встроенные ассеты / live-reload в dev-режиме)
	r.GET("/assets/{path...}", apis.Static(ui.AssetsFS(), false))

	// Клиентская панель управления (/client)
	r.GET("/client/login", clientAdminHandler.HandleLoginPage)
	r.POST("/client/login", clientAdminHandler.HandleLogin)
	r.POST("/client/setup", clientAdminHandler.HandleInitialSetup)
	r.GET("/client/logout", clientAdminHandler.HandleLogout)
	r.GET("/client", clientAdminHandler.HandleDashboard)
	r.GET("/client/tab/leads", clientAdminHandler.HandleTabLeads)
	r.GET("/client/tab/contacts", clientAdminHandler.HandleTabContacts)
	r.GET("/client/tab/articles", clientAdminHandler.HandleTabArticles)
	r.GET("/client/tab/security", clientAdminHandler.HandleTabSecurity)
	r.POST("/client/security/password", clientAdminHandler.HandleChangePassword)
	r.POST("/client/contacts", clientAdminHandler.HandleSaveContacts)
	r.POST("/client/leads/{id}/status", clientAdminHandler.HandleUpdateLeadStatus)
	r.GET("/client/export/leads.csv", clientAdminHandler.HandleExportLeadsCSV)
	r.GET("/client/articles/new", clientAdminHandler.HandleNewArticleForm)
	r.GET("/client/articles/edit/{id}", clientAdminHandler.HandleEditArticleForm)
	r.POST("/client/articles/save", clientAdminHandler.HandleSaveArticle)
	r.POST("/client/articles/toggle/{id}", clientAdminHandler.HandleToggleArticlePublish)
	r.POST("/client/articles/trash/{id}", clientAdminHandler.HandleTrashArticle)
	r.POST("/client/articles/restore/{id}", clientAdminHandler.HandleRestoreArticle)
	r.DELETE("/client/articles/permanent/{id}", clientAdminHandler.HandlePermanentDeleteArticle)
	r.POST("/client/articles/permanent/{id}", clientAdminHandler.HandlePermanentDeleteArticle)
	r.POST("/client/articles/trash/empty", clientAdminHandler.HandleEmptyTrashArticles)

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

	// Блок 4: Отраслевые модули, B2B, Вакансии и 152-ФЗ
	r.GET("/privacy", webHandler.HandlePrivacy)
	r.GET("/terms", webHandler.HandleTerms)
	r.GET("/careers", webHandler.HandleCareers)
	r.GET("/invoice", webHandler.HandleInvoice)
	r.GET("/invoice/{code}", webHandler.HandleInvoice)

	// TMA (Telegram Mini App): Ниша А (E-commerce / Shop) и Ниша В (Paywall / VIP Club)
	r.GET("/shop", tmaHandler.HandleShop)
	r.GET("/vip", tmaHandler.HandleVIPClub)
	r.POST("/api/tma/auth", tmaHandler.HandleTMAAuth)
	r.POST("/api/shop/checkout", tmaHandler.HandleShopCheckout)

	// API HTMX компоненты
	r.GET("/api/map-frame", webHandler.HandleMapFrame)
	r.GET("/api/slots", webHandler.HandleSlots)
	r.GET("/api/resources", webHandler.HandleResources)
	r.GET("/api/branches", webHandler.HandleBranches)
	r.GET("/api/booking/calendar.ics", bookingHandler.HandleDownloadICS)
	r.GET("/api/booking/action", bookingHandler.HandleBookingAction)
	r.GET("/api/payment/modal", webHandler.HandlePaymentModal)
	r.GET("/api/recent-activity", webHandler.HandleRecentActivity)

	// API Бизнес-действия
	r.POST("/api/leads", leadHandler.HandleCreateLead)
	r.POST("/api/careers/apply", leadHandler.HandleApplyVacancy)
	r.POST("/api/reviews", reviewHandler.HandleSubmitReview)
	r.POST("/api/cart/checkout", cartHandler.HandleCheckout)
	r.POST("/api/cart/validate-coupon", cartHandler.HandleValidateCoupon)
	r.POST("/api/payment/create", paymentHandler.HandleCreatePayment)
	r.GET("/api/lang/{code}", i18n.HandleSetLang)

	// Модерация отзывов в кабинете управления
	r.GET("/client/tab/reviews", clientAdminHandler.HandleTabReviews)
	r.POST("/client/reviews/{id}/toggle", reviewHandler.HandleTogglePublish)
	r.POST("/client/reviews/{id}/delete", reviewHandler.HandleDeleteReview)
	r.DELETE("/client/reviews/{id}", reviewHandler.HandleDeleteReview)

	// Модуль 6: Live Web-to-TG Chat
	r.GET("/api/chat/sse", chatHandler.HandleChatSSE)
	r.GET("/api/chat/history", chatHandler.HandleChatHistory)
	r.POST("/api/chat/send", chatHandler.HandleSendMessage)
	r.POST("/api/chat/telegram-webhook", chatHandler.HandleTelegramWebhook)

	// Модуль 20: Fleet Master Dashboard
	r.GET("/fleet", fleetHandler.HandleDashboard)
	r.GET("/client/tab/fleet", fleetHandler.HandleTabFleet)
	r.POST("/api/fleet/action", fleetHandler.HandleFleetAction)
	r.POST("/api/fleet/create", fleetHandler.HandleFleetCreate)
}
