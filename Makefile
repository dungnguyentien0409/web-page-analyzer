.PHONY: build docker-build run docker-run unit integration coverage clean

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
	go test ./test/integration/...

coverage:
	@echo "Running unit tests with coverage report..."
	go test ./... -coverprofile=cover.out
	go tool cover -func=cover.out

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -f cover.out