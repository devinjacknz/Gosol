.PHONY: build run test clean migrate-up migrate-down docker-up docker-down

# Development commands
build:
	cd backend && go build -o main

run:
	cd backend && go run main.go

test:
	cd backend && go test ./...
	cd ml-service && python -m pytest

clean:
	cd backend && rm -f main
	find . -type d -name "__pycache__" -exec rm -r {} +
	find . -type f -name "*.pyc" -delete

# Database migrations
migrate-up:
	cd backend && migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	cd backend && migrate -path migrations -database "$(DATABASE_URL)" down

# Docker commands
docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Development environment setup
setup-dev: setup-backend setup-ml

setup-backend:
	cd backend && go mod download && go mod tidy

setup-ml:
	cd ml-service && pip install -r requirements.txt

# Database commands
db-psql:
	docker-compose exec postgres psql -U postgres -d solmeme_trader

db-redis-cli:
	docker-compose exec redis redis-cli

# Service status
status:
	@echo "Checking service status..."
	@docker-compose ps

# Helper commands
backend-shell:
	docker-compose exec backend sh

ml-shell:
	docker-compose exec ml-service bash

# Default command
help:
	@echo "Available commands:"
	@echo "  make build         - Build the backend application"
	@echo "  make run          - Run the backend application locally"
	@echo "  make test         - Run tests for both backend and ML service"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make migrate-up   - Run database migrations up"
	@echo "  make migrate-down - Roll back database migrations"
	@echo "  make docker-up    - Start all services with Docker Compose"
	@echo "  make docker-down  - Stop all services"
	@echo "  make docker-logs  - View logs from all services"
	@echo "  make setup-dev    - Set up development environment"
	@echo "  make db-psql     - Connect to PostgreSQL database"
	@echo "  make db-redis-cli - Connect to Redis CLI"
	@echo "  make status      - Check status of all services"
	@echo "  make help        - Show this help message"
