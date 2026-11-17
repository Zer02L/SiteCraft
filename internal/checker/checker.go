package checker

import (
	"context"
	"database/sql"
	"strings"
	"sync"

	"sitecraft/internal/data"
	"sitecraft/internal/events"
	"sitecraft/internal/models"
)

// ChecklistItem represents a single readiness checkpoint for site launch.
type ChecklistItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	IsCompleted bool   `json:"is_completed"`
	IsCritical  bool   `json:"is_critical"`
	TabID       string `json:"tab_id"`
	ActionURL   string `json:"action_url"`
	ActionLabel string `json:"action_label"`
}

// ActionRoute returns the internal HTMX partial route for tab navigation.
func (i ChecklistItem) ActionRoute() string {
	return "/admin/tab/" + i.TabID
}

// ChecklistResult aggregates all checklist evaluation metrics.
type ChecklistResult struct {
	Percentage         int             `json:"percentage"`
	CompletedCount     int             `json:"completed_count"`
	TotalCount         int             `json:"total_count"`
	CriticalTotalCount int             `json:"critical_total_count"`
	CriticalDoneCount  int             `json:"critical_done_count"`
	IsReady            bool            `json:"is_ready"`
	Items              []ChecklistItem `json:"items"`
}

// PendingItems returns incomplete checklist items.
func (r ChecklistResult) PendingItems() []ChecklistItem {
	var list []ChecklistItem
	for _, item := range r.Items {
		if !item.IsCompleted {
			list = append(list, item)
		}
	}
	return list
}

// CompletedItems returns finished checklist items.
func (r ChecklistResult) CompletedItems() []ChecklistItem {
	var list []ChecklistItem
	for _, item := range r.Items {
		if item.IsCompleted {
			list = append(list, item)
		}
	}
	return list
}

// CriticalPendingCount returns the number of unfinished critical steps.
func (r ChecklistResult) CriticalPendingCount() int {
	count := 0
	for _, item := range r.Items {
		if item.IsCritical && !item.IsCompleted {
			count++
		}
	}
	return count
}

// ModuleChecker checks whether a specific administrative module is active in the current plan/registry.
type ModuleChecker func(id string, settings models.SiteSettings) bool

// Options contains optional dependencies for checklist evaluation.
type Options struct {
	Registry ModuleChecker
	Content  *models.SiteContent
	Repo     *data.Repository
}

// Option configures evaluation options.
type Option func(*Options)

// WithRegistry injects an administrative module registry.
func WithRegistry(r ModuleChecker) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

// WithContent injects pre-fetched site content.
func WithContent(c models.SiteContent) Option {
	return func(o *Options) {
		o.Content = &c
	}
}

// WithRepo injects a data repository.
func WithRepo(r *data.Repository) Option {
	return func(o *Options) {
		o.Repo = r
	}
}

