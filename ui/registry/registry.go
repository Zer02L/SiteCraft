package registry

import (
	"sitecraft/internal/models"
	"sitecraft/ui/components"
	"sitecraft/ui/components/cinematic"
	"sitecraft/ui/components/cinematic_2"
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
	Resources   []models.ResourceItem
	Reviews     []models.ReviewItem
	Branches    []models.BranchItem
	Vacancies   []models.VacancyItem
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
			case "legal":
				return hero.LegalHero(ctx.Content, ctx.Settings)
			case "classic":
				return hero.Hero(ctx.Content)
			}
		}
		if ctx.Settings.Theme == "cyber" {
			return hero.CyberHero(ctx.Content, ctx.Settings)
		}
		if ctx.Settings.Theme == "legal" {
			return hero.LegalHero(ctx.Content, ctx.Settings)
		}
		return hero.Hero(ctx.Content)
	},
	"cyber_hero": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return hero.CyberHero(ctx.Content, ctx.Settings)
	},
	"legal_hero": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return hero.LegalHero(ctx.Content, ctx.Settings)
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
		return sections.Booking(ctx.Content, ctx.BookedSlots, ctx.Settings, ctx.Resources)
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
		return sections.Reviews(ctx.Settings.Theme, ctx.Reviews)
	},
	"faq": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.Faq(ctx.Settings.Theme, ParseFAQPayload(b.Payload))
	},
	"gallery": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.Gallery(ctx.Settings.Theme)
	},
	"map": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.MapSection(ctx.Content, ctx.Settings, ctx.Branches)
	},
	"branches": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.MapSection(ctx.Content, ctx.Settings, ctx.Branches)
	},
	"careers": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.CareersSection(ctx.Vacancies, ctx.Settings)
	},
	"vacancies": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.CareersSection(ctx.Vacancies, ctx.Settings)
	},
	"quiz": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.Quiz(ParseQuizPayload(b.Payload, ctx.Settings.Theme), ctx.Settings)
	},
	"calculator": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.Quiz(ParseQuizPayload(b.Payload, ctx.Settings.Theme), ctx.Settings)
	},
	"before_after": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.BeforeAfter(ParseBeforeAfterPayload(b.Payload, ctx.Settings.Theme), ctx.Settings)
	},
	"comparison": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return sections.BeforeAfter(ParseBeforeAfterPayload(b.Payload, ctx.Settings.Theme), ctx.Settings)
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
	"cinematic_odometer": func(b models.PageBlock, ctx RenderContext) templ.Component {
		metrics := []cinematic.StatMetric{
			{Value: 98, Suffix: "%", Label: "Выигранных дел", Description: "Успешные исходы в судах всех инстанций"},
			{Value: 16, Suffix: "+", Label: "Лет непрерывной практики", Description: "Член Адвокатской палаты РФ"},
			{Value: 2400, Suffix: " млн ₽", Label: "Сохраненных активов", Description: "Защищенный капитал доверителей"},
			{Value: 1250, Suffix: "+", Label: "Закрытых споров", Description: "В пользу наших клиентов"},
		}
		return cinematic.OdometerStats(metrics)
	},
	"cinematic_spotlight": func(b models.PageBlock, ctx RenderContext) templ.Component {
		title := b.Title
		if title == "" {
			title = "Ключевые направления юридической защиты"
		}
		items := []cinematic.SpotlightCardItem{
			{
				Icon:        "⚖️",
				Tag:         "Арбитраж",
				Title:       "Сложные коммерческие споры",
				Description: "Взыскание задолженностей, расторжение кабальных договоров, споры между учредителями и защита репутации компании.",
				Link:        "#booking",
				LinkText:    "Консультация",
			},
			{
				Icon:        "🏢",
				Tag:         "Бизнес",
				Title:       "Защита при проверках ФНС и МВД",
				Description: "Сопровождение налоговых проверок, оперативных мероприятий, защита от субсидиарной ответственности руководителей.",
				Link:        "#booking",
				LinkText:    "Консультация",
			},
			{
				Icon:        "📜",
				Tag:         "Сделки",
				Title:       "Правовой аудит & M&A (Due Diligence)",
				Description: "Комплексная проверка чистоты активов перед покупкой, структурирование корпоративных сделок любой сложности.",
				Link:        "#booking",
				LinkText:    "Консультация",
			},
			{
				Icon:        "🛡️",
				Tag:         "Активы",
				Title:       "Банкротство и реструктуризация",
				Description: "Защита активов собственников от необоснованных претензий кредиторов, законное списание обязательств юрлиц.",
				Link:        "#booking",
				LinkText:    "Консультация",
			},
			{
				Icon:        "🏛️",
				Tag:         "Недвижимость",
				Title:       "Земельные и имущественные споры",
				Description: "Оспаривание кадастровой стоимости, защита права собственности на коммерческие объекты, градостроительные конфликты.",
				Link:        "#booking",
				LinkText:    "Консультация",
			},
			{
				Icon:        "💼",
				Tag:         "Уголовно-правовая защита",
				Title:       "Экономические преступления (ст. 159, 199)",
				Description: "Адвокатская защита первых лиц компании на всех стадиях доследственной проверки, следствия и судебного процесса.",
				Link:        "#booking",
				LinkText:    "Консультация",
			},
		}
		return cinematic.SpotlightBorder(title, "Практики бюро", items)
	},
	"cinematic_sticky_stack": func(b models.PageBlock, ctx RenderContext) templ.Component {
		title := b.Title
		if title == "" {
			title = "Стратегия защиты ваших интересов"
		}
		steps := []cinematic.StickyFeatureItem{
			{
				Number:      "01",
				Title:       "Правовой аудит и оценка судебных перспектив",
				Description: "Глубокий анализ документов доверителя, сопоставление с актуальной практикой Верховного Суда РФ и честный расчет шансов на победу.",
			},
			{
				Number:      "02",
				Title:       "Формирование доказательственной базы",
				Description: "Сбор неопровержимых фактов, адвокатские запросы, назначение судебных экспертиз и проведение досудебного урегулирования.",
			},
			{
				Number:      "03",
				Title:       "Личное процессуальное представительство",
				Description: "Бескомпромиссная защита позиции в заседаниях первой, апелляционной и кассационной инстанций арбитражных судов и судов общей юрисдикции.",
			},
			{
				Number:      "04",
				Title:       "Реальное исполнение судебного акта",
				Description: "Фактическое взыскание средств через ФССП и банковские счета, блокировка вывода активов оппонентов до полного исполнения.",
			},
		}
		return cinematic.StickyStack(title, "Порядок работы", steps)
	},
	"cinematic_sticky_cards": func(b models.PageBlock, ctx RenderContext) templ.Component {
		title := b.Title
		if title == "" {
			title = "Форматы юридического сопровождения"
		}
		cards := []cinematic.StackCardItem{
			{
				Badge:       "Судебный",
				Title:       "Разовое ведение дела в суде",
				Description: "Полный комплекс услуг под ключ: от составления искового заявления до получения исполнительного листа и победы в суде.",
				Tags:        []string{"Арбитраж", "Суды РФ", "Все инстанции"},
				StatNumber:  "от 75 000 ₽",
				StatLabel:   "+ % за успех",
			},
			{
				Badge:       "Для бизнеса",
				Title:       "Абонентское обслуживание (Аутсорсинг)",
				Description: "Заменяет целый юридический отдел. Экспертиза договоров, участие в переговорах, претензионная работа и проверка контрагентов 24/7.",
				Tags:        []string{"Неограниченно", "Выезд адвоката", "Личный куратор"},
				StatNumber:  "от 50 000 ₽",
				StatLabel:   "в месяц",
			},
			{
				Badge:       "VIP Защита",
				Title:       "Персональный правовой щит собственника",
				Description: "Экстренный выезд при следственных действиях и проверках, превентивная защита личных активов, налоговый комплаенс и конфиденциальность.",
				Tags:        []string{"24/7 выезд", "Адвокатская тайна", "Security"},
				StatNumber:  "от 150 000 ₽",
				StatLabel:   "индивидуально",
			},
		}
		return cinematic.StickyCards(title, "Тарифы и пакеты", cards)
	},
	"cinematic_marquee": func(b models.PageBlock, ctx RenderContext) templ.Component {
		items := []string{
			"АРБИТРАЖНЫЙ СУД", "ЗАЩИТА БИЗНЕСА", "НАЛОГОВЫЕ СПОРЫ", "БАНКРОТСТВО",
			"M&A СДЕЛКИ", "НЕДВИЖИМОСТЬ И ЗЕМЛЯ", "УГОЛОВНО-ПРАВОВАЯ ЗАЩИТА",
		}
		return cinematic.KineticMarquee(items, 25, false)
	},
	"cinematic_2_hero": func(b models.PageBlock, ctx RenderContext) templ.Component {
		props := cinematic_2.DefaultHeroProps()
		if b.Title != "" {
			props.BrandName = b.Title
		}
		if sub, ok := b.Payload["subtitle"].(string); ok && sub != "" {
			props.Subtitle = sub
		} else if b.Content != "" {
			props.Subtitle = b.Content
		}
		return cinematic_2.HeroScreen(props)
	},
	"cinematic_2_portfolio": func(b models.PageBlock, ctx RenderContext) templ.Component {
		props := cinematic_2.DefaultCollageProps()
		if b.Title != "" {
			props.Title = b.Title
		}
		return cinematic_2.PortfolioCollage(props)
	},
	"cinematic_2_showcase": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return cinematic_2.DefaultShowcase()
	},
	"jhill_hero": func(b models.PageBlock, ctx RenderContext) templ.Component {
		props := cinematic_2.DefaultHeroProps()
		if b.Title != "" {
			props.BrandName = b.Title
		}
		if sub, ok := b.Payload["subtitle"].(string); ok && sub != "" {
			props.Subtitle = sub
		} else if b.Content != "" {
			props.Subtitle = b.Content
		}
		return cinematic_2.HeroScreen(props)
	},
	"jhill_portfolio": func(b models.PageBlock, ctx RenderContext) templ.Component {
		props := cinematic_2.DefaultCollageProps()
		if b.Title != "" {
			props.Title = b.Title
		}
		return cinematic_2.PortfolioCollage(props)
	},
	"jhill_showcase": func(b models.PageBlock, ctx RenderContext) templ.Component {
		return cinematic_2.DefaultShowcase()
	},
}

