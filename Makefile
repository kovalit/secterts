IMAGE_REGISTRY ?= ttdocker.me
BACKEND_IMAGE ?= $(IMAGE_REGISTRY)/secrets-center-backend:latest
DEPLOY_HOST ?= 176.57.218.35
REMOTE_USER ?= root
REMOTE_DIR ?= /home/dockers/secrets-center
REMOTE_FRONTEND_DIR ?= /home/react/secrets
COMPOSE_FILE ?= docker-compose.yml

.PHONY: dev backend frontend extension docker-up docker-down migrate master-key backend-test frontend-build extension-build deploy check-env publish-image build-frontend deploy-frontend start logs

dev:
	docker compose --env-file .env -f docker-compose.yml up

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
	docker compose --env-file .env -f docker-compose.yml up -d

docker-down:
	docker compose --env-file .env -f docker-compose.yml down

master-key:
	./scripts/generate-master-key.sh

check-env:
	@test -f .env || { echo "Missing .env in project root"; exit 1; }

publish-image:
	docker buildx build --platform linux/amd64 --tag $(BACKEND_IMAGE) --push apps/backend

build-frontend: check-env
	@VITE_API_URL="$$(sed -n 's/^VITE_API_URL=//p' .env | tail -n 1)"; \
	test -n "$$VITE_API_URL" || { echo "VITE_API_URL is missing in .env"; exit 1; }; \
	cd apps/frontend && npm ci && VITE_API_URL="$$VITE_API_URL" npm run build

deploy-frontend: build-frontend
	ssh $(REMOTE_USER)@$(DEPLOY_HOST) "mkdir -p $(REMOTE_FRONTEND_DIR)"
	rsync -az --delete apps/frontend/dist/ $(REMOTE_USER)@$(DEPLOY_HOST):$(REMOTE_FRONTEND_DIR)/
	ssh $(REMOTE_USER)@$(DEPLOY_HOST) "chmod -R a+rX $(REMOTE_FRONTEND_DIR) && test -f $(REMOTE_FRONTEND_DIR)/index.html"

deploy: check-env publish-image build-frontend
	@set -e; \
	echo "Deploying to $(REMOTE_USER)@$(DEPLOY_HOST):$(REMOTE_DIR)"; \
	ssh $(REMOTE_USER)@$(DEPLOY_HOST) "mkdir -p $(REMOTE_DIR) $(REMOTE_FRONTEND_DIR)"; \
	scp $(COMPOSE_FILE) .env Makefile $(REMOTE_USER)@$(DEPLOY_HOST):$(REMOTE_DIR)/; \
	ssh $(REMOTE_USER)@$(DEPLOY_HOST) "chmod 600 $(REMOTE_DIR)/.env && make -C $(REMOTE_DIR) start"; \
	rsync -az --delete apps/frontend/dist/ $(REMOTE_USER)@$(DEPLOY_HOST):$(REMOTE_FRONTEND_DIR)/; \
	ssh $(REMOTE_USER)@$(DEPLOY_HOST) "chmod -R a+rX $(REMOTE_FRONTEND_DIR) && test -f $(REMOTE_FRONTEND_DIR)/index.html"

start:
	docker compose --env-file .env -f docker-compose.yml pull
	docker compose --env-file .env -f docker-compose.yml up -d --remove-orphans

logs:
	docker compose --env-file .env -f docker-compose.yml logs -f --tail=100
