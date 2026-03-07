#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

mkdir -p "$PROJECT_DIR/docs"

docker run --rm \
  --platform linux/amd64 \
  -v "$PROJECT_DIR/proto:/protos" \
  -v "$PROJECT_DIR/docs:/out" \
  pseudomuto/protoc-gen-doc:latest \
  -I /protos \
  --doc_opt=html,index.html \
  user/v1/user.proto \
  product/v1/product.proto \
  order/v1/order.proto \
  payment/v1/payment.proto

echo "Documentation generated at docs/index.html"