// BlockSourceMap maps block type to source .templ file for IDE element inspect
var BlockSourceMap = map[string]string{
	"hero":                      "ui/sections/hero/hero.templ",
	"cyber_hero":                "ui/sections/hero/cyber_hero.templ",
	"legal_hero":                "ui/sections/hero/legal_hero.templ",
	"split_hero":                "ui/sections/hero/split_hero.templ",
	"schedule":                  "ui/sections/schedule.templ",
	"services":                  "ui/sections/services.templ",
	"booking":                   "ui/sections/booking.templ",
	"features":                  "ui/sections/features.templ",
	"team":                      "ui/sections/team.templ",
	"steps":                     "ui/sections/steps.templ",
	"reviews":                   "ui/sections/reviews.templ",
	"faq":                       "ui/sections/faq.templ",
	"gallery":                   "ui/sections/gallery.templ",
	"map":                       "ui/sections/map.templ",
	"branches":                  "ui/sections/map.templ",
	"careers":                   "ui/sections/careers.templ",
	"vacancies":                 "ui/sections/careers.templ",
	"quiz":                      "ui/sections/quiz.templ",
	"calculator":                "ui/sections/quiz.templ",
	"before_after":              "ui/sections/before_after.templ",
	"comparison":                "ui/sections/before_after.templ",
	"showcase":                  "ui/components/ui_showcase.templ",
	"rich_text":                 "ui/components/dynamic_block.templ",
	"html":                      "ui/components/dynamic_block.templ",
	"cinematic_odometer":        "ui/components/cinematic/odometer.templ",
	"cinematic_spotlight":       "ui/components/cinematic/spotlight_border.templ",
	"cinematic_split_scroll":    "ui/components/cinematic/split_scroll.templ",
	"cinematic_accordion":       "ui/components/cinematic/accordion_slider.templ",
	"cinematic_horizontal":      "ui/components/cinematic/horizontal_scroll.templ",
	"cinematic_mesh":            "ui/components/cinematic/mesh_gradient.templ",
	"cinematic_coverflow":       "ui/components/cinematic/coverflow.templ",
	"cinematic_sticky_cards":    "ui/components/cinematic/sticky_cards.templ",
	"cinematic_image_reveal":    "ui/components/cinematic/image_reveal.templ",
	"cinematic_cursor":          "ui/components/cinematic/cursor_reactive.templ",
	"cinematic_curtain":         "ui/components/cinematic/curtain_reveal.templ",
	"cinematic_dock":            "ui/components/cinematic/dock_nav.templ",
	"cinematic_marquee":         "ui/components/cinematic/kinetic_marquee.templ",
	"cinematic_magnetic":        "ui/components/cinematic/magnetic_grid.templ",
	"cinematic_particles":       "ui/components/cinematic/particle_button.templ",
	"cinematic_scramble":        "ui/components/cinematic/text_scramble.templ",
	"cinematic_sticky_stack":    "ui/components/cinematic/sticky_stack.templ",
	"cinematic_svg_draw":        "ui/components/cinematic/svg_draw.templ",
	"cinematic_typewriter":      "ui/components/cinematic/typewriter.templ",
	"cinematic_color_shift":     "ui/components/cinematic/color_shift.templ",
	"cinematic_dynamic_island":  "ui/components/cinematic/dynamic_island.templ",
	"cinematic_flat_to_3d":      "ui/components/cinematic/flat_to_3d.templ",
	"cinematic_image_trail":     "ui/components/cinematic/image_trail.templ",
	"cinematic_ogl_waves":       "ui/components/cinematic/ogl_waves.templ",
	"cinematic_text_mask":       "ui/components/cinematic/text_mask.templ",
	"cinematic_view_transitions": "ui/components/cinematic/view_transitions.templ",
	"cinematic_zoom_parallax":   "ui/components/cinematic/zoom_parallax.templ",
	"jhill_hero":                "ui/components/cinematic_2/hero_screen.templ",
	"jhill_collage":             "ui/components/cinematic_2/portfolio_collage.templ",
	"jhill_showcase":            "ui/components/cinematic_2/showcase.templ",
}

