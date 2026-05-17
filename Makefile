.PHONY: build admin migrate docker-build docker-up docker-down help

help:
	@echo "Available targets:"
	@echo "  build        - Build the qalbum-server binary"
	@echo "  admin        - Build the qalbum-admin binary"
	@echo "  migrate      - Run database migrations"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-up    - Start containers with docker-compose"
	@echo "  docker-down  - Stop containers"

build:
	@echo "Building qalbum-server..."
	@go build -o bin/qalbum-server ./cmd/server

admin:
	@echo "Building qalbum-admin..."
	@go build -o bin/qalbum-admin ./cmd/admin

migrate:
	@echo "Running migrations..."
	@go run cmd/server/main.go --migrate

docker-build:
	@echo "Building Docker image..."
	@docker build -t qalbum:latest .

docker-up:
	@echo "Starting containers..."
	@docker-compose up -d

docker-down:
	@echo "Stopping containers..."
	@docker-compose down

.DEFAULT_GOAL := help
