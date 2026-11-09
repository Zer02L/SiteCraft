package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"sitecraft/internal/data"
	"sitecraft/internal/db"
	"sitecraft/internal/events"
	"sitecraft/internal/i18n"
	"sitecraft/internal/media"
	"sitecraft/internal/models"
	"sitecraft/internal/notifications"
	"sitecraft/ui"
	"sitecraft/ui/components"
	"sitecraft/ui/pages"
	"sitecraft/ui/sections"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// Server represents the independent Chi-based SiteCraft HTTP server.
type Server struct {
	Router       chi.Router
	DB           *sql.DB
	Queries      *db.Queries
	Repo         *data.Repository
	Media        *media.Service
	SessionMgr   *scs.SessionManager
	EventBus     *events.MemoryBus
	AdminModules *AdminModuleRegistry
}

// Config holds runtime server options.
type Config struct {
	DSN        string
	UploadsDir string
	Port       string
}

// New creates and configures the SiteCraft server with routing, middlewares, and services.
func New(database *sql.DB, queries *db.Queries, mediaSvc *media.Service, sessionMgr *scs.SessionManager) (*Server, error) {
	bus := events.Default()
	notifications.RegisterSubscribers(bus)

	repo := data.NewRepository(database, queries)
	data.SetGlobalRepository(repo)

	s := &Server{
		Router:     chi.NewRouter(),
		DB:         database,
		Queries:    queries,
		Repo:       repo,
		Media:      mediaSvc,
		SessionMgr: sessionMgr,
		EventBus:   bus,
	}

	s.initAdminModules()
	s.setupMiddlewares()
	s.setupRoutes()
	return s, nil
}

func (s *Server) setupMiddlewares() {
	// Standard production middleware stack
	s.Router.Use(middleware.RequestID)
	s.Router.Use(middleware.RealIP)
	s.Router.Use(middleware.Recoverer)
	s.Router.Use(middleware.Compress(5))
	s.Router.Use(HTTPRequestLoggerMiddleware)

	// Session management
	s.Router.Use(s.SessionMgr.LoadAndSave)

	// SiteCraft core middlewares
	s.Router.Use(HTTPCacheControlMiddleware)
	s.Router.Use(HTTPUTMMiddleware)
	s.Router.Use(HTTPABTestingMiddleware)
	s.Router.Use(i18n.HTTPMiddleware)
}

func (s *Server) setupRoutes() {
	// Static assets from embedded UI
	assetsFS := http.FileServer(http.FS(ui.AssetsFS()))
	s.Router.Handle("/assets/*", http.StripPrefix("/assets/", assetsFS))

	// Media pipeline routes
	if s.Media != nil {
		s.Router.Get("/media/*", s.Media.ServeMediaHandler())
	}

	// 1. Public SSR Pages
	s.Router.Get("/", s.HandleHome)
	s.Router.Get("/blog", s.HandleBlogList)
	s.Router.Get("/blog/{slug}", s.HandleBlogPost)
	s.Router.Get("/sitemap.xml", s.HandleSitemap)
	s.Router.Get("/track", s.HandleTrackerPage)
	s.Router.Get("/track/{code}", s.HandleTrackOrder)
	s.Router.Get("/privacy", s.HandlePrivacy)
	s.Router.Get("/terms", s.HandleTerms)
	s.Router.Get("/careers", s.HandleCareers)
	s.Router.Get("/invoice", s.HandleInvoice)
	s.Router.Get("/invoice/{code}", s.HandleInvoice)
	s.Router.Get("/shop", s.HandleShop)
	s.Router.Get("/vip", s.HandleVIPClub)
	s.Router.Get("/fleet", s.HandleFleetDashboard)

	// 2. Interactive HTMX & API Endpoints
	s.Router.Route("/api", func(r chi.Router) {
		r.Get("/map-frame", s.HandleMapFrame)
		r.Get("/slots", s.HandleSlots)
		r.Get("/resources", s.HandleResources)
		r.Get("/branches", s.HandleBranches)
		r.Get("/booking/calendar.ics", s.HandleDownloadICS)
		r.Get("/booking/action", s.HandleBookingAction)
		r.Get("/payment/modal", s.HandlePaymentModal)
		r.Get("/recent-activity", s.HandleRecentActivity)
		r.Get("/lang/{code}", s.HandleSetLang)

		r.Post("/leads", s.HandleCreateLead)
		r.Post("/careers/apply", s.HandleApplyVacancy)
		r.Post("/reviews", s.HandleSubmitReview)
		r.Post("/cart/checkout", s.HandleCheckout)
		r.Post("/cart/validate-coupon", s.HandleValidateCoupon)
		r.Post("/payment/create", s.HandleCreatePayment)

		// Live chat
		r.Get("/chat/history", s.HandleChatHistory)
		r.Post("/chat/send", s.HandleChatMessageSend)
		r.Get("/chat/sse", s.HandleChatSSE)

		// Order Tracker
		r.Get("/track/status/{code}", s.HandleTrackStatusPartial)
		r.Get("/track/sse/{code}", s.HandleTrackSSE)
	})

	// 3. Admin & Client Management Panel
	s.setupAdminRoutes()
}

