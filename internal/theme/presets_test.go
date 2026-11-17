package theme

import (
	"strings"
	"testing"
)

func TestBuiltinPresets(t *testing.T) {
	if len(BuiltinPresets) < 4 {
		t.Fatalf("expected at least 4 builtin presets, got %d", len(BuiltinPresets))
	}

	keys := []string{"modern_bright", "luxury_editorial", "dark_premium", "swiss_bold"}
	for _, k := range keys {
		p, ok := GetPreset(k)
		if !ok {
			t.Errorf("expected preset %q to exist", k)
		}
		if p.Name == "" || p.PrimaryColor == "" || p.AccentColor == "" {
			t.Errorf("preset %q has incomplete fields: %+v", k, p)
		}
	}
}

func TestLuminanceAndContrast(t *testing.T) {
	tests := []struct {
		hex          string
		isLight      bool
		expectedText string
	}{
		{"#ffffff", true, "#0f172a"},
		{"#facc15", true, "#0f172a"}, // Yellow-400
		{"#38bdf8", true, "#0f172a"}, // Sky-400
		{"#000000", false, "#ffffff"},
		{"#0f172a", false, "#ffffff"}, // Slate-900
		{"#1e293b", false, "#ffffff"}, // Slate-800
		{"#2563eb", false, "#ffffff"}, // Blue-600
		{"#dc2626", false, "#ffffff"}, // Red-600
	}

	for _, tt := range tests {
		light := IsLightHex(tt.hex)
		if light != tt.isLight {
			t.Errorf("IsLightHex(%q) = %v, expected %v (lum=%f)", tt.hex, light, tt.isLight, RelativeLuminance(tt.hex))
		}
		textColor := ContrastTextColor(tt.hex)
		if textColor != tt.expectedText {
			t.Errorf("ContrastTextColor(%q) = %s, expected %s", tt.hex, textColor, tt.expectedText)
		}
	}
}

func TestGenerateCSSVariables(t *testing.T) {
	p, ok := GetPreset("swiss_bold")
	if !ok {
		t.Fatal("swiss_bold preset not found")
	}

	css := GenerateCSSVariables(p)
	if !strings.Contains(css, "--color-primary: #000000") {
		t.Errorf("expected primary color in css, got: %s", css)
	}
	if !strings.Contains(css, "--radius-card: 0px") {
		t.Errorf("expected sharp 0px radius in css, got: %s", css)
	}
	if !strings.Contains(css, "--font-serif-title: \"Bebas Neue\"") {
		t.Errorf("expected Bebas Neue font in css, got: %s", css)
	}
	if !strings.Contains(css, "--color-on-primary: #ffffff") {
		t.Errorf("expected white on primary for black, got: %s", css)
	}
}