// EvaluateSiteHealth audits the current site readiness and returns the ChecklistResult.
func EvaluateSiteHealth(settings models.SiteSettings, db *sql.DB, opts ...Option) ChecklistResult {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	isModuleActive := func(moduleID string) bool {
		if options.Registry != nil {
			return options.Registry(moduleID, settings)
		}
		return settings.IsModuleActive(moduleID)
	}

	var content models.SiteContent
	if options.Content != nil {
		content = *options.Content
	} else if options.Repo != nil {
		c, err := options.Repo.GetSiteContent(context.Background())
		if err == nil {
			content = c
		}
	} else if db != nil {
		_ = db.QueryRow("SELECT phone1, phone2, address, address_sub FROM site_content LIMIT 1").Scan(
			&content.Phone1, &content.Phone2, &content.Address, &content.AddressSub,
		)
	}

	var servicesCount int
	var articlesCount int
	var reviewsCount int
	var blocksCount int

	if options.Repo != nil {
		ctx := context.Background()
		if isModuleActive("services") {
			if srvs, err := options.Repo.GetServices(ctx); err == nil {
				for _, s := range srvs {
					if s.Active && !s.IsDeleted {
						servicesCount++
					}
				}
			}
		}
		if isModuleActive("articles") && settings.EnableBlog {
			if arts, err := options.Repo.GetArticles(ctx); err == nil {
				for _, a := range arts {
					if a.IsPublished && !a.IsDeleted {
						articlesCount++
					}
				}
			}
		}
		if isModuleActive("reviews") {
			if revs, err := options.Repo.GetReviews(ctx); err == nil {
				for _, r := range revs {
					if r.IsPublished {
						reviewsCount++
					}
				}
			}
		}
		if blks, err := options.Repo.GetAllPageBlocks(ctx); err == nil {
			for _, b := range blks {
				if b.Active {
					blocksCount++
				}
			}
		}
	} else if db != nil {
		if isModuleActive("services") {
			_ = db.QueryRow("SELECT COUNT(*) FROM services WHERE is_deleted = 0 AND active = 1").Scan(&servicesCount)
		}
		if isModuleActive("articles") && settings.EnableBlog {
			_ = db.QueryRow("SELECT COUNT(*) FROM articles WHERE is_deleted = 0 AND is_published = 1").Scan(&articlesCount)
		}
		if isModuleActive("reviews") {
			_ = db.QueryRow("SELECT COUNT(*) FROM reviews WHERE is_published = 1").Scan(&reviewsCount)
		}
		_ = db.QueryRow("SELECT COUNT(*) FROM page_blocks WHERE active = 1").Scan(&blocksCount)
	}

	var items []ChecklistItem

	// 1. Contacts
	hasPhone := strings.TrimSpace(content.Phone1) != "" || strings.TrimSpace(content.Phone2) != "" || strings.TrimSpace(settings.WhatsappPhone) != ""
	hasEmail := strings.TrimSpace(settings.ContactEmail) != ""
	hasAddress := strings.TrimSpace(content.Address) != "" || strings.TrimSpace(settings.LegalAddress) != ""
	contactsCompleted := hasPhone && (hasEmail || hasAddress)

	items = append(items, ChecklistItem{
		ID:          "contacts",
		Title:       "Контактная информация",
		Description: "Укажите основной телефон, контактный email и адрес компании",
		Category:    "contacts",
		IsCompleted: contactsCompleted,
		IsCritical:  true,
		TabID:       "contacts",
		ActionURL:   "/admin?tab=contacts",
		ActionLabel: "Заполнить контакты",
	})

	// 2. Legal / 152-ФЗ
	legalCompleted := strings.TrimSpace(settings.LegalEntity) != "" &&
		strings.TrimSpace(settings.INN) != "" &&
		settings.EnableCookieConsent

	items = append(items, ChecklistItem{
		ID:          "legal",
		Title:       "Реквизиты и 152-ФЗ",
		Description: "Заполните юр. лицо, ИНН и включите согласие на обработку данных (152-ФЗ)",
		Category:    "legal",
		IsCompleted: legalCompleted,
		IsCritical:  true,
		TabID:       "contacts",
		ActionURL:   "/admin?tab=contacts",
		ActionLabel: "Настроить реквизиты",
	})

	// 3. Telegram Bot Notifications
	telegramCompleted := strings.TrimSpace(settings.TelegramBotToken) != "" &&
		strings.TrimSpace(settings.TelegramChatID) != ""

	items = append(items, ChecklistItem{
		ID:          "telegram",
		Title:       "Telegram-бот для заявок",
		Description: "Подключите Telegram бота и Chat ID для мгновенных уведомлений о заявках",
		Category:    "notifications",
		IsCompleted: telegramCompleted,
		IsCritical:  false,
		TabID:       "settings",
		ActionURL:   "/admin?tab=settings",
		ActionLabel: "Подключить Telegram",
	})

	// 4. Analytics (Яндекс Метрика)
	analyticsCompleted := strings.TrimSpace(settings.YandexMetrikaID) != ""

	items = append(items, ChecklistItem{
		ID:          "analytics",
		Title:       "Яндекс Метрика",
		Description: "Укажите номер счетчика Яндекс Метрики для веб-аналитики и конверсий",
		Category:    "analytics",
		IsCompleted: analyticsCompleted,
		IsCritical:  false,
		TabID:       "settings",
		ActionURL:   "/admin?tab=settings",
		ActionLabel: "Указать счетчик",
	})

	// 5. Services / Catalog (Only if services module is active)
	if isModuleActive("services") {
		servicesCompleted := servicesCount > 0
		items = append(items, ChecklistItem{
			ID:          "services",
			Title:       "Каталог услуг и цен",
			Description: "Добавьте хотя бы одну актуальную услугу с ценой в каталог",
			Category:    "catalog",
			IsCompleted: servicesCompleted,
			IsCritical:  true,
			TabID:       "services",
			ActionURL:   "/admin?tab=services",
			ActionLabel: "Добавить услугу",
		})
	}

	// 6. Online Payment (Only if payment module is enabled)
	if settings.EnablePayment {
		paymentCompleted := strings.TrimSpace(settings.PaymentShopID) != "" &&
			strings.TrimSpace(settings.PaymentSecret) != ""
		items = append(items, ChecklistItem{
			ID:          "payment",
			Title:       "Интернет-эквайринг",
			Description: "Укажите Shop ID и Secret Key платежного шлюза для онлайн-оплаты",
			Category:    "payment",
			IsCompleted: paymentCompleted,
			IsCritical:  true,
			TabID:       "settings",
			ActionURL:   "/admin?tab=settings",
			ActionLabel: "Настроить оплату",
		})
	}

	// 7. Blog & Articles (Only if blog is enabled and module active)
	if isModuleActive("articles") && settings.EnableBlog {
		articlesCompleted := articlesCount > 0
		items = append(items, ChecklistItem{
			ID:          "articles",
			Title:       "Статьи в блоге",
			Description: "Опубликуйте хотя бы одну статью или новость в блоге",
			Category:    "content",
			IsCompleted: articlesCompleted,
			IsCritical:  false,
			TabID:       "articles",
			ActionURL:   "/admin?tab=articles",
			ActionLabel: "Написать статью",
		})
	}

	// 8. Customer Reviews (Only if reviews module is active)
	if isModuleActive("reviews") {
		reviewsCompleted := reviewsCount > 0
		items = append(items, ChecklistItem{
			ID:          "reviews",
			Title:       "Отзывы клиентов",
			Description: "Добавьте реальные отзывы клиентов для повышения доверия",
			Category:    "reviews",
			IsCompleted: reviewsCompleted,
			IsCritical:  false,
			TabID:       "reviews",
			ActionURL:   "/admin?tab=reviews",
			ActionLabel: "Добавить отзыв",
		})
	}

	// 9. Page Blocks / Homepage Structure
	blocksCompleted := blocksCount > 0
	items = append(items, ChecklistItem{
		ID:          "blocks",
		Title:       "Структура главной страницы",
		Description: "Сконфигурируйте и опубликуйте блоки главной страницы в конструкторе",
		Category:    "design",
		IsCompleted: blocksCompleted,
		IsCritical:  true,
		TabID:       "blocks",
		ActionURL:   "/admin?tab=blocks",
		ActionLabel: "Настроить блоки",
	})

	totalCount := len(items)
	completedCount := 0
	criticalTotal := 0
	criticalDone := 0

	for _, item := range items {
		if item.IsCompleted {
			completedCount++
		}
		if item.IsCritical {
			criticalTotal++
			if item.IsCompleted {
				criticalDone++
			}
		}
	}

	percentage := 0
	if totalCount > 0 {
		percentage = (completedCount * 100) / totalCount
	}

	isReady := (criticalDone == criticalTotal && percentage >= 80) || percentage == 100

	return ChecklistResult{
		Percentage:         percentage,
		CompletedCount:     completedCount,
		TotalCount:         totalCount,
		CriticalTotalCount: criticalTotal,
		CriticalDoneCount:  criticalDone,
		IsReady:            isReady,
		Items:              items,
	}
}