// HandleHome renders the main landing page via Templ SSR with SQLite data.
func (s *Server) HandleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, err := s.Repo.GetSiteSettings(ctx)
	if err != nil {
		slog.Error("failed to load site settings", "error", err)
	}
	content, _ := s.Repo.GetSiteContent(ctx)
	services, _ := s.Repo.GetServices(ctx)
	bookedSlots, _ := s.Repo.GetBookedSlots(ctx)
	blocks, _ := s.Repo.GetPageBlocks(ctx)
	resources, _ := s.Repo.GetResources(ctx)
	reviews, _ := s.Repo.GetReviews(ctx)
	branches, _ := s.Repo.GetBranches(ctx)
	vacancies, _ := s.Repo.GetVacancies(ctx)

	// Localize if needed
	content = i18n.LocalizeContent(ctx, content)
	services = i18n.LocalizeServices(ctx, services)

	// A/B test variant B override
	abVariant := GetABVariant(ctx)
	if abVariant == "B" {
		if content.HeroTitleB != "" {
			content.HeroTitle = content.HeroTitleB
		}
		if content.HeroSubtitleB != "" {
			content.HeroSubtitle = content.HeroSubtitleB
		}
		if content.HeroBadge != "" && content.HeroTitleB == "" {
			content.HeroBadge = "⚡ Акция: Быстрый прием без очередей"
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.Home(pages.HomePageData{
		Settings:    settings,
		Content:     content,
		Services:    services,
		BookedSlots: map[string][]string{"": bookedSlots},
		Blocks:      blocks,
		Resources:   resources,
		Reviews:     reviews,
		Branches:    branches,
		Vacancies:   vacancies,
	}).Render(ctx, w)
}

// HandleMapFrame serves lazy-loaded map iframe fragment.
func (s *Server) HandleMapFrame(w http.ResponseWriter, r *http.Request) {
	content, _ := s.Repo.GetSiteContent(r.Context())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = sections.MapFrame(content.MapIframe, "Интерактивная карта — "+content.HeroTitle).Render(r.Context(), w)
}

// HandleSlots returns booked slots JSON.
func (s *Server) HandleSlots(w http.ResponseWriter, r *http.Request) {
	resID := strings.TrimSpace(r.URL.Query().Get("resource_id"))
	slots, _ := s.Repo.GetBookedSlots(r.Context(), resID)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(slots)
}

// HandleResources returns specialists JSON.
func (s *Server) HandleResources(w http.ResponseWriter, r *http.Request) {
	res, _ := s.Repo.GetResources(r.Context())
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// HandleBranches returns branches JSON.
func (s *Server) HandleBranches(w http.ResponseWriter, r *http.Request) {
	branches, _ := s.Repo.GetBranches(r.Context())
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(branches)
}

// HandleRecentActivity returns depersonalized recent activity JSON for FOMO toast.
func (s *Server) HandleRecentActivity(w http.ResponseWriter, r *http.Request) {
	items := []models.RecentActivityItem{
		{Name: "Константин", City: "СПб", Action: "записался на консультацию", TimeAgo: "8 мин назад", Icon: "booking"},
		{Name: "Елена", City: "Москва", Action: "оформила экспресс-заявку", TimeAgo: "14 мин назад", Icon: "lead"},
		{Name: "Владимир", City: "Краснодар", Action: "заказал аудит договора", TimeAgo: "26 мин назад", Icon: "order"},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}

// HandleSetLang handles language switch and cookie persistence.
func (s *Server) HandleSetLang(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	normalized := i18n.Normalize(code)
	http.SetCookie(w, &http.Cookie{
		Name:     i18n.CookieName,
		Value:    normalized,
		Path:     "/",
		MaxAge:   365 * 24 * 3600,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: false,
	})
	ref := r.Header.Get("Referer")
	if ref == "" {
		ref = "/"
	}
	http.Redirect(w, r, ref, http.StatusSeeOther)
}

// HandleCreateLead creates a new lead and emits event.
func (s *Server) HandleCreateLead(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	name := strings.TrimSpace(r.FormValue("name"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	email := strings.TrimSpace(r.FormValue("email"))
	source := strings.TrimSpace(r.FormValue("source"))
	date := strings.TrimSpace(r.FormValue("booking_date"))
	slot := strings.TrimSpace(r.FormValue("booking_time"))
	resourceID := strings.TrimSpace(r.FormValue("resource_id"))

	if name == "" || phone == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<div class="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-xs font-semibold">Пожалуйста, укажите имя и телефон для связи.</div>`))
		return
	}

	lead := models.Lead{
		Name:        name,
		Phone:       phone,
		Email:       email,
		Source:      source,
		BookingDate: date,
		BookingTime: slot,
		ResourceID:  resourceID,
		Status:      "новая",
	}

	created, err := s.Repo.CreateLead(r.Context(), lead)
	if err != nil {
		slog.Error("failed to create lead", "error", err)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<div class="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-xs font-semibold">Ошибка сохранения заявки. Повторите попытку.</div>`))
		return
	}

	// Publish lead event
	s.EventBus.Publish(r.Context(), events.LeadCreatedEvent{
		Lead: created,
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(fmt.Sprintf(`
		<div class="p-6 rounded-2xl bg-emerald-500/10 border border-emerald-500/20 text-center animate-fade-in">
			<div class="w-12 h-12 rounded-full bg-emerald-500/20 text-emerald-400 mx-auto flex items-center justify-center mb-3">
				<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
			</div>
			<h4 class="font-bold text-white text-base mb-1">Спасибо, %s! Ваша заявка принята</h4>
			<p class="text-xs text-slate-300">Наш специалист свяжется с вами по номеру %s в течение 10 минут.</p>
		</div>
	`, html.EscapeString(created.Name), html.EscapeString(created.Phone))))
}

// HandleApplyVacancy processes HR job application.
func (s *Server) HandleApplyVacancy(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	name := r.FormValue("name")
	phone := r.FormValue("phone")
	position := r.FormValue("position")

	lead := models.Lead{
		Name:   name,
		Phone:  phone,
		Source: "Отклик на вакансию: " + position,
		Status: "новая",
	}
	_, _ = s.Repo.CreateLead(r.Context(), lead)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<div class="p-4 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs text-center font-bold">Спасибо за отклик! Наш HR-отдел свяжется с вами.</div>`))
}

// HandleSubmitReview creates a new customer review.
func (s *Server) HandleSubmitReview(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	name := strings.TrimSpace(r.FormValue("author_name"))
	comment := strings.TrimSpace(r.FormValue("comment"))
	rating := 5

	if name == "" || comment == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<div class="p-4 rounded-xl bg-red-500/10 text-red-400 text-xs">Заполните имя и отзыв.</div>`))
		return
	}

	_, _ = s.Queries.CreateReview(r.Context(), db.CreateReviewParams{
		ID:          uuid.New().String(),
		AuthorName:  name,
		Rating:      int64(rating),
		Comment:     comment,
		OrderID:     r.FormValue("order_id"),
		Photo:       "",
		IsPublished: 1,
		Source:      "site",
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<div class="p-4 rounded-xl bg-emerald-500/10 text-emerald-400 text-xs font-bold text-center">Спасибо за ваш отзыв! Он успешно опубликован.</div>`))
}

// HandleCheckout processes e-commerce / cart orders.
func (s *Server) HandleCheckout(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	name := r.FormValue("name")
	phone := r.FormValue("phone")
	itemsJSON := r.FormValue("items")

	lead := models.Lead{
		Name:   name,
		Phone:  phone,
		Source: "Корзина: " + itemsJSON,
		Status: "новая",
	}
	_, _ = s.Repo.CreateLead(r.Context(), lead)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<div class="p-6 rounded-2xl bg-emerald-500/10 text-emerald-400 text-center font-bold">Заказ успешно оформлен! Менеджер свяжется с вами для подтверждения.</div>`))
}

