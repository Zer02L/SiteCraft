# Cookbook: HTMX Endpoint with Partial Templ Render

Canonical pattern for writing HTTP endpoints in PocketBase that render partial Templ components with proper HTTP status codes and OOB / Trigger headers.

## File: `internal/handlers/<domain>.go`

```go
package handlers

import (
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type FeatureHandler struct {
	app *pocketbase.PocketBase
}

func NewFeatureHandler(app *pocketbase.PocketBase) *FeatureHandler {
	return &FeatureHandler{app: app}
}

func (h *FeatureHandler) HandleAction(e *core.RequestEvent) error {
	// 1. Parsing and validation
	input := e.Request.FormValue("query")
	if len(input) == 0 {
		e.Response.Header().Set("HX-Retarget", "#error-container")
		return e.HTML(http.StatusUnprocessableEntity, `<div class="text-rose-600 text-xs">Field cannot be empty</div>`)
	}

	// 2. Business logic execution
	// ...

	// 3. Render partial Templ component directly to response writer
	component := ui_partials.SearchResult(input) // or inline Templ component
	e.Response.Header().Set("HX-Trigger", `{"featureUpdated": true}`)
	e.Response.WriteHeader(http.StatusOK)
	return component.Render(e.Request.Context(), e.Response)
}
```

## Important Rules:
1. Return HTTP 422 (Unprocessable Entity) on validation errors so HTMX triggers error state.
2. Use `e.Response.Header().Set("HX-Trigger", ...)` for client-side event notifications.
3. Use `component.Render(e.Request.Context(), e.Response)` to stream Templ output directly without intermediate buffers.
