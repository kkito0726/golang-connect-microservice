.PHONY: proto build up down logs clean test docs logs-web sqlc

sqlc:
	cd services/user    && sqlc generate
	cd services/product && sqlc generate
	cd services/order   && sqlc generate
	cd services/payment && sqlc generate

proto:
	cd proto && buf generate

lint:
	cd proto && buf lint

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

down-v:
	docker compose down -v

logs:
	docker compose logs -f

logs-user:
	docker compose logs -f user-service

logs-product:
	docker compose logs -f product-service

logs-order:
	docker compose logs -f order-service

logs-payment:
	docker compose logs -f payment-service

logs-web:
	docker compose logs -f web

docs:
	./scripts/gen-docs.sh

test:
	go test ./...

clean:
	docker compose down -v
	rm -rf gen/
