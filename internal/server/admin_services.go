package server

import (
	"log/slog"
	"net/http"
	"strconv"

	"sitecraft/internal/models"
	"sitecraft/ui/pages"

	"github.com/go-chi/chi/v5"
)

func (s *Server) HandleAdminTabServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	services, _ := s.Repo.GetServices(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientServicesTab(services, nil, "").Render(ctx, w)
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
	} else {
		s.publishSiteMutated(ctx, "services")
	}
	w.Header().Set("HX-Trigger", "refreshChecklist")
	s.HandleAdminTabServices(w, r)
}

func (s *Server) HandleAdminToggleService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if err := s.Repo.ToggleService(ctx, id); err != nil {
		slog.Error("failed to toggle service", "id", id, "error", err)
	} else {
		s.publishSiteMutated(ctx, "services")
	}
	w.Header().Set("HX-Trigger", "refreshChecklist")
	s.HandleAdminTabServices(w, r)
}

func (s *Server) HandleAdminTrashService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if err := s.Repo.SoftDeleteService(ctx, id); err != nil {
		slog.Error("failed to trash service", "id", id, "error", err)
	} else {
		s.publishSiteMutated(ctx, "services")
	}
	w.Header().Set("HX-Trigger", "refreshChecklist")
	s.HandleAdminTabServices(w, r)
}
