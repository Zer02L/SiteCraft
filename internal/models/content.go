package models

// SiteSettings represents general site configuration and meta
type SiteSettings struct {
	ID                 string `json:"id"`
	SiteName           string `json:"site_name"`
	Tagline            string `json:"tagline"`
	Theme              string `json:"theme"` // legal, medical, auto, tech, default
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

// PageBlock represents a dynamic section on the page (Block Reordering / Custom Content)
type PageBlock struct {
	ID      string `json:"id"`
	Type    string `json:"type"` // hero, schedule, services, booking, map, rich_text, html
	Title   string `json:"title"`
	Content string `json:"content"` // Raw HTML or sanitized rich text
	Order   int    `json:"order"`
	Active  bool   `json:"active"`
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
