package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sitecraft/internal/auth"
)

func TestModularAdmin_RegistryAndSecurity(t *testing.T) {
	srv, conn := setupTestServer(t)
	defer conn.Close()

	ctx := context.Background()

	// 1. Authenticate an admin session
	sessionCookie, err := srv.SessionMgr.Load(ctx, "")
	if err != nil {
		t.Fatalf("failed to init session: %v", err)
	}
	ctx = sessionCookie
	_ = auth.SetAdminSession(ctx, srv.SessionMgr, "admin@sitecraft.local")

	// Save session to cookie
	_, _, err = srv.SessionMgr.Commit(ctx)
	if err != nil {
		t.Fatalf("failed to commit session: %v", err)
	}

	// Helper to send authenticated request
	sendAuthRequest := func(method, path string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, nil)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()
		srv.Router.ServeHTTP(w, r)
		return w
	}

	// 2. Default settings test: Blog/Articles is disabled by default (EnableBlog=false)
	settings, err := srv.Repo.GetSiteSettings(ctx)
	if err != nil {
		t.Fatalf("GetSiteSettings failed: %v", err)
	}

	// Ensure EnableBlog is false
	settings.EnableBlog = false
	_ = srv.Repo.UpsertSiteSettings(ctx, settings)

	// Inactive module endpoint must return 404 Not Found
	wTabArticles := sendAuthRequest("GET", "/admin/tab/articles")
	if wTabArticles.Code != http.StatusNotFound {
		t.Errorf("expected 404 for disabled module tab /admin/tab/articles, got %d", wTabArticles.Code)
	}

	// Inactive module action must return 404 Not Found
	wActionArticles := sendAuthRequest("POST", "/admin/articles/save")
	if wActionArticles.Code != http.StatusNotFound {
		t.Errorf("expected 404 for disabled module action /admin/articles/save, got %d", wActionArticles.Code)
	}

	// 3. Enable Blog/Articles module
	settings.EnableBlog = true
	_ = srv.Repo.UpsertSiteSettings(ctx, settings)

	wTabArticlesEnabled := sendAuthRequest("GET", "/admin/tab/articles")
	if wTabArticlesEnabled.Code != http.StatusOK {
		t.Errorf("expected 200 for enabled module tab /admin/tab/articles, got %d", wTabArticlesEnabled.Code)
	}

	// 4. Test disabling via SetEnabled (e.g. disable "reviews" module)
	srv.AdminModules.SetEnabled("reviews", false)

	wTabReviews := sendAuthRequest("GET", "/admin/tab/reviews")
	if wTabReviews.Code != http.StatusNotFound {
		t.Errorf("expected 404 for disabled module reviews, got %d", wTabReviews.Code)
	}

	// 5. Test ResolveActiveTab fallback
	activeMods := srv.AdminModules.GetActive(ctx, srv, settings)
	fallbackTab := srv.ResolveActiveTab("reviews", activeMods)
	if fallbackTab == "reviews" {
		t.Errorf("ResolveActiveTab should not return disabled tab 'reviews', got %s", fallbackTab)
	}
	if fallbackTab == "" {
		t.Errorf("ResolveActiveTab should return a fallback tab, got empty")
	}

	// 6. Test Dashboard rendering with active modules
	wDashboard := sendAuthRequest("GET", "/admin")
	if wDashboard.Code != http.StatusOK {
		t.Errorf("expected 200 for /admin dashboard, got %d", wDashboard.Code)
	}
	body := wDashboard.Body.String()
	if !strings.Contains(body, "Клиентская панель") {
		t.Errorf("dashboard response missing header 'Клиентская панель'")
	}
	if strings.Contains(body, "Отзывы клиентов") {
		t.Errorf("dashboard should NOT display disabled 'Отзывы клиентов' tab")
	}
}

func TestAdminSaveThemeTokens_AndTypography(t *testing.T) {
	srv, conn := setupTestServer(t)
	defer conn.Close()

	ctx := context.Background()
	sessionCookie, err := srv.SessionMgr.Load(ctx, "")
	if err != nil {
		t.Fatalf("failed to init session: %v", err)
	}
	ctx = sessionCookie
	_ = auth.SetAdminSession(ctx, srv.SessionMgr, "admin@sitecraft.local")
	_, _, _ = srv.SessionMgr.Commit(ctx)

	// 1. Post to /admin/theme/tokens with custom tokens and flavor
	bodyStr := "theme=coral&primary_color=%230f172a&accent_color=%23e65a40&design_flavor=sharp&font_family=editorial&background_color=%23ffffff"
	req := httptest.NewRequest("POST", "/admin/theme/tokens", strings.NewReader(bodyStr))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from /admin/theme/tokens, got %d: %s", w.Code, w.Body.String())
	}

	// 2. Verify settings in repository
	settings, err := srv.Repo.GetSiteSettings(context.Background())
	if err != nil {
		t.Fatalf("failed to get site settings: %v", err)
	}

	if settings.Theme != "coral" {
		t.Errorf("expected theme 'coral', got %s", settings.Theme)
	}
	if settings.PrimaryColor != "#0f172a" {
		t.Errorf("expected primary color '#0f172a', got %s", settings.PrimaryColor)
	}
	if settings.GetDesignFlavor() != "sharp" {
		t.Errorf("expected design flavor 'sharp', got %s", settings.GetDesignFlavor())
	}
	if settings.GetFontPair() != "editorial" {
		t.Errorf("expected font pair 'editorial', got %s", settings.GetFontPair())
	}
	if settings.GetCustomBgColor() != "#ffffff" {
		t.Errorf("expected custom bg '#ffffff', got %s", settings.GetCustomBgColor())
	}

	// 3. Test block creation with typography auto-formatting
	blockReqBody := "type=hero&title=Юридические+услуги+в+Москве+и+области&content=Компания+\"Лидер\"+-+надежный+партнер"
	bReq := httptest.NewRequest("POST", "/admin/blocks/new", strings.NewReader(blockReqBody))
	bReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	bReq.Header.Set("HX-Request", "true")
	bReq = bReq.WithContext(ctx)

	bRec := httptest.NewRecorder()
	srv.Router.ServeHTTP(bRec, bReq)

	if bRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from /admin/blocks/new, got %d", bRec.Code)
	}

	blocks, err := srv.Repo.GetAllPageBlocks(context.Background())
	if err != nil {
		t.Fatalf("GetAllPageBlocks failed: %v", err)
	}

	var createdHero bool
	for _, b := range blocks {
		if b.Type == "hero" {
			createdHero = true
			if !strings.Contains(b.Title, "в\u00A0Москве") || !strings.Contains(b.Title, "и\u00A0области") {
				t.Errorf("block title missing typography nbsp: %q", b.Title)
			}
			if !strings.Contains(b.Content, "«Лидер»") {
				t.Errorf("block content missing normalized quotes: %q", b.Content)
			}
			if !strings.Contains(b.Content, "—") {
				t.Errorf("block content missing em-dash: %q", b.Content)
			}
		}
	}
	if !createdHero {
		t.Errorf("created hero block not found")
	}
}

