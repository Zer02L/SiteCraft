package theme

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"sort"
	"strings"

	_ "golang.org/x/image/webp"
)

// ExtractedColor represents an analyzed color with metric properties.
type ExtractedColor struct {
	Hex        string  `json:"hex"`
	R          uint8   `json:"r"`
	G          uint8   `json:"g"`
	B          uint8   `json:"b"`
	Luminance  float64 `json:"luminance"`  // WCAG relative luminance (0..1)
	Chroma     float64 `json:"chroma"`     // Approximate color saturation/chroma (0..1)
	Hue        float64 `json:"hue"`        // Hue angle in degrees (0..360)
	Population int     `json:"population"` // Number of sampled pixels in cluster
	Percent    float64 `json:"percent"`    // Approximate share of total pixels (0..100)
	IsDark     bool    `json:"is_dark"`    // True if luminance < 0.22
}

// CuratedPalette represents a 60-30-10 theme scheme derived from image colors.
type CuratedPalette struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Accent      string `json:"accent"`       // 10% CTA / Focus
	Primary     string `json:"primary"`      // 30% Headings / Text
	Surface     string `json:"surface"`      // 60% Background
	Theme       string `json:"theme"`        // Base CSS theme preset
	DesignFlavor string `json:"design_flavor"`// "soft", "corporate", "sharp"
}

// ImageThemePalette contains all extracted colors and curated palettes.
type ImageThemePalette struct {
	DominantColors  []ExtractedColor `json:"dominant_colors"`
	Accent          string           `json:"accent"`
	Primary         string           `json:"primary"`
	Surface         string           `json:"surface"`
	SuggestedTheme  string           `json:"suggested_theme"`
	SuggestedFlavor string           `json:"suggested_flavor"`
	Palettes        []CuratedPalette `json:"palettes"`
}

// rgbToHsl converts RGB (0..255) to H (0..360), S (0..1), L (0..1).
func rgbToHsl(r, g, b uint8) (float64, float64, float64) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	delta := max - min

	l := (max + min) / 2.0
	if delta == 0 {
		return 0, 0, l
	}

	var s float64
	if l > 0.5 {
		s = delta / (2.0 - max - min)
	} else {
		s = delta / (max + min)
	}

	var h float64
	switch max {
	case rf:
		h = (gf - bf) / delta
		if gf < bf {
			h += 6.0
		}
	case gf:
		h = (bf-rf)/delta + 2.0
	case bf:
		h = (rf-gf)/delta + 4.0
	}
	h *= 60.0

	return h, s, l
}

// hslToRgb converts H (0..360), S (0..1), L (0..1) back to RGB (0..255).
func hslToRgb(h, s, l float64) (uint8, uint8, uint8) {
	if s == 0 {
		v := uint8(math.Round(l * 255.0))
		return v, v, v
	}

	var q float64
	if l < 0.5 {
		q = l * (1.0 + s)
	} else {
		q = l + s - (l * s)
	}
	p := 2.0*l - q

	hk := h / 360.0
	tr := hk + 1.0/3.0
	tg := hk
	tb := hk - 1.0/3.0

	colorChannel := func(tc float64) uint8 {
		if tc < 0 {
			tc += 1.0
		}
		if tc > 1 {
			tc -= 1.0
		}
		var val float64
		switch {
		case tc < 1.0/6.0:
			val = p + ((q - p) * 6.0 * tc)
		case tc < 1.0/2.0:
			val = q
		case tc < 2.0/3.0:
			val = p + ((q - p) * (2.0/3.0 - tc) * 6.0)
		default:
			val = p
		}
		return uint8(math.Max(0, math.Min(255, math.Round(val*255.0))))
	}

	return colorChannel(tr), colorChannel(tg), colorChannel(tb)
}

