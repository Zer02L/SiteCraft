package cache

import (
	"sync"

	"sitecraft/internal/models"
)

type Store struct {
	mu           sync.RWMutex
	settings     models.SiteSettings
	content      models.SiteContent
	services     []models.ServiceItem
	settingsSet  bool
	contentSet   bool
	servicesSet  bool
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

func (s *Store) InvalidateAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settingsSet = false
	s.contentSet = false
	s.servicesSet = false
}
