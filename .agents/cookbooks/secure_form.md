# Cookbook: Secure Form Handler (Honeypot + Bluemonday XSS)

Canonical pattern for public web forms (leads, contact, reviews, booking) protecting against automated bots and XSS payload injection.

## File: `internal/handlers/<domain>_form.go`

```go
package handlers

import (
	"net/http"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

var xssPolicy = bluemonday.UGCPolicy()

func HandlePublicFormSubmit(app *pocketbase.PocketBase, e *core.RequestEvent) error {
	// 1. Honeypot check (hidden input rendered with opacity-0 / sr-only)
	// Bot fills field -> return fake success (HTTP 200) silently
	if honeypot := e.Request.FormValue("website_company_url"); honeypot != "" {
		return e.HTML(http.StatusOK, `<div class="p-3 bg-emerald-50 text-emerald-800 rounded-xl text-xs">Заявка отправлена!</div>`)
	}

	// 2. Extract and sanitize untrusted inputs
	rawName := strings.TrimSpace(e.Request.FormValue("name"))
	rawComment := strings.TrimSpace(e.Request.FormValue("comment"))

	safeName := xssPolicy.Sanitize(rawName)
	safeComment := xssPolicy.Sanitize(rawComment)

	// 3. Trust boundary validation
	if len(safeName) < 2 {
		return e.HTML(http.StatusUnprocessableEntity, `
			<div class="p-3 bg-rose-50 border border-rose-300 rounded-xl text-rose-800 text-xs">
				Имя должно содержать не менее 2 символов.
			</div>
		`)
	}

	// 4. Save to PocketBase collection
	col, err := app.FindCollectionByNameOrId("leads")
	if err != nil {
		return e.HTML(http.StatusInternalServerError, `<div>Ошибка сохранения</div>`)
	}
	record := core.NewRecord(col)
	record.Set("name", safeName)
	record.Set("message", safeComment)
	if err := app.Save(record); err != nil {
		return e.HTML(http.StatusInternalServerError, `<div>Ошибка сохранения</div>`)
	}

	return e.HTML(http.StatusOK, `<div class="p-3 bg-emerald-50 text-emerald-800 rounded-xl text-xs">Заявка успешно принята!</div>`)
}
```

## Honeypot Frontend Snippet (`.templ`):
```html
<div class="hidden" aria-hidden="true" tabindex="-1">
    <label for="website_company_url">Do not fill</label>
    <input type="text" id="website_company_url" name="website_company_url" autocomplete="off" />
</div>
```
