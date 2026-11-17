package ai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sitecraft/internal/ai"
	"sitecraft/internal/testutil"
)

func TestCrypto_EncryptDecrypt(t *testing.T) {
	plain := "sk-proj-super-secret-key-12345"

	encrypted, err := ai.Encrypt(plain)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if !strings.HasPrefix(encrypted, "enc:v1:") {
		t.Fatalf("expected prefix 'enc:v1:', got: %s", encrypted)
	}

	decrypted, err := ai.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if decrypted != plain {
		t.Fatalf("expected %s, got %s", plain, decrypted)
	}

	// Plaintext backward compatibility test
	plainDecrypted, err := ai.Decrypt("raw-plain-key-without-prefix")
	if err != nil {
		t.Fatalf("decryption of plain text failed: %v", err)
	}
	if plainDecrypted != "raw-plain-key-without-prefix" {
		t.Fatalf("expected plain key unchanged, got %s", plainDecrypted)
	}
}

func TestPII_Masking_152FZ(t *testing.T) {
	tests := []struct {
		input    string
		contains string
	}{
		{
			input:    "Контакты клиента: +7 (999) 123-45-67 или 8 (916) 789-01-23",
			contains: "+7 (999) ***-45-67",
		},
		{
			input:    "Наш телефон 89991234567, пишите в WhatsApp",
			contains: "+7 (999) ***-45-67",
		},
		{
			input:    "Отправьте запрос на ivan.petrov@mail.ru или support@company.com",
			contains: "i***v@mail.ru",
		},
	}

	for _, tc := range tests {
		masked := ai.MaskPII(tc.input)
		if !strings.Contains(masked, tc.contains) {
			t.Errorf("MaskPII(%q) = %q, expected to contain %q", tc.input, masked, tc.contains)
		}
	}
}

func TestProviders_PresetFactory(t *testing.T) {
	presetOpenAI := ai.GetPreset("openai")
	if presetOpenAI.DefaultModel != "gpt-4o-mini" {
		t.Errorf("expected default model gpt-4o-mini, got %s", presetOpenAI.DefaultModel)
	}

	presetYandex := ai.GetPreset("yandex")
	if !presetYandex.Compliant152FZ {
		t.Errorf("expected Yandex to be 152-FZ compliant")
	}

	// Instantiate OpenAI provider
	p, err := ai.NewProvider(ai.Config{
		Provider: "openai",
		APIKey:   "dummy-key",
	})
	if err != nil {
		t.Fatalf("failed to create openai provider: %v", err)
	}
	if p.Name() != "openai" {
		t.Errorf("expected name openai, got %s", p.Name())
	}
}

func TestMCPServer_ProtocolAndDraftFirst(t *testing.T) {
	repo, _, _ := testutil.NewTestRepo(t)
	mcpServer := ai.NewMCPServer(repo, func(ctx context.Context) (ai.LLMProvider, error) {
		return nil, nil // Fallback SEO testing
	})

	// 1. Initialize
	initReq := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
	req := httptest.NewRequest("POST", "/api/v1/mcp", bytes.NewBufferString(initReq))
	w := httptest.NewRecorder()
	mcpServer.HandleMessage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var initResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &initResp); err != nil {
		t.Fatalf("failed to parse JSON-RPC: %v", err)
	}
	res := initResp["result"].(map[string]interface{})
	if res["protocolVersion"] != "2024-11-05" {
		t.Errorf("expected protocolVersion 2024-11-05, got %v", res["protocolVersion"])
	}

	// 2. Tools List
	listReq := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`
	req = httptest.NewRequest("POST", "/api/v1/mcp", bytes.NewBufferString(listReq))
	w = httptest.NewRecorder()
	mcpServer.HandleMessage(w, req)

	if !strings.Contains(w.Body.String(), "sitecraft_create_article") {
		t.Errorf("expected tools list to contain sitecraft_create_article")
	}

	// 3. Tool Call: Create Article (Draft-First Verification)
	createReq := `{
		"jsonrpc": "2.0",
		"id": 3,
		"method": "tools/call",
		"params": {
			"name": "sitecraft_create_article",
			"arguments": {
				"title": "AI сгенерированная статья",
				"content": "<p>Текст статьи</p>",
				"slug": "ai-generated-article"
			}
		}
	}`
	req = httptest.NewRequest("POST", "/api/v1/mcp", bytes.NewBufferString(createReq))
	w = httptest.NewRecorder()
	mcpServer.HandleMessage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Draft-First") {
		t.Errorf("expected Draft-First notice in tool response: %s", w.Body.String())
	}

	// Verify article in DB is strictly draft (is_published = false)
	articles, err := repo.GetArticles(context.Background())
	if err != nil {
		t.Fatalf("failed to query articles: %v", err)
	}
	// Note: GetArticles returns published articles by default. If draft-first worked, it won't be returned by GetArticles!
	for _, a := range articles {
		if a.Slug == "ai-generated-article" {
			t.Fatalf("CRITICAL SECURITY VIOLATION: AI created article was automatically published!")
		}
	}

	// 4. Tool Call: Generate SEO
	seoReq := `{
		"jsonrpc": "2.0",
		"id": 4,
		"method": "tools/call",
		"params": {
			"name": "sitecraft_generate_seo",
			"arguments": {
				"text": "Ремонт стиральных машин в Москве с выездом мастера на дом. Гарантия 1 год."
			}
		}
	}`
	req = httptest.NewRequest("POST", "/api/v1/mcp", bytes.NewBufferString(seoReq))
	w = httptest.NewRecorder()
	mcpServer.HandleMessage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for SEO, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "title") || !strings.Contains(w.Body.String(), "description") {
		t.Errorf("expected SEO result with title and description: %s", w.Body.String())
	}

	// 5. Tool Call: Site Info (PII Masking verification)
	infoReq := `{
		"jsonrpc": "2.0",
		"id": 5,
		"method": "tools/call",
		"params": {
			"name": "sitecraft_get_site_info",
			"arguments": {}
		}
	}`
	req = httptest.NewRequest("POST", "/api/v1/mcp", bytes.NewBufferString(infoReq))
	w = httptest.NewRecorder()
	mcpServer.HandleMessage(w, req)

	if !strings.Contains(w.Body.String(), "phone_masked") || !strings.Contains(w.Body.String(), "152-FZ") {
		t.Errorf("expected site info with 152-FZ masking: %s", w.Body.String())
	}
}

func TestOpenAPI_Specification(t *testing.T) {
	spec := ai.GenerateOpenAPISpec("https://sitecraft.example.com")
	if spec["openapi"] != "3.1.0" {
		t.Errorf("expected openapi 3.1.0, got %v", spec["openapi"])
	}
	paths := spec["paths"].(map[string]interface{})
	if _, ok := paths["/api/v1/agent/articles"]; !ok {
		t.Errorf("expected /api/v1/agent/articles path")
	}
}
