package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"sitecraft/internal/models"
)

func TestStore_Settings(t *testing.T) {
	s := &Store{}

	// Initially empty
	_, ok := s.GetSettings()
	if ok {
		t.Fatal("expected empty settings initially")
	}

	// Set and Get
	sample := models.SiteSettings{
		SiteName: "Тестовый сайт",
		Tagline:  "Инновационные решения",
	}
	s.SetSettings(sample)

	got, ok := s.GetSettings()
	if !ok {
		t.Fatal("expected settings to be set")
	}
	if got.SiteName != sample.SiteName || got.Tagline != sample.Tagline {
		t.Errorf("expected %+v, got %+v", sample, got)
	}

	// Invalidate
	s.InvalidateSettings()
	_, ok = s.GetSettings()
	if ok {
		t.Fatal("expected settings to be invalidated")
	}
}

func TestStore_Content(t *testing.T) {
	s := &Store{}

	_, ok := s.GetContent()
	if ok {
		t.Fatal("expected empty content initially")
	}

	sample := models.SiteContent{
		HeroTitle: "Лучший сервис",
		Address:   "ул. Ленина, д. 1",
	}
	s.SetContent(sample)

	got, ok := s.GetContent()
	if !ok {
		t.Fatal("expected content to be set")
	}
	if got.HeroTitle != sample.HeroTitle || got.Address != sample.Address {
		t.Errorf("expected %+v, got %+v", sample, got)
	}

	s.InvalidateContent()
	_, ok = s.GetContent()
	if ok {
		t.Fatal("expected content to be invalidated")
	}
}

func TestStore_Services_Isolation(t *testing.T) {
	s := &Store{}

	_, ok := s.GetServices()
	if ok {
		t.Fatal("expected empty services initially")
	}

	original := []models.ServiceItem{
		{ID: "srv_1", Title: "Стрижка", Price: "1000 ₽"},
	}
	s.SetServices(original)

	// Mutating the slice that was passed in must not affect cache (deep copy check)
	original[0].Title = "Мутировано"
	got, ok := s.GetServices()
	if !ok || len(got) != 1 {
		t.Fatalf("expected 1 service item, got ok=%v, len=%d", ok, len(got))
	}
	if got[0].Title != "Стрижка" {
		t.Errorf("expected Title 'Стрижка', got %q (copy isolation violated on Set)", got[0].Title)
	}

	// Mutating the returned slice must not affect cache
	got[0].Title = "Мутировано снова"
	gotAgain, _ := s.GetServices()
	if gotAgain[0].Title != "Стрижка" {
		t.Errorf("expected Title 'Стрижка', got %q (copy isolation violated on Get)", gotAgain[0].Title)
	}

	s.InvalidateServices()
	_, ok = s.GetServices()
	if ok {
		t.Fatal("expected services to be invalidated")
	}
}

func TestStore_Blocks_Isolation(t *testing.T) {
	s := &Store{}

	_, ok := s.GetBlocks()
	if ok {
		t.Fatal("expected empty blocks initially")
	}

	original := []models.PageBlock{
		{ID: "b_1", Type: "hero", Order: 1},
	}
	s.SetBlocks(original)

	original[0].Type = "mutated"
	got, ok := s.GetBlocks()
	if !ok || len(got) != 1 {
		t.Fatalf("expected 1 block, got %d", len(got))
	}
	if got[0].Type != "hero" {
		t.Errorf("expected block type 'hero', got %q", got[0].Type)
	}

	s.InvalidateBlocks()
	_, ok = s.GetBlocks()
	if ok {
		t.Fatal("expected blocks to be invalidated")
	}
}

