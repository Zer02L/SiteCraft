-- name: ListResources :many
SELECT * FROM resources
ORDER BY order_index ASC;

-- name: ListActiveResources :many
SELECT * FROM resources
WHERE active = 1
ORDER BY order_index ASC;

-- name: CreateResource :one
INSERT INTO resources (
    id, name, role, avatar, active, order_index, service_ids, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING *;
