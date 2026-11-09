package auth_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"sitecraft/internal/auth"
	"sitecraft/internal/db"
)

func setupTestDB(t *testing.T) *sql.DB {
	conn, _, err := db.Open("file:authtest?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	return conn
}

func TestBcryptHashing(t *testing.T) {
	password := "SecretPass123!"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if !auth.CheckPassword(password, hash) {
		t.Errorf("expected password to match hash")
	}

	if auth.CheckPassword("WrongPass", hash) {
		t.Errorf("expected wrong password to not match")
	}
}

func TestSessionAndAuthMiddleware(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close()

	sm := auth.NewSessionManager(conn, false)

	// Protected handler
	protected := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Protected secret"))
	})

	handler := sm.LoadAndSave(auth.RequireAdminAuth(sm, "/admin/login")(protected))

	// 1. Unauthenticated request should be redirected
	req := httptest.NewRequest("GET", "/admin/dashboard", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/admin/login" {
		t.Errorf("expected redirect to /admin/login, got %s", loc)
	}

	// 2. Authenticate session
	loginHandler := sm.LoadAndSave(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := auth.SetAdminSession(r.Context(), sm, "admin@sitecraft.local"); err != nil {
			t.Fatalf("SetAdminSession failed: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))

	loginReq := httptest.NewRequest("POST", "/admin/login", nil)
	loginRR := httptest.NewRecorder()
	loginHandler.ServeHTTP(loginRR, loginReq)

	cookie := loginRR.Result().Cookies()
	if len(cookie) == 0 {
		t.Fatalf("expected session cookie to be set")
	}

	// 3. Authenticated request with cookie should succeed
	authReq := httptest.NewRequest("GET", "/admin/dashboard", nil)
	for _, c := range cookie {
		authReq.AddCookie(c)
	}
	authRR := httptest.NewRecorder()
	handler.ServeHTTP(authRR, authReq)

	if authRR.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, authRR.Code)
	}
	if authRR.Body.String() != "Protected secret" {
		t.Errorf("expected body 'Protected secret', got '%s'", authRR.Body.String())
	}
}