func TestStore_Articles_Isolation(t *testing.T) {
	s := &Store{}

	_, ok := s.GetArticles()
	if ok {
		t.Fatal("expected empty articles initially")
	}

	original := []models.ArticleItem{
		{ID: "art_1", Title: "Новость", Slug: "novost"},
	}
	s.SetArticles(original)

	original[0].Slug = "mutated"
	got, ok := s.GetArticles()
	if !ok || len(got) != 1 {
		t.Fatalf("expected 1 article, got %d", len(got))
	}
	if got[0].Slug != "novost" {
		t.Errorf("expected slug 'novost', got %q", got[0].Slug)
	}

	s.InvalidateArticles()
	_, ok = s.GetArticles()
	if ok {
		t.Fatal("expected articles to be invalidated")
	}
}

func TestStore_RecentActivity_TTL(t *testing.T) {
	s := &Store{}

	// Initially empty
	_, ok := s.GetRecentActivity()
	if ok {
		t.Fatal("expected empty recent activity initially")
	}

	items := []models.RecentActivityItem{
		{Name: "Иван", City: "Москва", Action: "оформил заявку", TimeAgo: "5 мин назад"},
	}

	// Short TTL of 50ms
	s.SetRecentActivity(items, 50*time.Millisecond)

	got, ok := s.GetRecentActivity()
	if !ok || len(got) != 1 {
		t.Fatalf("expected active items before TTL expiry, got ok=%v", ok)
	}
	if got[0].Name != items[0].Name || got[0].Action != items[0].Action {
		t.Errorf("expected item %+v, got %+v", items[0], got[0])
	}

	// Wait for TTL expiration
	time.Sleep(60 * time.Millisecond)
	_, ok = s.GetRecentActivity()
	if ok {
		t.Fatal("expected recent activity to expire after TTL")
	}

	// Test InvalidateRecentActivity
	s.SetRecentActivity(items, 1*time.Minute)
	s.InvalidateRecentActivity()
	_, ok = s.GetRecentActivity()
	if ok {
		t.Fatal("expected recent activity to be invalidated immediately")
	}
}

func TestStore_InvalidateAll(t *testing.T) {
	s := &Store{}

	s.SetSettings(models.SiteSettings{SiteName: "Test"})
	s.SetContent(models.SiteContent{HeroTitle: "Test"})
	s.SetServices([]models.ServiceItem{{Title: "Srv", Price: "100"}})
	s.SetBlocks([]models.PageBlock{{Type: "hero"}})
	s.SetArticles([]models.ArticleItem{{Title: "Art"}})
	s.SetRecentActivity([]models.RecentActivityItem{{Name: "Иван"}}, time.Minute)

	s.InvalidateAll()

	if _, ok := s.GetSettings(); ok {
		t.Error("settings not invalidated")
	}
	if _, ok := s.GetContent(); ok {
		t.Error("content not invalidated")
	}
	if _, ok := s.GetServices(); ok {
		t.Error("services not invalidated")
	}
	if _, ok := s.GetBlocks(); ok {
		t.Error("blocks not invalidated")
	}
	if _, ok := s.GetArticles(); ok {
		t.Error("articles not invalidated")
	}
	if _, ok := s.GetRecentActivity(); ok {
		t.Error("recent activity not invalidated")
	}
}

func TestGetStore_Singleton(t *testing.T) {
	st1 := GetStore()
	st2 := GetStore()
	if st1 == nil || st2 == nil {
		t.Fatal("expected non-nil default store")
	}
	if st1 != st2 {
		t.Errorf("expected GetStore to return the same singleton instance")
	}
}

func TestStore_ConcurrencyRace(t *testing.T) {
	s := &Store{}
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)

		// Writer
		go func(idx int) {
			defer wg.Done()
			s.SetSettings(models.SiteSettings{SiteName: "Concurrency"})
			s.SetServices([]models.ServiceItem{{Title: "Srv", Price: fmt.Sprintf("%d", idx)}})
			s.SetRecentActivity([]models.RecentActivityItem{{Name: "User", City: "SPb"}}, 100*time.Millisecond)
			if idx%3 == 0 {
				s.InvalidateServices()
			}
		}(i)

		// Reader
		go func() {
			defer wg.Done()
			_, _ = s.GetSettings()
			_, _ = s.GetServices()
			_, _ = s.GetRecentActivity()
		}()
	}

	wg.Wait()
}
