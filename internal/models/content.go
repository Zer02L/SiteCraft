package models

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

	// Новые модули (Этап 1: Блог и TMA: Telegram Mini App, Этап 2: Трекер заказов, Локализация)
	EnableBlog    bool   `json:"enable_blog"`
	EnableTMA     bool   `json:"enable_tma"`
	EnableTracker bool   `json:"enable_tracker"`
	EnableI18n    bool   `json:"enable_i18n"`
	DefaultLang   string `json:"default_lang"` // ru, en

	// Варианты шапки и подвала (Layout Variants: classic, minimal, centered / columns)
	HeaderVariant string `json:"header_variant"`
	FooterVariant string `json:"footer_variant"`
}

// SiteContent holds content blocks displayed on the website
type SiteContent struct {
	ID           string `json:"id"`
	HeroBadge    string `json:"hero_badge"`
	HeroTitle    string `json:"hero_title"`
	HeroSubtitle string `json:"hero_subtitle"`
	OwnerName    string `json:"owner_name"`
	OwnerTitle   string `json:"owner_title"`
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

// ServiceItem represents a service or pricing item
type ServiceItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Order       int    `json:"order"`
	Active      bool   `json:"active"`
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
	Payload     map[string]interface{} `json:"payload"`
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
	ViewsCount     int      `json:"views_count"`
	SEOTitle       string   `json:"seo_title"`
	SEODescription string   `json:"seo_description"`
	Tags           []string `json:"tags"`
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


