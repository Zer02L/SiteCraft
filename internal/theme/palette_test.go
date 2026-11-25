package theme

import (
	"image"
	"image/color"
	"testing"
)

func TestExtractPaletteFromImage(t *testing.T) {
	// 1. Nil image test
	_, err := ExtractPaletteFromImage(nil)
	if err == nil {
		t.Error("expected error for nil image, got nil")
	}

	// 2. Synthetic test image: 100x100
	// 60% White (#ffffff), 30% Dark Navy (#0f172a), 10% Bright Crimson (#dc2626)
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Fill with white
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	// 30% Dark Navy (#0f172a -> R:15, G:23, B:42) on top 30 rows
	for y := 0; y < 30; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 15, G: 23, B: 42, A: 255})
		}
	}

	// 10% Bright Crimson (#dc2626 -> R:220, G:38, B:38) on middle 10 rows
	for y := 30; y < 40; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 220, G: 38, B: 38, A: 255})
		}
	}

	res, err := ExtractPaletteFromImage(img)
	if err != nil {
		t.Fatalf("unexpected error extracting palette: %v", err)
	}

	if len(res.DominantColors) == 0 {
		t.Fatal("expected at least 1 dominant color extracted")
	}

	t.Logf("Extracted %d dominant colors", len(res.DominantColors))
	for i, c := range res.DominantColors {
		t.Logf("[%d] %s: pct=%.1f%%, lum=%.2f, chroma=%.2f, hue=%.0f°",
			i, c.Hex, c.Percent, c.Luminance, c.Chroma, c.Hue)
	}

	if res.Accent == "" || len(res.Accent) != 7 || res.Accent[0] != '#' {
		t.Errorf("invalid accent color format: %q", res.Accent)
	}

	if res.Primary == "" || len(res.Primary) != 7 || res.Primary[0] != '#' {
		t.Errorf("invalid primary color format: %q", res.Primary)
	}

	if res.Surface == "" || len(res.Surface) != 7 || res.Surface[0] != '#' {
		t.Errorf("invalid surface color format: %q", res.Surface)
	}

	if len(res.Palettes) < 4 {
		t.Errorf("expected 4 curated palettes, got %d", len(res.Palettes))
	}

	for _, p := range res.Palettes {
		if p.ID == "" || p.Name == "" || p.Accent == "" || p.Primary == "" || p.Surface == "" {
			t.Errorf("incomplete curated palette: %+v", p)
		}
	}
}

func TestHslConversion(t *testing.T) {
	tests := []struct {
		r, g, b uint8
	}{
		{255, 0, 0},     // Pure Red
		{0, 255, 0},     // Pure Green
		{0, 0, 255},     // Pure Blue
		{255, 255, 255}, // White
		{0, 0, 0},       // Black
		{37, 99, 235},   // Blue-600
	}

	for _, tt := range tests {
		h, s, l := rgbToHsl(tt.r, tt.g, tt.b)
		r2, g2, b2 := hslToRgb(h, s, l)

		// Check difference within rounding tolerance
		diffR := int(tt.r) - int(r2)
		diffG := int(tt.g) - int(g2)
		diffB := int(tt.b) - int(b2)
		if diffR < 0 {
			diffR = -diffR
		}
		if diffG < 0 {
			diffG = -diffG
		}
		if diffB < 0 {
			diffB = -diffB
		}

		if diffR > 2 || diffG > 2 || diffB > 2 {
			t.Errorf("roundtrip HSL failed for (%d,%d,%d) -> H:%.1f, S:%.2f, L:%.2f -> got (%d,%d,%d)",
				tt.r, tt.g, tt.b, h, s, l, r2, g2, b2)
		}
	}
}
