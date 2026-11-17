package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAICompatibleClient interacts with any OpenAI-compatible API endpoint
// (OpenAI, DeepSeek, anymodel.org, OpenRouter, Groq, Ollama, vLLM, LM Studio).
type OpenAICompatibleClient struct {
	baseURL    string
	apiKey     string
	model      string
	provider   string
	httpClient *http.Client
}

// NewOpenAICompatibleClient creates a client for OpenAI-compatible REST endpoints.
func NewOpenAICompatibleClient(baseURL, apiKey, model, provider string, client *http.Client) *OpenAICompatibleClient {
	baseURL = strings.TrimRight(baseURL, "/")
	if client == nil {
		client = http.DefaultClient
	}
	return &OpenAICompatibleClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		model:      model,
		provider:   provider,
		httpClient: client,
	}
}

func (c *OpenAICompatibleClient) Name() string {
	if c.provider != "" {
		return c.provider
	}
	return "openai-compatible"
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []openAIChatMessage `json:"messages"`
	Temperature float64             `json:"temperature,omitempty"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (c *OpenAICompatibleClient) GenerateContent(ctx context.Context, prompt string, opts GenerateOpts) (string, error) {
	model := opts.Model
	if model == "" {
		model = c.model
	}

	messages := make([]openAIChatMessage, 0, 2)
	if opts.SystemPrompt != "" {
		messages = append(messages, openAIChatMessage{Role: "system", Content: opts.SystemPrompt})
	}
	messages = append(messages, openAIChatMessage{Role: "user", Content: prompt})

	reqBody := openAIChatRequest{
		Model:       model,
		Messages:    messages,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode request: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI provider returned HTTP %d: %s", resp.StatusCode, string(respData))
	}

	var chatResp openAIChatResponse
	if err := json.Unmarshal(respData, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse AI response: %w", err)
	}

	if chatResp.Error != nil && chatResp.Error.Message != "" {
		return "", fmt.Errorf("AI error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("AI returned no completions")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}

func (c *OpenAICompatibleClient) GenerateSEO(ctx context.Context, text string) (SEOMeta, error) {
	systemPrompt := `Ты эксперт по поисковой оптимизации (SEO) для сайтов на русском языке.
Проанализируй предоставленный текст и сформируй привлекательные и эффективные метатеги.
Ответ верни СТРОГО в формате JSON без какого-либо дополнительного текста или markdown-разметки:
{
  "title": "SEO-заголовок (до 60-70 символов с главным ключом)",
  "description": "Привлекательное описание (до 150-160 символов с призывом к действию)",
  "keywords": ["ключ 1", "ключ 2", "ключ 3", "ключ 4", "ключ 5"]
}`

	raw, err := c.GenerateContent(ctx, text, GenerateOpts{
		SystemPrompt: systemPrompt,
		Temperature:  0.3,
		MaxTokens:    500,
	})
	if err != nil {
		return SEOMeta{}, err
	}

	return ParseSEOMeta(raw)
}
