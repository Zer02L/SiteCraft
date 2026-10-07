package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SchemaPostalAddress represents Schema.org PostalAddress
type SchemaPostalAddress struct {
	Type            string `json:"@type"`
	StreetAddress   string `json:"streetAddress,omitempty"`
	AddressLocality string `json:"addressLocality,omitempty"`
	AddressCountry  string `json:"addressCountry,omitempty"`
}

// SchemaLocalBusiness represents Schema.org LocalBusiness (or Organization)
type SchemaLocalBusiness struct {
	Context     string               `json:"@context"`
	Type        string               `json:"@type"`
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	Telephone   string               `json:"telephone,omitempty"`
	Address     *SchemaPostalAddress `json:"address,omitempty"`
	OpeningHours []string            `json:"openingHours,omitempty"`
	Image       string               `json:"image,omitempty"`
	URL         string               `json:"url,omitempty"`
}

// SchemaAuthor represents an author in Schema.org
type SchemaAuthor struct {
	Type string `json:"@type"`
	Name string `json:"name"`
}

// SchemaArticle represents Schema.org BlogPosting / Article
type SchemaArticle struct {
	Context       string        `json:"@context"`
	Type          string        `json:"@type"`
	Headline      string        `json:"headline"`
	Description   string        `json:"description,omitempty"`
	Image         string        `json:"image,omitempty"`
	DatePublished string        `json:"datePublished,omitempty"`
	Author        *SchemaAuthor `json:"author,omitempty"`
	MainEntityOfPage string     `json:"mainEntityOfPage,omitempty"`
}

// SchemaBreadcrumbItem represents an item in BreadcrumbList
type SchemaBreadcrumbItem struct {
	Type     string `json:"@type"`
	Position int    `json:"position"`
	Name     string `json:"name"`
	Item     string `json:"item,omitempty"`
}

// SchemaBreadcrumbList represents Schema.org BreadcrumbList
type SchemaBreadcrumbList struct {
	Context         string                 `json:"@context"`
	Type            string                 `json:"@type"`
	ItemListElement []SchemaBreadcrumbItem `json:"itemListElement"`
}

// SEOData holds complete SEO meta tags and structured data for a page.
type SEOData struct {
	Title         string
	Description   string
	CanonicalURL  string
	OGType        string // "website", "article"
	OGImage       string
	TwitterCard   string // "summary_large_image"
	NoIndex       bool
	JSONLDObjects []any
}

// JSONLDScripts serializes all registered Schema.org objects into JSON-LD strings.
func (s SEOData) JSONLDScripts() []string {
	var scripts []string
	for _, obj := range s.JSONLDObjects {
		if obj == nil {
			continue
		}
		data, err := json.MarshalIndent(obj, "", "  ")
		if err == nil && len(data) > 0 {
			scripts = append(scripts, string(data))
		}
	}
	return scripts
}

// NewHomeSEO constructs default LocalBusiness and WebSite SEO for the home page.
func NewHomeSEO(settings SiteSettings, content SiteContent, siteURL string) SEOData {
	title := settings.SiteName
	if content.HeroTitle != "" {
		title = fmt.Sprintf("%s — %s", settings.SiteName, content.HeroTitle)
	}

	desc := content.HeroSubtitle
	if desc == "" {
		desc = settings.Tagline
	}

	phone := content.Phone1
	if phone == "" {
		phone = content.Phone2
	}

	var address *SchemaPostalAddress
	if content.Address != "" {
		address = &SchemaPostalAddress{
			Type:          "PostalAddress",
			StreetAddress: content.Address,
		}
	}

	var openingHours []string
	if content.WorkHours != "" {
		openingHours = append(openingHours, content.WorkHours)
	}

	lb := SchemaLocalBusiness{
		Context:      "https://schema.org",
		Type:         "LocalBusiness",
		Name:         settings.SiteName,
		Description:  desc,
		Telephone:    phone,
		Address:      address,
		OpeningHours: openingHours,
		Image:        content.HeroPhoto,
		URL:          siteURL,
	}

	return SEOData{
		Title:         title,
		Description:   desc,
		CanonicalURL:  siteURL,
		OGType:        "website",
		OGImage:       content.HeroPhoto,
		TwitterCard:   "summary_large_image",
		JSONLDObjects: []any{lb},
	}
}

// NewArticleSEO constructs Schema.org BlogPosting and Breadcrumbs for a blog post.
func NewArticleSEO(settings SiteSettings, content SiteContent, article ArticleItem, siteURL string) SEOData {
	title := fmt.Sprintf("%s — %s", article.Title, settings.SiteName)
	desc := article.Excerpt
	if desc == "" {
		desc = article.SEODescription
	}

	pageURL := fmt.Sprintf("%s/blog/%s", strings.TrimRight(siteURL, "/"), article.Slug)
	blogURL := fmt.Sprintf("%s/blog", strings.TrimRight(siteURL, "/"))

	authorName := article.AuthorName
	if authorName == "" {
		authorName = content.OwnerName
	}
	if authorName == "" {
		authorName = settings.SiteName
	}

	artSchema := SchemaArticle{
		Context:          "https://schema.org",
		Type:             "BlogPosting",
		Headline:         article.Title,
		Description:      desc,
		Image:            article.CoverImage,
		DatePublished:    article.PublishedAt,
		MainEntityOfPage: pageURL,
		Author: &SchemaAuthor{
			Type: "Person",
			Name: authorName,
		},
	}

	breadcrumbs := SchemaBreadcrumbList{
		Context: "https://schema.org",
		Type:    "BreadcrumbList",
		ItemListElement: []SchemaBreadcrumbItem{
			{Type: "ListItem", Position: 1, Name: "Главная", Item: siteURL},
			{Type: "ListItem", Position: 2, Name: "Блог", Item: blogURL},
			{Type: "ListItem", Position: 3, Name: article.Title, Item: pageURL},
		},
	}

	image := article.CoverImage
	if image == "" {
		image = content.HeroPhoto
	}

	return SEOData{
		Title:         title,
		Description:   desc,
		CanonicalURL:  pageURL,
		OGType:        "article",
		OGImage:       image,
		TwitterCard:   "summary_large_image",
		JSONLDObjects: []any{artSchema, breadcrumbs},
	}
}
