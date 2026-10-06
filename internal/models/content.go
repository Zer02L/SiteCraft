package models

type SiteContent struct {
	ID          string
	HeroBadge   string
	HeroTitle   string
	HeroSubtitle string
	NotaryName  string
	NotaryTitle string
	Phone1      string
	Phone2      string
	WorkHours   string
	LunchBreak  string
	DaysOff     string
	Address     string
	AddressSub  string
	MapIframe   string
	HeroPhoto   string // URL to image (either PB file or fallback)
}

type ServiceItem struct {
	ID          string
	Title       string
	Description string
	Price       string
	Order       int
	Active      bool
}
