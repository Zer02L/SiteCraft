package data

import (
	"fmt"
	"strings"
	"time"

	"sitecraft/internal/cache"
	"sitecraft/internal/models"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// RecordToArticleItem maps PocketBase record to models.ArticleItem
func RecordToArticleItem(r *core.Record) models.ArticleItem {
	var tags []string
	if rawTags := r.GetString("tags"); rawTags != "" {
		parts := strings.Split(rawTags, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	cover := ""
	if file := r.GetString("cover_image"); file != "" {
		cover = "/api/files/articles/" + r.Id + "/" + file
	}

	return models.ArticleItem{
		ID:             r.Id,
		Title:          r.GetString("title"),
		Slug:           r.GetString("slug"),
		Excerpt:        r.GetString("excerpt"),
		Content:        htmlSanitizer.Sanitize(r.GetString("content")),
		CoverImage:     cover,
		AuthorName:     r.GetString("author_name"),
		PublishedAt:    r.GetString("published_at"),
		IsPublished:    r.GetBool("is_published"),
		IsDeleted:      r.GetBool("is_deleted"),
		DeletedAt:      r.GetString("deleted_at"),
		ViewsCount:     r.GetInt("views_count"),
		SEOTitle:       r.GetString("seo_title"),
		SEODescription: r.GetString("seo_description"),
		Tags:           tags,
	}
}

// GetArticles retrieves published and live articles (not scheduled in future, not in trash) from cache or PocketBase
func GetArticles(app core.App) []models.ArticleItem {
	c := cache.GetStore()
	if cached, ok := c.GetArticles(); ok {
		return cached
	}

	records, err := app.FindRecordsByFilter("articles", "is_published = true && (is_deleted = false || is_deleted = null)", "-published_at,-created", 0, 0)
	if err != nil || len(records) == 0 {
		return nil
	}

	var articles []models.ArticleItem
	for _, r := range records {
		art := RecordToArticleItem(r)
		if art.IsLive() {
			articles = append(articles, art)
		}
	}

	c.SetArticles(articles)
	return articles
}

// GetArticleBySlug finds a published live article by slug
func GetArticleBySlug(app core.App, slug string) (*models.ArticleItem, error) {
	articles := GetArticles(app)
	for _, a := range articles {
		if a.Slug == slug && a.IsLive() {
			return &a, nil
		}
	}

	rec, err := app.FindFirstRecordByFilter("articles", "slug = {:slug} && is_published = true && (is_deleted = false || is_deleted = null)", dbx.Params{"slug": slug})
	if err != nil || rec == nil {
		return nil, fmt.Errorf("article not found")
	}

	art := RecordToArticleItem(rec)
	if !art.IsLive() {
		return nil, fmt.Errorf("article is scheduled or not published")
	}
	return &art, nil
}

// IncrementArticleViews increments the views counter for an article
func IncrementArticleViews(app core.App, id string) {
	rec, err := app.FindRecordById("articles", id)
	if err == nil && rec != nil {
		rec.Set("views_count", rec.GetInt("views_count")+1)
		_ = app.Save(rec)
	}
}

// GetAllArticlesAdmin retrieves all active articles (not in trash) for client admin
func GetAllArticlesAdmin(app core.App) []models.ArticleItem {
	records, err := app.FindRecordsByFilter("articles", "is_deleted = false || is_deleted = null", "-created,-published_at", 0, 0)
	if err != nil {
		records, err = app.FindRecordsByFilter("articles", "", "-created", 0, 0)
		if err != nil {
			return nil
		}
	}

	var articles []models.ArticleItem
	for _, r := range records {
		if r.GetBool("is_deleted") {
			continue
		}
		articles = append(articles, RecordToArticleItem(r))
	}
	return articles
}

// GetTrashArticlesAdmin retrieves deleted articles (in trash) for client admin
func GetTrashArticlesAdmin(app core.App) []models.ArticleItem {
	records, err := app.FindRecordsByFilter("articles", "is_deleted = true", "-deleted_at,-updated", 0, 0)
	if err != nil || len(records) == 0 {
		return nil
	}

	var articles []models.ArticleItem
	for _, r := range records {
		articles = append(articles, RecordToArticleItem(r))
	}
	return articles
}

// ToggleArticlePublish toggles is_published status
func ToggleArticlePublish(app core.App, id string) (*models.ArticleItem, error) {
	rec, err := app.FindRecordById("articles", id)
	if err != nil || rec == nil {
		return nil, fmt.Errorf("статья не найдена")
	}
	current := rec.GetBool("is_published")
	rec.Set("is_published", !current)
	if !current && rec.GetString("published_at") == "" {
		rec.Set("published_at", time.Now().Format("2006-01-02 15:04"))
	}
	if err := app.Save(rec); err != nil {
		return nil, err
	}
	cache.GetStore().InvalidateArticles()
	art := RecordToArticleItem(rec)
	return &art, nil
}

// MoveArticleToTrash moves article to trash (soft delete)
func MoveArticleToTrash(app core.App, id string) error {
	rec, err := app.FindRecordById("articles", id)
	if err != nil || rec == nil {
		return fmt.Errorf("статья не найдена")
	}
	rec.Set("is_deleted", true)
	rec.Set("deleted_at", time.Now().Format("2006-01-02 15:04:05"))
	if err := app.Save(rec); err != nil {
		return err
	}
	cache.GetStore().InvalidateArticles()
	return nil
}

// RestoreArticleFromTrash restores article from trash
func RestoreArticleFromTrash(app core.App, id string) error {
	rec, err := app.FindRecordById("articles", id)
	if err != nil || rec == nil {
		return fmt.Errorf("статья не найдена")
	}
	rec.Set("is_deleted", false)
	rec.Set("deleted_at", "")
	if err := app.Save(rec); err != nil {
		return err
	}
	cache.GetStore().InvalidateArticles()
	return nil
}

// PermanentDeleteArticle permanently removes article from database
func PermanentDeleteArticle(app core.App, id string) error {
	rec, err := app.FindRecordById("articles", id)
	if err != nil || rec == nil {
		return fmt.Errorf("статья не найдена")
	}
	if err := app.Delete(rec); err != nil {
		return err
	}
	cache.GetStore().InvalidateArticles()
	return nil
}

// EmptyTrashArticles permanently deletes all articles in trash
func EmptyTrashArticles(app core.App) error {
	records, err := app.FindRecordsByFilter("articles", "is_deleted = true", "", 0, 0)
	if err != nil {
		return err
	}
	for _, rec := range records {
		_ = app.Delete(rec)
	}
	cache.GetStore().InvalidateArticles()
	return nil
}
