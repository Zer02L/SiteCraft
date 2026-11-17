package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"sitecraft/internal/ai"
	"sitecraft/internal/auth"
)

func TestAdminAI_EngineAndMCPIntegration(t *testing.T) {
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
	_, _, err = srv.SessionMgr.Commit(ctx)
	if err != nil {
		t.Fatalf("failed to commit session: %v", err)
	}

	sendAuthRequest := func(method, path string, body *bytes.Buffer, contentType string) *httptest.ResponseRecorder {
		var req *http.Request
		if body != nil {
			req = httptest.NewRequest(method, path, body)
		} else {
			req = httptest.NewRequest(method, path, nil)
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		srv.Router.ServeHTTP(w, req)
		return w
	}

	// 2. Test Tab AI rendering
	wTab := sendAuthRequest("GET", "/admin/tab/ai", nil, "")
	if wTab.Code != http.StatusOK {
		t.Fatalf("expected 200 for /admin/tab/ai, got %d", wTab.Code)
	}
	bodyStr := wTab.Body.String()
	if !strings.Contains(bodyStr, "AI Engine") || !strings.Contains(bodyStr, "152-ФЗ") {
		t.Errorf("expected tab to contain AI Engine and 152-ФЗ, got: %s", bodyStr[:200])
	}

	// 3. Test Save AI Settings with AES-256-GCM encryption
	formSettings := url.Values{}
	formSettings.Set("ai_provider", "anymodel")
	formSettings.Set("ai_base_url", "https://api.anymodel.org/v1")
	formSettings.Set("ai_model_name", "llama-3.3-70b")
	formSettings.Set("ai_api_key", "secret-test-api-key-999")
	formSettings.Set("ai_proxy_url", "http://127.0.0.1:8080")

	wSave := sendAuthRequest("POST", "/admin/ai/settings", bytes.NewBufferString(formSettings.Encode()), "application/x-www-form-urlencoded")
	if wSave.Code != http.StatusOK {
		t.Fatalf("expected 200 for /admin/ai/settings, got %d", wSave.Code)
	}
	if !strings.Contains(wSave.Body.String(), "сохранены") {
		t.Errorf("expected success alert: %s", wSave.Body.String())
	}

	// Verify encryption in DB (key must NOT be stored in plaintext)
	settings, err := srv.Repo.GetSiteSettings(context.Background())
	if err != nil {
		t.Fatalf("failed to query site settings: %v", err)
	}
	if settings.AIAPIKeyEncrypted == "secret-test-api-key-999" {
		t.Fatalf("CRITICAL SECURITY VIOLATION: API key stored in plaintext!")
	}
	if !strings.HasPrefix(settings.AIAPIKeyEncrypted, "enc:v1:") {
		t.Errorf("expected enc:v1: prefix in encrypted key, got: %s", settings.AIAPIKeyEncrypted)
	}
	decryptedKey, err := ai.Decrypt(settings.AIAPIKeyEncrypted)
	if err != nil || decryptedKey != "secret-test-api-key-999" {
		t.Errorf("failed to decrypt saved key: %v (got %s)", err, decryptedKey)
	}

	// 4. Test AI Copilot Article Generator with Draft-First guarantee
	formGen := url.Values{}
	formGen.Set("topic", "Преимущества бесшовной интеграции с 152-ФЗ")
	formGen.Set("tone", "expert")
	formGen.Set("keywords", "безопасность, комплаенс, персональные данные")

	wGen := sendAuthRequest("POST", "/admin/ai/generate-article", bytes.NewBufferString(formGen.Encode()), "application/x-www-form-urlencoded")
	if wGen.Code != http.StatusOK {
		t.Fatalf("expected 200 for /admin/ai/generate-article, got %d", wGen.Code)
	}
	if !strings.Contains(wGen.Body.String(), "Draft-First") {
		t.Errorf("expected Draft-First notice in response: %s", wGen.Body.String())
	}

	// Verify article draft status in repository: MUST NOT be published!
	allArticles, _ := srv.Repo.GetArticles(context.Background())
	for _, a := range allArticles {
		if strings.Contains(a.Title, "Преимущества") {
			t.Fatalf("CRITICAL ERROR: AI article draft was published live!")
		}
	}

	// 5. Test AI Copilot SEO Generator
	formSEO := url.Values{}
	formSEO.Set("text", "Комплексные юридические услуги для IT-компаний в Санкт-Петербурге. Защита интеллектуальной собственности, аудит договоров.")

	wSEO := sendAuthRequest("POST", "/admin/ai/generate-seo", bytes.NewBufferString(formSEO.Encode()), "application/x-www-form-urlencoded")
	if wSEO.Code != http.StatusOK {
		t.Fatalf("expected 200 for /admin/ai/generate-seo, got %d", wSEO.Code)
	}
	if !strings.Contains(wSEO.Body.String(), "Title") || !strings.Contains(wSEO.Body.String(), "Description") {
		t.Errorf("expected SEO result: %s", wSEO.Body.String())
	}

	// 6. Test External AI Agent Endpoints (Inbound Protocol)

	// A. OpenAPI 3.1 Schema
	reqOpenAPI := httptest.NewRequest("GET", "/api/v1/agent/openapi.json", nil)
	wOpenAPI := httptest.NewRecorder()
	srv.Router.ServeHTTP(wOpenAPI, reqOpenAPI)
	if wOpenAPI.Code != http.StatusOK {
		t.Fatalf("expected 200 for openapi.json, got %d", wOpenAPI.Code)
	}
	var spec map[string]interface{}
	if err := json.Unmarshal(wOpenAPI.Body.Bytes(), &spec); err != nil {
		t.Fatalf("invalid OpenAPI JSON: %v", err)
	}
	if spec["openapi"] != "3.1.0" {
		t.Errorf("expected openapi 3.1.0, got %v", spec["openapi"])
	}

	// B. Agent Info endpoint with 152-FZ PII masking
	reqInfo := httptest.NewRequest("GET", "/api/v1/agent/info", nil)
	wInfo := httptest.NewRecorder()
	srv.Router.ServeHTTP(wInfo, reqInfo)
	if wInfo.Code != http.StatusOK {
		t.Fatalf("expected 200 for agent info, got %d", wInfo.Code)
	}
	if !strings.Contains(wInfo.Body.String(), "phone_masked") || !strings.Contains(wInfo.Body.String(), "152-FZ") {
		t.Errorf("expected 152-FZ masked info: %s", wInfo.Body.String())
	}

	// C. Agent Create Article REST endpoint (Draft-First guarantee)
	artPayload := `{
		"title": "Новая статья от автономного AI-агента",
		"content": "<p>Контент статьи</p>",
		"slug": "autonomous-agent-article"
	}`
	reqArt := httptest.NewRequest("POST", "/api/v1/agent/articles", bytes.NewBufferString(artPayload))
	reqArt.Header.Set("Content-Type", "application/json")
	wArt := httptest.NewRecorder()
	srv.Router.ServeHTTP(wArt, reqArt)
	if wArt.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for agent article, got %d: %s", wArt.Code, wArt.Body.String())
	}
	if !strings.Contains(wArt.Body.String(), `"is_published":false`) {
		t.Errorf("expected is_published:false in agent response: %s", wArt.Body.String())
	}

	// D. MCP Protocol (Model Context Protocol over SSE and POST)
	mcpInit := `{"jsonrpc":"2.0","id":100,"method":"initialize","params":{}}`
	reqMCP := httptest.NewRequest("POST", "/api/v1/mcp", bytes.NewBufferString(mcpInit))
	wMCP := httptest.NewRecorder()
	srv.Router.ServeHTTP(wMCP, reqMCP)
	if wMCP.Code != http.StatusOK {
		t.Fatalf("expected 200 for MCP POST, got %d", wMCP.Code)
	}
	if !strings.Contains(wMCP.Body.String(), "sitecraft-mcp") {
		t.Errorf("expected sitecraft-mcp in serverInfo: %s", wMCP.Body.String())
	}
}