// formatHex formats RGB as lowercase 7-char string "#rrggbb".
func formatHex(r, g, b uint8) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// colorDistSqr computes squared Euclidean distance in RGB color space.
func colorDistSqr(r1, g1, b1, r2, g2, b2 uint8) float64 {
	dr := float64(r1) - float64(r2)
	dg := float64(g1) - float64(g2)
	db := float64(b1) - float64(b2)
	return dr*dr + dg*dg + db*db
}

// ExtractPaletteFromImage processes an image, quantizes dominant colors,
// and constructs balanced 60-30-10 website theme palettes.
func ExtractPaletteFromImage(img image.Image) (*ImageThemePalette, error) {
	if img == nil {
		return nil, fmt.Errorf("nil image provided")
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width == 0 || height == 0 {
		return nil, fmt.Errorf("empty image dimensions")
	}

	// 1. Downsample step to sample ~10,000 to 20,000 pixels for sub-10ms performance
	stepX := int(math.Max(1, float64(width)/120.0))
	stepY := int(math.Max(1, float64(height)/120.0))

	// 2. 5-bit color quantization bucket (32x32x32 bins)
	type binData struct {
		rSum, gSum, bSum int64
		count            int
	}
	buckets := make(map[uint32]*binData)
	totalSampled := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y += stepY {
		for x := bounds.Min.X; x < bounds.Max.X; x += stepX {
			c := img.At(x, y)
			r16, g16, b16, a16 := c.RGBA()

			// Skip transparent / semi-transparent pixels (threshold: alpha < 50%)
			if a16 < 0x8000 {
				continue
			}

			// Convert to standard 8-bit channels
			r := uint8(r16 >> 8)
			g := uint8(g16 >> 8)
			b := uint8(b16 >> 8)

			// 5-bit key
			key := (uint32(r>>3) << 10) | (uint32(g>>3) << 5) | uint32(b>>3)
			bd := buckets[key]
			if bd == nil {
				bd = &binData{}
				buckets[key] = bd
			}
			bd.rSum += int64(r)
			bd.gSum += int64(g)
			bd.bSum += int64(b)
			bd.count++
			totalSampled++
		}
	}

	if totalSampled == 0 {
		return nil, fmt.Errorf("image contains no opaque pixels")
	}

	// 3. Collect cluster averages
	type rawCluster struct {
		r, g, b uint8
		count   int
	}
	clusters := make([]rawCluster, 0, len(buckets))
	for _, bd := range buckets {
		cnt := bd.count
		if cnt == 0 {
			continue
		}
		clusters = append(clusters, rawCluster{
			r:     uint8(bd.rSum / int64(cnt)),
			g:     uint8(bd.gSum / int64(cnt)),
			b:     uint8(bd.bSum / int64(cnt)),
			count: cnt,
		})
	}

	// Sort by population descending
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].count > clusters[j].count
	})

	// 4. Merge visually indistinguishable clusters (threshold: distance < 28)
	const mergeThresholdSqr = 28.0 * 28.0
	merged := make([]rawCluster, 0, 16)
	for _, c := range clusters {
		isDuplicate := false
		for i := range merged {
			if colorDistSqr(c.r, c.g, c.b, merged[i].r, merged[i].g, merged[i].b) < mergeThresholdSqr {
				merged[i].count += c.count
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			merged = append(merged, c)
			if len(merged) >= 16 {
				break
			}
		}
	}

	// Re-sort merged by population
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].count > merged[j].count
	})

	// 5. Build ExtractedColor list
	dominantList := make([]ExtractedColor, 0, len(merged))
	for _, m := range merged {
		hex := formatHex(m.r, m.g, m.b)
		lum := RelativeLuminance(hex)
		hue, sat, _ := rgbToHsl(m.r, m.g, m.b)
		pct := math.Round((float64(m.count)/float64(totalSampled))*1000.0) / 10.0

		dominantList = append(dominantList, ExtractedColor{
			Hex:        hex,
			R:          m.r,
			G:          m.g,
			B:          m.b,
			Luminance:  lum,
			Chroma:     sat,
			Hue:        hue,
			Population: m.count,
			Percent:    pct,
			IsDark:     lum < 0.22,
		})
	}

	// Limit dominant swatches to top 8
	if len(dominantList) > 8 {
		dominantList = dominantList[:8]
	}

	// 6. Intelligent 60-30-10 Role Selection
	accentHex := pickBestAccent(dominantList)
	primaryHex := pickBestPrimary(dominantList, accentHex)
	surfaceHex := pickBestSurface(dominantList, primaryHex)

	suggestedTheme := matchSuggestedTheme(accentHex, primaryHex, surfaceHex)
	suggestedFlavor := "soft"
	if strings.Contains(suggestedTheme, "legal") {
		suggestedFlavor = "corporate"
	}

	// 7. Curated Palettes (4 ready presets)
	palettes := buildCuratedPalettes(dominantList, accentHex, primaryHex, surfaceHex)

	return &ImageThemePalette{
		DominantColors:  dominantList,
		Accent:          accentHex,
		Primary:         primaryHex,
		Surface:         surfaceHex,
		SuggestedTheme:  suggestedTheme,
		SuggestedFlavor: suggestedFlavor,
		Palettes:        palettes,
	}, nil
}

