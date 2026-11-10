package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"sitecraft/internal/auth"
	"sitecraft/internal/media"
	"sitecraft/internal/models"
	"sitecraft/ui/pages"

	"github.com/go-chi/chi/v5"
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
		target := "/admin"
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	})

	// 2. Protected Admin Group
	s.Router.Group(func(admin chi.Router) {
		admin.Use(auth.RequireAdminAuth(s.SessionMgr, "/admin/login"))

		// Dashboard and tabs
		admin.Get("/admin", s.HandleAdminDashboard)

		// Dynamic modular tab endpoints with security checks
		admin.Get("/admin/tab/leads", s.RequireModule("leads", s.HandleAdminTabLeads))
		admin.Get("/client/tab/leads", s.RequireModule("leads", s.HandleAdminTabLeads))
		admin.Get("/admin/tab/blocks", s.HandleAdminTabBlocks)
		admin.Get("/client/tab/blocks", s.HandleAdminTabBlocks)
		admin.Get("/admin/tab/media", s.HandleAdminTabMedia)
		admin.Get("/client/tab/media", s.HandleAdminTabMedia)
		admin.Get("/admin/tab/settings", s.HandleAdminTabSettings)
		admin.Get("/client/tab/settings", s.HandleAdminTabSettings)
		admin.Get("/admin/tab/services", s.RequireModule("services", s.HandleAdminTabServices))
		admin.Get("/client/tab/services", s.RequireModule("services", s.HandleAdminTabServices))
		admin.Get("/admin/tab/articles", s.RequireModule("articles", s.HandleAdminTabArticles))
		admin.Get("/client/tab/articles", s.RequireModule("articles", s.HandleAdminTabArticles))
		admin.Get("/admin/tab/contacts", s.RequireModule("contacts", s.HandleAdminTabContacts))
		admin.Get("/client/tab/contacts", s.RequireModule("contacts", s.HandleAdminTabContacts))
		admin.Get("/admin/tab/reviews", s.RequireModule("reviews", s.HandleAdminTabReviews))
		admin.Get("/client/tab/reviews", s.RequireModule("reviews", s.HandleAdminTabReviews))
		admin.Get("/admin/tab/fleet", s.RequireModule("fleet", s.HandleAdminTabFleet))
		admin.Get("/client/tab/fleet", s.RequireModule("fleet", s.HandleAdminTabFleet))
		admin.Get("/admin/tab/security", s.RequireModule("security", s.HandleAdminTabSecurity))
		admin.Get("/client/tab/security", s.RequireModule("security", s.HandleAdminTabSecurity))

		// Block Builder actions (Phase 5.2)
		admin.Post("/admin/blocks/new", s.HandleAdminCreateBlock)
		admin.Post("/client/blocks/new", s.HandleAdminCreateBlock)
		admin.Post("/admin/blocks/edit", s.HandleAdminUpdateBlock)
		admin.Post("/client/blocks/edit", s.HandleAdminUpdateBlock)
		admin.Post("/admin/blocks/{id}/edit", s.HandleAdminUpdateBlock)
		admin.Post("/client/blocks/{id}/edit", s.HandleAdminUpdateBlock)
		admin.Post("/admin/blocks/{id}/toggle", s.HandleAdminToggleBlock)
		admin.Post("/client/blocks/{id}/toggle", s.HandleAdminToggleBlock)
		admin.Post("/admin/blocks/{id}/move-up", s.HandleAdminMoveBlockUp)
		admin.Post("/client/blocks/{id}/move-up", s.HandleAdminMoveBlockUp)
		admin.Post("/admin/blocks/{id}/move-down", s.HandleAdminMoveBlockDown)
		admin.Post("/client/blocks/{id}/move-down", s.HandleAdminMoveBlockDown)
		admin.Delete("/admin/blocks/{id}", s.HandleAdminDeleteBlock)
		admin.Delete("/client/blocks/{id}", s.HandleAdminDeleteBlock)
		admin.Post("/admin/blocks/{id}/delete", s.HandleAdminDeleteBlock)
		admin.Post("/client/blocks/{id}/delete", s.HandleAdminDeleteBlock)

		// Leads actions (protected by RequireModuleGroup)
		admin.Group(func(ld chi.Router) {
			ld.Use(s.RequireModuleGroup("leads"))
			ld.Post("/admin/leads/{id}/status", s.HandleAdminUpdateLeadStatus)
			ld.Post("/client/leads/{id}/status", s.HandleAdminUpdateLeadStatus)
			ld.Get("/admin/export/leads.csv", s.HandleAdminExportLeadsCSV)
			ld.Get("/client/export/leads.csv", s.HandleAdminExportLeadsCSV)
		})

		// Media Library upload (Phase 5.4)
		if s.Media != nil {
			admin.Post("/admin/media/upload", s.HandleAdminUploadMedia)
		}

		// Settings actions (Phase 5.5)
		admin.Post("/admin/settings", s.HandleAdminSaveSettings)
		admin.Post("/admin/contacts", s.HandleAdminSaveContacts)
		admin.Post("/admin/security/password", s.HandleAdminChangePassword)

		// Service management (protected by RequireModuleGroup)
		admin.Group(func(srv chi.Router) {
			srv.Use(s.RequireModuleGroup("services"))
			srv.Get("/admin/services/new", s.HandleAdminNewServiceForm)
			srv.Get("/admin/services/edit/{id}", s.HandleAdminEditServiceForm)
			srv.Post("/admin/services/save", s.HandleAdminSaveService)
			srv.Post("/admin/services/{id}/toggle", s.HandleAdminToggleService)
			srv.Post("/admin/services/{id}/trash", s.HandleAdminTrashService)
		})

		// Article management (protected by RequireModuleGroup)
		admin.Group(func(art chi.Router) {
			art.Use(s.RequireModuleGroup("articles"))
			art.Get("/admin/articles/new", s.HandleAdminNewArticleForm)
			art.Get("/admin/articles/edit/{id}", s.HandleAdminEditArticleForm)
			art.Post("/admin/articles/save", s.HandleAdminSaveArticle)
			art.Post("/admin/articles/toggle/{id}", s.HandleAdminToggleArticle)
			art.Post("/admin/articles/trash/{id}", s.HandleAdminTrashArticle)
		})

		// Reviews (protected by RequireModuleGroup)
		admin.Group(func(rev chi.Router) {
			rev.Use(s.RequireModuleGroup("reviews"))
			rev.Post("/admin/reviews/{id}/toggle", s.HandleAdminToggleReview)
			rev.Delete("/admin/reviews/{id}", s.HandleAdminDeleteReview)
			rev.Post("/admin/reviews/{id}/delete", s.HandleAdminDeleteReview)
			rev.Post("/admin/reviews/{id}/reply", s.HandleAdminReplyReview)
			rev.Post("/client/reviews/{id}/toggle", s.HandleAdminToggleReview)
			rev.Delete("/client/reviews/{id}", s.HandleAdminDeleteReview)
			rev.Post("/client/reviews/{id}/delete", s.HandleAdminDeleteReview)
			rev.Post("/client/reviews/{id}/reply", s.HandleAdminReplyReview)
		})

		// Mirror under /client/* for full backwards compatibility
		admin.Get("/client/*", func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/client")
			target := "/admin" + path
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusTemporaryRedirect)
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
	blocks, _ := s.Repo.GetAllPageBlocks(ctx)

	// Fetch active modules dynamically from registry
	modules := s.AdminModules.GetActive(ctx, s, settings)

	activeTab := r.URL.Query().Get("tab")
	activeTab = s.ResolveActiveTab(activeTab, modules)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientDashboard(settings, content, leads, articles, nil, reviews, services, nil, activeTab, modules, blocks).Render(ctx, w)
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
	blocks, _ := s.Repo.GetAllPageBlocks(ctx)
	settings, _ := s.Repo.GetSiteSettings(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.AdminBlocksTab(blocks, settings).Render(ctx, w)
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

	blocks, _ := s.Repo.GetAllPageBlocks(ctx)
	order := len(blocks) + 1

	_, err := s.Repo.CreatePageBlock(ctx, models.PageBlock{
		Type:    blockType,
		Title:   title,
		Content: content,
		Order:   order,
		Active:  true,
		Payload: map[string]interface{}{},
	})
	if err != nil {
		slog.Error("failed to create page block", "error", err)
	}

	if r.Header.Get("HX-Request") == "true" {
		s.HandleAdminTabBlocks(w, r)
		return
	}
	http.Redirect(w, r, "/admin?tab=blocks", http.StatusSeeOther)
}

func (s *Server) HandleAdminUpdateBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	id := chi.URLParam(r, "id")
	if id == "" {
		id = r.FormValue("id")
	}
	title := r.FormValue("title")
	content := r.FormValue("content")
	blockType := r.FormValue("type")
	variant := r.FormValue("variant")
	slog.Info("HandleAdminUpdateBlock called", "id", id, "title", title, "type", blockType, "variant", variant)

	if id != "" {
		block, err := s.Repo.GetPageBlockByID(ctx, id)
		if err == nil {
			if title != "" {
				block.Title = title
			}
			block.Content = content
			if blockType != "" {
				block.Type = blockType
			}
			if variant != "" {
				if block.Payload == nil {
					block.Payload = make(map[string]interface{})
				}
				block.Payload["variant"] = variant
			}
			if err := s.Repo.UpdatePageBlock(ctx, block); err != nil {
				slog.Error("failed to update page block", "id", id, "error", err)
			}
		} else {
			slog.Error("page block not found for update", "id", id, "error", err)
		}
	}

	if r.Header.Get("HX-Request") == "true" {
		s.HandleAdminTabBlocks(w, r)
		return
	}
	http.Redirect(w, r, "/admin?tab=blocks", http.StatusSeeOther)
}

func (s *Server) HandleAdminToggleBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if err := s.Repo.TogglePageBlock(ctx, id); err != nil {
		slog.Error("failed to toggle page block", "id", id, "error", err)
	}
	if r.Header.Get("HX-Request") == "true" {
		s.HandleAdminTabBlocks(w, r)
		return
	}
	http.Redirect(w, r, "/admin?tab=blocks", http.StatusSeeOther)
}