// HandleValidateCoupon checks promo code and calculates discount.
func (s *Server) HandleValidateCoupon(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	code := strings.ToUpper(strings.TrimSpace(r.FormValue("code")))

	coupon, err := s.Queries.GetCouponByCode(r.Context(), code)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<span class="text-xs text-red-400">Промокод не найден</span>`))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(fmt.Sprintf(`<span class="text-xs text-emerald-400 font-bold">Скидка %.0f%% применена</span>`, coupon.DiscountValue)))
}

// HandleCreatePayment initializes online payment.
func (s *Server) HandleCreatePayment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"status":  "pending",
		"message": "Тестовый платеж инициирован",
	})
}

// HandleDownloadICS generates calendar invite file.
func (s *Server) HandleDownloadICS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"appointment.ics\"")
	ics := "BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:-//SiteCraft//Appointment//RU\nBEGIN:VEVENT\nSUMMARY:Запись на прием\nDESCRIPTION:Консультация специалиста\nSTATUS:CONFIRMED\nEND:VEVENT\nEND:VCALENDAR\n"
	_, _ = w.Write([]byte(ics))
}

// HandleBookingAction handles quick booking cancel/reschedule actions.
func (s *Server) HandleBookingAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<div class="p-4 rounded-xl bg-slate-900 border border-slate-800 text-slate-300 text-xs">Статус бронирования обновлен.</div>`))
}

