package models

import (
	"strings"
	"time"
)

// SiteSettings represents general site configuration and meta
type SiteSettings struct {
	ID                 string `json:"id"`
	SiteName           string `json:"site_name"`
	Tagline            string `json:"tagline"`
	Theme              string `json:"theme"` // legal, medical, auto, tech, beauty, dark, custom, default
	PrimaryColor       string `json:"primary_color"`
	AccentColor        string `json:"accent_color"`
	BackgroundColor    string `json:"background_color"`
	FontFamily         string `json:"font_family"`
	EnableA11y         bool   `json:"enable_a11y"`
	EnableAlpine       bool   `json:"enable_alpine"`
	EnableTurnstile    bool   `json:"enable_turnstile"`
	TurnstileSiteKey   string `json:"turnstile_site_key"`
	TurnstileSecretKey string `json:"turnstile_secret_key"`
	TelegramBotToken   string `json:"telegram_bot_token"`
	TelegramChatID     string `json:"telegram_chat_id"`
	WebhookURL         string `json:"webhook_url"`
	YandexMetrikaID    string `json:"yandex_metrika_id"`
	MapMode            string `json:"map_mode"`     // lazy, static, iframe
	BookingMode        string `json:"booking_mode"` // simple, slots

	// Модули интеграций (E-commerce / Оплата)
	EnablePayment  bool   `json:"enable_payment"`
	PaymentShopID  string `json:"payment_shop_id"`
	PaymentSecret  string `json:"payment_secret"`
	PaymentSandbox bool   `json:"payment_sandbox"`

	// Модули взаимодействия и конверсии (Квиз, Мульти-мессенджер, Веб-корзина)
	EnableFloatingContact bool   `json:"enable_floating_contact"`
	WhatsappPhone         string `json:"whatsapp_phone"`
	TelegramContact       string `json:"telegram_contact"`
	EnableCart            bool   `json:"enable_cart"`

	// Новые модули (Этап 1: Блог и TMA: Telegram Mini App, Этап 2: Трекер заказов, Локализация, Чат, Fleet, Фаза 3: CRO)
	EnableBlog           bool   `json:"enable_blog"`
	EnableTMA            bool   `json:"enable_tma"`
	EnableTracker        bool   `json:"enable_tracker"`
	EnableI18n           bool   `json:"enable_i18n"`
	EnableLiveChat       bool   `json:"enable_live_chat"`
	EnableFleet          bool   `json:"enable_fleet"`
	EnableRecentActivity bool   `json:"enable_recent_activity"`
	DefaultLang          string `json:"default_lang"` // ru, en

	// Варианты шапки и подвала (Layout Variants: classic, minimal, centered / columns)
	HeaderVariant string `json:"header_variant"`
	FooterVariant string `json:"footer_variant"`

	// Реквизиты компании и юридические параметры (Блок 4: 152-ФЗ и B2B Invoicing)
	LegalEntity         string `json:"legal_entity"`
	LegalAddress        string `json:"legal_address"`
	INN                 string `json:"inn"`
	KPP                 string `json:"kpp"`
	OGRN                string `json:"ogrn"`
	BankName            string `json:"bank_name"`
	BankBIK             string `json:"bank_bik"`
	BankAccount         string `json:"bank_account"`
	BankCorr            string `json:"bank_corr"`
	ContactEmail        string `json:"contact_email"`
	EnableCareers       bool   `json:"enable_careers"`
	EnableCookieConsent bool   `json:"enable_cookie_consent"`

	// Модули кабинета администратора: список явно отключенных модулей через запятую (например "reviews,fleet")
	DisabledModules string `json:"disabled_modules,omitempty"`

	// Пароль для входа в клиентский кабинет (/client)
	ClientAdminPassword     string `json:"client_admin_password"` // Legacy plain password for auto-migration
	ClientAdminPasswordHash string `json:"client_admin_password_hash"`
}

