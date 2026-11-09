-- name: ListServices :many
SELECT * FROM services
WHERE is_deleted = 0
ORDER BY order_index ASC;

-- name: ListActiveServices :many
SELECT * FROM services
WHERE is_deleted = 0 AND active = 1
ORDER BY order_index ASC;

-- name: GetServiceByID :one
SELECT * FROM services
WHERE id = ? AND is_deleted = 0
LIMIT 1;

-- name: CreateService :one
INSERT INTO services (
    id, title, description, price, order_index, active, stock_count, is_preorder, is_deleted, deleted_at, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, 0, '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING *;

-- name: UpdateService :one
UPDATE services
SET title = ?,
    description = ?,
    price = ?,
    order_index = ?,
    active = ?,
    stock_count = ?,
    is_preorder = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: SoftDeleteService :exec
UPDATE services
SET is_deleted = 1,
    deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteService :exec
DELETE FROM services
WHERE id = ?;

-- name: DeleteAllServices :exec
DELETE FROM services;
