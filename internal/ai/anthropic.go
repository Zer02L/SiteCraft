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

// AnthropicClient interacts with Anthropic Claude Messages API.
type AnthropicClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewAnthropicClient creates a client for Anthropic Claude API.
func NewAnthropicClient(baseURL, apiKey, model string, client *http.Client) *AnthropicClient {
	baseURL = strings.TrimRight(baseURL, "/")
	if client == nil {
		client = http.DefaultClient
	}
	return &AnthropicClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		model:      model,
		httpClient: client,
	}
}

func (c *AnthropicClient) Name() string {
	return "anthropic"
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature,omitempty"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *AnthropicClient) GenerateContent(ctx context.Context, prompt string, opts GenerateOpts) (string, error) {
	model := opts.Model
	if model == "" {
		model = c.model
	}

	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 2000
	}

	reqBody := anthropicRequest{
		Model:       model,
		Messages:    []anthropicMessage{{Role: "user", Content: prompt}},
		System:      opts.SystemPrompt,
		MaxTokens:   maxTokens,
		Temperature: opts.Temperature,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode request: %w", err)
	}

	url := c.baseURL + "/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

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
		return "", fmt.Errorf("anthropic API returned HTTP %d: %s", resp.StatusCode, string(respData))
	}

	var aResp anthropicResponse
	if err := json.Unmarshal(respData, &aResp); err != nil {
		return "", fmt.Errorf("failed to parse Anthropic response: %w", err)
	}

	if aResp.Error != nil && aResp.Error.Message != "" {
		return "", fmt.Errorf("anthropic error: %s", aResp.Error.Message)
	}

	for _, block := range aResp.Content {
		if block.Type == "text" {
			return strings.TrimSpace(block.Text), nil
		}
	}

	return "", fmt.Errorf("no text content in Anthropic response")
}

func (c *AnthropicClient) GenerateSEO(ctx context.Context, text string) (SEOMeta, error) {
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
