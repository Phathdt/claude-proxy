# Load environment variables from .env file if it exists
ifneq (,$(wildcard .env))
    include .env
    export
endif

.PHONY: run build clean sqlc-generate docker-up docker-down dev dev-setup format format-go format-check test test-unit test-integration test-coverage test-watch fe-install fe-dev fe-build fe be

run:
	go run .

be:
	@echo "Starting backend dev server on port 4000..."
	go run .

fe:
	@echo "Starting frontend dev server on port 5173..."
	cd frontend && pnpm dev

build: fe-build
	@echo "Building production binary with embedded frontend..."
	go build -o bin/claude-proxy .
	@echo "✅ Build complete: bin/claude-proxy"

fe-install:
	@echo "Installing frontend dependencies..."
	cd frontend && pnpm install

fe-dev:
	@echo "Starting frontend dev server..."
	cd frontend && pnpm dev

fe-build:
	@echo "Building frontend for production..."
	cd frontend && pnpm build

clean:
	rm -rf bin/
	rm -rf frontend/dist

sqlc-generate:
	sqlc generate

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

deps:
	go mod tidy

format:
	@echo "🎨 Formatting all Go files..."
	@find . -name "*.go" -type f -exec gofmt -w {} \;
	@echo "📦 Organizing imports..."
	@goimports -w .
	@echo "📏 Formatting line lengths..."
	@golines -w -m 120 .
	@echo "✨ Applying gofumpt formatting..."
	@gofumpt -extra -w .
	@echo "✅ Go files formatted successfully!"

format-go: format

format-check:
	@echo "🔍 Checking Go file formatting..."
	@if find . -name "*.go" -type f -exec gofmt -l {} \; | grep -q .; then \
		echo "❌ Some Go files are not properly formatted:"; \
		find . -name "*.go" -type f -exec gofmt -l {} \; | sed 's/^/  /'; \
		echo "Run 'make format' to fix formatting issues"; \
		exit 1; \
	else \
		echo "✅ All Go files are properly formatted"; \
	fi

test:
	go test ./... -v

test-unit:
	go test ./modules/... -v -short

test-integration:
	go test ./... -v -run Integration

test-coverage:
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

test-watch:
	@command -v entr >/dev/null 2>&1 || { echo "entr is required for watch mode. Install it first."; exit 1; }
	find . -name "*.go" | entr -r make test-unit

dev: docker-up
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Starting application..."
	@make run

dev-setup: format deps
	@echo "✅ Development environment setup complete"
