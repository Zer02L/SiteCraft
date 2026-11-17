package server

import (
	"log/slog"
	"net/http"

	"sitecraft/internal/models"
	"sitecraft/ui/pages"

	"github.com/go-chi/chi/v5"
)

func (s *Server) HandleAdminTabArticles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	articles, _ := s.Repo.GetArticles(ctx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientArticlesTab(articles, nil, false).Render(ctx, w)
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
	} else {
		s.publishSiteMutated(ctx, "articles")
	}
	w.Header().Set("HX-Trigger", "refreshChecklist")
	s.HandleAdminTabArticles(w, r)
}

func (s *Server) HandleAdminToggleArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if err := s.Repo.ToggleArticle(ctx, id); err != nil {
		slog.Error("failed to toggle article", "id", id, "error", err)
	} else {
		s.publishSiteMutated(ctx, "articles")
	}
	w.Header().Set("HX-Trigger", "refreshChecklist")
	s.HandleAdminTabArticles(w, r)
}

func (s *Server) HandleAdminTrashArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if err := s.Repo.SoftDeleteArticle(ctx, id); err != nil {
		slog.Error("failed to trash article", "id", id, "error", err)
	} else {
		s.publishSiteMutated(ctx, "articles")
	}
	w.Header().Set("HX-Trigger", "refreshChecklist")
	s.HandleAdminTabArticles(w, r)
}
