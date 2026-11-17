package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"sitecraft/internal/ai"
	"sitecraft/internal/models"

	"github.com/google/uuid"
)

// HandleAgentInfo returns site info and aggregate stats with 152-FZ PII masking.
func (s *Server) HandleAgentInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, err := s.Repo.GetSiteSettings(ctx)
	if err != nil {
		http.Error(w, "Failed to load settings", http.StatusInternalServerError)
		return
	}
	content, _ := s.Repo.GetSiteContent(ctx)
	articles, _ := s.Repo.GetArticles(ctx)
	services, _ := s.Repo.GetServices(ctx)

	info := map[string]interface{}{
		"site_name":      settings.SiteName,
		"tagline":        settings.Tagline,
		"theme":          settings.Theme,
		"phone_masked":   ai.MaskPII(content.Phone1),
		"email_masked":   ai.MaskPII(settings.ContactEmail),
		"articles_count": len(articles),
		"services_count": len(services),
		"draft_first":    true,
		"compliance":     "152-FZ Compliant",
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(info)
}

// HandleAgentListArticles returns articles for external AI agents.
func (s *Server) HandleAgentListArticles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	articles, err := s.Repo.GetArticles(ctx)
	if err != nil {
		http.Error(w, "Failed to list articles", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(articles)
}

// HandleAgentCreateArticle creates an article draft (Draft-First principle).
func (s *Server) HandleAgentCreateArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body struct {
		Title          string   `json:"title"`
		Content        string   `json:"content"`
		Excerpt        string   `json:"excerpt"`
		Slug           string   `json:"slug"`
		Tags           []string `json:"tags"`
		SEOTitle       string   `json:"seo_title"`
		SEODescription string   `json:"seo_description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	if body.Title == "" || body.Content == "" {
		http.Error(w, "Title and content are required", http.StatusUnprocessableEntity)
		return
	}

	slug := body.Slug
	if slug == "" {
		slug = strings.ToLower(body.Title)
		slug = strings.ReplaceAll(slug, " ", "-")
	}

	// DRAFT-FIRST GUARANTEE: is_published is strictly false
	art := models.ArticleItem{
		ID:             uuid.New().String(),
		Title:          body.Title,
		Slug:           slug,
		Excerpt:        body.Excerpt,
		Content:        body.Content,
		Tags:           body.Tags,
		SEOTitle:       body.SEOTitle,
		SEODescription: body.SEODescription,
		AuthorName:     "AI Agent",
		IsPublished:    false, // STRICTLY DRAFT
		PublishedAt:    time.Now().Format("2006-01-02 15:04"),
	}

	if err := s.Repo.SaveArticle(ctx, art); err != nil {
		http.Error(w, "Failed to save article draft", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "created",
		"id":           art.ID,
		"slug":         art.Slug,
		"is_published": false,
		"notice":       "Article created strictly as draft (Draft-First policy). Human review required before publishing.",
	})
}

// HandleAgentGenerateSEO generates SEO metadata via configured AI provider.
func (s *Server) HandleAgentGenerateSEO(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	if body.Text == "" {
		http.Error(w, "Text is required", http.StatusUnprocessableEntity)
		return
	}

	provider, err := s.GetAIProvider(ctx)
	if err != nil {
		http.Error(w, "AI provider not configured: "+err.Error(), http.StatusBadGateway)
		return
	}

	meta, err := provider.GenerateSEO(ctx, body.Text)
	if err != nil {
		http.Error(w, "SEO generation failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(meta)
}
