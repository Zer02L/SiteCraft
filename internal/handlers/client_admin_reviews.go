package handlers

import (
	"net/http"

	"sitecraft/internal/data"
	"sitecraft/ui/pages"

	"github.com/pocketbase/pocketbase/core"
)

// HandleTabReviews renders reviews moderation tab fragment for HTMX.
func (h *ClientAdminHandler) HandleTabReviews(e *core.RequestEvent) error {
	if !h.isAuthorized(e) {
		return e.Redirect(http.StatusFound, "/client/login")
	}
	reviews := data.GetAllReviewsAdmin(h.app)
	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return pages.ClientReviewsTab(reviews).Render(e.Request.Context(), e.Response)
}
