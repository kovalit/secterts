.PHONY: dev backend frontend extension docker-up docker-down migrate master-key backend-test frontend-build extension-build

dev:
	docker compose -f deploy/docker-compose.yml up

backend:
	cd apps/backend && go run ./cmd/api

migrate:
	cd apps/backend && go run ./cmd/migrate up

backend-test:
	cd apps/backend && go test ./...

frontend:
	cd apps/frontend && npm run dev

frontend-build:
	cd apps/frontend && npm run build

extension:
	cd apps/chrome-extension && npm run dev

extension-build:
	cd apps/chrome-extension && npm run build

docker-up:
	docker compose -f deploy/docker-compose.yml up -d

docker-down:
	docker compose -f deploy/docker-compose.yml down

master-key:
	./scripts/generate-master-key.sh