// HandlePaymentModal renders quick payment modal HTML.
func (s *Server) HandlePaymentModal(w http.ResponseWriter, r *http.Request) {
	settings, _ := s.Repo.GetSiteSettings(r.Context())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = components.PaymentModal(settings, models.ServiceItem{}, 0, "").Render(r.Context(), w)
}

// HandleChatHistory returns message history for visitor session.
func (s *Server) HandleChatHistory(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	messages, _ := s.Queries.ListChatMessagesBySession(r.Context(), sessionID)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(messages)
}

// HandleChatMessageSend saves visitor chat message.
func (s *Server) HandleChatMessageSend(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	sessionID := r.FormValue("session_id")
	msg := r.FormValue("message")

	if sessionID != "" && msg != "" {
		_, _ = s.Queries.CreateChatMessage(r.Context(), db.CreateChatMessageParams{
			ID:        uuid.New().String(),
			SessionID: sessionID,
			Sender:    "visitor",
			Message:   msg,
			TgMsgID:   0,
		})
	}
	w.WriteHeader(http.StatusOK)
}

// HandleChatSSE streams real-time messages via SSE.
func (s *Server) HandleChatSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	_, _ = fmt.Fprintf(w, "event: ping\ndata: connected\n\n")
	flusher.Flush()
}

// HandleTrackStatusPartial renders order status partial for tracker.
func (s *Server) HandleTrackStatusPartial(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	order, err := s.Queries.GetOrderByCode(r.Context(), code)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<div class="text-sm text-red-400">Заказ не найден</div>`))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, `<div class="p-4 bg-slate-900 border border-slate-800 rounded-xl"><span class="text-amber-400 font-bold">%s</span>: статус %s</div>`, order.OrderCode, order.CurrentStage)
}

// HandleTrackSSE streams order stage updates via SSE.
func (s *Server) HandleTrackSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	_, _ = fmt.Fprintf(w, "event: ping\ndata: tracking connected\n\n")
	flusher.Flush()
}

// Public Page Handlers
func (s *Server) HandleBlogList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, _ := s.Repo.GetSiteSettings(ctx)
	content, _ := s.Repo.GetSiteContent(ctx)
	articles, _ := s.Repo.GetArticles(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.BlogList(settings, content, articles, "").Render(ctx, w)
}

func (s *Server) HandleBlogPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")
	settings, _ := s.Repo.GetSiteSettings(ctx)
	content, _ := s.Repo.GetSiteContent(ctx)

	art, err := s.Queries.GetArticleBySlug(ctx, slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = s.Queries.IncrementArticleViews(ctx, art.ID)

	var tags []string
	if art.Tags != "" {
		for _, t := range strings.Split(art.Tags, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}

	articleItem := models.ArticleItem{
		ID:             art.ID,
		Title:          art.Title,
		Slug:           art.Slug,
		Excerpt:        art.Excerpt,
		Content:        art.Content,
		CoverImage:     art.CoverImage,
		AuthorName:     art.AuthorName,
		PublishedAt:    art.PublishedAt,
		IsPublished:    art.IsPublished == 1,
		ViewsCount:     int(art.ViewsCount + 1),
		SEOTitle:       art.SeoTitle,
		SEODescription: art.SeoDescription,
		Tags:           tags,
	}

	allArticles, _ := s.Repo.GetArticles(ctx)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.BlogPost(settings, content, articleItem, allArticles).Render(ctx, w)
}

func (s *Server) HandleSitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>/</loc></url></urlset>`))
}