// Service provides in-memory cached checklist evaluation with EventBus invalidation.
type Service struct {
	db       *sql.DB
	repo     *data.Repository
	bus      events.Bus
	registry ModuleChecker

	mu     sync.RWMutex
	cached *ChecklistResult
}

// NewService creates a new checklist service subscribed to site mutation events.
func NewService(db *sql.DB, repo *data.Repository, bus events.Bus, registry ModuleChecker) *Service {
	s := &Service{
		db:       db,
		repo:     repo,
		bus:      bus,
		registry: registry,
	}

	if bus != nil {
		bus.Subscribe(events.EventSiteMutated, func(ctx context.Context, e events.Event) error {
			s.Invalidate()
			return nil
		})
	}

	return s
}

// Invalidate clears the in-memory cached checklist result.
func (s *Service) Invalidate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cached = nil
}

// GetChecklist returns the cached checklist or computes it atomically.
func (s *Service) GetChecklist(ctx context.Context, settings models.SiteSettings) ChecklistResult {
	s.mu.RLock()
	if s.cached != nil {
		res := *s.cached
		s.mu.RUnlock()
		return res
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cached != nil {
		return *s.cached
	}

	res := EvaluateSiteHealth(settings, s.db, WithRegistry(s.registry), WithRepo(s.repo))
	s.cached = &res
	return res
}
