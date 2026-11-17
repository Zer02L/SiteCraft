package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SEOMeta represents extracted/generated SEO metadata.
type SEOMeta struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

// GenerateOpts configures LLM inference parameters.
type GenerateOpts struct {
	SystemPrompt string  `json:"system_prompt,omitempty"`
	Temperature  float64 `json:"temperature,omitempty"`
	MaxTokens    int     `json:"max_tokens,omitempty"`
	Model        string  `json:"model,omitempty"`
}

// LLMProvider is the universal facade for text and SEO generation.
type LLMProvider interface {
	GenerateContent(ctx context.Context, prompt string, opts GenerateOpts) (string, error)
	GenerateSEO(ctx context.Context, text string) (SEOMeta, error)
	Name() string
}

// Config specifies credentials and endpoints to construct an LLMProvider.
type Config struct {
	Provider    string `json:"provider"`   // openai, anthropic, gemini, deepseek, anymodel, yandex, gigachat, custom
	BaseURL     string `json:"base_url"`   // custom or preset endpoint URL
	APIKey      string `json:"api_key"`    // decrypted plain API key
	ModelName   string `json:"model_name"` // model identifier
	ProxyURL    string `json:"proxy_url"`  // optional HTTP/SOCKS5 proxy
	FolderID    string `json:"folder_id"`  // required for Yandex Cloud
}

// ProviderPreset defines configuration defaults for each supported AI service.
type ProviderPreset struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	DefaultBaseURL string   `json:"default_base_url"`
	DefaultModel   string   `json:"default_model"`
	AvailableModels []string `json:"available_models"`
	KeyPlaceholder string   `json:"key_placeholder"`
	Description    string   `json:"description"`
	NeedsFolderID  bool     `json:"needs_folder_id"`
	Compliant152FZ bool     `json:"compliant_152fz"`
}

var Presets = []ProviderPreset{
	{
		ID:             "openai",
		Name:           "OpenAI (GPT-4o)",
		DefaultBaseURL: "https://api.openai.com/v1",
		DefaultModel:   "gpt-4o-mini",
		AvailableModels: []string{"gpt-4o-mini", "gpt-4o", "gpt-4-turbo", "o3-mini"},
		KeyPlaceholder: "sk-proj-...",
		Description:    "Флагманские модели от OpenAI (требуется иностранный аккаунт или прокси)",
		NeedsFolderID:  false,
		Compliant152FZ: false,
	},
	{
		ID:             "anthropic",
		Name:           "Anthropic Claude",
		DefaultBaseURL: "https://api.anthropic.com/v1",
		DefaultModel:   "claude-3-5-sonnet-latest",
		AvailableModels: []string{"claude-3-5-sonnet-latest", "claude-3-5-haiku-latest"},
		KeyPlaceholder: "sk-ant-...",
		Description:    "Claude 3.5 Sonnet / Haiku — непревзойденное качество работы с текстом",
		NeedsFolderID:  false,
		Compliant152FZ: false,
	},
	{
		ID:             "gemini",
		Name:           "Google Gemini",
		DefaultBaseURL: "https://generativelanguage.googleapis.com/v1beta",
		DefaultModel:   "gemini-2.0-flash",
		AvailableModels: []string{"gemini-2.0-flash", "gemini-1.5-pro", "gemini-1.5-flash"},
		KeyPlaceholder: "AIzaSy...",
		Description:    "Google Gemini 2.0 Flash — сверхбыстрая и экономичная генерация",
		NeedsFolderID:  false,
		Compliant152FZ: false,
	},
	{
		ID:             "deepseek",
		Name:           "DeepSeek",
		DefaultBaseURL: "https://api.deepseek.com/v1",
		DefaultModel:   "deepseek-chat",
		AvailableModels: []string{"deepseek-chat", "deepseek-reasoner"},
		KeyPlaceholder: "sk-...",
		Description:    "DeepSeek V3 и R1 — мощная открытая модель по доступной стоимости",
		NeedsFolderID:  false,
		Compliant152FZ: false,
	},
	{
		ID:             "anymodel",
		Name:           "Anymodel.org (Агрегатор)",
		DefaultBaseURL: "https://api.anymodel.org/v1",
		DefaultModel:   "llama-3.3-70b",
		AvailableModels: []string{"llama-3.3-70b", "qwen-2.5-72b", "gpt-4o-mini", "claude-3-5-haiku"},
		KeyPlaceholder: "any-...",
		Description:    "Единый шлюз без блокировок, оплата российскими картами",
		NeedsFolderID:  false,
		Compliant152FZ: true,
	},
	{
		ID:             "yandex",
		Name:           "Яндекс GPT (152-ФЗ РФ)",
		DefaultBaseURL: "https://llm.api.cloud.yandex.net/foundationModels/v1",
		DefaultModel:   "yandexgpt-lite",
		AvailableModels: []string{"yandexgpt-lite", "yandexgpt"},
		KeyPlaceholder: "AQVN...",
		Description:    "Отечественная модель Яндекс Cloud. Полный комплаенс 152-ФЗ, без VPN и блокировок",
		NeedsFolderID:  true,
		Compliant152FZ: true,
	},
	{
		ID:             "gigachat",
		Name:           "Сбер GigaChat (152-ФЗ РФ)",
		DefaultBaseURL: "https://gigachat.devices.sberbank.ru/api/v1",
		DefaultModel:   "GigaChat",
		AvailableModels: []string{"GigaChat", "GigaChat-Plus", "GigaChat-Pro"},
		KeyPlaceholder: "Client Secret / Bearer ...",
		Description:    "Нейросеть от Сбера для корпоративных сайтов и госорганов",
		NeedsFolderID:  false,
		Compliant152FZ: true,
	},
	{
		ID:             "custom",
		Name:           "Кастомный / OpenAI-Compatible (OpenRouter, Groq, Ollama, vLLM)",
		DefaultBaseURL: "https://openrouter.ai/api/v1",
		DefaultModel:   "meta-llama/llama-3.3-70b-instruct",
		AvailableModels: []string{"meta-llama/llama-3.3-70b-instruct", "mistralai/mistral-large"},
		KeyPlaceholder: "sk-...",
		Description:    "Подключение любых совместимых шлюзов, локальных селф-хост моделей или агрегаторов",
		NeedsFolderID:  false,
		Compliant152FZ: false,
	},
}

