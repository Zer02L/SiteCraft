package server

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"sitecraft/internal/ai"
	"sitecraft/internal/models"
	"sitecraft/ui/pages"

	"github.com/google/uuid"
)

// HandleAdminTabAI renders the AI Engine configuration and copilot tab.
func (s *Server) HandleAdminTabAI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, err := s.Repo.GetSiteSettings(ctx)
	if err != nil {
		slog.Error("failed to load site settings", "error", err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ClientAITab(settings, ai.Presets, "", false).Render(ctx, w)
}

// HandleAdminSaveAISettings updates AI provider settings and securely encrypts API key.
func (s *Server) HandleAdminSaveAISettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()

	settings, err := s.Repo.GetSiteSettings(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pages.AIAlert("Ошибка загрузки настроек: "+err.Error(), false).Render(ctx, w)
		return
	}

	provider := strings.TrimSpace(r.FormValue("ai_provider"))
	baseURL := strings.TrimSpace(r.FormValue("ai_base_url"))
	modelName := strings.TrimSpace(r.FormValue("ai_model_name"))
	proxyURL := strings.TrimSpace(r.FormValue("ai_proxy_url"))
	folderID := strings.TrimSpace(r.FormValue("ai_folder_id"))
	rawKey := strings.TrimSpace(r.FormValue("ai_api_key"))

	if provider != "" {
		settings.AIProvider = provider
	}
	settings.AIBaseURL = baseURL
	settings.AIModelName = modelName
	settings.AIProxyURL = proxyURL
	settings.AIFolderID = folderID

	if rawKey != "" {
		encryptedKey, err := ai.Encrypt(rawKey)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = pages.AIAlert("Ошибка шифрования ключа: "+err.Error(), false).Render(ctx, w)
			return
		}
		settings.AIAPIKeyEncrypted = encryptedKey
	}

	if err := s.Repo.UpsertSiteSettings(ctx, settings); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pages.AIAlert("Ошибка сохранения: "+err.Error(), false).Render(ctx, w)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "refreshChecklist")
	_ = pages.AIAlert("Настройки ИИ успешно сохранены и зашифрованы (AES-256-GCM)", true).Render(ctx, w)
}

// HandleAdminTestAIConnection performs a ping test to the configured AI provider.
func (s *Server) HandleAdminTestAIConnection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()

	providerName := strings.TrimSpace(r.FormValue("ai_provider"))
	baseURL := strings.TrimSpace(r.FormValue("ai_base_url"))
	modelName := strings.TrimSpace(r.FormValue("ai_model_name"))
	proxyURL := strings.TrimSpace(r.FormValue("ai_proxy_url"))
	folderID := strings.TrimSpace(r.FormValue("ai_folder_id"))
	apiKey := strings.TrimSpace(r.FormValue("ai_api_key"))

	// If no key in form, try saved encrypted key
	if apiKey == "" {
		settings, _ := s.Repo.GetSiteSettings(ctx)
		if settings.AIAPIKeyEncrypted != "" {
			if dec, err := ai.Decrypt(settings.AIAPIKeyEncrypted); err == nil {
				apiKey = dec
			}
		}
		if providerName == "" {
			providerName = settings.AIProvider
		}
		if baseURL == "" {
			baseURL = settings.AIBaseURL
		}
		if modelName == "" {
			modelName = settings.AIModelName
		}
		if proxyURL == "" {
			proxyURL = settings.AIProxyURL
		}
		if folderID == "" {
			folderID = settings.AIFolderID
		}
	}

	if apiKey == "" && providerName != "custom" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pages.AIAlert("Укажите API ключ для проверки подключения", false).Render(ctx, w)
		return
	}

	provider, err := ai.NewProvider(ai.Config{
		Provider:  providerName,
		BaseURL:   baseURL,
		APIKey:    apiKey,
		ModelName: modelName,
		ProxyURL:  proxyURL,
		FolderID:  folderID,
	})
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pages.AIAlert("Ошибка инициализации провайдера: "+err.Error(), false).Render(ctx, w)
		return
	}

	testCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	resp, err := provider.GenerateContent(testCtx, "Ответь строго одной фразой: 'SiteCraft AI готов к работе.'", ai.GenerateOpts{
		MaxTokens: 25,
	})
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pages.AIAlert("Подключение не удалось: "+err.Error(), false).Render(ctx, w)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.AIAlert("Связь установлена успешно! Ответ: "+resp, true).Render(ctx, w)
}

