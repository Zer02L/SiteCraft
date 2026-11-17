package ai

import (
	"encoding/json"
	"net/http"
	"strings"
)

// GenerateOpenAPISpec returns OpenAPI 3.1.0 schema for OpenAI Custom GPTs, Gemini Extensions, and external AI agents.
func GenerateOpenAPISpec(host string) map[string]interface{} {
	if host == "" {
		host = "https://example.com"
	}

	return map[string]interface{}{
		"openapi": "3.1.0",
		"info": map[string]interface{}{
			"title":       "SiteCraft Universal AI Agent API",
			"description": "API для подключения автономных ИИ-агентов, OpenAI Custom GPTs и Claude к движку SiteCraft. Поддерживает принцип Draft-First и комплаенс 152-ФЗ.",
			"version":     "1.0.0",
		},
		"servers": []map[string]interface{}{
			{
				"url":         host,
				"description": "SiteCraft Production Server",
			},
		},
		"paths": map[string]interface{}{
			"/api/v1/agent/info": map[string]interface{}{
				"get": map[string]interface{}{
					"operationId": "getSiteInfo",
					"summary":     "Получить статус сайта, настройки и агрегированные метрики",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Успешный ответ (ПДн замаскированы по 152-ФЗ)",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"site_name":      map[string]interface{}{"type": "string"},
											"phone_masked":   map[string]interface{}{"type": "string"},
											"email_masked":   map[string]interface{}{"type": "string"},
											"articles_count": map[string]interface{}{"type": "integer"},
											"services_count": map[string]interface{}{"type": "integer"},
										},
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/agent/articles": map[string]interface{}{
				"get": map[string]interface{}{
					"operationId": "listArticles",
					"summary":     "Получить список всех опубликованных и черновых статей",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Список статей",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "array",
										"items": map[string]interface{}{
											"type": "object",
											"properties": map[string]interface{}{
												"id":           map[string]interface{}{"type": "string"},
												"title":        map[string]interface{}{"type": "string"},
												"slug":         map[string]interface{}{"type": "string"},
												"is_published": map[string]interface{}{"type": "boolean"},
											},
										},
									},
								},
							},
						},
					},
				},
				"post": map[string]interface{}{
					"operationId": "createArticleDraft",
					"summary":     "Создать черновик новой статьи в блоге (Принцип Draft-First: всегда сохраняется как is_published = false)",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"required": []string{"title", "content"},
									"properties": map[string]interface{}{
										"title":           map[string]interface{}{"type": "string", "description": "Заголовок"},
										"content":         map[string]interface{}{"type": "string", "description": "Текст статьи"},
										"excerpt":         map[string]interface{}{"type": "string", "description": "Краткое описание"},
										"slug":            map[string]interface{}{"type": "string", "description": "URL slug"},
										"tags":            map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
										"seo_title":       map[string]interface{}{"type": "string"},
										"seo_description": map[string]interface{}{"type": "string"},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Черновик статьи создан",
						},
					},
				},
			},
			"/api/v1/agent/seo/generate": map[string]interface{}{
				"post": map[string]interface{}{
					"operationId": "generateSEO",
					"summary":     "Сгенерировать SEO-метатеги (Title, Description, Keywords) по тексту",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"required": []string{"text"},
									"properties": map[string]interface{}{
										"text": map[string]interface{}{"type": "string"},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Сгенерированные метатеги",
						},
					},
				},
			},
		},
	}
}

// HandleOpenAPI returns the JSON specification for OpenAPI 3.1.
func HandleOpenAPI(w http.ResponseWriter, r *http.Request) {
	scheme := "https"
	if r.TLS == nil && !strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "http"
	}
	host := scheme + "://" + r.Host

	spec := GenerateOpenAPISpec(host)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(spec)
}
