.PHONY: build test clean install lint fmt vet

BINARY_NAME=giwo
BUILD_DIR=bin
INSTALL_PATH=/usr/local/bin

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .

test:
	@echo "Running tests..."
	@go test -v ./...

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@go clean

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "$(BINARY_NAME) installed successfully!"

uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)

lint:
	@echo "Running linter..."
	@golangci-lint run

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Running go vet..."
	@go vet ./...

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

dev: deps fmt vet build test

release: clean deps fmt vet test build
	@echo "Release build complete!"

help:
	@echo "Available targets:"
	@echo "  build     - Build the binary"
	@echo "  test      - Run tests"
	@echo "  clean     - Clean build artifacts"
	@echo "  install   - Install binary to system"
	@echo "  uninstall - Remove binary from system"
	@echo "  lint      - Run linter"
	@echo "  fmt       - Format code"
	@echo "  vet       - Run go vet"
	@echo "  deps      - Download and tidy dependencies"
	@echo "  dev       - Development build (deps + fmt + vet + build + test)"
	@echo "  release   - Release build (clean + deps + fmt + vet + test + build)"
	@echo "  help      - Show this help"