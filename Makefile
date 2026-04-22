.PHONY: build docker-build run docker-run unit integration test coverage bench lint clean help
.PHONY: docker-compose-up docker-compose-down docker-compose-logs docker-compose-restart
.PHONY: fmt vet check metrics dev

BINARY_DIR=bin
BINARY_NAME=$(BINARY_DIR)/analyzer
DOCKER_IMAGE=web-page-analyzer

# Build
build:
	@echo "Building local..."
	@mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_NAME) ./cmd/server/main.go

docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

# Run
run: build
	@echo "Running local..."
	CONFIG_PATH=configs/development.json ./$(BINARY_NAME)

docker-run: docker-build
	@echo "Running container..."
	docker run -p 8080:8080 $(DOCKER_IMAGE)

dev:
	@echo "Running with hot reload..."
	air

# Tests
unit:
	@echo "Running unit tests..."
	go test ./... -count=1 -parallel 4

integration:
	@echo "Running integration tests..."
	@echo "Select an integration test to run:"
	@echo "1. All integration tests"
	@echo "2. TestAnalyzeIntegration"
	@echo "3. All RateLimiter tests (Inbound + Outbound)"
	@read -p "Enter your choice: " choice; \
	case $$choice in \
		1) go test ./test/integration/... -count=1 -parallel 4 ;; \
		2) go test ./test/integration/... -run TestAnalyzeIntegration -count=1 ;; \
		3) go test ./test/integration/... -run RateLimiter -count=1 -parallel 4 ;; \
		*) echo "Invalid choice." ;; \
	esac

test: unit
	@echo "Running integration tests..."
	go test ./test/integration/... -count=1 -parallel 4

coverage:
	@echo "Running tests with coverage report..."
	go test ./... -coverprofile=cover.out -count=1
	go tool cover -func=cover.out

coverage-html: coverage
	go tool cover -html=cover.out

bench:
	@echo "Running all benchmarks..."
	go test -bench=. -benchmem ./...

# Code quality
fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Running go vet..."
	go vet ./...

lint:
	@echo "Running linter..."
	golangci-lint run ./...

check: fmt vet lint unit
	@echo "All checks passed!"

# Docker Compose
docker-compose-up:
	docker compose up --build -d
	@echo "Services running:"
	@echo "  App:       http://localhost:8080"
	@echo "  Prometheus: http://localhost:9090"
	@echo "  Grafana:    http://localhost:3000"

docker-compose-down:
	docker compose down

docker-compose-logs:
	docker compose logs -f

docker-compose-restart:
	docker compose restart

docker-compose-ps:
	docker compose ps

# Utilities
metrics:
	@echo "Fetching metrics from localhost:8080/metrics..."
	@curl -s http://localhost:8080/metrics | grep -E "^(web_analyzer|TYPE|HELP)" | head -40

clean:
	@echo "Cleaning up..."
	rm -rf $(BINARY_DIR)
	rm -f cover.out

# Help
help:
	@echo "Available commands:"
	@echo ""
	@echo "  Build & Run:"
	@echo "    build              - Build binary locally"
	@echo "    docker-build       - Build Docker image"
	@echo "    run                - Run locally (build first)"
	@echo "    docker-run         - Run in Docker container"
	@echo "    dev                - Run with hot reload (requires air)"
	@echo ""
	@echo "  Tests:"
	@echo "    unit               - Run unit tests"
	@echo "    integration        - Run integration tests (interactive)"
	@echo "    test               - Run all tests (unit + integration)"
	@echo "    coverage           - Run tests with coverage"
	@echo "    coverage-html      - Run tests with HTML coverage report"
	@echo "    bench              - Run benchmarks"
	@echo ""
	@echo "  Code Quality:"
	@echo "    fmt                - Format code (go fmt)"
	@echo "    vet                - Run go vet"
	@echo "    lint               - Run golangci-lint"
	@echo "    check              - Run all checks (fmt + vet + lint + unit)"
	@echo ""
	@echo "  Docker Compose:"
	@echo "    docker-compose-up     - Start all services (app + prometheus + grafana)"
	@echo "    docker-compose-down   - Stop all services"
	@echo "    docker-compose-logs   - View logs (follow mode)"
	@echo "    docker-compose-restart - Restart all services"
	@echo "    docker-compose-ps     - Show service status"
	@echo ""
	@echo "  Utilities:"
	@echo "    metrics            - Fetch metrics from /metrics endpoint"
	@echo "    clean              - Remove build artifacts"