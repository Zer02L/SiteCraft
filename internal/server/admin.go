package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"sitecraft/internal/auth"
	"sitecraft/internal/db"
	"sitecraft/internal/media"
	"sitecraft/internal/models"
	"sitecraft/ui/pages"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (s *Server) setupAdminRoutes() {
	// 1. Authentication routes (Public)
	s.Router.Get("/admin/login", s.HandleAdminLoginPage)
	s.Router.Post("/admin/login", s.HandleAdminLogin)
	s.Router.Get("/admin/logout", s.HandleAdminLogout)

	// Backwards compatibility for /client/login, /client/logout
	s.Router.Get("/client/login", s.HandleAdminLoginPage)
	s.Router.Post("/client/login", s.HandleAdminLogin)
	s.Router.Get("/client/logout", s.HandleAdminLogout)
	s.Router.Get("/client", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin"+r.URL.Path[len("/client"):], http.StatusPermanentRedirect)
	})

	// 2. Protected Admin Group
	s.Router.Group(func(admin chi.Router) {
		admin.Use(auth.RequireAdminAuth(s.SessionMgr, "/admin/login"))

		// Dashboard and tabs
		admin.Get("/admin", s.HandleAdminDashboard)
		admin.Get("/admin/tab/leads", s.HandleAdminTabLeads)
		admin.Get("/admin/tab/blocks", s.HandleAdminTabBlocks)
		admin.Get("/admin/tab/media", s.HandleAdminTabMedia)
		admin.Get("/admin/tab/settings", s.HandleAdminTabSettings)
		admin.Get("/admin/tab/services", s.HandleAdminTabServices)
		admin.Get("/admin/tab/articles", s.HandleAdminTabArticles)
		admin.Get("/admin/tab/contacts", s.HandleAdminTabContacts)
		admin.Get("/admin/tab/reviews", s.HandleAdminTabReviews)
		admin.Get("/admin/tab/fleet", s.HandleAdminTabFleet)
		admin.Get("/admin/tab/security", s.HandleAdminTabSecurity)

		// Block Builder actions (Phase 5.2)
		admin.Post("/admin/blocks/new", s.HandleAdminCreateBlock)
		admin.Post("/admin/blocks/{id}/toggle", s.HandleAdminToggleBlock)
		admin.Post("/admin/blocks/{id}/move-up", s.HandleAdminMoveBlockUp)
		admin.Post("/admin/blocks/{id}/move-down", s.HandleAdminMoveBlockDown)
		admin.Delete("/admin/blocks/{id}", s.HandleAdminDeleteBlock)

		// Leads actions (Phase 5.3)
		admin.Post("/admin/leads/{id}/status", s.HandleAdminUpdateLeadStatus)
		admin.Get("/admin/export/leads.csv", s.HandleAdminExportLeadsCSV)

		// Media Library upload (Phase 5.4)
		if s.Media != nil {
			admin.Post("/admin/media/upload", s.HandleAdminUploadMedia)
		}

		// Settings actions (Phase 5.5)
		admin.Post("/admin/settings", s.HandleAdminSaveSettings)
		admin.Post("/admin/contacts", s.HandleAdminSaveContacts)
		admin.Post("/admin/security/password", s.HandleAdminChangePassword)

		// Service management
		admin.Get("/admin/services/new", s.HandleAdminNewServiceForm)
		admin.Get("/admin/services/edit/{id}", s.HandleAdminEditServiceForm)
		admin.Post("/admin/services/save", s.HandleAdminSaveService)
		admin.Post("/admin/services/{id}/toggle", s.HandleAdminToggleService)
		admin.Post("/admin/services/{id}/trash", s.HandleAdminTrashService)

		// Article management
		admin.Get("/admin/articles/new", s.HandleAdminNewArticleForm)
		admin.Get("/admin/articles/edit/{id}", s.HandleAdminEditArticleForm)
		admin.Post("/admin/articles/save", s.HandleAdminSaveArticle)
		admin.Post("/admin/articles/toggle/{id}", s.HandleAdminToggleArticle)
		admin.Post("/admin/articles/trash/{id}", s.HandleAdminTrashArticle)

		// Reviews
		admin.Post("/admin/reviews/{id}/toggle", s.HandleAdminToggleReview)
		admin.Delete("/admin/reviews/{id}", s.HandleAdminDeleteReview)
		admin.Post("/admin/reviews/{id}/delete", s.HandleAdminDeleteReview)

		// Mirror under /client/* for full backwards compatibility
		admin.Get("/client/*", func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/client")
			http.Redirect(w, r, "/admin"+path, http.StatusTemporaryRedirect)
		})
	})
}

