package theme

import (
	"fmt"
	"math"
	"strings"
)

// ThemePreset represents a design system preset with typography, color tokens, and geometry flavor.
type ThemePreset struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Category        string `json:"category"` // "global", "industry"
	Theme           string `json:"theme"`
	PrimaryColor    string `json:"primary_color"`
	AccentColor     string `json:"accent_color"`
	BackgroundColor string `json:"background_color"` // format: "flavor|#hex" or "#hex"
	FontFamily      string `json:"font_family"`      // "modern", "editorial", "condensed", "tech"
	DesignFlavor    string `json:"design_flavor"`    // "soft", "corporate", "sharp", "brutalist"
}

// BuiltinPresets holds the 4 global styles and key industry presets.
var BuiltinPresets = []ThemePreset{
	// 4 Key Global Styles
	{
		ID:              "modern_bright",
		Name:            "Modern Bright",
		Description:     "Свежий, технологичный стиль с мягкими скруглениями и яркими акцентами",
		Category:        "global",
		Theme:           "coral",
		PrimaryColor:    "#0f172a",
		AccentColor:     "#2563eb",
		BackgroundColor: "soft|#ffffff",
		FontFamily:      "modern",
		DesignFlavor:    "soft",
	},
	{
		ID:              "luxury_editorial",
		Name:            "Luxury Editorial",
		Description:     "Премиальный редакционный стиль с утонченной засечечной типографикой",
		Category:        "global",
		Theme:           "beauty",
		PrimaryColor:    "#1c1917",
		AccentColor:     "#b45309",
		BackgroundColor: "corporate|#fdfbf7",
		FontFamily:      "editorial",
		DesignFlavor:    "corporate",
	},
	{
		ID:              "dark_premium",
		Name:            "Dark Premium",
		Description:     "Глубокая ночная тема с неоновыми статусными акцентами для хай-тек и премиума",
		Category:        "global",
		Theme:           "main-night",
		PrimaryColor:    "#f8fafc",
		AccentColor:     "#38bdf8",
		BackgroundColor: "soft|#090d16",
		FontFamily:      "tech",
		DesignFlavor:    "soft",
	},
	{
		ID:              "swiss_bold",
		Name:            "Swiss Bold",
		Description:     "Строгий швейцарский интернациональный стиль, нулевые радиусы, строгая сетка",
		Category:        "global",
		Theme:           "legal-swiss",
		PrimaryColor:    "#000000",
		AccentColor:     "#dc2626",
		BackgroundColor: "sharp|#ffffff",
		FontFamily:      "condensed",
		DesignFlavor:    "sharp",
	},

	// Industry Presets
	{
		ID:              "b2b_legal",
		Name:            "B2B Юриспруденция и Консалтинг",
		Description:     "Сдержанный деловой тон, доверие и авторитет",
		Category:        "industry",
		Theme:           "legal",
		PrimaryColor:    "#0f172a",
		AccentColor:     "#c2410c",
		BackgroundColor: "corporate|#f8fafc",
		FontFamily:      "editorial",
		DesignFlavor:    "corporate",
	},
	{
		ID:              "b2c_medical",
		Name:            "B2C Медицина и Здоровье",
		Description:     "Стерильный, чистый и спокойный дизайн, вызывающий доверие пациентов",
		Category:        "industry",
		Theme:           "medical",
		PrimaryColor:    "#0f766e",
		AccentColor:     "#0284c7",
		BackgroundColor: "soft|#f0fdfa",
		FontFamily:      "modern",
		DesignFlavor:    "soft",
	},
	{
		ID:              "b2c_auto",
		Name:            "B2C Автосервис и Детейлинг",
		Description:     "Контрастный динамичный стиль с акцентом на скорость и надежность",
		Category:        "industry",
		Theme:           "auto",
		PrimaryColor:    "#18181b",
		AccentColor:     "#ea580c",
		BackgroundColor: "sharp|#ffffff",
		FontFamily:      "condensed",
		DesignFlavor:    "sharp",
	},
	{
		ID:              "b2b_cyber",
		Name:            "B2B Кибербезопасность и IT",
		Description:     "Футуристичный высокотехнологичный стиль с темным фоном и бирюзовым неоном",
		Category:        "industry",
		Theme:           "cyber",
		PrimaryColor:    "#f1f5f9",
		AccentColor:     "#06b6d4",
		BackgroundColor: "soft|#0a0f1d",
		FontFamily:      "tech",
		DesignFlavor:    "soft",
	},
}

