-- name: ListBranches :many
SELECT * FROM branches
ORDER BY order_index ASC;

-- name: ListActiveBranches :many
SELECT * FROM branches
WHERE active = 1
ORDER BY order_index ASC;

-- name: CreateBranch :one
INSERT INTO branches (
    id, name, city, address, phone, work_hours, lat, lng, metro, services, is_main, active, order_index, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING *;
