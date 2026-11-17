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

// YandexClient interacts with Yandex Cloud Foundation Models API (YandexGPT).
type YandexClient struct {
	baseURL    string
	apiKey     string
	folderID   string
	model      string
	httpClient *http.Client
}

// NewYandexClient creates a client for Yandex Cloud Foundation Models API.
func NewYandexClient(baseURL, apiKey, folderID, model string, client *http.Client) *YandexClient {
	baseURL = strings.TrimRight(baseURL, "/")
	if client == nil {
		client = http.DefaultClient
	}
	return &YandexClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		folderID:   folderID,
		model:      model,
		httpClient: client,
	}
}

func (c *YandexClient) Name() string {
	return "yandex"
}

type yandexMessage struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type yandexRequest struct {
	ModelURI          string          `json:"modelUri"`
	CompletionOptions struct {
		Stream      bool    `json:"stream"`
		Temperature float64 `json:"temperature,omitempty"`
		MaxTokens   string  `json:"maxTokens,omitempty"`
	} `json:"completionOptions"`
	Messages []yandexMessage `json:"messages"`
}

type yandexResponse struct {
	Result struct {
		Alternatives []struct {
			Message struct {
				Role string `json:"role"`
				Text string `json:"text"`
			} `json:"message"`
		} `json:"alternatives"`
	} `json:"result"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *YandexClient) GenerateContent(ctx context.Context, prompt string, opts GenerateOpts) (string, error) {
	model := opts.Model
	if model == "" {
		model = c.model
	}

	modelURI := fmt.Sprintf("gpt://%s/%s/latest", c.folderID, model)
	if strings.HasPrefix(model, "gpt://") {
		modelURI = model
	}

	messages := make([]yandexMessage, 0, 2)
	if opts.SystemPrompt != "" {
		messages = append(messages, yandexMessage{Role: "system", Text: opts.SystemPrompt})
	}
	messages = append(messages, yandexMessage{Role: "user", Text: prompt})

	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 2000
	}

	reqBody := yandexRequest{
		ModelURI: modelURI,
		Messages: messages,
	}
	reqBody.CompletionOptions.Stream = false
	reqBody.CompletionOptions.Temperature = opts.Temperature
	reqBody.CompletionOptions.MaxTokens = fmt.Sprintf("%d", maxTokens)

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode Yandex request: %w", err)
	}

	url := c.baseURL + "/completion"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create Yandex request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if strings.HasPrefix(c.apiKey, "t1.") || strings.HasPrefix(c.apiKey, "AQVN") {
		req.Header.Set("Authorization", "Api-Key "+c.apiKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	if c.folderID != "" {
		req.Header.Set("x-folder-id", c.folderID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request to Yandex failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Yandex response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("yandex API returned HTTP %d: %s", resp.StatusCode, string(respData))
	}

	var yResp yandexResponse
	if err := json.Unmarshal(respData, &yResp); err != nil {
		return "", fmt.Errorf("failed to parse Yandex response: %w", err)
	}

	if yResp.Error != nil && yResp.Error.Message != "" {
		return "", fmt.Errorf("yandex error: %s", yResp.Error.Message)
	}

	if len(yResp.Result.Alternatives) > 0 {
		return strings.TrimSpace(yResp.Result.Alternatives[0].Message.Text), nil
	}

	return "", fmt.Errorf("no alternatives returned from YandexGPT")
}

func (c *YandexClient) GenerateSEO(ctx context.Context, text string) (SEOMeta, error) {
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