// IsModuleActive checks whether an admin/business module is active for the current site.
func (s SiteSettings) IsModuleActive(moduleID string) bool {
	if s.DisabledModules != "" {
		for _, disabled := range strings.Split(s.DisabledModules, ",") {
			if strings.TrimSpace(disabled) == moduleID {
				return false
			}
		}
	}
	switch moduleID {
	case "articles", "blog":
		return s.EnableBlog
	case "fleet":
		return s.EnableFleet
	case "cart", "orders":
		return s.EnableCart
	case "live_chat", "chat":
		return s.EnableLiveChat
	default:
		return true
	}
}

// AdminTabModule represents a client admin panel tab/module passed to UI templates.
type AdminTabModule struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Route      string `json:"route"`
	BadgeCount int    `json:"badge_count"`
	BadgeClass string `json:"badge_class"`
	Order      int    `json:"order"`
}

// SiteContent holds content blocks displayed on the website
type SiteContent struct {
	ID           string `json:"id"`
	HeroBadge    string `json:"hero_badge"`
	HeroTitle    string `json:"hero_title"`
	HeroSubtitle string `json:"hero_subtitle"`
	// A/B Testing Variant B overrides (Server-Side A/B testing)
	HeroTitleB    string `json:"hero_title_b,omitempty"`
	HeroSubtitleB string `json:"hero_subtitle_b,omitempty"`
	HeroCTAB      string `json:"hero_cta_b,omitempty"`
	OwnerName     string `json:"owner_name"`
	OwnerTitle    string `json:"owner_title"`
	// Aliases for template backwards compatibility
	NotaryName   string `json:"notary_name"`
	NotaryTitle  string `json:"notary_title"`
	Phone1       string `json:"phone1"`
	Phone2       string `json:"phone2"`
	WorkHours    string `json:"work_hours"`
	LunchBreak   string `json:"lunch_break"`
	DaysOff      string `json:"days_off"`
	Address      string `json:"address"`
	AddressSub   string `json:"address_sub"`
	MapIframe    string `json:"map_iframe"`
	HeroPhoto    string `json:"hero_photo"` // URL to image (either PB file or fallback)
}

// PageBlock represents a dynamic section on the page (Block Reordering / Custom Content / Hybrid JSON payload)
type PageBlock struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"` // hero, cyber_hero, schedule, services, booking, map, features, faq, steps, reviews, team, gallery, rich_text, html
	Title   string                 `json:"title"`
	Content string                 `json:"content"` // Raw HTML or sanitized rich text
	Order   int                    `json:"order"`
	Active  bool                   `json:"active"`
	Payload map[string]interface{} `json:"payload,omitempty"` // Flexible data-driven content for features, faq, steps, etc.
}

// FeatureItem represents a single card in Features section
type FeatureItem struct {
	Icon        string `json:"icon"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// FeaturesPayload represents data-driven Features section
type FeaturesPayload struct {
	Badge       string        `json:"badge"`
	Title       string        `json:"title"`
	Subtitle    string        `json:"subtitle"`
	Items       []FeatureItem `json:"items"`
}

// FAQItem represents a single question/answer
type FAQItem struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// FAQPayload represents data-driven FAQ section
type FAQPayload struct {
	Badge    string    `json:"badge"`
	Title    string    `json:"title"`
	Subtitle string    `json:"subtitle"`
	Items    []FAQItem `json:"items"`
}

// StepItem represents a workflow stage
type StepItem struct {
	Number      string `json:"number"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// StepsPayload represents data-driven Steps section
type StepsPayload struct {
	Badge    string     `json:"badge"`
	Title    string     `json:"title"`
	Subtitle string     `json:"subtitle"`
	Items    []StepItem `json:"items"`
}

// QuizOption represents a selectable answer in a quiz step
type QuizOption struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Description   string  `json:"description,omitempty"`
	PriceModifier float64 `json:"price_modifier"`
	Icon          string  `json:"icon,omitempty"`
}

// QuizStep represents an individual question in the interactive quiz
type QuizStep struct {
	ID          string       `json:"id"`
	Question    string       `json:"question"`
	Subtitle    string       `json:"subtitle,omitempty"`
	MultiSelect bool         `json:"multi_select"`
	Options     []QuizOption `json:"options"`
}

