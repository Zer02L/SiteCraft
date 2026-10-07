package layouts

import (
	"context"
	"io"
	"sitecraft/internal/models"

	"github.com/a-h/templ"
)

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
