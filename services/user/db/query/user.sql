-- name: CreateUser :one
INSERT INTO users (email, name, role, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING id, email, name, role, password_hash, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, email, name, role, password_hash, created_at, updated_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, email, name, role, password_hash, created_at, updated_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: ListUsers :many
SELECT id, email, name, role, password_hash, created_at, updated_at
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users WHERE deleted_at IS NULL;

-- name: UpdateUser :one
UPDATE users
SET name = $1, email = $2, updated_at = now()
WHERE id = $3 AND deleted_at IS NULL
RETURNING id, email, name, role, password_hash, created_at, updated_at;

-- name: SoftDeleteUser :execrows
UPDATE users
SET deleted_at = now(), updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;
