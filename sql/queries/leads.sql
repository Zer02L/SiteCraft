-- name: ListLeads :many
SELECT * FROM leads
ORDER BY created_at DESC;

-- name: ListLeadsPaged :many
SELECT * FROM leads
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountLeads :one
SELECT COUNT(*) FROM leads;

-- name: GetLeadByID :one
SELECT * FROM leads
WHERE id = ? LIMIT 1;

-- name: CreateLead :one
INSERT INTO leads (
    id, name, phone, email, source, booking_date, booking_time, resource_id, status, payload, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING *;

-- name: UpdateLeadStatus :one
UPDATE leads
SET status = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteLead :exec
DELETE FROM leads
WHERE id = ?;
