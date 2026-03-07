#!/bin/sh
set -e

echo "Running user migrations..."
migrate -path=/migrations/user -database "postgres://postgres:postgres@postgres:5432/user_db?sslmode=disable" up

echo "Running product migrations..."
migrate -path=/migrations/product -database "postgres://postgres:postgres@postgres:5432/product_db?sslmode=disable" up

echo "Running order migrations..."
migrate -path=/migrations/order -database "postgres://postgres:postgres@postgres:5432/order_db?sslmode=disable" up

echo "Running payment migrations..."
migrate -path=/migrations/payment -database "postgres://postgres:postgres@postgres:5432/payment_db?sslmode=disable" up

echo "All migrations completed successfully."
