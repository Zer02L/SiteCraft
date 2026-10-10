package handlers

import (
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
	// Middleware: локализация (ru/en определение из cookie, query, Accept-Language)
	r.BindFunc(i18n.Middleware())

	bus := events.Default()
	notifications.RegisterSubscribers(bus)

	webHandler := NewWebHandler(app)
	leadHandler := NewLeadHandler(app, bus)
	paymentHandler := NewPaymentHandler(app, WithPaymentEventBus(bus))
	tmaHandler := NewTMAHandler(app)
	trackerHandler := NewTrackerHandler(app)
	cartHandler := NewCartHandler(app, bus)
	clientAdminHandler := NewClientAdminHandler(app)

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
	r.POST("/api/cart/checkout", cartHandler.HandleCheckout)
	r.POST("/api/payment/create", paymentHandler.HandleCreatePayment)
	r.GET("/api/lang/{code}", i18n.HandleSetLang)
}
