.PHONY: proto build up down logs clean test docs

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

docs:
	./scripts/gen-docs.sh

test:
	go test ./...

clean:
	docker compose down -v
	rm -rf gen/
