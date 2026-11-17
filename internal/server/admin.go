package server

import (
	"context"
	"net/http"

	"sitecraft/internal/checker"
	"sitecraft/internal/events"
	"sitecraft/ui/pages"
)

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

	var checklist checker.ChecklistResult
	if s.Checker != nil {
		checklist = s.Checker.GetChecklist(ctx, settings)
	} else {
		checklist = checker.EvaluateSiteHealth(settings, s.DB, checker.WithRepo(s.Repo))
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientDashboard(settings, content, leads, articles, nil, reviews, services, nil, activeTab, modules, blocks, checklist).Render(ctx, w)
}

// HandleAdminChecklist renders setup checklist widget for partial HTMX reload.
func (s *Server) HandleAdminChecklist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, _ := s.Repo.GetSiteSettings(ctx)
	var checklist checker.ChecklistResult
	if s.Checker != nil {
		checklist = s.Checker.GetChecklist(ctx, settings)
	} else {
		checklist = checker.EvaluateSiteHealth(settings, s.DB, checker.WithRepo(s.Repo))
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.SetupChecklistWidget(checklist).Render(ctx, w)
}

// publishSiteMutated invalidates the in-memory checklist and notifies event bus subscribers.
func (s *Server) publishSiteMutated(ctx context.Context, scope string) {
	if s.Checker != nil {
		s.Checker.Invalidate()
	}
	if s.EventBus != nil {
		s.EventBus.Publish(ctx, events.SiteMutatedEvent{Scope: scope})
	}
}

// HandleAdminTabFleet renders the fleet management tab placeholder.
func (s *Server) HandleAdminTabFleet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<div class="p-6 bg-slate-900 border border-slate-800 rounded-2xl text-slate-300"><h3 class="font-bold text-white mb-2">Fleet Management</h3><p class="text-xs">Управление кластером микросайтов активно в фоновом режиме.</p></div>`))
}

// HandleAdminTabSecurity renders the security settings tab.
func (s *Server) HandleAdminTabSecurity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<div class="p-6 bg-slate-900 border border-slate-800 rounded-2xl text-slate-300"><h3 class="font-bold text-white mb-2">Безопасность кабинета</h3><form hx-post="/admin/security/password" class="space-y-4 max-w-sm"><input type="password" name="new_password" placeholder="Новый пароль..." required minlength="8" class="w-full px-3 py-2 rounded-xl bg-slate-800 border border-slate-700 text-white text-sm"/><button type="submit" class="px-4 py-2 bg-emerald-500 text-slate-950 font-bold rounded-xl text-xs">Обновить пароль</button></form></div>`))
}
