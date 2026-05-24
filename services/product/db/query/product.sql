-- name: CreateProduct :one
INSERT INTO products (sku, name, description, price_cents, stock_quantity, category)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, sku, name, description, price_cents, stock_quantity, category, created_at, updated_at;

-- name: GetProductByID :one
SELECT id, sku, name, description, price_cents, stock_quantity, category, created_at, updated_at
FROM products
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListProducts :many
SELECT id, sku, name, description, price_cents, stock_quantity, category, created_at, updated_at
FROM products
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListProductsByCategory :many
SELECT id, sku, name, description, price_cents, stock_quantity, category, created_at, updated_at
FROM products
WHERE deleted_at IS NULL AND category = $3
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountProducts :one
SELECT COUNT(*) FROM products WHERE deleted_at IS NULL;

-- name: CountProductsByCategory :one
SELECT COUNT(*) FROM products WHERE deleted_at IS NULL AND category = $1;

-- name: UpdateProduct :one
UPDATE products
SET name = $1, description = $2, price_cents = $3, category = $4, updated_at = now()
WHERE id = $5 AND deleted_at IS NULL
RETURNING id, sku, name, description, price_cents, stock_quantity, category, created_at, updated_at;

-- name: SoftDeleteProduct :execrows
UPDATE products
SET deleted_at = now(), updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetProductForUpdate :one
SELECT id, sku, name, description, price_cents, stock_quantity, category, created_at, updated_at
FROM products
WHERE id = $1 AND deleted_at IS NULL
FOR UPDATE;

-- name: UpdateProductStock :one
UPDATE products
SET stock_quantity = $1, updated_at = now()
WHERE id = $2
RETURNING id, sku, name, description, price_cents, stock_quantity, category, created_at, updated_at;

-- name: CreateStockMovement :one
INSERT INTO stock_movements (product_id, delta, reason, reference_id)
VALUES ($1, $2, $3, sqlc.narg('reference_id')::uuid)
RETURNING id, product_id, delta, reason, reference_id, created_at;

-- name: GetStockMovements :many
SELECT id, product_id, delta, reason, reference_id, created_at
FROM stock_movements
WHERE product_id = $1
ORDER BY created_at DESC
LIMIT $2;