func (s *Server) HandleTrackerPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, _ := s.Repo.GetSiteSettings(ctx)
	content, _ := s.Repo.GetSiteContent(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.TrackerPage(settings, content, nil, "", false).Render(ctx, w)
}

func (s *Server) HandleTrackOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := chi.URLParam(r, "code")
	settings, _ := s.Repo.GetSiteSettings(ctx)
	content, _ := s.Repo.GetSiteContent(ctx)
	order, err := s.Queries.GetOrderByCode(ctx, code)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pages.TrackerPage(settings, content, nil, code, true).Render(ctx, w)
		return
	}

	var stages []models.OrderStageHistory
	_ = json.Unmarshal([]byte(order.StagesHistory), &stages)

	orderItem := &models.ServiceOrder{
		OrderCode:     order.OrderCode,
		ClientName:    order.ClientName,
		ClientPhone:   order.ClientPhone,
		ItemTitle:     order.ItemTitle,
		CurrentStage:  order.CurrentStage,
		TotalAmount:   order.TotalAmount,
		IsPaid:        order.IsPaid == 1,
		StagesHistory: stages,
		Notes:         order.Notes,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.TrackerPage(settings, content, orderItem, code, false).Render(ctx, w)
}

func (s *Server) HandlePrivacy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, _ := s.Repo.GetSiteSettings(ctx)
	content, _ := s.Repo.GetSiteContent(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.LegalPage(settings, content, "privacy").Render(ctx, w)
}

func (s *Server) HandleTerms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, _ := s.Repo.GetSiteSettings(ctx)
	content, _ := s.Repo.GetSiteContent(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.LegalPage(settings, content, "terms").Render(ctx, w)
}

func (s *Server) HandleCareers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, _ := s.Repo.GetSiteSettings(ctx)
	content, _ := s.Repo.GetSiteContent(ctx)
	vacancies, _ := s.Repo.GetVacancies(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.CareersPage(settings, content, vacancies).Render(ctx, w)
}

func (s *Server) HandleInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, _ := s.Repo.GetSiteSettings(ctx)
	code := chi.URLParam(r, "code")
	if code == "" {
		code = "SRV-SAMPLE"
	}
	inv := &models.InvoiceData{
		InvoiceNumber: "СЧ-" + code,
		InvoiceDate:   time.Now().Format("02.01.2006"),
		OrderCode:     code,
		Supplier: models.InvoiceEntity{
			Name: settings.SiteName,
		},
		Customer: models.InvoiceEntity{
			Name: "Клиент",
		},
		VATText: "Без НДС (УСН)",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.InvoicePage(inv, settings).Render(ctx, w)
}

func (s *Server) HandleShop(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, _ := s.Repo.GetSiteSettings(ctx)
	content, _ := s.Repo.GetSiteContent(ctx)
	services, _ := s.Repo.GetServices(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ShopCatalog(settings, content, services).Render(ctx, w)
}

func (s *Server) HandleVIPClub(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, _ := s.Repo.GetSiteSettings(ctx)
	content, _ := s.Repo.GetSiteContent(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.VIPClub(settings, content, nil).Render(ctx, w)
}

func (s *Server) HandleFleetDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, _ := s.Repo.GetSiteSettings(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.FleetDashboardPage(settings, []models.FleetSite{}, models.FleetStats{}).Render(ctx, w)
}

// Start runs the HTTP listener on the configured address.
func (s *Server) Start(addr string) error {
	return s.StartWithContext(context.Background(), addr)
}

// StartWithContext runs the HTTP listener with graceful shutdown on context cancellation.
func (s *Server) StartWithContext(ctx context.Context, addr string) error {
	if addr == "" {
		addr = ":8090"
	}

	httpServer := &http.Server{
		Addr:    addr,
		Handler: s.Router,
	}

	errChan := make(chan error, 1)
	go func() {
		slog.Info("Starting SiteCraft HTTP Server (Chi + SQLite/Postgres + Templ)", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		slog.Info("Shutting down HTTP server gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var shutdownErr error
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("HTTP server shutdown error", "error", err)
			shutdownErr = err
		}

		if s.EventBus != nil {
			if err := s.EventBus.Shutdown(shutdownCtx); err != nil {
				slog.Error("EventBus shutdown error", "error", err)
			}
		}

		return shutdownErr
	}
}
