BINARY_NAME=food-recipe
APP_MAIN=./cmd/app/main.go
GO_MOD_FLAG=$(if $(wildcard vendor),-mod=vendor,)

.PHONY: all help run dev build test test-cover fmt tidy clean docker-up docker-down docker-build docker-hasura docker-logs docker-logs-app docker-logs-hasura docker-clean docker-restart

all: help

help:
	@echo "Available commands:"
	@echo "  make run             - Run Go app locally"
	@echo "  make build           - Build Go application binary into bin/"
	@echo "  make test            - Run all unit and integration tests"
	@echo "  make test-cover      - Run tests with code coverage report"
	@echo "  make fmt             - Format all Go source files"
	@echo "  make tidy            - Tidy and vendor Go module dependencies"
	@echo "  make clean           - Remove build artifacts and temp binaries"
	@echo ""
	@echo "Docker & Hasura commands:"
	@echo "  make docker-up       - Start the full stack (Postgres + Hasura + App) in background"
	@echo "  make docker-hasura   - Start only Postgres + Hasura (for local app development)"
	@echo "  make docker-down     - Stop and remove all containers"
	@echo "  make docker-build    - Rebuild app container image"
	@echo "  make docker-logs     - Follow logs from all running containers"
	@echo "  make docker-clean    - Stop containers and remove Postgres database volume (fresh start)"
	@echo "  make docker-restart  - Restart all containers"

##### APPLICATION TARGETS #####

run:
	go run $(GO_MOD_FLAG) $(APP_MAIN)

dev: run

build:
	go build $(GO_MOD_FLAG) -o bin/$(BINARY_NAME) $(APP_MAIN)

test:
	go test $(GO_MOD_FLAG) -count=1 -v ./...

test-cover:
	go test $(GO_MOD_FLAG) -count=1 -cover ./...

fmt:
	gofmt -s -w cmd/ internal/

tidy:
	go mod tidy
	go mod vendor

ifeq ($(OS),Windows_NT)
CLEAN_CMD = (if exist bin rmdir /s /q bin) & (if exist tmp rmdir /s /q tmp) & (if exist $(BINARY_NAME) del /f /q $(BINARY_NAME)) & (if exist $(BINARY_NAME).exe del /f /q $(BINARY_NAME).exe) & (if exist app.exe del /f /q app.exe)
else
CLEAN_CMD = rm -rf bin/ tmp/ $(BINARY_NAME) $(BINARY_NAME).exe app.exe
endif

clean:
	@$(CLEAN_CMD)

##### DOCKER & HASURA TARGETS #####

docker-up:
	docker compose up -d

docker-hasura:
	docker compose up -d postgres hasura

docker-down:
	docker compose down

docker-build:
	docker compose build

docker-logs:
	docker compose logs -f

docker-logs-app:
	docker compose logs -f app

docker-logs-hasura:
	docker compose logs -f hasura

docker-clean:
	docker compose down -v

docker-restart:
	docker compose restart
