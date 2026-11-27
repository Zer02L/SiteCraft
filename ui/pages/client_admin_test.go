package pages

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"sitecraft/internal/ai"
	"sitecraft/internal/checker"
	"sitecraft/internal/models"
)

func TestAdminThemeToggleUsesCircularReveal(t *testing.T) {
	var buf bytes.Buffer
	if err := AdminThemeToggle().Render(context.Background(), &buf); err != nil {
		t.Fatalf("render AdminThemeToggle: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"startViewTransition",
		"clipPath",
		"::view-transition-new(root)",
		"prefers-reduced-motion",
		"theme-freeze",
		"theme-crossfade",
		"transition.finished",
		`toggle($event)`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("AdminThemeToggle markup missing %q", want)
		}
	}
}

func TestClientDashboardWorkbenchContracts(t *testing.T) {
	settings := models.SiteSettings{SiteName: "Craft Studio"}
	content := models.SiteContent{Phone1: "+7 (999) 123-45-67"}
	modules := []models.AdminTabModule{
		{ID: "leads", Title: "Заявки", Route: "/admin/tab/leads", Order: 10},
		{ID: "blocks", Title: "Конструктор блоков", Route: "/admin/tab/blocks", Order: 20},
		{ID: "settings", Title: "Настройки", Route: "/admin/tab/settings", Order: 65},
	}
	cl := checker.ChecklistResult{
		CompletedCount: 3,
		TotalCount:     5,
		Percentage:     60,
		Items: []checker.ChecklistItem{
			{ID: "phone", Title: "Указан телефон", IsCompleted: true, TabID: "contacts"},
		},
	}

	var buf bytes.Buffer
	if err := ClientDashboard(settings, content, nil, nil, nil, nil, nil, nil, "leads", modules, nil, cl).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render ClientDashboard: %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		"Операции",
		"Сайт и контент",
		"Система",
		`id="admin-checklist-dialog"`,
		`id="admin-mobile-nav-dialog"`,
		`id="admin-command-palette"`,
		`id="admin-htmx-indicator"`,
		`id="lead-detail-dialog"`,
		"toggleSidebar()",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("ClientDashboard missing expected workbench contract %q", want)
		}
	}

	// Inactive modules must never leak into rendered shell or command palette.
	if strings.Contains(out, "Отзывы клиентов") {
		t.Errorf("disabled module title leaked into ClientDashboard HTML")
	}

	// AI Engine tab uses semantic admin-* tokens without dark-mode slate/purple overrides.
	var aiBuf bytes.Buffer
	if err := ClientAITab(settings, ai.Presets, "", false).Render(context.Background(), &aiBuf); err != nil {
		t.Fatalf("render ClientAITab: %v", err)
	}
	aiOut := aiBuf.String()
	for _, forbidden := range []string{"bg-slate-900", "bg-purple-500"} {
		if strings.Contains(aiOut, forbidden) {
			t.Errorf("ClientAITab still contains forbidden raw color class %q", forbidden)
		}
	}
}
