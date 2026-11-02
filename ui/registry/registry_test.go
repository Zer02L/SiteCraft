package registry

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"sitecraft/internal/models"
)

func TestResolveBlockSource(t *testing.T) {
	ctx := RenderContext{
		Settings: models.SiteSettings{Theme: "cyber"},
	}

	// 1. Cyber hero via theme
	b1 := models.PageBlock{Type: "hero"}
	if src := ResolveBlockSource(b1, ctx); src != "ui/sections/hero/cyber_hero.templ" {
		t.Errorf("expected cyber_hero.templ, got %s", src)
	}

	// 2. Split hero via payload variant
	b2 := models.PageBlock{
		Type:    "hero",
		Payload: map[string]interface{}{"variant": "split"},
	}
	if src := ResolveBlockSource(b2, ctx); src != "ui/sections/hero/split_hero.templ" {
		t.Errorf("expected split_hero.templ, got %s", src)
	}

	// 3. Known static block
	b3 := models.PageBlock{Type: "booking"}
	if src := ResolveBlockSource(b3, ctx); src != "ui/sections/booking.templ" {
		t.Errorf("expected booking.templ, got %s", src)
	}

	// 4. Fallback unknown block
	b4 := models.PageBlock{Type: "unknown_custom_block"}
	if src := ResolveBlockSource(b4, ctx); src != "ui/components/dynamic_block.templ" {
		t.Errorf("expected dynamic_block.templ, got %s", src)
	}
}

func TestRenderBlock_SourceMapping(t *testing.T) {
	ctx := RenderContext{
		Settings: models.SiteSettings{Theme: "classic"},
	}

	b := models.PageBlock{
		Type:  "services",
		Title: "Our Services",
	}

	comp := RenderBlock(b, ctx)
	if comp == nil {
		t.Fatal("expected non-nil component from RenderBlock")
	}

	var buf bytes.Buffer
	if err := comp.Render(context.Background(), &buf); err != nil {
		t.Fatalf("failed to render component: %v", err)
	}

	rendered := buf.String()
	if !strings.Contains(rendered, `data-block="services"`) {
		t.Errorf("expected rendered HTML to contain data-block=\"services\", got:\n%s", rendered)
	}
	if !strings.Contains(rendered, `data-source="ui/sections/services.templ"`) {
		t.Errorf("expected rendered HTML to contain data-source=\"ui/sections/services.templ\", got:\n%s", rendered)
	}
	if !strings.Contains(rendered, `class="contents"`) {
		t.Errorf("expected rendered HTML to contain class=\"contents\", got:\n%s", rendered)
	}
}
