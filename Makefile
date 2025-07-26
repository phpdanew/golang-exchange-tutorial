.PHONY: build run test clean docker-up docker-down docker-logs api-gen model-gen

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=exchange
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v .

# Run the application
run:
	$(GOBUILD) -o $(BINARY_NAME) -v .
	./$(BINARY_NAME) -f etc/exchange-api.yaml

# Test the application
test:
	$(GOTEST) -v ./...

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Generate API code from .api file
api-gen:
	goctl api go -api api/exchange.api -dir . --style=goZero

# Generate model code from SQL
model-gen:
	goctl model pg datasource -url="postgres://exchange_user:exchange_pass@localhost:5432/exchange?sslmode=disable" -table="*" -dir="./model" --style=goZero

# Docker commands
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-build:
	docker-compose build

# Database commands
db-init:
	docker-compose exec postgres psql -U exchange_user -d exchange -f /docker-entrypoint-initdb.d/init.sql

db-connect:
	docker-compose exec postgres psql -U exchange_user -d exchange

# Development setup
dev-setup: docker-up
	@echo "Waiting for database to be ready..."
	@sleep 10
	@echo "Development environment is ready!"
	@echo "Database: postgres://exchange_user:exchange_pass@localhost:5432/exchange"
	@echo "Redis: localhost:6379"

# Install dependencies
deps:
	$(GOGET) github.com/lib/pq
	$(GOGET) github.com/shopspring/decimal
	$(GOGET) github.com/zeromicro/go-zero/tools/goctl
	$(GOMOD) tidy

# Format code
fmt:
	$(GOCMD) fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v .

# Help
help:
	@echo "Available commands:"
	@echo "  build       - Build the application"
	@echo "  run         - Build and run the application"
	@echo "  test        - Run tests"
	@echo "  clean       - Clean build files"
	@echo "  tidy        - Tidy go modules"
	@echo "  api-gen     - Generate API code from .api file"
	@echo "  model-gen   - Generate model code from database"
	@echo "  docker-up   - Start Docker containers"
	@echo "  docker-down - Stop Docker containers"
	@echo "  docker-logs - Show Docker logs"
	@echo "  dev-setup   - Setup development environment"
	@echo "  deps        - Install dependencies"
	@echo "  fmt         - Format code"
	@echo "  lint        - Lint code"
	@echo "  help        - Show this help"