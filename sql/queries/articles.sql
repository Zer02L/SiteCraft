-- name: ListArticles :many
SELECT * FROM articles
WHERE is_deleted = 0
ORDER BY created_at DESC;

-- name: ListPublishedArticles :many
SELECT * FROM articles
WHERE is_deleted = 0 AND is_published = 1
ORDER BY published_at DESC, created_at DESC;

-- name: GetArticleBySlug :one
SELECT * FROM articles
WHERE slug = ? AND is_deleted = 0
LIMIT 1;

-- name: GetArticleByID :one
SELECT * FROM articles
WHERE id = ? AND is_deleted = 0
LIMIT 1;

-- name: CreateArticle :one
INSERT INTO articles (
    id, title, slug, excerpt, content, cover_image, author_name, published_at, is_published,
    is_deleted, deleted_at, views_count, seo_title, seo_description, tags, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?,
    0, '', 0, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING *;

-- name: UpdateArticle :one
UPDATE articles
SET title = ?,
    slug = ?,
    excerpt = ?,
    content = ?,
    cover_image = ?,
    author_name = ?,
    published_at = ?,
    is_published = ?,
    seo_title = ?,
    seo_description = ?,
    tags = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: SoftDeleteArticle :exec
UPDATE articles
SET is_deleted = 1,
    deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: IncrementArticleViews :exec
UPDATE articles
SET views_count = views_count + 1
WHERE id = ?;

-- name: DeleteArticle :exec
DELETE FROM articles
WHERE id = ?;
