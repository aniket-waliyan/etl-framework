.PHONY: build test clean install lint

# Build variables
BINARY_NAME=etl-cli
BUILD_DIR=build
GO_FILES=$(shell find . -name '*.go')

# Build the CLI application
build:
	@echo "Building ETL CLI..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/etl-cli

# Install the CLI globally
install: build
	@echo "Installing ETL CLI..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

# Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Generate mocks for testing
mocks:
	@echo "Generating mocks..."
	@mockgen -source=internal/pipeline/interfaces.go -destination=internal/pipeline/mocks/mocks.go

# Initialize a new pipeline
new-pipeline:
	@echo "Creating new pipeline..."
	@$(BUILD_DIR)/$(BINARY_NAME) generate --name $(name)

# Validate a pipeline configuration
validate-pipeline:
	@echo "Validating pipeline configuration..."
	@$(BUILD_DIR)/$(BINARY_NAME) validate --config $(config) 