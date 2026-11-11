package layouts

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"

	"sitecraft/internal/models"

	"github.com/a-h/templ"
)

func srgbToLinear(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// isLightHex checks if a color hex is perceptually light (relative luminance > 0.22)
func isLightHex(hex string) bool {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return false
	}
	var r, g, b uint8
	if _, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b); err != nil {
		return false
	}
	rL := srgbToLinear(float64(r) / 255.0)
	gL := srgbToLinear(float64(g) / 255.0)
	bL := srgbToLinear(float64(b) / 255.0)
	lum := 0.2126*rL + 0.7152*gL + 0.0722*bL
	return lum > 0.22
}

// resolveSEO normalizes SEO parameters and merges defaults.
func resolveSEO(title, description string, settings models.SiteSettings, seoOpt ...models.SEOData) models.SEOData {
	if len(seoOpt) > 0 {
		s := seoOpt[0]
		if s.Title == "" {
			s.Title = title
		}
		if s.Description == "" {
			s.Description = description
		}
		if s.OGType == "" {
			s.OGType = "website"
		}
		if s.TwitterCard == "" {
			s.TwitterCard = "summary_large_image"
		}
		return s
	}

	return models.SEOData{
		Title:       title,
		Description: description,
		OGType:      "website",
		TwitterCard: "summary_large_image",
	}
}

// RenderJSONLD renders a raw application/ld+json script block safely without templ escaping.
func RenderJSONLD(jsonStr string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<script type=\"application/ld+json\">"+jsonStr+"</script>\n")
		return err
	})
}

// RenderCustomTokens generates an inline <style> block containing custom CSS variables from settings.
func RenderCustomTokens(settings models.SiteSettings) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		primary := settings.PrimaryColor
		accent := settings.AccentColor
		bg := settings.GetCustomBgColor()
		if primary == "" && accent == "" && bg == "" {
			return nil
		}
		var css string
		css += "<style id=\"sitecraft-custom-tokens\">:root {"
		if primary != "" {
			css += "--color-primary: " + primary + "; "
			css += "--color-content: " + primary + "; "
			css += "--color-content-muted: color-mix(in srgb, " + primary + " 65%, transparent); "
			css += "--color-primary-dark: color-mix(in srgb, var(--color-primary), black 30%); "
			css += "--color-primary-light: color-mix(in srgb, var(--color-primary), white 25%); "
			if isLightHex(primary) {
				css += "--color-on-primary: #0f172a; "
			} else {
				css += "--color-on-primary: #ffffff; "
			}
		}
		if accent != "" {
			css += "--color-accent: " + accent + "; "
			css += "--color-accent-hover: color-mix(in srgb, var(--color-accent), black 15%); "
			css += "--color-accent-light: color-mix(in srgb, var(--color-accent) 12%, white); "
			if isLightHex(accent) {
				css += "--color-on-accent: #0f172a; "
			} else {
				css += "--color-on-accent: #ffffff; "
			}
		}
		if bg != "" {
			css += "--color-surface: " + bg + "; "
			css += "--color-surface-muted: color-mix(in srgb, var(--color-primary, #000) 4%, " + bg + "); "
		}
		css += "}</style>\n"
		_, err := io.WriteString(w, css)
		return err
	})
}

