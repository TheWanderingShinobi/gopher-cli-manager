# Go parameters
GO=go
BUILD_DIR=bin
BINARY_NAME=gopher-cli-manager

# Default build command
all: build

# Build the binary
build:
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd

# Run the program
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run tests
test:
	$(GO) test ./...

# Clean up build artifacts
clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME)

# Tidy the Go module
tidy:
	$(GO) mod tidy

# Format code
fmt:
	$(GO) fmt ./...

# Generate any Go files, e.g. mock files
generate:
	$(GO) generate ./...

# Add a help command to list Makefile options
help:
	@echo "Makefile options:"
	@echo "  make build       - Build the Go binary"
	@echo "  make run         - Build and run the binary"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Remove the binary and clean up"
	@echo "  make tidy        - Tidy the Go module"
	@echo "  make fmt         - Format the code"
	@echo "  make generate    - Run go generate"
	@echo "  make help        - Display this help message"


