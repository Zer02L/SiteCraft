package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sitecraft/internal/cache"
	"sitecraft/internal/models"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
)

func TestABContextKey(t *testing.T) {
	ctx := context.Background()

	// Default fallback should be "A"
	if v := GetABVariant(ctx); v != "A" {
		t.Errorf("Expected default variant A, got %s", v)
	}

	// Set variant "B"
	ctxB := WithABVariant(ctx, "B")
	if v := GetABVariant(ctxB); v != "B" {
		t.Errorf("Expected variant B, got %s", v)
	}

	// Set variant "A"
	ctxA := WithABVariant(ctx, "A")
	if v := GetABVariant(ctxA); v != "A" {
		t.Errorf("Expected variant A, got %s", v)
	}
}

func TestABTestingMiddleware(t *testing.T) {
	r := router.NewRouter(func(w http.ResponseWriter, req *http.Request) (*core.RequestEvent, router.EventCleanupFunc) {
		re := &core.RequestEvent{}
		re.Request = req
		re.Response = w
		return re, nil
	})
	r.BindFunc(ABTestingMiddleware())

	var capturedVariant string
	r.GET("/test-ab", func(e *core.RequestEvent) error {
		capturedVariant = GetABVariant(e.Request.Context())
		return e.String(http.StatusOK, "ok")
	})

	mux, err := r.BuildMux()
	if err != nil {
		t.Fatalf("Failed to build router mux: %v", err)
	}

	t.Run("Query parameter ?ab=B forces variant B and sets cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test-ab?ab=B", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if capturedVariant != "B" {
			t.Errorf("Expected captured variant B, got %s", capturedVariant)
		}

		cookies := rec.Result().Cookies()
		var foundCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == ABCookieName {
				foundCookie = c
				break
			}
		}

		if foundCookie == nil {
			t.Fatal("Expected sitecraft_ab cookie to be set")
		}
		if foundCookie.Value != "B" {
			t.Errorf("Expected cookie value B, got %s", foundCookie.Value)
		}
	})

	t.Run("Existing cookie sitecraft_ab=A is honored", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test-ab", nil)
		req.AddCookie(&http.Cookie{
			Name:  ABCookieName,
			Value: "A",
		})
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if capturedVariant != "A" {
			t.Errorf("Expected captured variant A, got %s", capturedVariant)
		}
	})

	t.Run("First visit without cookie generates A or B and sets 30-day cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test-ab", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if capturedVariant != "A" && capturedVariant != "B" {
			t.Errorf("Expected variant A or B, got %s", capturedVariant)
		}

		cookies := rec.Result().Cookies()
		var foundCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == ABCookieName {
				foundCookie = c
				break
			}
		}

		if foundCookie == nil {
			t.Fatal("Expected sitecraft_ab cookie on first visit")
		}
		if foundCookie.Value != capturedVariant {
			t.Errorf("Expected cookie value %s, got %s", capturedVariant, foundCookie.Value)
		}
		if foundCookie.MaxAge != 30*24*3600 {
			t.Errorf("Expected 30-day cookie max age, got %d", foundCookie.MaxAge)
		}
	})
}

func TestRecentActivityCacheAndTTL(t *testing.T) {
	store := cache.GetStore()
	store.InvalidateRecentActivity()

	// Initial check should be empty
	if _, ok := store.GetRecentActivity(); ok {
		t.Fatal("Expected empty cache initially")
	}

	testItems := []models.RecentActivityItem{
		{
			Name:    "Константин",
			City:    "СПб",
			Action:  "заказал аудит",
			TimeAgo: "12 мин назад",
			Icon:    "order",
		},
		{
			Name:    "Елена",
			City:    "Москва",
			Action:  "записалась на прием",
			TimeAgo: "25 мин назад",
			Icon:    "booking",
		},
	}

	// Cache with positive TTL
	store.SetRecentActivity(testItems, 5*time.Minute)

	cached, ok := store.GetRecentActivity()
	if !ok {
		t.Fatal("Expected cache hit after SetRecentActivity")
	}
	if len(cached) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(cached))
	}
	if cached[0].Name != "Константин" || cached[1].Name != "Елена" {
		t.Errorf("Cached item names mismatch: %+v", cached)
	}

	// Test invalidation
	store.InvalidateRecentActivity()
	if _, ok := store.GetRecentActivity(); ok {
		t.Fatal("Expected cache miss after InvalidateRecentActivity")
	}
}