// HandleAdminLoginPage renders admin login screen.
func (s *Server) HandleAdminLoginPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.SessionMgr.GetBool(ctx, auth.SessionKeyAdminAuth) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	settings, _ := s.Repo.GetSiteSettings(ctx)
	errMsg := s.SessionMgr.PopString(ctx, auth.SessionKeyFlashError)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientLogin(settings, errMsg).Render(ctx, w)
}

// HandleAdminLogin authenticates administrator.
func (s *Server) HandleAdminLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	password := strings.TrimSpace(r.FormValue("password"))

	settings, _ := s.Repo.GetSiteSettings(ctx)

	authenticated := false
	if settings.ClientAdminPasswordHash != "" {
		authenticated = auth.CheckPassword(password, settings.ClientAdminPasswordHash)
	} else if settings.ClientAdminPassword != "" {
		authenticated = (password == settings.ClientAdminPassword)
	} else {
		// Default password if not configured
		authenticated = (password == "admin" || password == "sitecraft" || password == "admin123")
	}

	if !authenticated {
		s.SessionMgr.Put(ctx, auth.SessionKeyFlashError, "Неверный пароль доступа")
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	_ = auth.SetAdminSession(ctx, s.SessionMgr, "admin@sitecraft.local")
	slog.Info("Admin logged in successfully", "ip", r.RemoteAddr)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// HandleAdminLogout clears session.
func (s *Server) HandleAdminLogout(w http.ResponseWriter, r *http.Request) {
	_ = auth.ClearAdminSession(r.Context(), s.SessionMgr)
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

// HandleAdminDashboard renders base admin panel layout with selected tab.
func (s *Server) HandleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, _ := s.Repo.GetSiteSettings(ctx)
	content, _ := s.Repo.GetSiteContent(ctx)
	leads, _ := s.Repo.GetLeads(ctx)
	articles, _ := s.Repo.GetArticles(ctx)
	reviews, _ := s.Repo.GetReviews(ctx)
	services, _ := s.Repo.GetServices(ctx)

	activeTab := r.URL.Query().Get("tab")
	if activeTab == "" {
		activeTab = "leads"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientDashboard(settings, content, leads, articles, nil, reviews, services, nil, activeTab).Render(ctx, w)
}

// Tab handlers for HTMX partial swap
func (s *Server) HandleAdminTabLeads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	leads, _ := s.Repo.GetLeads(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientLeadsTab(leads).Render(ctx, w)
}

func (s *Server) HandleAdminTabBlocks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	blocks, _ := s.Repo.GetPageBlocks(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.AdminBlocksTab(blocks).Render(ctx, w)
}

func (s *Server) HandleAdminTabMedia(w http.ResponseWriter, r *http.Request) {
	files := s.scanMediaFiles()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.AdminMediaTab(files).Render(r.Context(), w)
}

func (s *Server) scanMediaFiles() []pages.MediaFileItem {
	var items []pages.MediaFileItem
	uploadsDir := "data/uploads/originals"
	entries, err := os.ReadDir(uploadsDir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				items = append(items, pages.MediaFileItem{
					Filename: e.Name(),
					URL:      "/media/raw/" + e.Name(),
					ThumbURL: "/media/thumb/" + e.Name(),
				})
			}
		}
	}
	// Also fallback public/ images
	pubEntries, err := os.ReadDir("public")
	if err == nil {
		for _, e := range pubEntries {
			if !e.IsDir() && (strings.HasSuffix(e.Name(), ".png") || strings.HasSuffix(e.Name(), ".jpg")) {
				items = append(items, pages.MediaFileItem{
					Filename: e.Name(),
					URL:      "/media/raw/" + e.Name(),
					ThumbURL: "/media/thumb/" + e.Name(),
				})
			}
		}
	}
	return items
}

func (s *Server) HandleAdminTabSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, _ := s.Repo.GetSiteSettings(ctx)
	content, _ := s.Repo.GetSiteContent(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.AdminSettingsTab(settings, content, "").Render(ctx, w)
}

func (s *Server) HandleAdminTabServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	services, _ := s.Repo.GetServices(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientServicesTab(services, nil, "").Render(ctx, w)
}

func (s *Server) HandleAdminTabArticles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	articles, _ := s.Repo.GetArticles(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientArticlesTab(articles, nil, false).Render(ctx, w)
}

func (s *Server) HandleAdminTabContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, _ := s.Repo.GetSiteSettings(ctx)
	content, _ := s.Repo.GetSiteContent(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientContactsTab(content, settings).Render(ctx, w)
}