// pickBestAccent selects the most chromatic / punchy brand color.
func pickBestAccent(colors []ExtractedColor) string {
	if len(colors) == 0 {
		return "#2563eb"
	}

	var bestHex string
	bestScore := -1.0

	for _, c := range colors {
		// Score based on saturation (chroma) and balanced luminance (avoid pure black or pure white)
		if c.Luminance > 0.03 && c.Luminance < 0.88 {
			// Favor colors with high saturation
			score := c.Chroma * 3.0
			// Add weight for significant pixel presence (up to +1.0)
			score += math.Min(1.0, c.Percent/10.0)
			if score > bestScore {
				bestScore = score
				bestHex = c.Hex
			}
		}
	}

	if bestHex == "" {
		// Fallback to most frequent non-extreme color
		for _, c := range colors {
			if c.Luminance > 0.02 && c.Luminance < 0.96 {
				return c.Hex
			}
		}
		return colors[0].Hex
	}

	return bestHex
}

// pickBestPrimary selects a deep, readable text color (contrast with light background >= 4.5:1).
func pickBestPrimary(colors []ExtractedColor, accentHex string) string {
	// Look for a dark color among extracted colors
	var bestDark string
	lowestLum := 1.0

	for _, c := range colors {
		if c.Hex == accentHex {
			continue
		}
		if c.Luminance < 0.15 && c.Luminance < lowestLum {
			lowestLum = c.Luminance
			bestDark = c.Hex
		}
	}

	if bestDark != "" && lowestLum <= 0.12 {
		return bestDark
	}

	// If no natural deep dark color in image, tint a deep charcoal (#0f172a / #18181b)
	// with the hue of the accent color for cohesive harmony
	lumAccent := RelativeLuminance(accentHex)
	_ = lumAccent
	hue, _, _ := rgbToHslFromHex(accentHex)
	r, g, b := hslToRgb(hue, 0.25, 0.10) // 10% lightness, 25% saturation
	return formatHex(r, g, b)
}

// pickBestSurface selects an elegant clean background (60%).
func pickBestSurface(colors []ExtractedColor, primaryHex string) string {
	// Look for a very light neutral tone in the image
	for _, c := range colors {
		if c.Luminance > 0.88 && c.Chroma < 0.20 {
			return c.Hex
		}
	}
	// Default to crisp white
	return "#ffffff"
}

// rgbToHslFromHex parses hex and returns H, S, L.
func rgbToHslFromHex(hex string) (float64, float64, float64) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return 0, 0, 0
	}
	var r, g, b uint8
	if _, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b); err != nil {
		return 0, 0, 0
	}
	return rgbToHsl(r, g, b)
}

