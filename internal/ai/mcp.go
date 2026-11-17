package ai

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"sitecraft/internal/data"
	"sitecraft/internal/models"
)

// MCP JSON-RPC 2.0 structures
type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *jsonRPCError `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type mcpToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type mcpToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type mcpContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type mcpToolCallResult struct {
	Content []mcpContentItem `json:"content"`
	IsError bool             `json:"isError,omitempty"`
}

// MCPServer manages incoming Model Context Protocol connections over SSE and HTTP.
type MCPServer struct {
	repo         *data.Repository
	getProvider  func(ctx context.Context) (LLMProvider, error)
	sessions     sync.Map // sessionId -> chan []byte
	tools        []mcpToolDefinition
}

// NewMCPServer instantiates an MCP Server instance.
func NewMCPServer(repo *data.Repository, getProvider func(ctx context.Context) (LLMProvider, error)) *MCPServer {
	server := &MCPServer{
		repo:        repo,
		getProvider: getProvider,
	}
	server.initTools()
	return server
}

func (s *MCPServer) initTools() {
	s.tools = []mcpToolDefinition{
		{
			Name:        "sitecraft_get_site_info",
			Description: "Получить общее состояние сайта, настройки и агрегированную статистику. Персональные данные маскируются по 152-ФЗ РФ.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "sitecraft_list_articles",
			Description: "Получить список статей блога (ID, заголовок, slug, статус публикации, просмотры, теги).",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "sitecraft_get_article",
			Description: "Получить полный текст и метаданные конкретной статьи по ID или slug.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":   map[string]interface{}{"type": "string", "description": "ID статьи"},
					"slug": map[string]interface{}{"type": "string", "description": "URL-slug статьи"},
				},
			},
		},
		{
			Name:        "sitecraft_create_article",
			Description: "Создать новую статью в блоге. ПРИНЦИП DRAFT-FIRST: статья сохраняется строго как черновик (is_published = false) для обязательной проверки человеком.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title":           map[string]interface{}{"type": "string", "description": "Заголовок статьи"},
					"content":         map[string]interface{}{"type": "string", "description": "HTML или Markdown контент статьи"},
					"excerpt":         map[string]interface{}{"type": "string", "description": "Краткое превью"},
					"slug":            map[string]interface{}{"type": "string", "description": "URL slug (если не указан, сформируется автоматически)"},
					"tags":            map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Теги"},
					"seo_title":       map[string]interface{}{"type": "string", "description": "SEO Title"},
					"seo_description": map[string]interface{}{"type": "string", "description": "SEO Description"},
				},
				"required": []string{"title", "content"},
			},
		},
		{
			Name:        "sitecraft_update_article",
			Description: "Обновить существующую статью (заголовок, контент, теги, SEO). Не публикует автоматически черновики.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":              map[string]interface{}{"type": "string", "description": "ID статьи"},
					"title":           map[string]interface{}{"type": "string", "description": "Новый заголовок"},
					"content":         map[string]interface{}{"type": "string", "description": "Новый контент"},
					"excerpt":         map[string]interface{}{"type": "string", "description": "Новое превью"},
					"tags":            map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"seo_title":       map[string]interface{}{"type": "string", "description": "SEO Title"},
					"seo_description": map[string]interface{}{"type": "string", "description": "SEO Description"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "sitecraft_list_services",
			Description: "Получить каталог услуг/товаров с ценами и описанием.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "sitecraft_create_service",
			Description: "Создать услугу/товар в каталоге. ПРИНЦИП DRAFT-FIRST: создается в неактивном режиме (is_active = false).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title":       map[string]interface{}{"type": "string", "description": "Название услуги/товара"},
					"price":       map[string]interface{}{"type": "number", "description": "Цена в рублях"},
					"description": map[string]interface{}{"type": "string", "description": "Описание услуги"},
					"category":    map[string]interface{}{"type": "string", "description": "Категория"},
				},
				"required": []string{"title", "price"},
			},
		},
		{
			Name:        "sitecraft_generate_seo",
			Description: "Сгенерировать SEO Title, Meta Description и Keywords для переданного текста.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{"type": "string", "description": "Исходный текст статьи или страницы"},
				},
				"required": []string{"text"},
			},
		},
	}
}

// HandleSSE handles incoming SSE stream for MCP clients (Cursor, Claude Desktop, Windsurf, Zed).
func (s *MCPServer) HandleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	sessionID := generateSessionID()
	msgChan := make(chan []byte, 16)
	s.sessions.Store(sessionID, msgChan)
	defer s.sessions.Delete(sessionID)

	// Send endpoint event with session ID
	endpointURL := fmt.Sprintf("/api/v1/mcp/message?sessionId=%s", sessionID)
	_, _ = fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", endpointURL)
	flusher.Flush()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg := <-msgChan:
			_, _ = fmt.Fprintf(w, "event: message\ndata: %s\n\n", string(msg))
			flusher.Flush()
		case <-ticker.C:
			// Ping keep-alive
			_, _ = fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

// HandleMessage handles incoming JSON-RPC 2.0 messages from MCP clients via POST.
func (s *MCPServer) HandleMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeJSONRPCError(w, nil, -32700, "Parse error: unable to read body")
		return
	}

	var req jsonRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.writeJSONRPCError(w, nil, -32700, "Parse error: invalid JSON")
		return
	}

	res := s.processRequest(r.Context(), req)
	if res != nil {
		respBytes, _ := json.Marshal(res)
		// If sessionId is provided and channel exists, we can optionally broadcast
		sessionID := r.URL.Query().Get("sessionId")
		if sessionID != "" {
			if chVal, ok := s.sessions.Load(sessionID); ok {
				ch := chVal.(chan []byte)
				select {
				case ch <- respBytes:
				default:
				}
			}
		}
		_, _ = w.Write(respBytes)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *MCPServer) processRequest(ctx context.Context, req jsonRPCRequest) *jsonRPCResponse {
	switch req.Method {
	case "initialize":
		return &jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{
						"listChanged": false,
					},
				},
				"serverInfo": map[string]interface{}{
					"name":    "sitecraft-mcp",
					"version": "1.0.0",
				},
			},
		}

	case "notifications/initialized":
		return nil

	case "ping":
		return &jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]interface{}{},
		}

	case "tools/list":
		return &jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"tools": s.tools,
			},
		}

	case "tools/call":
		var params mcpToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return &jsonRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &jsonRPCError{
					Code:    -32602,
					Message: "Invalid params for tools/call",
				},
			}
		}

		callRes, err := s.executeTool(ctx, params.Name, params.Arguments)
		if err != nil {
			return &jsonRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: mcpToolCallResult{
					Content: []mcpContentItem{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
					IsError: true,
				},
			}
		}

		return &jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  callRes,
		}

	default:
		return &jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &jsonRPCError{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
		}
	}
}

func (s *MCPServer) executeTool(ctx context.Context, name string, argsRaw json.RawMessage) (mcpToolCallResult, error) {
	switch name {
	case "sitecraft_get_site_info":
		settings, err := s.repo.GetSiteSettings(ctx)
		if err != nil {
			return mcpToolCallResult{}, fmt.Errorf("failed to get site settings: %w", err)
		}
		content, _ := s.repo.GetSiteContent(ctx)
		articles, _ := s.repo.GetArticles(ctx)
		services, _ := s.repo.GetServices(ctx)
		leads, _ := s.repo.GetLeads(ctx)

		info := map[string]interface{}{
			"site_name":      settings.SiteName,
			"tagline":        settings.Tagline,
			"theme":          settings.Theme,
			"phone_masked":   MaskPII(content.Phone1),
			"email_masked":   MaskPII(settings.ContactEmail),
			"articles_count": len(articles),
			"services_count": len(services),
			"leads_count":    len(leads),
			"modules": map[string]bool{
				"blog":     settings.EnableBlog,
				"cart":     settings.EnableCart,
				"chat":     settings.EnableLiveChat,
				"fleet":    settings.EnableFleet,
				"tracking": settings.EnableTracker,
			},
			"compliance": "152-FZ PII Masked, Draft-First Policy Active",
		}

		outBytes, _ := json.MarshalIndent(info, "", "  ")
		return mcpToolCallResult{
			Content: []mcpContentItem{{Type: "text", Text: string(outBytes)}},
		}, nil

	case "sitecraft_list_articles":
		articles, err := s.repo.GetArticles(ctx)
		if err != nil {
			return mcpToolCallResult{}, fmt.Errorf("failed to list articles: %w", err)
		}
		outBytes, _ := json.MarshalIndent(articles, "", "  ")
		return mcpToolCallResult{
			Content: []mcpContentItem{{Type: "text", Text: string(outBytes)}},
		}, nil

	case "sitecraft_get_article":
		var p struct {
			ID   string `json:"id"`
			Slug string `json:"slug"`
		}
		_ = json.Unmarshal(argsRaw, &p)

		if p.ID == "" && p.Slug == "" {
			return mcpToolCallResult{}, fmt.Errorf("either 'id' or 'slug' must be provided")
		}

		var found models.ArticleItem
		var err error
		if p.ID != "" {
			found, err = s.repo.GetArticleByID(ctx, p.ID)
		} else {
			// Find by slug
			all, listErr := s.repo.GetArticles(ctx)
			if listErr != nil {
				return mcpToolCallResult{}, listErr
			}
			matched := false
			for _, a := range all {
				if a.Slug == p.Slug {
					found = a
					matched = true
					break
				}
			}
			if !matched {
				return mcpToolCallResult{}, fmt.Errorf("article with slug %q not found", p.Slug)
			}
		}

		if err != nil {
			return mcpToolCallResult{}, err
		}

		outBytes, _ := json.MarshalIndent(found, "", "  ")
		return mcpToolCallResult{
			Content: []mcpContentItem{{Type: "text", Text: string(outBytes)}},
		}, nil

	case "sitecraft_create_article":
		var p struct {
			Title          string   `json:"title"`
			Content        string   `json:"content"`
			Excerpt        string   `json:"excerpt"`
			Slug           string   `json:"slug"`
			Tags           []string `json:"tags"`
			SEOTitle       string   `json:"seo_title"`
			SEODescription string   `json:"seo_description"`
		}
		if err := json.Unmarshal(argsRaw, &p); err != nil {
			return mcpToolCallResult{}, fmt.Errorf("invalid arguments: %w", err)
		}

		if p.Title == "" || p.Content == "" {
			return mcpToolCallResult{}, fmt.Errorf("title and content are required")
		}

		slug := p.Slug
		if slug == "" {
			slug = generateSlug(p.Title)
		}

		// DRAFT-FIRST GUARANTEE: is_published is strictly forced to false!
		art := models.ArticleItem{
			Title:          p.Title,
			Slug:           slug,
			Excerpt:        p.Excerpt,
			Content:        p.Content,
			Tags:           p.Tags,
			SEOTitle:       p.SEOTitle,
			SEODescription: p.SEODescription,
			AuthorName:     "AI Assistant",
			IsPublished:    false, // STRICTLY DRAFT
			PublishedAt:    time.Now().Format("2006-01-02 15:04"),
		}

		if err := s.repo.SaveArticle(ctx, art); err != nil {
			return mcpToolCallResult{}, fmt.Errorf("failed to save article draft: %w", err)
		}

		res := map[string]interface{}{
			"status":       "created",
			"slug":         slug,
			"is_published": false,
			"draft_policy": "Draft-First enforcement: Article saved strictly as draft (is_published=false) requiring admin moderation.",
		}
		outBytes, _ := json.MarshalIndent(res, "", "  ")
		return mcpToolCallResult{
			Content: []mcpContentItem{{Type: "text", Text: string(outBytes)}},
		}, nil

	case "sitecraft_update_article":
		var p struct {
			ID             string   `json:"id"`
			Title          string   `json:"title"`
			Content        string   `json:"content"`
			Excerpt        string   `json:"excerpt"`
			Tags           []string `json:"tags"`
			SEOTitle       string   `json:"seo_title"`
			SEODescription string   `json:"seo_description"`
		}
		if err := json.Unmarshal(argsRaw, &p); err != nil {
			return mcpToolCallResult{}, fmt.Errorf("invalid arguments: %w", err)
		}

		existing, err := s.repo.GetArticleByID(ctx, p.ID)
		if err != nil {
			return mcpToolCallResult{}, fmt.Errorf("article not found: %w", err)
		}

		if p.Title != "" {
			existing.Title = p.Title
		}
		if p.Content != "" {
			existing.Content = p.Content
		}
		if p.Excerpt != "" {
			existing.Excerpt = p.Excerpt
		}
		if len(p.Tags) > 0 {
			existing.Tags = p.Tags
		}
		if p.SEOTitle != "" {
			existing.SEOTitle = p.SEOTitle
		}
		if p.SEODescription != "" {
			existing.SEODescription = p.SEODescription
		}

		if err := s.repo.SaveArticle(ctx, existing); err != nil {
			return mcpToolCallResult{}, fmt.Errorf("failed to update article: %w", err)
		}

		res := map[string]interface{}{
			"status":       "updated",
			"id":           existing.ID,
			"is_published": existing.IsPublished,
		}
		outBytes, _ := json.MarshalIndent(res, "", "  ")
		return mcpToolCallResult{
			Content: []mcpContentItem{{Type: "text", Text: string(outBytes)}},
		}, nil

	case "sitecraft_list_services":
		services, err := s.repo.GetServices(ctx)
		if err != nil {
			return mcpToolCallResult{}, fmt.Errorf("failed to list services: %w", err)
		}
		outBytes, _ := json.MarshalIndent(services, "", "  ")
		return mcpToolCallResult{
			Content: []mcpContentItem{{Type: "text", Text: string(outBytes)}},
		}, nil

	case "sitecraft_create_service":
		var p struct {
			Title       string      `json:"title"`
			Price       interface{} `json:"price"`
			Description string      `json:"description"`
		}
		if err := json.Unmarshal(argsRaw, &p); err != nil {
			return mcpToolCallResult{}, fmt.Errorf("invalid arguments: %w", err)
		}

		priceStr := fmt.Sprintf("%v", p.Price)
		if p.Price == nil || priceStr == "" || priceStr == "<nil>" {
			priceStr = "0 ₽"
		}

		srv := models.ServiceItem{
			Title:       p.Title,
			Price:       priceStr,
			Description: p.Description,
			Active:      false, // STRICTLY DRAFT
		}

		if err := s.repo.SaveService(ctx, srv); err != nil {
			return mcpToolCallResult{}, fmt.Errorf("failed to save service: %w", err)
		}

		res := map[string]interface{}{
			"status":       "created",
			"title":        p.Title,
			"is_active":    false,
			"draft_policy": "Draft-First enforcement: Service created in inactive/draft mode.",
		}
		outBytes, _ := json.MarshalIndent(res, "", "  ")
		return mcpToolCallResult{
			Content: []mcpContentItem{{Type: "text", Text: string(outBytes)}},
		}, nil

	case "sitecraft_generate_seo":
		var p struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(argsRaw, &p); err != nil {
			return mcpToolCallResult{}, fmt.Errorf("invalid arguments: %w", err)
		}

		if p.Text == "" {
			return mcpToolCallResult{}, fmt.Errorf("text cannot be empty")
		}

		provider, err := s.getProvider(ctx)
		if err == nil && provider != nil {
			meta, err := provider.GenerateSEO(ctx, p.Text)
			if err == nil {
				outBytes, _ := json.MarshalIndent(meta, "", "  ")
				return mcpToolCallResult{
					Content: []mcpContentItem{{Type: "text", Text: string(outBytes)}},
				}, nil
			}
			slog.Warn("AI provider failed for SEO generation, using fallback", "error", err)
		}

		// Fallback heuristic SEO generator if AI provider is not connected
		fallbackMeta := generateFallbackSEO(p.Text)
		outBytes, _ := json.MarshalIndent(fallbackMeta, "", "  ")
		return mcpToolCallResult{
			Content: []mcpContentItem{{Type: "text", Text: string(outBytes)}},
		}, nil

	default:
		return mcpToolCallResult{}, fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *MCPServer) writeJSONRPCError(w http.ResponseWriter, id interface{}, code int, message string) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &jsonRPCError{
			Code:    code,
			Message: message,
		},
	}
	respBytes, _ := json.Marshal(resp)
	_, _ = w.Write(respBytes)
}

func generateSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "/", "-")
	slug = strings.ReplaceAll(slug, "?", "")
	slug = strings.ReplaceAll(slug, "!", "")
	slug = strings.Trim(slug, "-")
	if len(slug) > 80 {
		slug = slug[:80]
	}
	if slug == "" {
		slug = fmt.Sprintf("post-%d", time.Now().Unix())
	}
	return slug
}

func generateFallbackSEO(text string) SEOMeta {
	lines := strings.Split(text, "\n")
	firstLine := ""
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			firstLine = l
			break
		}
	}
	if len(firstLine) > 60 {
		firstLine = firstLine[:57] + "..."
	}

	clean := strings.ReplaceAll(text, "\n", " ")
	clean = strings.TrimSpace(clean)
	if len(clean) > 150 {
		clean = clean[:147] + "..."
	}

	return SEOMeta{
		Title:       firstLine,
		Description: clean,
		Keywords:    []string{"sitecraft", "бизнес", "услуги", "статья"},
	}
}