func (s *Server) HandleAdminTabReviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reviews, _ := s.Repo.GetReviews(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientReviewsTab(reviews).Render(ctx, w)
}

func (s *Server) HandleAdminTabFleet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<div class="p-6 bg-slate-900 border border-slate-800 rounded-2xl text-slate-300"><h3 class="font-bold text-white mb-2">Fleet Management</h3><p class="text-xs">Управление кластером микросайтов активно в фоновом режиме.</p></div>`))
}

func (s *Server) HandleAdminTabSecurity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<div class="p-6 bg-slate-900 border border-slate-800 rounded-2xl text-slate-300"><h3 class="font-bold text-white mb-2">Безопасность кабинета</h3><form hx-post="/admin/security/password" class="space-y-4 max-w-sm"><input type="password" name="new_password" placeholder="Новый пароль..." required minlength="8" class="w-full px-3 py-2 rounded-xl bg-slate-800 border border-slate-700 text-white text-sm"/><button type="submit" class="px-4 py-2 bg-emerald-500 text-slate-950 font-bold rounded-xl text-xs">Обновить пароль</button></form></div>`))
}

// Block Builder actions (Phase 5.2)
func (s *Server) HandleAdminCreateBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	blockType := r.FormValue("type")
	title := r.FormValue("title")
	content := r.FormValue("content")

	blocks, _ := s.Repo.GetPageBlocks(ctx)
	order := int64(len(blocks) + 1)

	_, _ = s.Queries.CreatePageBlock(ctx, db.CreatePageBlockParams{
		ID:         uuid.New().String(),
		PageID:     "",
		Type:       blockType,
		Title:      title,
		Content:    content,
		OrderIndex: order,
		Active:     1,
		Payload:    "{}",
	})

	s.HandleAdminTabBlocks(w, r)
}

func (s *Server) HandleAdminToggleBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	block, err := s.Queries.GetPageBlockByID(ctx, id)
	if err == nil {
		newActive := int64(1)
		if block.Active == 1 {
			newActive = 0
		}
		_, _ = s.Queries.UpdatePageBlock(ctx, db.UpdatePageBlockParams{
			ID:         block.ID,
			Type:       block.Type,
			Title:      block.Title,
			Content:    block.Content,
			OrderIndex: block.OrderIndex,
			Active:     newActive,
			Payload:    block.Payload,
		})
	}
	s.HandleAdminTabBlocks(w, r)
}

func (s *Server) HandleAdminMoveBlockUp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	blocks, _ := s.Queries.ListPageBlocks(ctx, "")
	for i, b := range blocks {
		if b.ID == id && i > 0 {
			prev := blocks[i-1]
			_ = s.Queries.UpdatePageBlockOrder(ctx, db.UpdatePageBlockOrderParams{ID: b.ID, OrderIndex: prev.OrderIndex})
			_ = s.Queries.UpdatePageBlockOrder(ctx, db.UpdatePageBlockOrderParams{ID: prev.ID, OrderIndex: b.OrderIndex})
			break
		}
	}
	s.HandleAdminTabBlocks(w, r)
}

func (s *Server) HandleAdminMoveBlockDown(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	blocks, _ := s.Queries.ListPageBlocks(ctx, "")
	for i, b := range blocks {
		if b.ID == id && i < len(blocks)-1 {
			next := blocks[i+1]
			_ = s.Queries.UpdatePageBlockOrder(ctx, db.UpdatePageBlockOrderParams{ID: b.ID, OrderIndex: next.OrderIndex})
			_ = s.Queries.UpdatePageBlockOrder(ctx, db.UpdatePageBlockOrderParams{ID: next.ID, OrderIndex: b.OrderIndex})
			break
		}
	}
	s.HandleAdminTabBlocks(w, r)
}

func (s *Server) HandleAdminDeleteBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	_ = s.Queries.DeletePageBlock(ctx, id)
	s.HandleAdminTabBlocks(w, r)
}

// Leads actions (Phase 5.3)
func (s *Server) HandleAdminUpdateLeadStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	_ = r.ParseForm()
	status := r.FormValue("status")

	_, _ = s.Queries.UpdateLeadStatus(ctx, db.UpdateLeadStatusParams{
		ID:     id,
		Status: status,
	})
	s.HandleAdminTabLeads(w, r)
}

func (s *Server) HandleAdminExportLeadsCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	leads, _ := s.Repo.GetLeads(ctx)

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"leads.csv\"")

	_, _ = w.Write([]byte("ID,Name,Phone,Email,Source,BookingDate,BookingTime,Status,Created\n"))
	for _, l := range leads {
		line := fmt.Sprintf("%s,\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n",
			l.ID, l.Name, l.Phone, l.Email, l.Source, l.BookingDate, l.BookingTime, l.Status, l.Created)
		_, _ = w.Write([]byte(line))
	}
}

// Media upload handler (Phase 5.4)
func (s *Server) HandleAdminUploadMedia(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(media.MaxUploadSize); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		_, _ = s.Media.SaveUpload(file, header)
	}

	s.HandleAdminTabMedia(w, r)
}

// Settings actions (Phase 5.5)
func (s *Server) HandleAdminSaveSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()

	settings, _ := s.Repo.GetSiteSettings(ctx)
	content, _ := s.Repo.GetSiteContent(ctx)

	if val := r.FormValue("site_name"); val != "" {
		settings.SiteName = val
	}
	settings.Tagline = r.FormValue("tagline")
	if val := r.FormValue("theme"); val != "" {
		settings.Theme = val
	}
	settings.HeaderVariant = r.FormValue("header_variant")
	settings.YandexMetrikaID = r.FormValue("yandex_metrika_id")
	settings.WebhookURL = r.FormValue("webhook_url")
	settings.TelegramBotToken = r.FormValue("telegram_bot_token")
	settings.TelegramChatID = r.FormValue("telegram_chat_id")
	settings.ContactEmail = r.FormValue("contact_email")
	settings.LegalEntity = r.FormValue("legal_entity")
	settings.INN = r.FormValue("inn")

	if val := r.FormValue("phone1"); val != "" {
		content.Phone1 = val
		_ = s.Repo.UpsertSiteContent(ctx, content)
	}

	_ = s.Repo.UpsertSiteSettings(ctx, settings)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.AdminSettingsTab(settings, content, "Настройки успешно сохранены!").Render(ctx, w)
}

func (s *Server) HandleAdminSaveContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	content, _ := s.Repo.GetSiteContent(ctx)

	content.Phone1 = r.FormValue("phone1")
	content.Phone2 = r.FormValue("phone2")
	content.WorkHours = r.FormValue("work_hours")
	content.Address = r.FormValue("address")
	_ = s.Repo.UpsertSiteContent(ctx, content)

	s.HandleAdminTabContacts(w, r)
}

