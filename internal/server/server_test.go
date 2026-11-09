package server_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"sitecraft/internal/auth"
	"sitecraft/internal/db"
	"sitecraft/internal/media"
	"sitecraft/internal/server"
)

func setupTestServer(t *testing.T) (*server.Server, *sql.DB) {
	conn, queries, err := db.Open("file:srvtest?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	mediaSvc, err := media.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("failed to init media service: %v", err)
	}

	sessionMgr := auth.NewSessionManager(conn, false)

	srv, err := server.New(conn, queries, mediaSvc, sessionMgr)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	return srv, conn
}

func TestPublicRoutes(t *testing.T) {
	srv, conn := setupTestServer(t)
	defer conn.Close()

	routes := []string{
		"/",
		"/blog",
		"/privacy",
		"/terms",
		"/careers",
		"/invoice",
		"/track",
		"/shop",
		"/vip",
		"/fleet",
		"/sitemap.xml",
	}

	for _, route := range routes {
		req := httptest.NewRequest("GET", route, nil)
		rr := httptest.NewRecorder()
		srv.Router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("route %s returned status %d, expected 200", route, rr.Code)
		}
	}
}

func TestHTMXLeadSubmission(t *testing.T) {
	srv, conn := setupTestServer(t)
	defer conn.Close()

	form := url.Values{}
	form.Set("name", "Алексей Смирнов")
	form.Set("phone", "+7 (911) 222-33-44")
	form.Set("source", "Онлайн-запись")

	req := httptest.NewRequest("POST", "/api/leads", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Алексей Смирнов") {
		t.Errorf("expected response to contain lead name, got: %s", rr.Body.String())
	}
}

func TestAdminAuthProtection(t *testing.T) {
	srv, conn := setupTestServer(t)
	defer conn.Close()

	// 1. Unauthenticated request to /admin should redirect to /admin/login
	req := httptest.NewRequest("GET", "/admin", nil)
	rr := httptest.NewRecorder()
	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected redirect status 303, got %d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/admin/login" {
		t.Errorf("expected redirect to /admin/login, got %s", loc)
	}

	// 2. Login with password
	loginForm := url.Values{}
	loginForm.Set("password", "admin123")
	loginReq := httptest.NewRequest("POST", "/admin/login", strings.NewReader(loginForm.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRR := httptest.NewRecorder()

	srv.Router.ServeHTTP(loginRR, loginReq)

	if loginRR.Code != http.StatusSeeOther {
		t.Fatalf("login failed, status %d", loginRR.Code)
	}

	cookies := loginRR.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie after login")
	}

	// 3. Authenticated request to /admin should return 200
	authReq := httptest.NewRequest("GET", "/admin", nil)
	for _, c := range cookies {
		authReq.AddCookie(c)
	}
	authRR := httptest.NewRecorder()
	srv.Router.ServeHTTP(authRR, authReq)

	if authRR.Code != http.StatusOK {
		t.Errorf("expected status 200 for authenticated admin, got %d", authRR.Code)
	}
}
