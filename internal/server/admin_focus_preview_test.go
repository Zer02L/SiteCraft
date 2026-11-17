package server_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sitecraft/internal/checker"
	"sitecraft/internal/db"
	"sitecraft/internal/models"
	"sitecraft/ui/pages"
)

// TestFeature5_FocusMode_DashboardAndForm verifies Focus Mode integration in the admin panel and article editor.
func TestFeature5_FocusMode_DashboardAndForm(t *testing.T) {
	ctx := context.Background()

	settings := models.SiteSettings{
		SiteName:   "SiteCraft Test",
		EnableBlog: true,
	}

	modules := []models.AdminTabModule{
		{ID: "articles", Title: "Статьи", Route: "/client/tab/articles"},
	}

	checklist := checker.ChecklistResult{
		TotalCount:     5,
		CompletedCount: 3,
		Percentage:     60,
	}

	// 1. Check ClientDashboard render includes focusMode Alpine state & window listeners
	var dashBuf bytes.Buffer
	err := pages.ClientDashboard(settings, models.SiteContent{}, nil, nil, nil, nil, nil, nil, "articles", modules, nil, checklist).Render(ctx, &dashBuf)
	if err != nil {
		t.Fatalf("failed to render ClientDashboard: %v", err)
	}
	dashHTML := dashBuf.String()

	if !strings.Contains(dashHTML, "focusMode: false") {
		t.Errorf("ClientDashboard missing 'focusMode: false' in x-data")
	}
	if !strings.Contains(dashHTML, "@toggle-focus-mode.window") {
		t.Errorf("ClientDashboard missing '@toggle-focus-mode.window' event listener")
	}
	if !strings.Contains(dashHTML, "admin-focus-mode") {
		t.Errorf("ClientDashboard missing 'admin-focus-mode' class binding")
	}
	if !strings.Contains(dashHTML, "Режим Focus Mode") {
		t.Errorf("ClientDashboard missing minimal Focus Mode bar")
	}

	// 2. Check ClientArticleForm render includes Focus Mode toggle & Quill focus integration
	art := models.ArticleItem{
		ID:          "art-123",
		Title:       "Как настроить Focus Mode",
		Slug:        "kak-nastroit-focus-mode",
		Content:     "<p>Чистый текст для фокус-режима</p>",
		AuthorName:  "Редактор",
		IsPublished: true,
	}

	var formBuf bytes.Buffer
	err = pages.ClientArticleForm(art, false).Render(ctx, &formBuf)
	if err != nil {
		t.Fatalf("failed to render ClientArticleForm: %v", err)
	}
	formHTML := formBuf.String()

	if !strings.Contains(formHTML, "$dispatch('toggle-focus-mode')") {
		t.Errorf("ClientArticleForm missing $dispatch('toggle-focus-mode') toggle button")
	}
	if !strings.Contains(formHTML, "Focus Mode") {
		t.Errorf("ClientArticleForm missing Focus Mode button text/label")
	}
	if !strings.Contains(formHTML, "quill-dark-wrapper") {
		t.Errorf("ClientArticleForm missing quill-dark-wrapper")
	}
	if !strings.Contains(formHTML, "showExtraSettings") {
		t.Errorf("ClientArticleForm missing showExtraSettings accordion for metadata")
	}
}

// TestFeature5_LiveHoverPreview_ArticlesTab verifies the single-instance hover preview container with debounce.
func TestFeature5_LiveHoverPreview_ArticlesTab(t *testing.T) {
	ctx := context.Background()

	articles := []models.ArticleItem{
		{
			ID:          "post-1",
			Title:       "Секреты юридического маркетинга",
			Slug:        "sekrety-marketinga",
			Excerpt:     "Короткий анонс статьи",
			PublishedAt: "2026-05-10T12:00:00Z",
			IsPublished: true,
		},
		{
			ID:          "post-2",
			Title:       "Обзор судебной практики 2026",
			Slug:        "obzor-praktiki",
			Excerpt:     "Обзор ключевых дел",
			PublishedAt: "2026-06-01T10:00:00Z",
			IsPublished: false,
		},
	}

	var buf bytes.Buffer
	err := pages.ClientArticlesTab(articles, nil, false).Render(ctx, &buf)
	if err != nil {
		t.Fatalf("failed to render ClientArticlesTab: %v", err)
	}
	html := buf.String()

	// 1. Verify Alpine singleton preview controller
	if !strings.Contains(html, `x-data="articleHoverPreview()"`) {
		t.Errorf("ClientArticlesTab missing x-data='articleHoverPreview()'")
	}

	// 2. Verify 400ms debounce hover on article row
	if !strings.Contains(html, "@mouseenter.debounce.400ms") {
		t.Errorf("ClientArticlesTab missing @mouseenter.debounce.400ms hover handler on rows")
	}
	if !strings.Contains(html, "@mouseleave=\"hidePreview()\"") {
		t.Errorf("ClientArticlesTab missing @mouseleave='hidePreview()' on rows")
	}

	// 3. Verify singleton preview container & template x-if
	if !strings.Contains(html, "hover-preview-singleton") {
		t.Errorf("ClientArticlesTab missing 'hover-preview-singleton' container")
	}
	if !strings.Contains(html, "<template x-if=\"previewUrl\">") {
		t.Errorf("ClientArticlesTab missing '<template x-if=\"previewUrl\">' to prevent multiple iframes")
	}
	if !strings.Contains(html, "<iframe") {
		t.Errorf("ClientArticlesTab missing iframe element inside template")
	}

	// 4. Verify touch screen guard media query
	if !strings.Contains(html, "@media not all and (hover: hover) and (pointer: fine)") {
		t.Errorf("ClientArticlesTab missing CSS touch-screen override '@media not all and (hover: hover) and (pointer: fine)'")
	}

	// 5. Verify manual preview button
	if !strings.Contains(html, "openPreviewNow") {
		t.Errorf("ClientArticlesTab missing openPreviewNow manual trigger")
	}
}

