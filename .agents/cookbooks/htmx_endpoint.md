# Cookbook: HTMX Endpoint with Partial Templ Render

Canonical pattern for writing HTTP endpoints in Chi that render partial Templ components with proper HTTP status codes and OOB / Trigger headers.

## File: `internal/handlers/<domain>.go`

```go
package handlers

import (
	"net/http"

	"sitecraft/internal/db"
)

type FeatureHandler struct {
	db db.Querier
}

func NewFeatureHandler(database db.Querier) *FeatureHandler {
	return &FeatureHandler{db: database}
}

func (h *FeatureHandler) HandleAction(w http.ResponseWriter, r *http.Request) {
	// 1. Parsing and validation
	input := r.FormValue("query")
	if len(input) == 0 {
		w.Header().Set("HX-Retarget", "#error-container")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`<div class="text-rose-600 text-xs">Field cannot be empty</div>`))
		return
	}

	// 2. Business logic execution via sqlc
	// ...

	// 3. Render partial Templ component directly to response writer
	component := ui_partials.SearchResult(input) // or inline Templ component
	w.Header().Set("HX-Trigger", `{"featureUpdated": true}`)
	w.WriteHeader(http.StatusOK)
	_ = component.Render(r.Context(), w)
}
```

## Important Rules:
1. Return HTTP 422 (Unprocessable Entity) on validation errors so HTMX triggers error state.
2. Use `w.Header().Set("HX-Trigger", ...)` for client-side event notifications.
3. Use `component.Render(r.Context(), w)` to stream Templ output directly without intermediate buffers.
