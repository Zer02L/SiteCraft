package handlers

import (
	"fmt"
	"html"
	"net/http"
	"strings"

	"sitecraft/internal/data"
	"sitecraft/internal/i18n"
	"sitecraft/internal/models"
	"sitecraft/ui/components"
	"sitecraft/ui/pages"
	"sitecraft/ui/sections"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// WebHandler handles public web pages and fragment components.
type WebHandler struct {
	app *pocketbase.PocketBase
}

// NewWebHandler creates a new instance of WebHandler.
func NewWebHandler(app *pocketbase.PocketBase) *WebHandler {
	return &WebHandler{app: app}
}

// HandleHome renders the main landing page via Templ SSR with cached data.
func (h *WebHandler) HandleHome(e *core.RequestEvent) error {
	settings := data.GetSiteSettings(h.app)
	content := data.GetSiteContent(h.app)
	services := data.GetServices(h.app)
	bookedSlots := data.GetBookedSlots(h.app)
	blocks := data.GetPageBlocks(h.app)

	// Автоматическая локализация контента и сервисов при переключении на EN
	ctx := e.Request.Context()
	content = i18n.LocalizeContent(ctx, content)
	services = i18n.LocalizeServices(ctx, services)

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.Home(settings, content, services, bookedSlots, blocks).Render(ctx, e.Response)
}

// HandleMapFrame serves lazy-loaded interactive map iframe via HTMX.
func (h *WebHandler) HandleMapFrame(e *core.RequestEvent) error {
	content := data.GetSiteContent(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return sections.MapFrame(content.MapIframe, "Интерактивная карта — "+content.HeroTitle).Render(e.Request.Context(), e.Response)
}

// HandleSlots returns booked time slots as JSON.
func (h *WebHandler) HandleSlots(e *core.RequestEvent) error {
	bookedSlots := data.GetBookedSlots(h.app)
	return e.JSON(http.StatusOK, bookedSlots)
}

// HandlePaymentModal renders the payment modal popup dialog.
func (h *WebHandler) HandlePaymentModal(e *core.RequestEvent) error {
	settings := data.GetSiteSettings(h.app)
	orderCode := strings.TrimSpace(e.Request.URL.Query().Get("order_code"))
	serviceID := strings.TrimSpace(e.Request.URL.Query().Get("service_id"))

	var selectedService models.ServiceItem
	var fixedAmount float64

	// Если передан номер заказа из трекера — фиксируем сумму и название строго из базы
	if orderCode != "" {
		order, err := data.GetServiceOrderByCode(h.app, orderCode)
		if err == nil && order != nil {
			selectedService = models.ServiceItem{
				ID:          order.ID,
				Title:       order.ItemTitle,
				Description: order.Notes,
				Price:       fmt.Sprintf("%.0f ₽", order.TotalAmount),
			}
			fixedAmount = order.TotalAmount
		}
	}

	if fixedAmount == 0 {
		services := data.GetServices(h.app)
		for _, s := range services {
			if s.ID == serviceID {
				selectedService = s
				break
			}
		}
		if selectedService.ID == "" && len(services) > 0 {
			selectedService = services[0]
		}
	}

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return components.PaymentModal(settings, selectedService, fixedAmount, orderCode).Render(e.Request.Context(), e.Response)
}

// RenderErrorHTML helper formats a styled error box.
func RenderErrorHTML(message string) string {
	return fmt.Sprintf(`<div class="p-4 bg-rose-50 border border-rose-300 rounded-xl text-rose-800 text-sm"><div class="font-bold">Ошибка</div><p class="mt-1 text-xs">%s</p></div>`, html.EscapeString(message))
}

// HandleBlogList renders the list of published articles with optional live search.
func (h *WebHandler) HandleBlogList(e *core.RequestEvent) error {
	settings := data.GetSiteSettings(h.app)
	content := data.GetSiteContent(h.app)
	allArticles := data.GetArticles(h.app)

	searchQuery := strings.TrimSpace(e.Request.URL.Query().Get("q"))
	filtered := allArticles
	if searchQuery != "" {
		qLower := strings.ToLower(searchQuery)
		var matches []models.ArticleItem
		for _, a := range allArticles {
			titleMatch := strings.Contains(strings.ToLower(a.Title), qLower)
			excerptMatch := strings.Contains(strings.ToLower(a.Excerpt), qLower)
			tagMatch := false
			for _, t := range a.Tags {
				if strings.Contains(strings.ToLower(t), qLower) {
					tagMatch = true
					break
				}
			}
			if titleMatch || excerptMatch || tagMatch {
				matches = append(matches, a)
			}
		}
		filtered = matches
	}

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.BlogList(settings, content, filtered, searchQuery).Render(e.Request.Context(), e.Response)
}

// HandleBlogPost renders an individual article by slug with related articles.
func (h *WebHandler) HandleBlogPost(e *core.RequestEvent) error {
	settings := data.GetSiteSettings(h.app)
	content := data.GetSiteContent(h.app)
	slug := e.Request.PathValue("slug")

	article, err := data.GetArticleBySlug(h.app, slug)
	if err != nil || article == nil {
		e.Response.Header().Set("Location", "/blog")
		e.Response.WriteHeader(http.StatusTemporaryRedirect)
		return nil
	}

	// Increment view count asynchronously
	go data.IncrementArticleViews(h.app, article.ID)

	allArticles := data.GetArticles(h.app)
	var related []models.ArticleItem
	for _, a := range allArticles {
		if a.ID != article.ID {
			related = append(related, a)
			if len(related) >= 2 {
				break
			}
		}
	}

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.BlogPost(settings, content, *article, related).Render(e.Request.Context(), e.Response)
}

// HandleSitemap serves dynamically generated sitemap.xml for search engines.
func (h *WebHandler) HandleSitemap(e *core.RequestEvent) error {
	host := e.Request.Host
	proto := "https"
	if e.Request.TLS == nil && !strings.Contains(e.Request.Header.Get("X-Forwarded-Proto"), "https") {
		proto = "http"
	}
	baseURL := fmt.Sprintf("%s://%s", proto, host)

	articles := data.GetArticles(h.app)

	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sb.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	// Home page
	sb.WriteString(fmt.Sprintf("  <url>\n    <loc>%s/</loc>\n    <changefreq>daily</changefreq>\n    <priority>1.0</priority>\n  </url>\n", baseURL))

	// Blog list
	sb.WriteString(fmt.Sprintf("  <url>\n    <loc>%s/blog</loc>\n    <changefreq>daily</changefreq>\n    <priority>0.8</priority>\n  </url>\n", baseURL))

	// Articles
	for _, a := range articles {
		sb.WriteString(fmt.Sprintf("  <url>\n    <loc>%s/blog/%s</loc>\n    <changefreq>weekly</changefreq>\n    <priority>0.7</priority>\n  </url>\n", baseURL, a.Slug))
	}

	sb.WriteString(`</urlset>`)

	e.Response.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, err := e.Response.Write([]byte(sb.String()))
	return err
}

