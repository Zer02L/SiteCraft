package cache

import (
	"sync"
	"time"

	"sitecraft/internal/models"
)

type Store struct {
	mu                     sync.RWMutex
	settings               models.SiteSettings
	content                models.SiteContent
	services               []models.ServiceItem
	blocks                 []models.PageBlock
	articles               []models.ArticleItem
	recentActivity         []models.RecentActivityItem
	recentActivityExpires  time.Time
	settingsSet            bool
	contentSet             bool
	servicesSet            bool
	blocksSet              bool
	articlesSet            bool
}

var defaultStore = &Store{}

func GetStore() *Store {
	return defaultStore
}

func (s *Store) GetSettings() (models.SiteSettings, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings, s.settingsSet
}

func (s *Store) SetSettings(settings models.SiteSettings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = settings
	s.settingsSet = true
}

func (s *Store) InvalidateSettings() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settingsSet = false
}

func (s *Store) GetContent() (models.SiteContent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.content, s.contentSet
}

func (s *Store) SetContent(content models.SiteContent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.content = content
	s.contentSet = true
}

func (s *Store) InvalidateContent() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.contentSet = false
}

func (s *Store) GetServices() ([]models.ServiceItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.servicesSet {
		return nil, false
	}
	out := make([]models.ServiceItem, len(s.services))
	copy(out, s.services)
	return out, true
}

func (s *Store) SetServices(services []models.ServiceItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services = make([]models.ServiceItem, len(services))
	copy(s.services, services)
	s.servicesSet = true
}

func (s *Store) InvalidateServices() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.servicesSet = false
}

func (s *Store) GetBlocks() ([]models.PageBlock, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.blocksSet {
		return nil, false
	}
	out := make([]models.PageBlock, len(s.blocks))
	copy(out, s.blocks)
	return out, true
}

func (s *Store) SetBlocks(blocks []models.PageBlock) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blocks = make([]models.PageBlock, len(blocks))
	copy(s.blocks, blocks)
	s.blocksSet = true
}

func (s *Store) InvalidateBlocks() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blocksSet = false
}

func (s *Store) GetArticles() ([]models.ArticleItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.articlesSet {
		return nil, false
	}
	out := make([]models.ArticleItem, len(s.articles))
	copy(out, s.articles)
	return out, true
}

func (s *Store) SetArticles(articles []models.ArticleItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.articles = make([]models.ArticleItem, len(articles))
	copy(s.articles, articles)
	s.articlesSet = true
}

func (s *Store) InvalidateArticles() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.articlesSet = false
}

func (s *Store) GetRecentActivity() ([]models.RecentActivityItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if time.Now().After(s.recentActivityExpires) || len(s.recentActivity) == 0 {
		return nil, false
	}
	out := make([]models.RecentActivityItem, len(s.recentActivity))
	copy(out, s.recentActivity)
	return out, true
}

func (s *Store) SetRecentActivity(items []models.RecentActivityItem, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recentActivity = make([]models.RecentActivityItem, len(items))
	copy(s.recentActivity, items)
	s.recentActivityExpires = time.Now().Add(ttl)
}

func (s *Store) InvalidateRecentActivity() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recentActivity = nil
	s.recentActivityExpires = time.Time{}
}

func (s *Store) InvalidateAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settingsSet = false
	s.contentSet = false
	s.servicesSet = false
	s.blocksSet = false
	s.articlesSet = false
	s.recentActivity = nil
	s.recentActivityExpires = time.Time{}
}
