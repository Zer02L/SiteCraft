package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"sitecraft/internal/data"

	"github.com/microcosm-cc/bluemonday"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// TestSanitizeCSVField verifies CSV Formula Injection mitigation (OWASP).
func TestSanitizeCSVField(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal text", "Иван Иванов", "Иван Иванов"},
		{"Formula with equal sign", "=cmd|' /C calc'!A0", "'=cmd|' /C calc'!A0"},
		{"Formula with plus", "+12345", "'+12345"},
		{"Formula with minus", "-5+10", "'-5+10"},
		{"Formula with at", "@SUM(1+1)", "'@SUM(1+1)"},
		{"Formula with tab prefix", "\t=cmd", "'\t=cmd"},
		{"Empty string", "", ""},
		{"Whitespace only", "   ", "   "},
		{"Leading spaces before formula", "  =SUM(A1:B1)", "'  =SUM(A1:B1)"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizeCSVField(tc.input)
			if result != tc.expected {
				t.Errorf("sanitizeCSVField(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestSessionStoreTrustBoundary verifies authentication token security and TTL.
func TestSessionStoreTrustBoundary(t *testing.T) {
	store := NewSessionStore()

	t.Run("Empty token is invalid", func(t *testing.T) {
		if store.IsValid("") {
			t.Error("Expected empty token to be invalid")
		}
	})

	t.Run("Unknown token is invalid", func(t *testing.T) {
		if store.IsValid("random_non_existent_token_12345") {
			t.Error("Expected unknown token to be invalid")
		}
	})

	t.Run("Created token is valid", func(t *testing.T) {
		token, err := store.CreateSession()
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
		if token == "" {
			t.Fatal("Expected non-empty token")
		}
		if !store.IsValid(token) {
			t.Error("Expected newly created token to be valid")
		}
	})

	t.Run("Revoked token becomes immediately invalid", func(t *testing.T) {
		token, _ := store.CreateSession()
		store.RevokeSession(token)
		if store.IsValid(token) {
			t.Error("Expected revoked token to be invalid")
		}
	})

	t.Run("Expired token is cleaned up and invalid", func(t *testing.T) {
		token, _ := store.CreateSession()

		// Manually backdate the expiration time to the past
		store.mu.Lock()
		info := store.sessions[token]
		info.ExpiresAt = time.Now().Add(-1 * time.Hour)
		store.sessions[token] = info
		store.mu.Unlock()

		if store.IsValid(token) {
			t.Error("Expected expired token to be invalid")
		}

		// Ensure it was purged from the internal map
		store.mu.RLock()
		_, exists := store.sessions[token]
		store.mu.RUnlock()
		if exists {
			t.Error("Expected expired token to be purged from session map")
		}
	})
}

// TestDoubleBookingPrevention verifies the trust boundary preventing conflicting slot reservations.
func TestDoubleBookingPrevention(t *testing.T) {
	testDir := t.TempDir()
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: testDir,
	})
	defer app.ResetBootstrapState()

	if err := app.Bootstrap(); err != nil {
		t.Fatalf("Bootstrap error: %v", err)
	}
	if err := data.InitSchema(app); err != nil {
		t.Fatalf("InitSchema error: %v", err)
	}

	handler := NewLeadHandler(app)

	// Helper to make a lead request
	makeLeadReq := func(name, phone, date, timeSlot string) *httptest.ResponseRecorder {
		form := url.Values{}
		form.Set("name", name)
		form.Set("phone", phone)
		form.Set("booking_date", date)
		form.Set("booking_time", timeSlot)
		form.Set("consent", "on")
		form.Set("company_website_url", "") // honeypot empty

		req := httptest.NewRequest(http.MethodPost, "/api/leads", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		re := &core.RequestEvent{App: app}
		re.Request = req
		re.Response = rec

		_ = handler.HandleCreateLead(re)
		return rec
	}

	date := "2026-11-15"
	timeSlot := "14:30"

	// 1. Первая запись на слот должна пройти успешно (HTTP 200)
	rec1 := makeLeadReq("Иван Иванов", "+7 (999) 111-22-33", date, timeSlot)
	if rec1.Code != http.StatusOK {
		t.Fatalf("First booking expected 200 OK, got %d: %s", rec1.Code, rec1.Body.String())
	}

	// 2. Вторая попытка записаться на тот же слот должна вернуть 409 Conflict
	rec2 := makeLeadReq("Петр Петров", "+7 (999) 444-55-66", date, timeSlot)
	if rec2.Code != http.StatusConflict {
		t.Fatalf("Second booking for same slot expected 409 Conflict, got %d", rec2.Code)
	}

	// 3. Запись на другое время в тот же день должна пройти успешно
	rec3 := makeLeadReq("Сидор Сидоров", "+7 (999) 777-88-99", date, "15:30")
	if rec3.Code != http.StatusOK {
		t.Fatalf("Booking for different slot expected 200 OK, got %d", rec3.Code)
	}

	// 4. Если первый лид отменен (status = 'отменена'), слот освобождается для новой записи
	recToCancel, err := app.FindFirstRecordByFilter("leads", "booking_date = {:d} && booking_time = {:t}", map[string]interface{}{
		"d": date,
		"t": timeSlot,
	})
	if err != nil || recToCancel == nil {
		t.Fatalf("Failed to find lead to cancel: %v", err)
	}
	recToCancel.Set("status", "отменена")
	if err := app.Save(recToCancel); err != nil {
		t.Fatalf("Failed to save cancelled lead: %v", err)
	}

	rec4 := makeLeadReq("Алексей Смирнов", "+7 (999) 000-11-22", date, timeSlot)
	if rec4.Code != http.StatusOK {
		t.Fatalf("Booking for cancelled slot expected 200 OK, got %d: %s", rec4.Code, rec4.Body.String())
	}
}

// TestXSSSanitizationTrustBoundary verifies bluemonday UGC policy strips malicious scripts.
func TestXSSSanitizationTrustBoundary(t *testing.T) {
	policy := bluemonday.UGCPolicy()

	maliciousPayloads := []struct {
		name     string
		input    string
		forbidden string
	}{
		{"Script tag injection", "Добрый день! <script>alert('xss')</script> Жду звонка", "<script>"},
		{"Onerror handler in img", "Комментарий <img src=x onerror=alert(1)>", "onerror"},
		{"Iframe injection", "Посмотрите <iframe src='http://evil.com'></iframe>", "<iframe"},
		{"Javascript URI", `<a href="javascript:alert('XSS')">Кликните здесь</a>`, "javascript:"},
	}

	for _, tc := range maliciousPayloads {
		t.Run(tc.name, func(t *testing.T) {
			sanitized := policy.Sanitize(tc.input)
			if strings.Contains(strings.ToLower(sanitized), strings.ToLower(tc.forbidden)) {
				t.Errorf("Sanitizer failed to remove forbidden payload %q from result: %q", tc.forbidden, sanitized)
			}
		})
	}
}

// TestExportLeadsCSVInjection verifies that CSV export neutralizes formula injection attacks.
func TestExportLeadsCSVInjection(t *testing.T) {
	testDir := t.TempDir()
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: testDir,
	})
	defer app.ResetBootstrapState()

	if err := app.Bootstrap(); err != nil {
		t.Fatalf("Bootstrap error: %v", err)
	}
	if err := data.InitSchema(app); err != nil {
		t.Fatalf("InitSchema error: %v", err)
	}

	// Create a lead with formula injection in name and comment
	col, err := app.FindCollectionByNameOrId("leads")
	if err != nil {
		t.Fatalf("FindCollection error: %v", err)
	}

	rec := core.NewRecord(col)
	rec.Set("name", "=cmd|' /C calc'!A0")
	rec.Set("phone", "+7 (999) 123-45-67")
	rec.Set("email", "attacker@test.local")
	rec.Set("status", "новая")
	rec.Set("source", "website")
	rec.Set("payload", map[string]interface{}{
		"comment": "@SUM(1+1)*cmd",
	})
	if err := app.Save(rec); err != nil {
		t.Fatalf("Failed to save malicious lead: %v", err)
	}

	adminHandler := NewClientAdminHandler(app)
	token, _ := adminHandler.SessionStore().CreateSession()

	httpRec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodGet, "/client/export/leads.csv", nil)
	httpReq.AddCookie(&http.Cookie{
		Name:  ClientAuthCookieName,
		Value: token,
	})

	re := &core.RequestEvent{App: app}
	re.Request = httpReq
	re.Response = httpRec

	if err := adminHandler.HandleExportLeadsCSV(re); err != nil {
		t.Fatalf("HandleExportLeadsCSV error: %v", err)
	}

	csvOutput := httpRec.Body.String()

	// Verify that Excel formula injection triggers are escaped with single quote '
	if strings.Contains(csvOutput, ";=cmd") {
		t.Errorf("CSV contains unescaped formula: %s", csvOutput)
	}
	if !strings.Contains(csvOutput, ";'=cmd") {
		t.Errorf("CSV expected escaped formula ';'=cmd', got: %s", csvOutput)
	}
	if strings.Contains(csvOutput, ";@SUM") {
		t.Errorf("CSV contains unescaped @SUM: %s", csvOutput)
	}
	if !strings.Contains(csvOutput, ";'@SUM") {
		t.Errorf("CSV expected escaped '@SUM', got: %s", csvOutput)
	}
}

