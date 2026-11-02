package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"sitecraft/internal/data"
	"sitecraft/internal/models"
	"sitecraft/ui/pages"

	"ganigeorgiev.com/pocketbase/pocketbase"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

// TrackerHandler manages B2B repair/service order tracking (Ниша Г).
type TrackerHandler struct {
	app *pocketbase.PocketBase
}

// NewTrackerHandler creates a new instance of TrackerHandler.
func NewTrackerHandler(app *pocketbase.PocketBase) *TrackerHandler {
	return &TrackerHandler{app: app}
}

// HandleTrackerPage renders the tracking search page and search results.
func (h *TrackerHandler) HandleTrackerPage(e *core.RequestEvent) error {
	settings := data.GetSiteSettings(h.app)
	content := data.GetSiteContent(h.app)

	code := strings.TrimSpace(e.Request.URL.Query().Get("code"))
	var order *models.ServiceOrder
	notFound := false

	if code != "" {
		o, err := data.GetServiceOrderByCode(h.app, code)
		if err != nil || o == nil {
			notFound = true
		} else {
			order = o
		}
	}

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.TrackerPage(settings, content, order, code, notFound).Render(e.Request.Context(), e.Response)
}

// HandleTrackOrder handles direct URLs like /track/{code}.
func (h *TrackerHandler) HandleTrackOrder(e *core.RequestEvent) error {
	settings := data.GetSiteSettings(h.app)
	content := data.GetSiteContent(h.app)

	code := strings.TrimSpace(e.Request.PathValue("code"))
	var order *models.ServiceOrder
	notFound := false

	if code != "" {
		o, err := data.GetServiceOrderByCode(h.app, code)
		if err != nil || o == nil {
			notFound = true
		} else {
			order = o
		}
	}

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.TrackerPage(settings, content, order, code, notFound).Render(e.Request.Context(), e.Response)
}

// HandleTrackStatusPartial returns just the TrackerStatusCard HTML for HTMX polling.
func (h *TrackerHandler) HandleTrackStatusPartial(e *core.RequestEvent) error {
	settings := data.GetSiteSettings(h.app)
	code := strings.TrimSpace(e.Request.PathValue("code"))

	order, err := data.GetServiceOrderByCode(h.app, code)
	if err != nil || order == nil {
		return e.HTML(http.StatusNotFound, `<div class="p-4 text-center text-xs text-stone-500">Заказ не найден</div>`)
	}

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.TrackerStatusCard(settings, *order).Render(e.Request.Context(), e.Response)
}

// HandleTrackSSE streams realtime Server-Sent Events when order status updates.
func (h *TrackerHandler) HandleTrackSSE(e *core.RequestEvent) error {
	settings := data.GetSiteSettings(h.app)
	code := strings.TrimSpace(e.Request.PathValue("code"))

	w := e.Response
	flusher, ok := w.(http.Flusher)
	if !ok {
		return e.HTML(http.StatusInternalServerError, "SSE streaming unsupported")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx, cancel := context.WithTimeout(e.Request.Context(), 5*time.Minute)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	lastStage := ""

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			order, err := data.GetServiceOrderByCode(h.app, code)
			if err == nil && order != nil {
				if order.CurrentStage != lastStage {
					lastStage = order.CurrentStage

					var buf bytes.Buffer
					_ = pages.TrackerStatusCard(settings, *order).Render(ctx, &buf)

					htmlContent := strings.ReplaceAll(buf.String(), "\n", "")
					fmt.Fprintf(w, "event: OrderUpdated\ndata: %s\n\n", htmlContent)
					flusher.Flush()
				}
			}
		}
	}
}
