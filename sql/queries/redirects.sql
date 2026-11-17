-- name: ListRedirects :many
SELECT * FROM redirects
ORDER BY created_at DESC;

-- name: GetRedirectBySource :one
SELECT * FROM redirects
WHERE source_path = ? LIMIT 1;

-- name: GetRedirectByID :one
SELECT * FROM redirects
WHERE id = ? LIMIT 1;

-- name: CreateRedirect :one
INSERT INTO redirects (
    id, source_path, target_path, status_code, created_at
) VALUES (
    ?, ?, ?, ?, CURRENT_TIMESTAMP
)
RETURNING *;

-- name: UpsertRedirect :one
INSERT INTO redirects (
    id, source_path, target_path, status_code, created_at
) VALUES (
    ?, ?, ?, ?, CURRENT_TIMESTAMP
)
ON CONFLICT(source_path) DO UPDATE SET
    target_path = excluded.target_path,
    status_code = excluded.status_code
RETURNING *;

-- name: DeleteRedirect :exec
DELETE FROM redirects
WHERE id = ?;

-- name: DeleteRedirectBySource :exec
DELETE FROM redirects
WHERE source_path = ?;

-- name: CountRedirects :one
SELECT COUNT(*) FROM redirects;
