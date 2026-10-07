package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sitecraft/internal/i18n"
)

func TestI18nWorkflow(t *testing.T) {
	t.Run("Default language is ru", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
		lang := i18n.DetectLang(req)
		if lang != "ru" {
			t.Errorf("expected 'ru', got %q", lang)
		}
	})

	t.Run("Query param ?lang=en is detected and normalized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://localhost/?lang=en", nil)
		lang := i18n.DetectLang(req)
		if lang != "en" {
			t.Errorf("expected 'en', got %q", lang)
		}
	})

	t.Run("Cookie lang=en is detected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
		req.AddCookie(&http.Cookie{Name: "lang", Value: "en"})
		lang := i18n.DetectLang(req)
		if lang != "en" {
			t.Errorf("expected 'en', got %q", lang)
		}
	})

	t.Run("Accept-Language header is detected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9,ru;q=0.8")
		lang := i18n.DetectLang(req)
		if lang != "en" {
			t.Errorf("expected 'en', got %q", lang)
		}
	})

	t.Run("Translation helper T works with context", func(t *testing.T) {
		ctxRu := i18n.WithLocale(context.Background(), "ru")
		ctxEn := i18n.WithLocale(context.Background(), "en")

		if val := i18n.T(ctxRu, "nav.about"); val != "О нас" {
			t.Errorf("expected 'О нас', got %q", val)
		}
		if val := i18n.T(ctxEn, "nav.about"); val != "About" {
			t.Errorf("expected 'About', got %q", val)
		}
		if val := i18n.T(ctxRu, "btn.submit"); val != "Отправить заявку" {
			t.Errorf("expected 'Отправить заявку', got %q", val)
		}
		if val := i18n.T(ctxEn, "btn.submit"); val != "Submit Request" {
			t.Errorf("expected 'Submit Request', got %q", val)
		}
	})

	t.Run("Lead validation returns localized errors", func(t *testing.T) {
		input := LeadFormInput{
			Name:  "",
			Phone: "",
		}
		errsRu := ValidateLeadFormLocalized(input, "ru")
		errsEn := ValidateLeadFormLocalized(input, "en")

		if !strings.Contains(errsRu[0], "Укажите ваше имя") {
			t.Errorf("expected Russian error, got %v", errsRu)
		}
		if !strings.Contains(errsEn[0], "Please provide your name") {
			t.Errorf("expected English error, got %v", errsEn)
		}
	})
}
