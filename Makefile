.PHONY: build docker-build run docker-run unit integration coverage bench lint clean docker-compose-up

BINARY_DIR=bin
BINARY_NAME=$(BINARY_DIR)/analyzer
DOCKER_IMAGE=web-page-analyzer

build:
	@echo "Building local..."
	@mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_NAME) ./cmd/server/main.go

docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

run: build
	@echo "Running local..."
	CONFIG_PATH=configs/development.json ./$(BINARY_NAME)

docker-run: docker-build
	@echo "Running container..."
	docker run -p 8080:8080 $(DOCKER_IMAGE)

unit:
	@echo "Running unit tests..."
	go test ./...

integration:
	@echo "Running integration tests..."
	@echo "Select an integration test to run:"
	@echo "1. All integration tests"
	@echo "2. TestAnalyzeIntegration"
	@echo "3. TestRateLimiterIntegration"
	@read -p "Enter your choice: " choice; \
	case $$choice in \
		1) go test ./test/integration/... ;; \
		2) go test ./test/integration/... -run TestAnalyzeIntegration ;; \
		3) go test ./test/integration/... -run TestRateLimiterIntegration ;; \
		*) echo "Invalid choice." ;; \
	esac

coverage:
	@echo "Running unit tests with coverage report..."
	go test ./... -coverprofile=cover.out
	go tool cover -func=cover.out

bench:
	@echo "Running all benchmarks..."
	go test -bench=. -benchmem ./...

lint:
	@echo "Running linter..."
	golangci-lint run ./...

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -f cover.out

docker-compose-up:
	docker compose up --build