// GetPreset finds a preset by its ID.
func GetPreset(id string) (ThemePreset, bool) {
	for _, p := range BuiltinPresets {
		if p.ID == id {
			return p, true
		}
	}
	return ThemePreset{}, false
}

// srgbToLinear converts sRGB channel (0..1) to linear luminance component.
func srgbToLinear(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// RelativeLuminance calculates standard WCAG relative luminance for a hex color.
func RelativeLuminance(hex string) float64 {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return 0
	}
	var r, g, b uint8
	if _, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b); err != nil {
		return 0
	}
	rL := srgbToLinear(float64(r) / 255.0)
	gL := srgbToLinear(float64(g) / 255.0)
	bL := srgbToLinear(float64(b) / 255.0)
	return 0.2126*rL + 0.7152*gL + 0.0722*bL
}

// IsLightHex checks whether a color hex is perceptually light (relative luminance > 0.22).
func IsLightHex(hex string) bool {
	return RelativeLuminance(hex) > 0.22
}

// ContrastTextColor returns appropriate readable button text color ("#0f172a" or "#ffffff")
// based on relative luminance of the background hex.
func ContrastTextColor(bgHex string) string {
	if IsLightHex(bgHex) {
		return "#0f172a"
	}
	return "#ffffff"
}

// GenerateCSSVariables produces inline CSS custom properties for zero-rebuild theme injection.
func GenerateCSSVariables(p ThemePreset) string {
	var sb strings.Builder
	sb.WriteString(":root { ")

	if p.PrimaryColor != "" {
		sb.WriteString(fmt.Sprintf("--color-primary: %s; ", p.PrimaryColor))
		sb.WriteString(fmt.Sprintf("--color-content: %s; ", p.PrimaryColor))
		sb.WriteString(fmt.Sprintf("--color-content-muted: color-mix(in srgb, %s 65%%, transparent); ", p.PrimaryColor))
		sb.WriteString("--color-primary-dark: color-mix(in srgb, var(--color-primary), black 30%); ")
		sb.WriteString("--color-primary-light: color-mix(in srgb, var(--color-primary), white 25%); ")
		sb.WriteString(fmt.Sprintf("--color-on-primary: %s; ", ContrastTextColor(p.PrimaryColor)))
	}

	if p.AccentColor != "" {
		sb.WriteString(fmt.Sprintf("--color-accent: %s; ", p.AccentColor))
		sb.WriteString("--color-accent-hover: color-mix(in srgb, var(--color-accent), black 15%); ")
		sb.WriteString("--color-accent-light: color-mix(in srgb, var(--color-accent) 12%, white); ")
		sb.WriteString(fmt.Sprintf("--color-on-accent: %s; ", ContrastTextColor(p.AccentColor)))
	}

	// Surface / background
	bg := p.BackgroundColor
	if strings.Contains(bg, "|") {
		parts := strings.SplitN(bg, "|", 2)
		bg = parts[1]
	}
	if bg != "" {
		sb.WriteString(fmt.Sprintf("--color-surface: %s; ", bg))
		sb.WriteString(fmt.Sprintf("--color-surface-muted: color-mix(in srgb, var(--color-primary, #000) 4%%, %s); ", bg))
	}

	// Geometry radius tokens based on design flavor
	switch p.DesignFlavor {
	case "sharp", "brutalist":
		sb.WriteString("--radius-card: 0px; --radius-card-sm: 0px; --radius-btn: 0px; --radius-input: 0px; --radius-badge: 0px; ")
	case "corporate":
		sb.WriteString("--radius-card: 0.5rem; --radius-card-sm: 0.375rem; --radius-btn: 0.375rem; --radius-input: 0.375rem; --radius-badge: 9999px; ")
	case "soft":
		fallthrough
	default:
		sb.WriteString("--radius-card: 1.5rem; --radius-card-sm: 1rem; --radius-btn: 0.75rem; --radius-input: 0.75rem; --radius-badge: 9999px; ")
	}

	// Typography font pair
	switch p.FontFamily {
	case "editorial":
		sb.WriteString(`--font-serif-title: "Merriweather", Georgia, serif; `)
	case "condensed":
		sb.WriteString(`--font-serif-title: "Bebas Neue", "Barlow Condensed", Impact, sans-serif; `)
	case "tech":
		sb.WriteString(`--font-serif-title: "Plus Jakarta Sans", "Inter", sans-serif; `)
	case "modern":
		fallthrough
	default:
		sb.WriteString(`--font-serif-title: "Inter", system-ui, sans-serif; `)
	}

	sb.WriteString("}")
	return sb.String()
}
