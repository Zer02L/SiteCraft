package server

import (
	"context"
	"net/http"
	"sort"

	"sitecraft/internal/models"
)

// AdminModule defines an independent pluggable client admin panel module.
type AdminModule struct {
	ID         string
	Title      string
	Order      int
	Route      string
	IsEnabled  func(s models.SiteSettings) bool
	Badge      func(ctx context.Context, s *Server) (int, string)
	TabHandler http.HandlerFunc
}

// AdminModuleRegistry manages registered administrative modules.
type AdminModuleRegistry struct {
	modules   map[string]AdminModule
	overrides map[string]bool
}

// NewAdminModuleRegistry creates an empty registry.
func NewAdminModuleRegistry() *AdminModuleRegistry {
	return &AdminModuleRegistry{
		modules:   make(map[string]AdminModule),
		overrides: make(map[string]bool),
	}
}

// Register adds a module to the registry.
func (r *AdminModuleRegistry) Register(m AdminModule) {
	if m.Route == "" {
		m.Route = "/client/tab/" + m.ID
	}
	r.modules[m.ID] = m
}

// SetEnabled explicitly enables or disables a module programmatically or via configuration.
func (r *AdminModuleRegistry) SetEnabled(id string, enabled bool) {
	r.overrides[id] = enabled
}

// Get finds a module by its identifier.
func (r *AdminModuleRegistry) Get(id string) (AdminModule, bool) {
	m, ok := r.modules[id]
	return m, ok
}

// IsActive checks if a module is registered and enabled for the given settings.
func (r *AdminModuleRegistry) IsActive(id string, settings models.SiteSettings) bool {
	m, ok := r.modules[id]
	if !ok {
		return false
	}
	if enabled, exists := r.overrides[id]; exists && !enabled {
		return false
	}
	if !settings.IsModuleActive(id) {
		return false
	}
	if m.IsEnabled != nil && !m.IsEnabled(settings) {
		return false
	}
	return true
}

// GetActive returns a sorted list of active UI tab modules for rendering.
func (r *AdminModuleRegistry) GetActive(ctx context.Context, s *Server, settings models.SiteSettings) []models.AdminTabModule {
	var list []models.AdminTabModule
	for _, m := range r.modules {
		if !r.IsActive(m.ID, settings) {
			continue
		}

		badgeCount := 0
		badgeClass := ""
		if m.Badge != nil {
			badgeCount, badgeClass = m.Badge(ctx, s)
		}

		list = append(list, models.AdminTabModule{
			ID:         m.ID,
			Title:      m.Title,
			Route:      m.Route,
			BadgeCount: badgeCount,
			BadgeClass: badgeClass,
			Order:      m.Order,
		})
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Order < list[j].Order
	})

	return list
}

// initAdminModules registers the default built-in client modules.
func (s *Server) initAdminModules() {
	if s.AdminModules == nil {
		s.AdminModules = NewAdminModuleRegistry()
	}

	// 1. Leads
	s.AdminModules.Register(AdminModule{
		ID:    "leads",
		Title: "Входящие заявки",
		Order: 10,
		Badge: func(ctx context.Context, s *Server) (int, string) {
			leads, err := s.Repo.GetLeads(ctx)
			if err != nil {
				return 0, ""
			}
			newCount := 0
			for _, l := range leads {
				if l.Status == "новая" || l.Status == "" {
					newCount++
				}
			}
			return len(leads), "bg-amber-500/10 text-amber-400 border-amber-500/30"
		},
		TabHandler: s.HandleAdminTabLeads,
	})

	// 2. Visual Editor & Blocks (Конструктор & Дизайн)
	s.AdminModules.Register(AdminModule{
		ID:    "blocks",
		Title: "Конструктор & Дизайн",
		Order: 15,
		Badge: func(ctx context.Context, s *Server) (int, string) {
			blocks, err := s.Repo.GetAllPageBlocks(ctx)
			if err != nil {
				return 0, ""
			}
			return len(blocks), "bg-slate-800/90 text-amber-400 border-amber-500/20"
		},
		TabHandler: s.HandleAdminTabBlocks,
	})

	// 3. Contacts
	s.AdminModules.Register(AdminModule{
		ID:         "contacts",
		Title:      "Контакты и реквизиты",
		Order:      20,
		TabHandler: s.HandleAdminTabContacts,
	})

	// 3. Services
	s.AdminModules.Register(AdminModule{
		ID:    "services",
		Title: "Услуги и цены",
		Order: 30,
		Badge: func(ctx context.Context, s *Server) (int, string) {
			services, err := s.Repo.GetServices(ctx)
			if err != nil {
				return 0, ""
			}
			return len(services), "bg-slate-800/90 text-slate-300 border-slate-700/60"
		},
		TabHandler: s.HandleAdminTabServices,
	})

	// 4. Articles / Blog
	s.AdminModules.Register(AdminModule{
		ID:    "articles",
		Title: "Блог и статьи",
		Order: 40,
		IsEnabled: func(st models.SiteSettings) bool {
			return st.EnableBlog
		},
		Badge: func(ctx context.Context, s *Server) (int, string) {
			articles, err := s.Repo.GetArticles(ctx)
			if err != nil {
				return 0, ""
			}
			return len(articles), "bg-slate-800/90 text-slate-300 border-slate-700/60"
		},
		TabHandler: s.HandleAdminTabArticles,
	})

	// 5. Reviews
	s.AdminModules.Register(AdminModule{
		ID:    "reviews",
		Title: "Отзывы клиентов",
		Order: 50,
		Badge: func(ctx context.Context, s *Server) (int, string) {
			reviews, err := s.Repo.GetReviews(ctx)
			if err != nil {
				return 0, ""
			}
			return len(reviews), "bg-slate-800/90 text-slate-300 border-slate-700/60"
		},
		TabHandler: s.HandleAdminTabReviews,
	})

	// 6. Security
	s.AdminModules.Register(AdminModule{
		ID:         "security",
		Title:      "Безопасность",
		Order:      60,
		TabHandler: s.HandleAdminTabSecurity,
	})

	// 7. Fleet / Cluster Integrations
	s.AdminModules.Register(AdminModule{
		ID:    "fleet",
		Title: "Интеграции и Fleet",
		Order: 70,
		IsEnabled: func(st models.SiteSettings) bool {
			return st.EnableFleet
		},
		TabHandler: s.HandleAdminTabFleet,
	})
}

// RequireModule wraps an HTTP handler and checks if the given module is enabled.
// Returns 404 Not Found if disabled, hiding the module existence for security.
func (s *Server) RequireModule(moduleID string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		settings, err := s.Repo.GetSiteSettings(r.Context())
		if err != nil || !s.AdminModules.IsActive(moduleID, settings) {
			http.NotFound(w, r)
			return
		}
		handler(w, r)
	}
}

// RequireModuleGroup returns a Chi middleware ensuring the module is active.
func (s *Server) RequireModuleGroup(moduleID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			settings, err := s.Repo.GetSiteSettings(r.Context())
			if err != nil || !s.AdminModules.IsActive(moduleID, settings) {
				http.NotFound(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ResolveActiveTab selects the requested tab if active, or falls back to the first available active module.
func (s *Server) ResolveActiveTab(requested string, active []models.AdminTabModule) string {
	for _, m := range active {
		if m.ID == requested {
			return requested
		}
	}
	if len(active) > 0 {
		return active[0].ID
	}
	return "leads"
}
