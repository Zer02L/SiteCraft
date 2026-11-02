package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sitecraft/internal/testutil"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

func TestFlatTo3DDemoEndpoint(t *testing.T) {
	app := testutil.NewTestApp(t)
	handler := NewWebHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/demo/flat-to-3d", nil)
	rec := httptest.NewRecorder()

	re := &core.RequestEvent{
		App:      app,
		Request:  req,
		Response: rec,
	}

	if err := handler.HandleFlatTo3DDemo(re); err != nil {
		t.Fatalf("HandleFlatTo3DDemo returned error: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "NORDIC TIMBER") {
		t.Errorf("expected body to contain 'NORDIC TIMBER'")
	}
	if !strings.Contains(body, "data-flat-to-3d") {
		t.Errorf("expected body to contain 'data-flat-to-3d'")
	}
	if !strings.Contains(body, "/assets/js/ogl.mjs") {
		t.Errorf("expected body to contain OGL script import '/assets/js/ogl.mjs'")
	}
	if !strings.Contains(body, "ГОСТ Р 52872-2019") {
		t.Errorf("expected body to contain accessible fallback mention of 'ГОСТ Р 52872-2019'")
	}
}