func (s *Server) HandleAdminMoveBlockUp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	blocks, _ := s.Repo.GetAllPageBlocks(ctx)
	for i, b := range blocks {
		if b.ID == id && i > 0 {
			prev := blocks[i-1]
			if err := s.Repo.SwapPageBlockOrder(ctx, b.ID, int64(b.Order), prev.ID, int64(prev.Order)); err != nil {
				slog.Error("failed to move block up", "id", id, "error", err)
			}
			break
		}
	}
	if r.Header.Get("HX-Request") == "true" {
		s.HandleAdminTabBlocks(w, r)
		return
	}
	http.Redirect(w, r, "/admin?tab=blocks", http.StatusSeeOther)
}

func (s *Server) HandleAdminMoveBlockDown(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	blocks, _ := s.Repo.GetAllPageBlocks(ctx)
	for i, b := range blocks {
		if b.ID == id && i < len(blocks)-1 {
			next := blocks[i+1]
			if err := s.Repo.SwapPageBlockOrder(ctx, b.ID, int64(b.Order), next.ID, int64(next.Order)); err != nil {
				slog.Error("failed to move block down", "id", id, "error", err)
			}
			break
		}
	}
	if r.Header.Get("HX-Request") == "true" {
		s.HandleAdminTabBlocks(w, r)
		return
	}
	http.Redirect(w, r, "/admin?tab=blocks", http.StatusSeeOther)
}

