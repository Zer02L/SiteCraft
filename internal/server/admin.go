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

// HandleAdminTabFleet renders the fleet management tab.
func (s *Server) HandleAdminTabFleet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientFleetTab().Render(r.Context(), w)
}

// HandleAdminTabSecurity renders the security settings tab.
func (s *Server) HandleAdminTabSecurity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientSecurityTab("", "").Render(r.Context(), w)
}
