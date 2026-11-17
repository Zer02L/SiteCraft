package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// GeminiClient interacts with Google Gemini REST API.
type GeminiClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewGeminiClient creates a client for Google Gemini API.
func NewGeminiClient(baseURL, apiKey, model string, client *http.Client) *GeminiClient {
	baseURL = strings.TrimRight(baseURL, "/")
	if client == nil {
		client = http.DefaultClient
	}
	return &GeminiClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		model:      model,
		httpClient: client,
	}
}

func (c *GeminiClient) Name() string {
	return "gemini"
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type geminiRequest struct {
	Contents          []geminiContent `json:"contents"`
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
	GenerationConfig  *struct {
		Temperature     float64 `json:"temperature,omitempty"`
		MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	} `json:"generationConfig,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

func (c *GeminiClient) GenerateContent(ctx context.Context, prompt string, opts GenerateOpts) (string, error) {
	model := opts.Model
	if model == "" {
		model = c.model
	}
	if !strings.HasPrefix(model, "models/") && !strings.Contains(model, "/") {
		model = "models/" + model
	}

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
	}

	if opts.SystemPrompt != "" {
		reqBody.SystemInstruction = &geminiContent{
			Parts: []geminiPart{{Text: opts.SystemPrompt}},
		}
	}

	if opts.Temperature > 0 || opts.MaxTokens > 0 {
		reqBody.GenerationConfig = &struct {
			Temperature     float64 `json:"temperature,omitempty"`
			MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
		}{
			Temperature:     opts.Temperature,
			MaxOutputTokens: opts.MaxTokens,
		}
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode Gemini request: %w", err)
	}

	reqURL := fmt.Sprintf("%s/%s:generateContent?key=%s", c.baseURL, model, url.QueryEscape(c.apiKey))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request to Gemini failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini API returned HTTP %d: %s", resp.StatusCode, string(respData))
	}

	var gResp geminiResponse
	if err := json.Unmarshal(respData, &gResp); err != nil {
		return "", fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	if gResp.Error != nil && gResp.Error.Message != "" {
		return "", fmt.Errorf("gemini error: %s", gResp.Error.Message)
	}

	if len(gResp.Candidates) > 0 && len(gResp.Candidates[0].Content.Parts) > 0 {
		return strings.TrimSpace(gResp.Candidates[0].Content.Parts[0].Text), nil
	}

	return "", fmt.Errorf("no content returned from Gemini")
}

func (c *GeminiClient) GenerateSEO(ctx context.Context, text string) (SEOMeta, error) {
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