// QuizPayload represents the configuration for the interactive quiz/calculator block
type QuizPayload struct {
	Badge     string     `json:"badge"`
	Title     string     `json:"title"`
	Subtitle  string     `json:"subtitle"`
	BasePrice float64    `json:"base_price"`
	Currency  string     `json:"currency"`
	BonusText string     `json:"bonus_text,omitempty"`
	Steps     []QuizStep `json:"steps"`
}

// BeforeAfterItem represents a single case comparison item
type BeforeAfterItem struct {
	Title         string   `json:"title"`
	Description   string   `json:"description,omitempty"`
	BeforeImage   string   `json:"before_image"`
	AfterImage    string   `json:"after_image"`
	BeforeLabel   string   `json:"before_label,omitempty"`
	AfterLabel    string   `json:"after_label,omitempty"`
	Duration      string   `json:"duration,omitempty"`       // e.g. "3 рабочих дня"
	Cost          string   `json:"cost,omitempty"`           // e.g. "48 500 ₽"
	MaterialsCost string   `json:"materials_cost,omitempty"` // e.g. "32 000 ₽"
	WorksCost     string   `json:"works_cost,omitempty"`     // e.g. "16 500 ₽"
	Stages        []string `json:"stages,omitempty"`         // e.g. ["1. Дефектовка", "2. Стапельные работы"]
}

// BeforeAfterPayload represents the data-driven Before/After block
type BeforeAfterPayload struct {
	Badge    string            `json:"badge"`
	Title    string            `json:"title"`
	Subtitle string            `json:"subtitle"`
	Items    []BeforeAfterItem `json:"items"`
}

// CartItem represents an item placed in the client's web cart
type CartItem struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Price    float64 `json:"price"`
	PriceStr string  `json:"price_str,omitempty"`
	Qty      int     `json:"qty"`
	Image    string  `json:"image,omitempty"`
	IsUpsell bool    `json:"is_upsell,omitempty"`
}

// CartCheckoutRequest represents the payload when customer places an order from cart
type CartCheckoutRequest struct {
	Name        string     `json:"name"`
	Phone       string     `json:"phone"`
	Email       string     `json:"email,omitempty"`
	Address     string     `json:"address,omitempty"`
	Comment     string     `json:"comment,omitempty"`
	PayOnline   bool       `json:"pay_online"`
	Items       []CartItem `json:"items"`
	TotalAmount float64    `json:"total_amount"`
}

// ServiceItem represents a service or pricing item
type ServiceItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Order       int    `json:"order"`
	Active      bool   `json:"active"`
	StockCount  int    `json:"stock_count"` // -1 for unlimited, 0 for out of stock / preorder, >0 for limited stock
	IsPreorder  bool   `json:"is_preorder"`
	IsDeleted   bool   `json:"is_deleted"`
	DeletedAt   string `json:"deleted_at"`
}

// Lead represents a form submission
type Lead struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Phone       string                 `json:"phone"`
	Email       string                 `json:"email"`
	Status      string                 `json:"status"` // новая, обработана, отменена
	Source      string                 `json:"source"`
	BookingDate string                 `json:"booking_date"`
	BookingTime string                 `json:"booking_time"`
	ResourceID  string                 `json:"resource_id,omitempty"`
	Payload     map[string]interface{} `json:"payload"`
	Created     string                 `json:"created,omitempty"`
}

// PaymentRequest contains information to initiate an online payment
type PaymentRequest struct {
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	ReturnURL   string  `json:"return_url"`
	Email       string  `json:"email"`
	Phone       string  `json:"phone"`
}

// PaymentResponse is returned after payment registration
type PaymentResponse struct {
	PaymentID   string `json:"payment_id"`
	Status      string `json:"status"`
	PaymentURL  string `json:"payment_url"`
	IsMock      bool   `json:"is_mock"`
}

