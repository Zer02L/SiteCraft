package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"sitecraft/internal/auth"
	"sitecraft/internal/models"
)

// Regression: the status action must refresh only the tab content. Swapping the
// root x-data element breaks Alpine's scope for the rows' x-show expressions.
func TestAdminUpdateLeadStatusReturnsContentOnly(t *testing.T) {
	srv, conn := setupTestServer(t)
	defer conn.Close()

	ctx, err := srv.SessionMgr.Load(context.Background(), "")
	if err != nil {
		t.Fatalf("failed to init session: %v", err)
	}
	_ = auth.SetAdminSession(ctx, srv.SessionMgr, "admin@sitecraft.local")
	if _, _, err := srv.SessionMgr.Commit(ctx); err != nil {
		t.Fatalf("failed to commit session: %v", err)
	}

	lead, err := srv.Repo.CreateLead(ctx, models.Lead{Name: "Тест", Phone: "+79990000000"})
	if err != nil {
		t.Fatalf("CreateLead failed: %v", err)
	}

	form := url.Values{"status": {"в обработке"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/leads/"+lead.ID+"/status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, `id="lead-row-`+lead.ID+`"`) {
		t.Errorf("response must contain the updated lead row")
	}
	if !strings.Contains(body, "в обработке") {
		t.Errorf("response must contain the new status")
	}
	if strings.Contains(body, "x-data=") {
		t.Errorf("response must not replace the Alpine x-data root")
	}

	leads, err := srv.Repo.GetLeads(ctx)
	if err != nil {
		t.Fatalf("GetLeads failed: %v", err)
	}
	found := false
	for _, l := range leads {
		if l.ID == lead.ID && l.Status == "в обработке" {
			found = true
		}
	}
	if !found {
		t.Errorf("lead status was not persisted")
	}
}
