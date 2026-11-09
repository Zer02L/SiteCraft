-- name: GetOrderByCode :one
SELECT * FROM service_orders
WHERE order_code = ? LIMIT 1;

-- name: ListOrders :many
SELECT * FROM service_orders
ORDER BY created_at DESC;

-- name: CreateOrder :one
INSERT INTO service_orders (
    id, order_code, client_name, client_phone, item_title, current_stage, total_amount, is_paid, stages_history, notes, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING *;

-- name: UpdateOrderStatus :one
UPDATE service_orders
SET current_stage = ?,
    stages_history = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE order_code = ?
RETURNING *;

-- name: UpdateOrderPayment :one
UPDATE service_orders
SET is_paid = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE order_code = ?
RETURNING *;
