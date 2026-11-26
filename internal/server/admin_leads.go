package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"sitecraft/ui/pages"

	"github.com/go-chi/chi/v5"
)

func (s *Server) HandleAdminTabLeads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	leads, _ := s.Repo.GetLeads(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientLeadsTab(leads).Render(ctx, w)
}

func (s *Server) HandleAdminUpdateLeadStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	_ = r.ParseForm()
	status := r.FormValue("status")

	if err := s.Repo.UpdateLeadStatus(ctx, id, status); err != nil {
		slog.Error("failed to update lead status", "id", id, "error", err)
	}
	// Refresh only the tab content. Replacing the root x-data element would make
	// Alpine evaluate the rows' x-show before their parent scope is available.
	leads, _ := s.Repo.GetLeads(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.LeadsTabContent(leads).Render(ctx, w)
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
