package data

import (
	"context"
	"database/sql"
	"strings"

	"sitecraft/internal/db"
	"sitecraft/internal/models"

	"github.com/google/uuid"
)

// GetArticles returns published articles.
func (r *Repository) GetArticles(ctx context.Context) ([]models.ArticleItem, error) {
	c := r.Cache()
	if cached, ok := c.GetArticles(); ok {
		return cached, nil
	}

	rows, err := r.queries.ListPublishedArticles(ctx)
	if err != nil {
		return nil, err
	}

	articles := make([]models.ArticleItem, len(rows))
	for i, row := range rows {
		articles[i] = models.ArticleItem{
			ID:             row.ID,
			Title:          row.Title,
			Slug:           row.Slug,
			Excerpt:        row.Excerpt,
			Content:        row.Content,
			CoverImage:     row.CoverImage,
			AuthorName:     row.AuthorName,
			PublishedAt:    row.PublishedAt,
			IsPublished:    row.IsPublished == 1,
			ViewsCount:     int(row.ViewsCount),
			SEOTitle:       row.SeoTitle,
			SEODescription: row.SeoDescription,
			Tags:           strings.Split(row.Tags, ","),
		}
	}
	c.SetArticles(articles)
	return articles, nil
}

// GetArticleByID returns an article by ID.
func (r *Repository) GetArticleByID(ctx context.Context, id string) (models.ArticleItem, error) {
	art, err := r.queries.GetArticleByID(ctx, id)
	if err != nil {
		return models.ArticleItem{}, err
	}
	return models.ArticleItem{
		ID:             art.ID,
		Title:          art.Title,
		Slug:           art.Slug,
		Excerpt:        art.Excerpt,
		Content:        art.Content,
		CoverImage:     art.CoverImage,
		AuthorName:     art.AuthorName,
		PublishedAt:    art.PublishedAt,
		IsPublished:    art.IsPublished == 1,
		ViewsCount:     int(art.ViewsCount),
		SEOTitle:       art.SeoTitle,
		SEODescription: art.SeoDescription,
		Tags:           strings.Split(art.Tags, ","),
	}, nil
}

// SaveArticle creates or updates an article and invalidates articles cache.
func (r *Repository) SaveArticle(ctx context.Context, item models.ArticleItem) error {
	pubInt := int64(0)
	if item.IsPublished {
		pubInt = 1
	}
	tags := strings.Join(item.Tags, ",")

	var err error
	if item.ID != "" {
		_, err = r.queries.UpdateArticle(ctx, db.UpdateArticleParams{
			ID:             item.ID,
			Title:          item.Title,
			Slug:           item.Slug,
			Excerpt:        item.Excerpt,
			Content:        item.Content,
			CoverImage:     item.CoverImage,
			AuthorName:     item.AuthorName,
			PublishedAt:    item.PublishedAt,
			IsPublished:    pubInt,
			SeoTitle:       item.SEOTitle,
			SeoDescription: item.SEODescription,
			Tags:           tags,
		})
		if err == sql.ErrNoRows {
			_, err = r.queries.CreateArticle(ctx, db.CreateArticleParams{
				ID:             item.ID,
				Title:          item.Title,
				Slug:           item.Slug,
				Excerpt:        item.Excerpt,
				Content:        item.Content,
				CoverImage:     item.CoverImage,
				AuthorName:     item.AuthorName,
				PublishedAt:    item.PublishedAt,
				IsPublished:    pubInt,
				SeoTitle:       item.SEOTitle,
				SeoDescription: item.SEODescription,
				Tags:           tags,
			})
		}
	} else {
		_, err = r.queries.CreateArticle(ctx, db.CreateArticleParams{
			ID:             uuid.New().String(),
			Title:          item.Title,
			Slug:           item.Slug,
			Excerpt:        item.Excerpt,
			Content:        item.Content,
			CoverImage:     item.CoverImage,
			AuthorName:     item.AuthorName,
			PublishedAt:    item.PublishedAt,
			IsPublished:    pubInt,
			SeoTitle:       item.SEOTitle,
			SeoDescription: item.SEODescription,
			Tags:           tags,
		})
	}
	if err != nil {
		return err
	}
	r.Cache().InvalidateArticles()
	return nil
}

// ToggleArticle toggles publication status of an article and invalidates articles cache.
func (r *Repository) ToggleArticle(ctx context.Context, id string) error {
	art, err := r.queries.GetArticleByID(ctx, id)
	if err != nil {
		return err
	}
	newPub := int64(1)
	if art.IsPublished == 1 {
		newPub = 0
	}
	_, err = r.queries.UpdateArticle(ctx, db.UpdateArticleParams{
		ID:             art.ID,
		Title:          art.Title,
		Slug:           art.Slug,
		Excerpt:        art.Excerpt,
		Content:        art.Content,
		CoverImage:     art.CoverImage,
		AuthorName:     art.AuthorName,
		PublishedAt:    art.PublishedAt,
		IsPublished:    newPub,
		SeoTitle:       art.SeoTitle,
		SeoDescription: art.SeoDescription,
		Tags:           art.Tags,
	})
	if err != nil {
		return err
	}
	r.Cache().InvalidateArticles()
	return nil
}

// SoftDeleteArticle soft deletes an article and invalidates articles cache.
func (r *Repository) SoftDeleteArticle(ctx context.Context, id string) error {
	if err := r.queries.SoftDeleteArticle(ctx, id); err != nil {
		return err
	}
	r.Cache().InvalidateArticles()
	return nil
}