// ArticleItem represents a blog post or knowledge base article (Этап 1: SEO-Блог)
type ArticleItem struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Slug           string   `json:"slug"`
	Excerpt        string   `json:"excerpt"`
	Content        string   `json:"content"`
	CoverImage     string   `json:"cover_image"`
	AuthorName     string   `json:"author_name"`
	PublishedAt    string   `json:"published_at"`
	IsPublished    bool     `json:"is_published"`
	IsDeleted      bool     `json:"is_deleted"`
	DeletedAt      string   `json:"deleted_at,omitempty"`
	ViewsCount     int      `json:"views_count"`
	SEOTitle       string   `json:"seo_title"`
	SEODescription string   `json:"seo_description"`
	Tags           []string `json:"tags"`
}

// ParsePublishDate parses a date or datetime string in common formats
func ParsePublishDate(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		time.RFC3339,
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, s, time.Local); err == nil {
			return t
		}
	}
	return time.Time{}
}

// IsScheduled returns true if article is published but publish time is in the future
func (a ArticleItem) IsScheduled() bool {
	if !a.IsPublished || a.IsDeleted || a.PublishedAt == "" {
		return false
	}
	t := ParsePublishDate(a.PublishedAt)
	if t.IsZero() {
		return false
	}
	return t.After(time.Now())
}

// IsLive returns true if article is published and published_at is reached or empty, and not deleted
func (a ArticleItem) IsLive() bool {
	if !a.IsPublished || a.IsDeleted {
		return false
	}
	if a.PublishedAt == "" {
		return true
	}
	t := ParsePublishDate(a.PublishedAt)
	if t.IsZero() {
		return true
	}
	return !t.After(time.Now())
}

// FormatPublishedAt returns human-readable date/time representation
func (a ArticleItem) FormatPublishedAt() string {
	if a.PublishedAt == "" {
		return "—"
	}
	t := ParsePublishDate(a.PublishedAt)
	if t.IsZero() {
		return a.PublishedAt
	}
	if t.Hour() == 0 && t.Minute() == 0 && len(a.PublishedAt) <= 10 {
		return t.Format("02.01.2006")
	}
	return t.Format("02.01.2006 15:04")
}

// PublishedAtInput returns HTML5 datetime-local compatible format (YYYY-MM-DDTHH:MM)
func (a ArticleItem) PublishedAtInput() string {
	if a.PublishedAt == "" {
		return time.Now().Format("2006-01-02T15:04")
	}
	t := ParsePublishDate(a.PublishedAt)
	if t.IsZero() {
		return a.PublishedAt
	}
	return t.Format("2006-01-02T15:04")
}

// TMAUser represents authenticated Telegram WebApp user (TMA для Ниш А и В)
type TMAUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	IsVIP     bool   `json:"is_vip"`
}

// OrderStageHistory represents a single milestone in order progress
type OrderStageHistory struct {
	Stage     string `json:"stage"`     // received, diagnostics, in_progress, waiting_parts, ready, completed
	Title     string `json:"title"`     // e.g. "Автомобиль принят в сервисный центр"
	Timestamp string `json:"timestamp"` // e.g. "07.10 09:30"
	Comment   string `json:"comment"`
	Done      bool   `json:"done"`
}

// ServiceOrder represents a trackable repair or service ticket (Этап 2: Ниша Г - B2B-трекер)
type ServiceOrder struct {
	ID            string              `json:"id"`
	OrderCode     string              `json:"order_code"`    // e.g. "SRV-4182"
	ClientName    string              `json:"client_name"`
	ClientPhone   string              `json:"client_phone"`
	ItemTitle     string              `json:"item_title"`    // e.g. "BMW X5 2020 (х819ок123)"
	CurrentStage  string              `json:"current_stage"` // received, diagnostics, in_progress, ready, completed
	TotalAmount   float64             `json:"total_amount"`
	IsPaid        bool                `json:"is_paid"`
	StagesHistory []OrderStageHistory `json:"stages_history"`
	Notes         string              `json:"notes"`
	CreatedAt     string              `json:"created_at"`
	UpdatedAt     string              `json:"updated_at"`
}


