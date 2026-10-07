package pages

import (
	"sitecraft/internal/models"
	"sitecraft/ui/components"
	"sitecraft/ui/sections"
	"sitecraft/ui/sections/hero"

	"github.com/a-h/templ"
)

// RenderContext contains all state required by block renderers
type RenderContext struct {
	Settings    models.SiteSettings
	Content     models.SiteContent
	Services    []models.ServiceItem
	BookedSlots map[string][]string
}

// BlockRenderer defines the function signature for a pluggable section renderer
type BlockRenderer func(block models.PageBlock, ctx RenderContext) templ.Component

// BlockRegistry maps block type string identifier to its pluggable renderer
var BlockRegistry = map[string]BlockRenderer{
	"hero": func(b models.PageBlock, ctx RenderContext) templ.Component {
		if variant, ok := b.Payload["variant"].(string); ok {
			switch variant {
			case "split":
				return hero.SplitHero(ctx.Content, ctx.Settings)
			case "cyber":
				return hero.CyberHero(ctx.Content, ctx.Settings)
			case "classic":
				return hero.Hero(ctx.Content)
			}
		}
		if ctx.Settings.Theme == "cyber" {
			return hero.CyberHero(ctx.Content, ctx.Settings)
		}
		return hero.Hero(ctx.Content)
	},
	"cyber_hero": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return hero.CyberHero(ctx.Content, ctx.Settings)
	},
	"split_hero": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return hero.SplitHero(ctx.Content, ctx.Settings)
	},
	"schedule": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.ScheduleBar(ctx.Content)
	},
	"services": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.Services(ctx.Services, ctx.Settings)
	},
	"booking": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.Booking(ctx.Content, ctx.BookedSlots, ctx.Settings)
	},
	"features": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.Features(ctx.Settings.Theme, ParseFeaturesPayload(b.Payload))
	},
	"team": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.Team(ctx.Settings.Theme, ctx.Content)
	},
	"steps": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.Steps(ctx.Settings.Theme)
	},
	"reviews": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.Reviews(ctx.Settings.Theme)
	},
	"faq": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.Faq(ctx.Settings.Theme, ParseFAQPayload(b.Payload))
	},
	"gallery": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.Gallery(ctx.Settings.Theme)
	},
	"map": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.MapSection(ctx.Content, ctx.Settings)
	},
	"showcase": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return components.UIShowcase()
	},
	"rich_text": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return components.DynamicBlock(b)
	},
	"html": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return components.DynamicBlock(b)
	},
}

// RegisterBlock allows registering custom block types dynamically
func RegisterBlock(blockType string, renderer BlockRenderer) {
	BlockRegistry[blockType] = renderer
}

// RenderBlock executes registered renderer or fallbacks to dynamic block
func RenderBlock(b models.PageBlock, ctx RenderContext) templ.Component {
	if renderer, ok := BlockRegistry[b.Type]; ok {
		return renderer(b, ctx)
	}
	return components.DynamicBlock(b)
}
