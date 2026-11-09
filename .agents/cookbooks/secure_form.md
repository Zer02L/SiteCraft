# Cookbook: Secure Form Handler (Honeypot + Bluemonday XSS)

Canonical pattern for public web forms (leads, contact, reviews, booking) protecting against automated bots and XSS payload injection.

## File: `internal/handlers/<domain>_form.go`

```go
package handlers

import (
	"net/http"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"sitecraft/internal/db"
)

var xssPolicy = bluemonday.UGCPolicy()

func HandlePublicFormSubmit(queries db.Querier, w http.ResponseWriter, r *http.Request) {
	// 1. Honeypot check (hidden input rendered with opacity-0 / sr-only)
	// Bot fills field -> return fake success (HTTP 200) silently
	if honeypot := r.FormValue("website_company_url"); honeypot != "" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<div class="p-3 bg-emerald-50 text-emerald-800 rounded-xl text-xs">Заявка отправлена!</div>`))
		return
	}

	// 2. Extract and sanitize untrusted inputs
	rawName := strings.TrimSpace(r.FormValue("name"))
	rawComment := strings.TrimSpace(r.FormValue("comment"))

	safeName := xssPolicy.Sanitize(rawName)
	safeComment := xssPolicy.Sanitize(rawComment)

	// 3. Trust boundary validation
	if len(safeName) < 2 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`
			<div class="p-3 bg-rose-50 border border-rose-300 rounded-xl text-rose-800 text-xs">
				Имя должно содержать не менее 2 символов.
			</div>
		`))
		return
	}

	// 4. Save to database via sqlc Querier
	err := queries.CreateLead(r.Context(), db.CreateLeadParams{
		ID:      newULID(),
		Name:    safeName,
		Message: safeComment,
		Status:  "new",
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<div>Ошибка сохранения</div>`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="p-3 bg-emerald-50 text-emerald-800 rounded-xl text-xs">Заявка успешно принята!</div>`))
}
```

## Honeypot Frontend Snippet (`.templ`):
```html
<div class="hidden" aria-hidden="true" tabindex="-1">
    <label for="website_company_url">Do not fill</label>
    <input type="text" id="website_company_url" name="website_company_url" autocomplete="off" />
</div>
```
