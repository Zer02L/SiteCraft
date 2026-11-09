-- name: GetPageBySlug :one
SELECT * FROM pages
WHERE slug = ? LIMIT 1;

-- name: ListPages :many
SELECT * FROM pages
ORDER BY created_at DESC;

-- name: CreatePage :one
INSERT INTO pages (
    id, slug, title, seo_title, seo_description, is_published, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING *;

-- name: UpdatePage :one
UPDATE pages
SET slug = ?,
    title = ?,
    seo_title = ?,
    seo_description = ?,
    is_published = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeletePage :exec
DELETE FROM pages
WHERE id = ?;

-- name: ListPageBlocks :many
SELECT * FROM page_blocks
WHERE (page_id = ? OR page_id = '') AND active = 1
ORDER BY order_index ASC;

-- name: ListAllPageBlocks :many
SELECT * FROM page_blocks
ORDER BY order_index ASC;

-- name: GetPageBlockByID :one
SELECT * FROM page_blocks
WHERE id = ? LIMIT 1;

-- name: CreatePageBlock :one
INSERT INTO page_blocks (
    id, page_id, type, title, content, order_index, active, payload, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING *;

-- name: UpdatePageBlock :one
UPDATE page_blocks
SET type = ?,
    title = ?,
    content = ?,
    order_index = ?,
    active = ?,
    payload = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: UpdatePageBlockOrder :exec
UPDATE page_blocks
SET order_index = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeletePageBlock :exec
DELETE FROM page_blocks
WHERE id = ?;

-- name: DeleteAllPageBlocks :exec
DELETE FROM page_blocks;
