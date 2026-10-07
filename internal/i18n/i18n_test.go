package i18n

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"sitecraft/internal/models"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", LangRU},
		{"ru", LangRU},
		{"RU", LangRU},
		{"ru-RU", LangRU},
		{"russian", LangRU},
		{"en", LangEN},
		{"EN", LangEN},
		{"en-US", LangEN},
		{"en-GB", LangEN},
		{"english", LangEN},
		{"de", LangRU},     // unsupported defaults to RU
		{"fr-FR", LangRU},  // unsupported defaults to RU
		{"   en   ", LangEN},
	}

	for _, tt := range tests {
		got := Normalize(tt.input)
		if got != tt.expected {
			t.Errorf("Normalize(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestContextLocale(t *testing.T) {
	ctx := context.Background()
	if got := GetLocale(ctx); got != LangRU {
		t.Errorf("GetLocale(background) = %q, expected %q", got, LangRU)
	}

	ctxEn := WithLocale(ctx, "en")
	if got := GetLocale(ctxEn); got != LangEN {
		t.Errorf("GetLocale(ctxEn) = %q, expected %q", got, LangEN)
	}

	ctxRu := WithLocale(ctxEn, "ru")
	if got := GetLocale(ctxRu); got != LangRU {
		t.Errorf("GetLocale(ctxRu) = %q, expected %q", got, LangRU)
	}
}

func TestDetectLang(t *testing.T) {
	// 1. Fallback to default
	req1 := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	if got := DetectLang(req1); got != LangRU {
		t.Errorf("DetectLang(empty) = %q, expected %q", got, LangRU)
	}

	// 2. Accept-Language header
	req2 := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	req2.Header.Set("Accept-Language", "en-US,en;q=0.9,ru;q=0.8")
	if got := DetectLang(req2); got != LangEN {
		t.Errorf("DetectLang(header en) = %q, expected %q", got, LangEN)
	}

	// 3. Cookie overrides header
	req3 := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	req3.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req3.AddCookie(&http.Cookie{Name: CookieName, Value: "ru"})
	if got := DetectLang(req3); got != LangRU {
		t.Errorf("DetectLang(cookie ru) = %q, expected %q", got, LangRU)
	}

	// 4. Query param overrides cookie and header
	req4 := httptest.NewRequest(http.MethodGet, "http://localhost/?lang=en", nil)
	req4.Header.Set("Accept-Language", "ru-RU,ru;q=0.9")
	req4.AddCookie(&http.Cookie{Name: CookieName, Value: "ru"})
	if got := DetectLang(req4); got != LangEN {
		t.Errorf("DetectLang(query en) = %q, expected %q", got, LangEN)
	}
}

func TestTranslations(t *testing.T) {
	ctxRu := WithLocale(context.Background(), "ru")
	ctxEn := WithLocale(context.Background(), "en")

	// Key in RU
	if got := T(ctxRu, "nav.about"); got != "О нас" {
		t.Errorf("T(ru, nav.about) = %q, expected 'О нас'", got)
	}

	// Key in EN
	if got := T(ctxEn, "nav.about"); got != "About" {
		t.Errorf("T(en, nav.about) = %q, expected 'About'", got)
	}

	// Interpolation with args
	if got := T(ctxRu, "blog.views", 42); got != "42 просмотров" {
		t.Errorf("T(ru, blog.views, 42) = %q, expected '42 просмотров'", got)
	}
	if got := T(ctxEn, "blog.views", 42); got != "42 views" {
		t.Errorf("T(en, blog.views, 42) = %q, expected '42 views'", got)
	}

	// Missing key fallback
	if got := T(ctxEn, "non.existing.key"); got != "non.existing.key" {
		t.Errorf("T(missing) = %q, expected 'non.existing.key'", got)
	}
}

func TestPlural(t *testing.T) {
	ctxRu := WithLocale(context.Background(), "ru")
	ctxEn := WithLocale(context.Background(), "en")

	// Russian: 1 отзыв, 2 отзыва, 5 отзывов, 11 отзывов, 21 отзыв
	if got := Plural(ctxRu, 1, "отзыв", "отзыва", "отзывов"); got != "отзыв" {
		t.Errorf("Plural(ru, 1) = %q, expected 'отзыв'", got)
	}
	if got := Plural(ctxRu, 2, "отзыв", "отзыва", "отзывов"); got != "отзыва" {
		t.Errorf("Plural(ru, 2) = %q, expected 'отзыва'", got)
	}
	if got := Plural(ctxRu, 5, "отзыв", "отзыва", "отзывов"); got != "отзывов" {
		t.Errorf("Plural(ru, 5) = %q, expected 'отзывов'", got)
	}
	if got := Plural(ctxRu, 11, "отзыв", "отзыва", "отзывов"); got != "отзывов" {
		t.Errorf("Plural(ru, 11) = %q, expected 'отзывов'", got)
	}
	if got := Plural(ctxRu, 21, "отзыв", "отзыва", "отзывов"); got != "отзыв" {
		t.Errorf("Plural(ru, 21) = %q, expected 'отзыв'", got)
	}

	// English: 1 review, 2 reviews, 5 reviews
	if got := Plural(ctxEn, 1, "review", "reviews", "reviews"); got != "review" {
		t.Errorf("Plural(en, 1) = %q, expected 'review'", got)
	}
	if got := Plural(ctxEn, 2, "review", "reviews", "reviews"); got != "reviews" {
		t.Errorf("Plural(en, 2) = %q, expected 'reviews'", got)
	}
	if got := Plural(ctxEn, 5, "review", "reviews", "reviews"); got != "reviews" {
		t.Errorf("Plural(en, 5) = %q, expected 'reviews'", got)
	}
}

func TestLocalizeContent(t *testing.T) {
	ctxEn := WithLocale(context.Background(), "en")
	content := LocalizeContent(ctxEn, models.SiteContent{
		HeroTitle: "Нотариус",
		HeroBadge: "Нотариальная контора открыта",
		WorkHours: "Понедельник — Пятница: 9:00 - 17:30",
	})

	if content.HeroTitle != "Notary Public" {
		t.Errorf("LocalizeContent HeroTitle = %q, expected 'Notary Public'", content.HeroTitle)
	}
	if content.HeroBadge != "Office is open" {
		t.Errorf("LocalizeContent HeroBadge = %q, expected 'Office is open'", content.HeroBadge)
	}
	if content.WorkHours != "Monday — Friday: 9:00 AM - 5:30 PM" {
		t.Errorf("LocalizeContent WorkHours = %q, expected 'Monday — Friday: 9:00 AM - 5:30 PM'", content.WorkHours)
	}

	// Russian content should remain untouched
	ctxRu := WithLocale(context.Background(), "ru")
	contentRu := LocalizeContent(ctxRu, models.SiteContent{
		HeroTitle: "Нотариус",
	})
	if contentRu.HeroTitle != "Нотариус" {
		t.Errorf("LocalizeContent in RU should not alter content: got %q", contentRu.HeroTitle)
	}
}

func TestOGLocaleAndOtherLang(t *testing.T) {
	ctxRu := WithLocale(context.Background(), "ru")
	ctxEn := WithLocale(context.Background(), "en")

	if got := OGLocale(ctxRu); got != "ru_RU" {
		t.Errorf("OGLocale(ru) = %q, expected 'ru_RU'", got)
	}
	if got := OGLocale(ctxEn); got != "en_US" {
		t.Errorf("OGLocale(en) = %q, expected 'en_US'", got)
	}

	if got := OtherLang("ru"); got != "en" {
		t.Errorf("OtherLang(ru) = %q, expected 'en'", got)
	}
	if got := OtherLang("en"); got != "ru" {
		t.Errorf("OtherLang(en) = %q, expected 'ru'", got)
	}
}