func (s *Server) HandleAdminChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	newPass := strings.TrimSpace(r.FormValue("new_password"))

	if len(newPass) >= 8 {
		hash, err := auth.HashPassword(newPass)
		if err == nil {
			settings, _ := s.Repo.GetSiteSettings(ctx)
			settings.ClientAdminPasswordHash = hash
			settings.ClientAdminPassword = ""
			_ = s.Repo.UpsertSiteSettings(ctx, settings)
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<div class="p-4 rounded-xl bg-emerald-500/10 text-emerald-400 text-xs font-bold">Пароль успешно обновлен!</div>`))
}

// Service CRUD
func (s *Server) HandleAdminNewServiceForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientServiceForm(models.ServiceItem{}, true).Render(r.Context(), w)
}

func (s *Server) HandleAdminEditServiceForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	svc, err := s.Queries.GetServiceByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	item := models.ServiceItem{
		ID:          svc.ID,
		Title:       svc.Title,
		Description: svc.Description,
		Price:       svc.Price,
		Order:       int(svc.OrderIndex),
		Active:      svc.Active == 1,
		StockCount:  int(svc.StockCount),
		IsPreorder:  svc.IsPreorder == 1,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientServiceForm(item, false).Render(ctx, w)
}

func (s *Server) HandleAdminSaveService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	id := r.FormValue("id")
	title := r.FormValue("title")
	desc := r.FormValue("description")
	price := r.FormValue("price")
	order, _ := strconv.Atoi(r.FormValue("order"))
	active := r.FormValue("active") == "on" || r.FormValue("active") == "true"
	actInt := int64(0)
	if active {
		actInt = 1
	}

	if id != "" {
		_, _ = s.Queries.UpdateService(ctx, db.UpdateServiceParams{
			ID:          id,
			Title:       title,
			Description: desc,
			Price:       price,
			OrderIndex:  int64(order),
			Active:      actInt,
			StockCount:  -1,
			IsPreorder:  0,
		})
	} else {
		_, _ = s.Queries.CreateService(ctx, db.CreateServiceParams{
			ID:          uuid.New().String(),
			Title:       title,
			Description: desc,
			Price:       price,
			OrderIndex:  int64(order),
			Active:      actInt,
			StockCount:  -1,
			IsPreorder:  0,
		})
	}
	s.HandleAdminTabServices(w, r)
}

func (s *Server) HandleAdminToggleService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	svc, err := s.Queries.GetServiceByID(ctx, id)
	if err == nil {
		newActive := int64(1)
		if svc.Active == 1 {
			newActive = 0
		}
		_, _ = s.Queries.UpdateService(ctx, db.UpdateServiceParams{
			ID:          svc.ID,
			Title:       svc.Title,
			Description: svc.Description,
			Price:       svc.Price,
			OrderIndex:  svc.OrderIndex,
			Active:      newActive,
			StockCount:  svc.StockCount,
			IsPreorder:  svc.IsPreorder,
		})
	}
	s.HandleAdminTabServices(w, r)
}

func (s *Server) HandleAdminTrashService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	_ = s.Queries.SoftDeleteService(ctx, id)
	s.HandleAdminTabServices(w, r)
}

// Article CRUD
func (s *Server) HandleAdminNewArticleForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientArticleForm(models.ArticleItem{}, true).Render(r.Context(), w)
}

func (s *Server) HandleAdminEditArticleForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	art, err := s.Queries.GetArticleByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	item := models.ArticleItem{
		ID:             art.ID,
		Title:          art.Title,
		Slug:           art.Slug,
		Excerpt:        art.Excerpt,
		Content:        art.Content,
		CoverImage:     art.CoverImage,
		AuthorName:     art.AuthorName,
		PublishedAt:    art.PublishedAt,
		IsPublished:    art.IsPublished == 1,
		SEOTitle:       art.SeoTitle,
		SEODescription: art.SeoDescription,
		Tags:           strings.Split(art.Tags, ","),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientArticleForm(item, false).Render(ctx, w)
}

func (s *Server) HandleAdminSaveArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	id := r.FormValue("id")
	title := r.FormValue("title")
	slug := r.FormValue("slug")
	excerpt := r.FormValue("excerpt")
	content := r.FormValue("content")
	author := r.FormValue("author_name")
	date := r.FormValue("published_at")
	isPub := r.FormValue("is_published") == "on" || r.FormValue("is_published") == "true"
	pubInt := int64(0)
	if isPub {
		pubInt = 1
	}

	if id != "" {
		_, _ = s.Queries.UpdateArticle(ctx, db.UpdateArticleParams{
			ID:             id,
			Title:          title,
			Slug:           slug,
			Excerpt:        excerpt,
			Content:        content,
			CoverImage:     "",
			AuthorName:     author,
			PublishedAt:    date,
			IsPublished:    pubInt,
			SeoTitle:       title,
			SeoDescription: excerpt,
			Tags:           "",
		})
	} else {
		_, _ = s.Queries.CreateArticle(ctx, db.CreateArticleParams{
			ID:             uuid.New().String(),
			Title:          title,
			Slug:           slug,
			Excerpt:        excerpt,
			Content:        content,
			CoverImage:     "",
			AuthorName:     author,
			PublishedAt:    date,
			IsPublished:    pubInt,
			SeoTitle:       title,
			SeoDescription: excerpt,
			Tags:           "",
		})
	}
	s.HandleAdminTabArticles(w, r)
}

func (s *Server) HandleAdminToggleArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	art, err := s.Queries.GetArticleByID(ctx, id)
	if err == nil {
		newPub := int64(1)
		if art.IsPublished == 1 {
			newPub = 0
		}
		_, _ = s.Queries.UpdateArticle(ctx, db.UpdateArticleParams{
			ID:             art.ID,
			Title:          art.Title,
			Slug:           art.Slug,
			Excerpt:        art.Excerpt,
			Content:        art.Content,
			CoverImage:     art.CoverImage,
			AuthorName:     art.AuthorName,
			PublishedAt:    art.PublishedAt,
			IsPublished:    newPub,
			SeoTitle:       art.SeoTitle,
			SeoDescription: art.SeoDescription,
			Tags:           art.Tags,
		})
	}
	s.HandleAdminTabArticles(w, r)
}

func (s *Server) HandleAdminTrashArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	_ = s.Queries.SoftDeleteArticle(ctx, id)
	s.HandleAdminTabArticles(w, r)
}

func (s *Server) HandleAdminToggleReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	_, _ = s.Queries.ApproveReview(ctx, id)
	s.HandleAdminTabReviews(w, r)
}

func (s *Server) HandleAdminDeleteReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	_, _ = s.DB.ExecContext(ctx, "DELETE FROM reviews WHERE id = ?", id)
	s.HandleAdminTabReviews(w, r)
}
