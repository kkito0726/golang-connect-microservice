-- name: CreateOrder :one
INSERT INTO orders (user_id, status, total_cents)
VALUES ($1, $2, $3)
RETURNING id, user_id, status, total_cents, created_at, updated_at;

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, product_name, quantity, unit_price_cents)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, order_id, product_id, product_name, quantity, unit_price_cents;

-- name: GetOrderByID :one
SELECT id, user_id, status, total_cents, created_at, updated_at
FROM orders
WHERE id = $1;

-- name: GetOrderItems :many
SELECT id, order_id, product_id, product_name, quantity, unit_price_cents
FROM order_items
WHERE order_id = $1;

-- name: ListOrders :many
SELECT id, user_id, status, total_cents, created_at, updated_at
FROM orders
WHERE (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id')::uuid)
  AND (sqlc.narg('status')::varchar IS NULL OR status = sqlc.narg('status')::varchar)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountOrders :one
SELECT COUNT(*)
FROM orders
WHERE (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id')::uuid)
  AND (sqlc.narg('status')::varchar IS NULL OR status = sqlc.narg('status')::varchar);

-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $1, updated_at = now()
WHERE id = $2
RETURNING id, user_id, status, total_cents, created_at, updated_at;
