package server

import (
	"log/slog"
	"net/http"
	"strings"

	"sitecraft/ui/pages"

	"github.com/go-chi/chi/v5"
)

func (s *Server) HandleAdminTabReviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reviews, _ := s.Repo.GetReviews(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientReviewsTab(reviews).Render(ctx, w)
}

func (s *Server) HandleAdminToggleReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if err := s.Repo.ToggleReview(ctx, id); err != nil {
		slog.Error("failed to toggle review", "id", id, "error", err)
	} else {
		s.publishSiteMutated(ctx, "reviews")
	}
	w.Header().Set("HX-Trigger", "refreshChecklist")
	s.HandleAdminTabReviews(w, r)
}

func (s *Server) HandleAdminDeleteReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if err := s.Repo.DeleteReview(ctx, id); err != nil {
		slog.Error("failed to delete review", "id", id, "error", err)
	} else {
		s.publishSiteMutated(ctx, "reviews")
	}
	w.Header().Set("HX-Trigger", "refreshChecklist")
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