// ResolveBlockSource determines the exact .templ source file for a given block
func ResolveBlockSource(b models.PageBlock, ctx RenderContext) string {
	if b.Type == "hero" {
		if variant, ok := b.Payload["variant"].(string); ok {
			switch variant {
			case "split":
				return "ui/sections/hero/split_hero.templ"
			case "cyber":
				return "ui/sections/hero/cyber_hero.templ"
			case "legal":
				return "ui/sections/hero/legal_hero.templ"
			case "classic":
				return "ui/sections/hero/hero.templ"
			}
		}
		if ctx.Settings.Theme == "cyber" {
			return "ui/sections/hero/cyber_hero.templ"
		}
		if ctx.Settings.Theme == "legal" {
			return "ui/sections/hero/legal_hero.templ"
		}
		return "ui/sections/hero/hero.templ"
	}
	if src, ok := BlockSourceMap[b.Type]; ok {
		return src
	}
	return "ui/components/dynamic_block.templ"
}

// RegisterBlock allows registering custom block types dynamically
func RegisterBlock(blockType string, renderer BlockRenderer) {
	BlockRegistry[blockType] = renderer
}

// RenderBlock executes registered renderer or fallbacks to dynamic block,
// wrapping it in BlockWrapper (display: contents) to provide data-block and data-source DOM context.
func RenderBlock(b models.PageBlock, ctx RenderContext) templ.Component {
	var comp templ.Component
	if renderer, ok := BlockRegistry[b.Type]; ok {
		comp = renderer(b, ctx)
	} else {
		comp = components.DynamicBlock(b)
	}
	src := ResolveBlockSource(b, ctx)
	return BlockWrapper(b.Type, src, comp)
}
