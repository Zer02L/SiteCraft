package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sitecraft/internal/testutil"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

func TestJHillDemoEndpoint(t *testing.T) {
	app := testutil.NewTestApp(t)
	handler := NewWebHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/demo/jhill", nil)
	rec := httptest.NewRecorder()

	re := &core.RequestEvent{
		App:      app,
		Request:  req,
		Response: rec,
	}

	if err := handler.HandleJHillDemo(re); err != nil {
		t.Fatalf("HandleJHillDemo returned error: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "JHILL") {
		t.Errorf("expected body to contain 'JHILL'")
	}
	if !strings.Contains(body, "Portfolio") {
		t.Errorf("expected body to contain 'Portfolio'")
	}
	if !strings.Contains(body, "3D Character Artist") {
		t.Errorf("expected body to contain '3D Character Artist'")
	}
}
