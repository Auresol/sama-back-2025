.PHONY: build run test clean deps dev

# Build the application
build:
	go build -o bin/api cmd/api/main.go

# Run the application
run:
	go run cmd/api/main.go

# Run in development mode
dev:
	go run cmd/api/main.go

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod tidy
	go mod download

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate API documentation (if you add swagger later)
docs:
	# swag init -g cmd/api/main.go

# Database migrations
migrate:
	# Add your migration commands here when you implement them

# Docker commands
docker-build:
	docker build -t sama-backend .

docker-run:
	docker run -p 8080:8080 --env-file .env sama-backend

docker-compose-up:
	docker compose up -d

docker-compose-down:
	docker compose down

docker-compose-logs:
	docker compose logs -f

docker-compose-restart:
	docker compose restart

# Swagger documentation
swagger-init:
	swag init -g cmd/api/main.go

swagger-serve:
	swag init -g cmd/api/main.go && swag serve

# Help
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  dev          - Run in development mode"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  docker-compose-up    - Start all services"
	@echo "  docker-compose-down  - Stop all services"
	@echo "  docker-compose-logs  - View logs"
	@echo "  docker-compose-restart - Restart services"
	@echo "  swagger-init - Initialize Swagger docs"
	@echo "  swagger-serve - Serve Swagger docs"
	@echo "  help         - Show this help" 