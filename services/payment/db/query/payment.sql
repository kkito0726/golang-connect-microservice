-- name: CreatePayment :one
INSERT INTO payments (order_id, user_id, amount_cents, status, method)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, order_id, user_id, amount_cents, status, method, created_at, updated_at;

-- name: GetPaymentByID :one
SELECT id, order_id, user_id, amount_cents, status, method, created_at, updated_at
FROM payments
WHERE id = $1;

-- name: ListPayments :many
SELECT id, order_id, user_id, amount_cents, status, method, created_at, updated_at
FROM payments
WHERE (sqlc.narg('order_id')::uuid IS NULL OR order_id = sqlc.narg('order_id')::uuid)
  AND (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id')::uuid)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountPayments :one
SELECT COUNT(*)
FROM payments
WHERE (sqlc.narg('order_id')::uuid IS NULL OR order_id = sqlc.narg('order_id')::uuid)
  AND (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id')::uuid);

-- name: UpdatePaymentStatus :one
UPDATE payments
SET status = $1, updated_at = now()
WHERE id = $2
RETURNING id, order_id, user_id, amount_cents, status, method, created_at, updated_at;