// matchSuggestedTheme maps colors to one of the built-in SiteCraft CSS theme presets.
func matchSuggestedTheme(accentHex, primaryHex, surfaceHex string) string {
	hue, sat, _ := rgbToHslFromHex(accentHex)
	surfLum := RelativeLuminance(surfaceHex)

	if surfLum < 0.2 {
		return "main-night"
	}

	switch {
	case hue >= 345 || hue <= 25:
		if sat > 0.4 {
			return "coral"
		}
		return "beauty"
	case hue > 25 && hue <= 50:
		return "auto"
	case hue > 50 && hue <= 165:
		return "medical"
	case hue > 165 && hue <= 205:
		return "cyber"
	case hue > 205 && hue <= 255:
		return "coral"
	case hue > 255 && hue <= 345:
		return "beauty"
	default:
		return "main"
	}
}

// buildCuratedPalettes generates 4 distinct ready-to-use schemes.
func buildCuratedPalettes(dominant []ExtractedColor, accent, primary, surface string) []CuratedPalette {
	accentHue, accentSat, _ := rgbToHslFromHex(accent)

	// 1. Balanced 60-30-10
	palettes := []CuratedPalette{
		{
			ID:           "balanced",
			Name:         "Сбалансированная (60-30-10)",
			Description:  "Чистый фон, четкий читаемый текст и выразительный брендовый акцент",
			Accent:       accent,
			Primary:      primary,
			Surface:      surface,
			Theme:        matchSuggestedTheme(accent, primary, surface),
			DesignFlavor: "soft",
		},
	}

	// 2. Vibrant / Bold
	vibrantAccent := accent
	for _, c := range dominant {
		if c.Chroma > accentSat && c.Luminance > 0.15 && c.Luminance < 0.75 {
			vibrantAccent = c.Hex
			break
		}
	}
	palettes = append(palettes, CuratedPalette{
		ID:           "vibrant",
		Name:         "Яркая / Брендовая",
		Description:  "Максимальная сочность цветов для технологичных и динамичных проектов",
		Accent:       vibrantAccent,
		Primary:      primary,
		Surface:      "#ffffff",
		Theme:        "coral",
		DesignFlavor: "sharp",
	})

	// 3. Editorial / Luxury
	warmR, warmG, warmB := hslToRgb(40, 0.15, 0.98) // Warm ivory parchment
	textR, textG, textB := hslToRgb(accentHue, 0.12, 0.11)
	palettes = append(palettes, CuratedPalette{
		ID:           "editorial",
		Name:         "Мягкая / Редакционная",
		Description:  "Теплый благородный фон с утонченным контрастом в стиле премиум-журналов",
		Accent:       accent,
		Primary:      formatHex(textR, textG, textB),
		Surface:      formatHex(warmR, warmG, warmB),
		Theme:        "beauty",
		DesignFlavor: "corporate",
	})

	// 4. Dark Mode / Night
	darkR, darkG, darkB := hslToRgb(accentHue, 0.20, 0.06) // Deep midnight tinted with brand hue
	textLightR, textLightG, textLightB := hslToRgb(accentHue, 0.05, 0.96)
	// Brightened accent for dark backgrounds
	brightAccR, brightAccG, brightAccB := hslToRgb(accentHue, math.Max(0.7, accentSat), 0.58)

	palettes = append(palettes, CuratedPalette{
		ID:           "dark",
		Name:         "Глубокая ночная (Dark)",
		Description:  "Премиальный темный фон с неоновым свечением акцентных элементов",
		Accent:       formatHex(brightAccR, brightAccG, brightAccB),
		Primary:      formatHex(textLightR, textLightG, textLightB),
		Surface:      formatHex(darkR, darkG, darkB),
		Theme:        "main-night",
		DesignFlavor: "soft",
	})

	return palettes
}
