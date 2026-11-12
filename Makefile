include ./.env
export
export PATH := $(PATH):$(shell go env GOPATH 2>/dev/null)/bin
DATABASE_SERVICE_ADDRESS ?= localhost:$(DATABASE_SERVICE_PORT)
export DATABASE_SERVICE_ADDRESS

build-database-service:
	go -C internal/database-service build -o app .
run-database-service:
	./internal/database-service/app

build-task-service:
	go -C internal/task-service build -o app .
run-task-service:
	./internal/task-service/app

build-auth-service:
	go -C internal/auth-service build -o app .
run-auth-service:
	./internal/auth-service/app

test:
	go -C internal/core test ./...
	go -C internal/database-service test ./...
	go -C internal/task-service test ./...
	go -C internal/auth-service test ./...

test-docker:
	docker run --rm -v "$(CURDIR):/workspace" -w /workspace golang:1.27-bookworm sh -c 'set -e; for module in internal/core internal/database-service internal/task-service internal/auth-service; do (cd "$$module" && go test ./...); done'

generate-auth-keys:
	mkdir -p secrets
	docker run --rm --user "$(shell id -u):$(shell id -g)" -v "$(CURDIR)/secrets:/secrets" golang:1.27-bookworm openssl genpkey -algorithm ED25519 -out /secrets/auth-private.pem
	docker run --rm --user "$(shell id -u):$(shell id -g)" -v "$(CURDIR)/secrets:/secrets" golang:1.27-bookworm openssl pkey -in /secrets/auth-private.pem -pubout -out /secrets/auth-public.pem
	chmod 600 secrets/auth-private.pem

MIGRATION_DATABASE_URL := $(subst @$(DB_HOST):,@$(MIGRATION_DB_HOST):,$(DATABASE_URL))
migrate-up:
	@migrate -path "./internal/database-service/migrations" -database "$(MIGRATION_DATABASE_URL)" up

generate-grpc-code-database-service:
	protoc \
      -I . \
      --go_out=. \
      --go_opt=paths=source_relative \
      --go-grpc_out=. \
      --go-grpc_opt=paths=source_relative \
      internal/database-service/api/task/v1/task.proto \
      internal/database-service/api/board/v1/board.proto \
      internal/database-service/api/user/v1/user.proto \
      internal/database-service/api/auth/v1/auth.proto

deploy-service:
	docker compose --env-file .env -f docker-compose.yaml up -d --build --wait
	$(MAKE) migrate-up

.PHONY: build-database-service run-database-service build-task-service run-task-service build-auth-service run-auth-service test test-docker generate-auth-keys migrate-up generate-grpc-code-database-service deploy-service
