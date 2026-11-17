package ai

import (
	"context"
	"net/http"
	"strings"
)

// GigaChatClient interacts with Sber GigaChat API (which uses an OpenAI-compatible /chat/completions structure).
type GigaChatClient struct {
	openAIClient *OpenAICompatibleClient
}

// NewGigaChatClient creates a client for Sber GigaChat.
func NewGigaChatClient(baseURL, apiKey, model string, client *http.Client) *GigaChatClient {
	if baseURL == "" {
		baseURL = "https://gigachat.devices.sberbank.ru/api/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if model == "" {
		model = "GigaChat"
	}
	return &GigaChatClient{
		openAIClient: NewOpenAICompatibleClient(baseURL, apiKey, model, "gigachat", client),
	}
}

func (c *GigaChatClient) Name() string {
	return "gigachat"
}

func (c *GigaChatClient) GenerateContent(ctx context.Context, prompt string, opts GenerateOpts) (string, error) {
	return c.openAIClient.GenerateContent(ctx, prompt, opts)
}

func (c *GigaChatClient) GenerateSEO(ctx context.Context, text string) (SEOMeta, error) {
	return c.openAIClient.GenerateSEO(ctx, text)
}
