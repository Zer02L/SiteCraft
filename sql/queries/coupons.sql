-- name: GetCouponByCode :one
SELECT * FROM coupons
WHERE code = ? AND active = 1
LIMIT 1;

-- name: ListCoupons :many
SELECT * FROM coupons
ORDER BY created_at DESC;

-- name: CreateCoupon :one
INSERT INTO coupons (
    id, code, discount_type, discount_value, min_order_amount, valid_until, usage_limit, used_count, active, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING *;

-- name: IncrementCouponUsed :exec
UPDATE coupons
SET used_count = used_count + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;
