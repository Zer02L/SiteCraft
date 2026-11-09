-- name: ListReviews :many
SELECT * FROM reviews
ORDER BY created_at DESC;

-- name: ListPublishedReviews :many
SELECT * FROM reviews
WHERE is_published = 1
ORDER BY created_at DESC;

-- name: CreateReview :one
INSERT INTO reviews (
    id, author_name, rating, comment, order_id, photo, is_published, source, created_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP
)
RETURNING *;

-- name: ApproveReview :one
UPDATE reviews
SET is_published = 1
WHERE id = ?
RETURNING *;
