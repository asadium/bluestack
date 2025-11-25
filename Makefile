.PHONY: build run test docker-build docker-run clean help

# Variables
BINARY_NAME=bluestack
VERSION?=dev
DOCKER_IMAGE=bluestack:$(VERSION)
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_TEST=$(GO_CMD) test
GO_RUN=$(GO_CMD) run

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO_BUILD) -ldflags="-X github.com/asad/bluestack/internal/cli.Version=$(VERSION)" -o bin/$(BINARY_NAME) ./cmd/bluestack
	@echo "Build complete: bin/$(BINARY_NAME)"

# Run the application locally
run:
	@echo "Running $(BINARY_NAME)..."
	$(GO_RUN) ./cmd/bluestack start

# Run tests
test:
	@echo "Running tests..."
	$(GO_TEST) -v ./...

# Build Docker image
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE) -f docker/Dockerfile --build-arg VERSION=$(VERSION) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 4566:4566 -v $(PWD)/data:/app/data $(DOCKER_IMAGE)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf data/
	@echo "Clean complete"

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the Go binary"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run tests"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  clean        - Remove build artifacts"
	@echo "  help         - Show this help message"