// TestFeature5_HandleBlogPost_PreviewAndFrameHeaders verifies that:
// - preview=1 or X-Preview: 1 sets X-Frame-Options: SAMEORIGIN
// - preview requests do NOT increment article view counter
// - normal public requests DO increment article view counter
func TestFeature5_HandleBlogPost_PreviewAndFrameHeaders(t *testing.T) {
	srv, conn := setupTestServer(t)
	defer conn.Close()

	ctx := context.Background()

	// Enable blog in settings
	settings, err := srv.Repo.GetSiteSettings(ctx)
	if err != nil {
		t.Fatalf("GetSiteSettings failed: %v", err)
	}
	settings.EnableBlog = true
	_ = srv.Repo.UpsertSiteSettings(ctx, settings)

	// Create test article in DB
	art, err := srv.Queries.CreateArticle(ctx, db.CreateArticleParams{
		ID:             "preview-post-id",
		Title:          "Превью тест статьи",
		Slug:           "preview-test-post",
		Excerpt:        "Краткий анонс",
		Content:        "<p>Контент</p>",
		CoverImage:     "",
		AuthorName:     "Автор",
		PublishedAt:    "2026-01-01T00:00:00Z",
		IsPublished:    1,
		SeoTitle:       "SEO Title",
		SeoDescription: "SEO Desc",
		Tags:           "тест",
	})
	if err != nil {
		t.Fatalf("CreateArticle failed: %v", err)
	}

	// 1. First test: preview request via query parameter ?preview=1
	reqPreview := httptest.NewRequest("GET", "/blog/"+art.Slug+"?preview=1", nil)
	wPreview := httptest.NewRecorder()
	srv.Router.ServeHTTP(wPreview, reqPreview)

	if wPreview.Code != http.StatusOK {
		t.Fatalf("expected 200 for preview request, got %d", wPreview.Code)
	}
	if xfo := wPreview.Header().Get("X-Frame-Options"); xfo != "SAMEORIGIN" {
		t.Errorf("expected X-Frame-Options: SAMEORIGIN, got %q", xfo)
	}

	// Check that views count was NOT incremented (remains 0)
	artAfterPreview, err := srv.Queries.GetArticleBySlug(ctx, art.Slug)
	if err != nil {
		t.Fatalf("GetArticleBySlug failed: %v", err)
	}
	if artAfterPreview.ViewsCount != 0 {
		t.Errorf("expected views_count to remain 0 on preview request, got %d", artAfterPreview.ViewsCount)
	}

	// 2. Second test: preview request via header X-Preview: 1
	reqHeaderPreview := httptest.NewRequest("GET", "/blog/"+art.Slug, nil)
	reqHeaderPreview.Header.Set("X-Preview", "1")
	wHeaderPreview := httptest.NewRecorder()
	srv.Router.ServeHTTP(wHeaderPreview, reqHeaderPreview)

	if wHeaderPreview.Code != http.StatusOK {
		t.Fatalf("expected 200 for header preview request, got %d", wHeaderPreview.Code)
	}
	artAfterHeaderPreview, _ := srv.Queries.GetArticleBySlug(ctx, art.Slug)
	if artAfterHeaderPreview.ViewsCount != 0 {
		t.Errorf("expected views_count to remain 0 on header preview request, got %d", artAfterHeaderPreview.ViewsCount)
	}

	// 3. Third test: normal public visit without preview flag
	reqNormal := httptest.NewRequest("GET", "/blog/"+art.Slug, nil)
	wNormal := httptest.NewRecorder()
	srv.Router.ServeHTTP(wNormal, reqNormal)

	if wNormal.Code != http.StatusOK {
		t.Fatalf("expected 200 for normal request, got %d", wNormal.Code)
	}

	// Allow query execution to finish
	time.Sleep(10 * time.Millisecond)

	artAfterNormal, _ := srv.Queries.GetArticleBySlug(ctx, art.Slug)
	if artAfterNormal.ViewsCount != 1 {
		t.Errorf("expected views_count to be incremented to 1 on normal request, got %d", artAfterNormal.ViewsCount)
	}
}