// TestRequireAuthMiddleware verifies that RequireAuth handles browser vs HTMX unauthenticated requests.
func TestRequireAuthMiddleware(t *testing.T) {
	testDir := t.TempDir()
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: testDir,
	})
	defer app.ResetBootstrapState()

	adminHandler := NewClientAdminHandler(app)
	middleware := adminHandler.RequireAuth()

	t.Run("Standard browser request without auth redirects with 302", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/client/dashboard", nil)
		re := &core.RequestEvent{App: app}
		re.Request = req
		re.Response = rec

		err := middleware(re)
		if err != nil {
			t.Fatalf("Unexpected error from middleware: %v", err)
		}
		if rec.Code != http.StatusFound {
			t.Errorf("Expected status 302 Found, got %d", rec.Code)
		}
		if loc := rec.Header().Get("Location"); loc != "/client/login" {
			t.Errorf("Expected Location /client/login, got %q", loc)
		}
	})

	t.Run("HTMX request without auth returns HX-Redirect header to prevent UI injection", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/client/tab/leads", nil)
		req.Header.Set("HX-Request", "true")
		re := &core.RequestEvent{App: app}
		re.Request = req
		re.Response = rec

		err := middleware(re)
		if err != nil {
			t.Fatalf("Unexpected error from middleware: %v", err)
		}
		if hxRedir := rec.Header().Get("HX-Redirect"); hxRedir != "/client/login" {
			t.Errorf("Expected HX-Redirect /client/login, got %q", hxRedir)
		}
	})

	t.Run("Authenticated request proceeds with Next", func(t *testing.T) {
		token, _ := adminHandler.SessionStore().CreateSession()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/client/dashboard", nil)
		req.AddCookie(&http.Cookie{
			Name:  ClientAuthCookieName,
			Value: token,
		})
		re := &core.RequestEvent{App: app}
		re.Request = req
		re.Response = rec

		err := middleware(re)
		if err != nil {
			t.Fatalf("Expected nil from Next(), got %v", err)
		}
		if rec.Code == http.StatusFound {
			t.Error("Authenticated request should not be redirected")
		}
	})
}
