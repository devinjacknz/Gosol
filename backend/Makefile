.PHONY: all dev build test clean docker-up docker-down

# Default target
all: build

# Development
dev:
	go run main.go

# Build
build:
	go build -o bin/solmeme-trader main.go

# Test
test:
	go test -v ./...

# Clean
clean:
	rm -rf bin/
	go clean

# Docker commands
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# Database migrations
migrate-up:
	psql -h localhost -U postgres -d solmeme -f migrations/001_initial_schema.sql

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run linter
lint:
	go vet ./...

# Generate mocks (if needed)
mocks:
	go generate ./...

# Help
help:
	@echo "Available targets:"
	@echo "  dev          - Run the application in development mode"
	@echo "  build        - Build the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-up    - Start Docker containers"
	@echo "  docker-down  - Stop Docker containers"
	@echo "  migrate-up   - Run database migrations"
	@echo "  deps         - Install dependencies"
	@echo "  lint         - Run linter"
	@echo "  mocks        - Generate mocks"