// GetPreset returns the preset metadata by identifier.
func GetPreset(id string) ProviderPreset {
	for _, p := range Presets {
		if strings.EqualFold(p.ID, id) {
			return p
		}
	}
	return Presets[0] // fallback to OpenAI
}

// buildHTTPClient creates an http.Client configured with optional proxy and timeout.
func buildHTTPClient(proxyURL string, timeout time.Duration) *http.Client {
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	if proxyURL != "" {
		if parsed, err := url.Parse(proxyURL); err == nil {
			transport.Proxy = http.ProxyURL(parsed)
		}
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// NewProvider instantiates the appropriate LLMProvider implementation according to config.
func NewProvider(cfg Config) (LLMProvider, error) {
	if cfg.Provider == "" {
		cfg.Provider = "openai"
	}

	preset := GetPreset(cfg.Provider)
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = preset.DefaultBaseURL
	}
	model := cfg.ModelName
	if model == "" {
		model = preset.DefaultModel
	}

	httpClient := buildHTTPClient(cfg.ProxyURL, 90*time.Second)

	switch strings.ToLower(cfg.Provider) {
	case "anthropic":
		return NewAnthropicClient(baseURL, cfg.APIKey, model, httpClient), nil
	case "gemini":
		return NewGeminiClient(baseURL, cfg.APIKey, model, httpClient), nil
	case "yandex":
		return NewYandexClient(baseURL, cfg.APIKey, cfg.FolderID, model, httpClient), nil
	case "gigachat":
		return NewGigaChatClient(baseURL, cfg.APIKey, model, httpClient), nil
	case "openai", "deepseek", "anymodel", "custom":
		fallthrough
	default:
		// Universal OpenAI-compatible client works with OpenAI, DeepSeek, anymodel.org, OpenRouter, Groq, Ollama, vLLM
		return NewOpenAICompatibleClient(baseURL, cfg.APIKey, model, cfg.Provider, httpClient), nil
	}
}

// cleanJSONString extracts and sanitizes JSON payload from LLM responses containing code fences.
func cleanJSONString(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		// remove starting ```json or ```
		lines := strings.Split(raw, "\n")
		if len(lines) > 2 {
			end := len(lines) - 1
			if strings.HasPrefix(strings.TrimSpace(lines[end]), "```") {
				raw = strings.Join(lines[1:end], "\n")
			} else {
				raw = strings.Join(lines[1:], "\n")
			}
		}
	}
	return strings.TrimSpace(raw)
}

// ParseSEOMeta extracts SEOMeta from raw JSON text or provides a fallback structure.
func ParseSEOMeta(raw string) (SEOMeta, error) {
	cleaned := cleanJSONString(raw)

	var meta SEOMeta
	if err := json.Unmarshal([]byte(cleaned), &meta); err == nil && (meta.Title != "" || meta.Description != "") {
		return meta, nil
	}

	// Fallback heuristic if LLM didn't return strict JSON
	lines := strings.Split(cleaned, "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		lower := strings.ToLower(l)
		if strings.HasPrefix(lower, "title:") || strings.HasPrefix(lower, "заголовок:") {
			parts := strings.SplitN(l, ":", 2)
			if len(parts) == 2 {
				meta.Title = strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			}
		} else if strings.HasPrefix(lower, "description:") || strings.HasPrefix(lower, "описание:") {
			parts := strings.SplitN(l, ":", 2)
			if len(parts) == 2 {
				meta.Description = strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			}
		}
	}

	if meta.Title != "" || meta.Description != "" {
		return meta, nil
	}

	return SEOMeta{}, fmt.Errorf("unable to parse SEO metadata from response: %w", errors.New("invalid JSON"))
}
