package handlers

import (
	"context"

	"sitecraft/internal/events"
	"sitecraft/internal/i18n"
	"sitecraft/internal/notifications"
	"sitecraft/ui"

	"ganigeorgiev.com/pocketbase/pocketbase"
	"ganigeorgiev.com/pocketbase/pocketbase/apis"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/router"
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
	fleetHandler := NewFleetHandler(app, clientAdminHandler.SessionStore())
	bookingHandler := NewBookingHandler(app)
	reviewHandler := NewReviewHandler(app)

	// Graceful lifecycle: запуск фонового поллера чата с корректной остановкой при завершении приложения
	chatCtx, cancelChat := context.WithCancel(context.Background())
	app.OnTerminate().BindFunc(func(te *core.TerminateEvent) error {
		cancelChat()
		return te.Next()
	})
	chatHandler.StartBackgroundPoller(chatCtx)

	// ==========================================
	// 1. Статические файлы и публичные страницы (SSR)
	// ==========================================
	r.GET("/assets/{path...}", apis.Static(ui.AssetsFS(), false))
	r.GET("/", webHandler.HandleHome)
	r.GET("/blog", webHandler.HandleBlogList)
	r.GET("/blog/{slug}", webHandler.HandleBlogPost)
	r.GET("/sitemap.xml", webHandler.HandleSitemap)
	r.GET("/track", trackerHandler.HandleTrackerPage)
	r.GET("/track/{code}", trackerHandler.HandleTrackOrder)
	r.GET("/privacy", webHandler.HandlePrivacy)
	r.GET("/terms", webHandler.HandleTerms)
	r.GET("/careers", webHandler.HandleCareers)
	r.GET("/invoice", webHandler.HandleInvoice)
	r.GET("/invoice/{code}", webHandler.HandleInvoice)
	r.GET("/shop", tmaHandler.HandleShop)
	r.GET("/vip", tmaHandler.HandleVIPClub)
	r.GET("/fleet", fleetHandler.HandleDashboard)
	r.GET("/demo/jhill", webHandler.HandleJHillDemo)
	r.GET("/demo/flat-to-3d", webHandler.HandleFlatTo3DDemo)

	// ==========================================
	// 2. Группа API маршрутов (/api/*)
	// ==========================================
	api := r.Group("/api")

	// Динамические HTMX-фрагменты
	api.GET("/map-frame", webHandler.HandleMapFrame)
	api.GET("/slots", webHandler.HandleSlots)
	api.GET("/resources", webHandler.HandleResources)
	api.GET("/branches", webHandler.HandleBranches)
	api.GET("/booking/calendar.ics", bookingHandler.HandleDownloadICS)
	api.GET("/booking/action", bookingHandler.HandleBookingAction)
	api.GET("/payment/modal", webHandler.HandlePaymentModal)
	api.GET("/recent-activity", webHandler.HandleRecentActivity)

	// Интерактивные действия и бизнес-операции
	api.POST("/leads", leadHandler.HandleCreateLead)
	api.POST("/careers/apply", leadHandler.HandleApplyVacancy)
	api.POST("/reviews", reviewHandler.HandleSubmitReview)
	api.POST("/cart/checkout", cartHandler.HandleCheckout)
	api.POST("/cart/validate-coupon", cartHandler.HandleValidateCoupon)
	api.POST("/payment/create", paymentHandler.HandleCreatePayment)
	api.GET("/lang/{code}", i18n.HandleSetLang)

	// TMA (Telegram Mini App)
	api.POST("/tma/auth", tmaHandler.HandleTMAAuth)
	api.POST("/shop/checkout", tmaHandler.HandleShopCheckout)

	// B2B Трекер заказов
	api.GET("/track/status/{code}", trackerHandler.HandleTrackStatusPartial)
	api.GET("/track/sse/{code}", trackerHandler.HandleTrackSSE)

	// Live Web-to-TG Чат
	api.GET("/chat/sse", chatHandler.HandleChatSSE)
	api.GET("/chat/history", chatHandler.HandleChatHistory)
	api.POST("/chat/send", chatHandler.HandleSendMessage)
	api.POST("/chat/telegram-webhook", chatHandler.HandleTelegramWebhook)

	// Fleet Управление
	api.POST("/fleet/action", fleetHandler.HandleFleetAction)
	api.POST("/fleet/create", fleetHandler.HandleFleetCreate)

	// ==========================================
	// 3. Группа клиентской панели управления (/client/*)
	// ==========================================
	client := r.Group("/client")

	// 3.1. Публичные маршруты (аутентификация и первичная настройка)
	client.GET("/login", clientAdminHandler.HandleLoginPage)
	client.POST("/login", clientAdminHandler.HandleLogin)
	client.POST("/setup", clientAdminHandler.HandleInitialSetup)
	client.GET("/logout", clientAdminHandler.HandleLogout)

	// 3.2. Защищенная подгруппа с Middleware авторизации и HTMX full-window редиректом
	admin := client.Group("")
	admin.BindFunc(clientAdminHandler.RequireAuth())

	// Дашборд и вкладки
	admin.GET("", clientAdminHandler.HandleDashboard)
	admin.GET("/tab/leads", clientAdminHandler.HandleTabLeads)
	admin.GET("/tab/contacts", clientAdminHandler.HandleTabContacts)
	admin.GET("/tab/articles", clientAdminHandler.HandleTabArticles)
	admin.GET("/tab/security", clientAdminHandler.HandleTabSecurity)
	admin.GET("/tab/reviews", clientAdminHandler.HandleTabReviews)
	admin.GET("/tab/services", clientAdminHandler.HandleTabServices)
	admin.GET("/tab/fleet", fleetHandler.HandleTabFleet)

	// Управление лидами и контактами
	admin.POST("/contacts", clientAdminHandler.HandleSaveContacts)
	admin.POST("/security/password", clientAdminHandler.HandleChangePassword)
	admin.POST("/leads/{id}/status", clientAdminHandler.HandleUpdateLeadStatus)
	admin.GET("/export/leads.csv", clientAdminHandler.HandleExportLeadsCSV)

	// Управление каталогом услуг и корзина
	admin.GET("/services/new", clientAdminHandler.HandleNewServiceForm)
	admin.GET("/services/edit/{id}", clientAdminHandler.HandleEditServiceForm)
	admin.POST("/services/save", clientAdminHandler.HandleSaveService)
	admin.POST("/services/{id}/toggle", clientAdminHandler.HandleToggleServiceActive)
	admin.POST("/services/{id}/trash", clientAdminHandler.HandleTrashService)
	admin.POST("/services/{id}/restore", clientAdminHandler.HandleRestoreService)
	admin.DELETE("/services/{id}/permanent", clientAdminHandler.HandlePermanentDeleteService)
	admin.POST("/services/{id}/permanent", clientAdminHandler.HandlePermanentDeleteService)
	admin.POST("/services/trash/empty", clientAdminHandler.HandleEmptyTrashServices)
	admin.POST("/services/{id}/reorder", clientAdminHandler.HandleReorderService)

	// Модерация отзывов
	admin.POST("/reviews/{id}/toggle", reviewHandler.HandleTogglePublish)
	admin.POST("/reviews/{id}/delete", reviewHandler.HandleDeleteReview)
	admin.DELETE("/reviews/{id}", reviewHandler.HandleDeleteReview)

	// CRUD статей и корзина
	admin.GET("/articles/new", clientAdminHandler.HandleNewArticleForm)
	admin.GET("/articles/edit/{id}", clientAdminHandler.HandleEditArticleForm)
	admin.POST("/articles/save", clientAdminHandler.HandleSaveArticle)
	admin.POST("/articles/toggle/{id}", clientAdminHandler.HandleToggleArticlePublish)
	admin.POST("/articles/trash/{id}", clientAdminHandler.HandleTrashArticle)
	admin.POST("/articles/restore/{id}", clientAdminHandler.HandleRestoreArticle)
	admin.DELETE("/articles/permanent/{id}", clientAdminHandler.HandlePermanentDeleteArticle)
	admin.POST("/articles/permanent/{id}", clientAdminHandler.HandlePermanentDeleteArticle)
	admin.POST("/articles/trash/empty", clientAdminHandler.HandleEmptyTrashArticles)
}
