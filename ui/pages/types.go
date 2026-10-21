package pages

import (
	"sitecraft/internal/models"
	"sitecraft/ui/registry"
)

// HomePageData encapsulates all inputs required for rendering the Home page.
type HomePageData struct {
	Settings    models.SiteSettings
	Content     models.SiteContent
	Services    []models.ServiceItem
	BookedSlots map[string][]string
	Blocks      []models.PageBlock
	Resources   []models.ResourceItem
	Reviews     []models.ReviewItem
	Branches    []models.BranchItem
	Vacancies   []models.VacancyItem
}

// RenderContext converts HomePageData into registry.RenderContext for block rendering.
func (d HomePageData) RenderContext() registry.RenderContext {
	return registry.RenderContext{
		Settings:    d.Settings,
		Content:     d.Content,
		Services:    d.Services,
		BookedSlots: d.BookedSlots,
		Resources:   d.Resources,
		Reviews:     d.Reviews,
		Branches:    d.Branches,
		Vacancies:   d.Vacancies,
	}
}
