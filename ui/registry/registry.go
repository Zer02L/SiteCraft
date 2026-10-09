package registry

import (
	"sitecraft/internal/models"
	"sitecraft/ui/components"
	"sitecraft/ui/components/cinematic"
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
