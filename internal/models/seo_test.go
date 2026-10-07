package models

import (
	"strings"
	"testing"
)

func TestHomeSEO(t *testing.T) {
	settings := SiteSettings{
		SiteName: "Нотариус Иванов",
		Tagline:  "Профессиональные нотариальные услуги",
	}
	content := SiteContent{
		HeroTitle:    "Нотариальная контора в Москве",
		HeroSubtitle: "Оформление сделок любой сложности",
		Phone1:       "+7 (999) 000-00-00",
		Address:      "г. Москва, ул. Тверская, д. 1",
		WorkHours:    "Пн-Пт 09:00 - 18:00",
		HeroPhoto:    "/assets/hero.jpg",
	}

	seo := NewHomeSEO(settings, content, "https://notary.example.com")

	if !strings.Contains(seo.Title, "Нотариус Иванов") {
		t.Fatalf("expected title to contain site name, got: %s", seo.Title)
	}

	scripts := seo.JSONLDScripts()
	if len(scripts) != 1 {
		t.Fatalf("expected 1 JSON-LD script, got: %d", len(scripts))
	}

	script := scripts[0]
	if !strings.Contains(script, `"@type": "LocalBusiness"`) {
		t.Errorf("missing LocalBusiness type in JSON-LD: %s", script)
	}
	if !strings.Contains(script, "г. Москва, ул. Тверская, д. 1") {
		t.Errorf("missing address in JSON-LD: %s", script)
	}
	if !strings.Contains(script, "+7 (999) 000-00-00") {
		t.Errorf("missing telephone in JSON-LD: %s", script)
	}
}

func TestArticleSEO(t *testing.T) {
	settings := SiteSettings{
		SiteName: "SiteCraft",
	}
	content := SiteContent{
		OwnerName: "Алексей Смирнов",
		HeroPhoto: "/assets/default.jpg",
	}
	article := ArticleItem{
		Title:       "Как зарегистрировать ООО в 2026 году",
		Slug:        "kak-zaregistrirovat-ooo-2026",
		Excerpt:     "Пошаговая инструкция для начинающих предпринимателей",
		CoverImage:  "/assets/blog/ooo.jpg",
		AuthorName:  "Алексей Смирнов",
		PublishedAt: "2026-10-07",
	}

	seo := NewArticleSEO(settings, content, article, "https://example.com")

	if seo.OGType != "article" {
		t.Errorf("expected OGType 'article', got '%s'", seo.OGType)
	}
	if seo.OGImage != "/assets/blog/ooo.jpg" {
		t.Errorf("expected cover image, got '%s'", seo.OGImage)
	}

	scripts := seo.JSONLDScripts()
	if len(scripts) != 2 {
		t.Fatalf("expected 2 JSON-LD scripts (Article + BreadcrumbList), got: %d", len(scripts))
	}

	articleScript := scripts[0]
	if !strings.Contains(articleScript, `"@type": "BlogPosting"`) {
		t.Errorf("missing BlogPosting type: %s", articleScript)
	}
	if !strings.Contains(articleScript, "Алексей Смирнов") {
		t.Errorf("missing AuthorName: %s", articleScript)
	}

	breadcrumbScript := scripts[1]
	if !strings.Contains(breadcrumbScript, `"@type": "BreadcrumbList"`) {
		t.Errorf("missing BreadcrumbList type: %s", breadcrumbScript)
	}
}
