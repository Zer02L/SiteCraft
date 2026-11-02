package cinematic_2

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestDefaultHeroProps(t *testing.T) {
	props := DefaultHeroProps()
	if props.BrandName != "JHILL" {
		t.Fatalf("expected BrandName to be 'JHILL', got %q", props.BrandName)
	}
	if props.Subtitle != "3D Character Artist" {
		t.Fatalf("expected Subtitle to be '3D Character Artist', got %q", props.Subtitle)
	}
	if props.VideoURL == "" || props.PosterURL == "" {
		t.Fatal("expected VideoURL and PosterURL to be non-empty")
	}
}

func TestHeroScreenRender(t *testing.T) {
	props := DefaultHeroProps()
	var buf bytes.Buffer
	err := HeroScreen(props).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("HeroScreen render failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "JHILL") {
		t.Errorf("rendered output missing brand name 'JHILL'")
	}
	if !strings.Contains(out, "3D Character Artist") {
		t.Errorf("rendered output missing subtitle '3D Character Artist'")
	}
	if !strings.Contains(out, "<video") {
		t.Errorf("rendered output missing <video> element")
	}
	if !strings.Contains(out, "prefers-reduced-motion") {
		t.Errorf("rendered output missing reduced-motion accessibility check")
	}
}

func TestPortfolioCollageRender(t *testing.T) {
	props := DefaultCollageProps()
	var buf bytes.Buffer
	err := PortfolioCollage(props).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("PortfolioCollage render failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Portfolio") {
		t.Errorf("rendered output missing title 'Portfolio'")
	}
	if !strings.Contains(out, "https://www.artstation.com/jhill") {
		t.Errorf("rendered output missing artstation portfolio link")
	}
	if !strings.Contains(out, "Mom - Back 4 Blood") {
		t.Errorf("rendered output missing artwork title 'Mom - Back 4 Blood'")
	}
}

func TestShowcaseRender(t *testing.T) {
	var buf bytes.Buffer
	err := DefaultShowcase().Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("DefaultShowcase render failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "JHILL") || !strings.Contains(out, "Portfolio") {
		t.Errorf("rendered showcase missing Hero or Collage content")
	}
}
