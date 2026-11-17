package server

import (
	"net/http"
	"strings"

	"sitecraft/internal/auth"

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

		// Setup Checklist API (Feature 2)
		admin.Get("/admin/api/checklist", s.HandleAdminChecklist)
		admin.Get("/client/api/checklist", s.HandleAdminChecklist)

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
		admin.Get("/admin/tab/importer", s.HandleAdminTabImporter)
		admin.Get("/client/tab/importer", s.HandleAdminTabImporter)
		admin.Get("/admin/tab/ai", s.HandleAdminTabAI)
		admin.Get("/client/tab/ai", s.HandleAdminTabAI)

		// AI Engine & Copilot actions (Feature 4)
		admin.Post("/admin/ai/settings", s.HandleAdminSaveAISettings)
		admin.Post("/client/ai/settings", s.HandleAdminSaveAISettings)
		admin.Post("/admin/ai/test", s.HandleAdminTestAIConnection)
		admin.Post("/client/ai/test", s.HandleAdminTestAIConnection)
		admin.Post("/admin/ai/generate-article", s.HandleAdminAIGenerateArticle)
		admin.Post("/client/ai/generate-article", s.HandleAdminAIGenerateArticle)
		admin.Post("/admin/ai/generate-seo", s.HandleAdminAIGenerateSEO)
		admin.Post("/client/ai/generate-seo", s.HandleAdminAIGenerateSEO)

		// Importer & Redirect actions (Feature 3: WordPress Migration & 301 Redirects)
		admin.Post("/admin/importer/upload", s.HandleAdminImporterUpload)
		admin.Post("/client/importer/upload", s.HandleAdminImporterUpload)
		admin.Get("/admin/importer/sse/{jobID}", s.HandleAdminImporterSSE)
		admin.Get("/client/importer/sse/{jobID}", s.HandleAdminImporterSSE)
		admin.Post("/admin/redirects/new", s.HandleAdminCreateRedirect)
		admin.Post("/client/redirects/new", s.HandleAdminCreateRedirect)
		admin.Post("/admin/redirects/{id}/delete", s.HandleAdminDeleteRedirect)
		admin.Delete("/admin/redirects/{id}", s.HandleAdminDeleteRedirect)
		admin.Post("/client/redirects/{id}/delete", s.HandleAdminDeleteRedirect)
		admin.Delete("/client/redirects/{id}", s.HandleAdminDeleteRedirect)

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

		// Settings & Theme actions (Phase 1: Quick Wins - Theme Presets Switcher)
		admin.Post("/admin/settings", s.HandleAdminSaveSettings)
		admin.Post("/admin/theme/tokens", s.HandleAdminSaveThemeTokens)
		admin.Post("/client/theme/tokens", s.HandleAdminSaveThemeTokens)
		admin.Post("/admin/api/theme/apply", s.HandleAdminApplyThemePreset)
		admin.Post("/client/api/theme/apply", s.HandleAdminApplyThemePreset)
		admin.Post("/admin/api/theme/rollback", s.HandleAdminRollbackThemePreset)
		admin.Post("/client/api/theme/rollback", s.HandleAdminRollbackThemePreset)
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