func (s *Server) HandleAdminDeleteBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if err := s.Repo.DeletePageBlock(ctx, id); err != nil {
		slog.Error("failed to delete page block", "id", id, "error", err)
	}
	if r.Header.Get("HX-Request") == "true" {
		s.HandleAdminTabBlocks(w, r)
		return
	}
	http.Redirect(w, r, "/admin?tab=blocks", http.StatusSeeOther)
}

// Leads actions (Phase 5.3)
func (s *Server) HandleAdminUpdateLeadStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	_ = r.ParseForm()
	status := r.FormValue("status")

	if err := s.Repo.UpdateLeadStatus(ctx, id, status); err != nil {
		slog.Error("failed to update lead status", "id", id, "error", err)
	}
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
	item, err := s.Repo.GetServiceByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
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

	item := models.ServiceItem{
		ID:          id,
		Title:       title,
		Description: desc,
		Price:       price,
		Order:       order,
		Active:      active,
		StockCount:  -1,
		IsPreorder:  false,
	}

	if err := s.Repo.SaveService(ctx, item); err != nil {
		slog.Error("failed to save service", "error", err)
	}
	s.HandleAdminTabServices(w, r)
}

func (s *Server) HandleAdminToggleService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if err := s.Repo.ToggleService(ctx, id); err != nil {
		slog.Error("failed to toggle service", "id", id, "error", err)
	}
	s.HandleAdminTabServices(w, r)
}

func (s *Server) HandleAdminTrashService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if err := s.Repo.SoftDeleteService(ctx, id); err != nil {
		slog.Error("failed to trash service", "id", id, "error", err)
	}
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
	item, err := s.Repo.GetArticleByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
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

	item := models.ArticleItem{
		ID:             id,
		Title:          title,
		Slug:           slug,
		Excerpt:        excerpt,
		Content:        content,
		CoverImage:     "",
		AuthorName:     author,
		PublishedAt:    date,
		IsPublished:    isPub,
		SEOTitle:       title,
		SEODescription: excerpt,
		Tags:           nil,
	}

	if err := s.Repo.SaveArticle(ctx, item); err != nil {
		slog.Error("failed to save article", "error", err)
	}
	s.HandleAdminTabArticles(w, r)
}

func (s *Server) HandleAdminToggleArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if err := s.Repo.ToggleArticle(ctx, id); err != nil {
		slog.Error("failed to toggle article", "id", id, "error", err)
	}
	s.HandleAdminTabArticles(w, r)
}

func (s *Server) HandleAdminTrashArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if err := s.Repo.SoftDeleteArticle(ctx, id); err != nil {
		slog.Error("failed to trash article", "id", id, "error", err)
	}
	s.HandleAdminTabArticles(w, r)
}

func (s *Server) HandleAdminToggleReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if err := s.Repo.ToggleReview(ctx, id); err != nil {
		slog.Error("failed to toggle review", "id", id, "error", err)
	}
	s.HandleAdminTabReviews(w, r)
}

func (s *Server) HandleAdminDeleteReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if err := s.Repo.DeleteReview(ctx, id); err != nil {
		slog.Error("failed to delete review", "id", id, "error", err)
	}
	s.HandleAdminTabReviews(w, r)
}

func (s *Server) HandleAdminReplyReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	_ = r.ParseForm()
	replyText := strings.TrimSpace(r.FormValue("reply"))
	slog.InfoContext(ctx, "moderator replied to review", "id", id, "reply_len", len(replyText))
	s.HandleAdminTabReviews(w, r)
}