// HandleAdminAIGenerateArticle creates a draft blog post using AI (Draft-First principle).
func (s *Server) HandleAdminAIGenerateArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()

	topic := strings.TrimSpace(r.FormValue("topic"))
	tone := strings.TrimSpace(r.FormValue("tone"))
	keywords := strings.TrimSpace(r.FormValue("keywords"))

	if topic == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pages.AIAlert("Укажите тему статьи", false).Render(ctx, w)
		return
	}

	provider, _ := s.GetAIProvider(ctx)

	var generatedContent string
	var err error
	if provider != nil {
		prompt := fmt.Sprintf(`Напиши подробную, качественную и структурированную статью для корпоративного блога на русском языке.
Тема: %s
Тон: %s
Ключевые слова для интеграции: %s

Используй чистую HTML разметку: <h2>, <h3>, <p>, <ul>, <li>, <strong>, <blockquote>. Не используй <html>, <head>, <body> или <h1>.`, topic, tone, keywords)

		generatedContent, err = provider.GenerateContent(ctx, prompt, ai.GenerateOpts{
			SystemPrompt: "Ты профессиональный копирайтер и эксперт по контент-маркетингу для малого и среднего бизнеса.",
			Temperature:  0.7,
			MaxTokens:    2500,
		})
	}

	if err != nil || generatedContent == "" {
		// Fallback clean template if external provider is not reachable
		generatedContent = fmt.Sprintf(`<h2>Введение в тему: %s</h2>
<p>В современных условиях грамотный подход к данному вопросу позволяет оптимизировать ресурсы и получить прогнозируемый результат.</p>
<h3>Ключевые аспекты</h3>
<ul>
  <li>Анализ текущей ситуации и выявление узких мест</li>
  <li>Внедрение лучших практик и проверенных инструментов</li>
  <li>Контроль результатов на каждом этапе</li>
</ul>
<p>Своевременное планирование снижает сопутствующие риски на 40%%.</p>`, html.EscapeString(topic))
	}

	slug := strings.ToLower(topic)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "?", "")
	slug = strings.ReplaceAll(slug, "!", "")
	if len(slug) > 60 {
		slug = slug[:60]
	}
	slug = fmt.Sprintf("%s-%d", strings.Trim(slug, "-"), time.Now().Unix()%10000)

	excerpt := "Экспертный материал по теме: " + topic
	if len(excerpt) > 160 {
		excerpt = excerpt[:157] + "..."
	}

	var tagsList []string
	if keywords != "" {
		for _, k := range strings.Split(keywords, ",") {
			k = strings.TrimSpace(k)
			if k != "" {
				tagsList = append(tagsList, k)
			}
		}
	} else {
		tagsList = []string{"AI", "Бизнес", "Практика"}
	}

	// CRITICAL DRAFT-FIRST GUARANTEE: is_published is strictly false!
	art := models.ArticleItem{
		ID:             uuid.New().String(),
		Title:          topic,
		Slug:           slug,
		Excerpt:        excerpt,
		Content:        generatedContent,
		Tags:           tagsList,
		SEOTitle:       topic,
		SEODescription: excerpt,
		AuthorName:     "SiteCraft AI Copilot",
		IsPublished:    false, // STRICTLY DRAFT
		PublishedAt:    time.Now().Format("2006-01-02 15:04"),
	}

	if err := s.Repo.SaveArticle(ctx, art); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pages.AIAlert("Ошибка сохранения черновика: "+err.Error(), false).Render(ctx, w)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "refreshChecklist")
	_ = pages.AICopilotArticleResult(art.Title, art.Slug, art.Excerpt).Render(ctx, w)
}

// HandleAdminAIGenerateSEO extracts and suggests SEO metadata for given content.
func (s *Server) HandleAdminAIGenerateSEO(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()

	text := strings.TrimSpace(r.FormValue("text"))
	if text == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pages.AIAlert("Введите текст для анализа", false).Render(ctx, w)
		return
	}

	provider, _ := s.GetAIProvider(ctx)
	var meta ai.SEOMeta
	var err error

	if provider != nil {
		meta, err = provider.GenerateSEO(ctx, text)
	}

	if err != nil || (meta.Title == "" && meta.Description == "") {
		// Heuristic fallback
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
		if len(clean) > 155 {
			clean = clean[:152] + "..."
		}
		meta = ai.SEOMeta{
			Title:       firstLine,
			Description: clean,
			Keywords:    []string{"услуги", "онлайн", "качество", "гарантия"},
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.AICopilotSEOResult(meta).Render(ctx, w)
}
