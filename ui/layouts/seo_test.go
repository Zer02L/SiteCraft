package layouts

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"sitecraft/internal/models"
)

func TestIsLightHex(t *testing.T) {
	tests := []struct {
		hex      string
		expected bool
	}{
		{"#ffffff", true},
		{"#facc15", true}, // Tailwind yellow-400
		{"#38bdf8", true}, // Tailwind sky-400
		{"#000000", false},
		{"#0f172a", false}, // slate-900
		{"#2563eb", false}, // blue-600
		{"#1e293b", false}, // slate-800
	}

	for _, tt := range tests {
		got := isLightHex(tt.hex)
		if got != tt.expected {
			t.Errorf("isLightHex(%q) = %v, expected %v", tt.hex, got, tt.expected)
		}
	}
}

func TestRenderCustomTokensOnAccent(t *testing.T) {
	// 1. Light accent (yellow) -> should produce dark text on-accent
	settingsLight := models.SiteSettings{
		AccentColor: "#facc15",
	}
	var bufLight bytes.Buffer
	err := RenderCustomTokens(settingsLight).Render(context.Background(), &bufLight)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	outLight := bufLight.String()
	if !strings.Contains(outLight, "--color-on-accent: #0f172a;") {
		t.Errorf("expected dark text on light accent, got: %s", outLight)
	}

	// 2. Dark accent (blue) -> should produce white text on-accent
	settingsDark := models.SiteSettings{
		AccentColor: "#2563eb",
	}
	var bufDark bytes.Buffer
	err = RenderCustomTokens(settingsDark).Render(context.Background(), &bufDark)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	outDark := bufDark.String()
	if !strings.Contains(outDark, "--color-on-accent: #ffffff;") {
		t.Errorf("expected white text on dark accent, got: %s", outDark)
	}
}
