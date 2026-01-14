.PHONY: build install clean test

# Binary name
BINARY_NAME=gw
INSTALL_PATH=/usr/local/bin

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) .
	@echo "✓ Binary built at bin/$(BINARY_NAME)"

# Build and install to /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@cp bin/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME) 2>/dev/null || \
		(echo "Need sudo permissions..." && sudo cp bin/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME))
	@chmod +x $(INSTALL_PATH)/$(BINARY_NAME) 2>/dev/null || sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "✓ $(BINARY_NAME) installed to $(INSTALL_PATH)"
	@echo "✓ You can now run 'gw' from anywhere"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean
	@echo "✓ Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Uninstall from /usr/local/bin
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "✓ $(BINARY_NAME) uninstalled"

# Show help
help:
	@echo "Available targets:"
	@echo "  make build      - Build the binary to bin/gw"
	@echo "  make install    - Build and install to /usr/local/bin"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make test       - Run tests"
	@echo "  make uninstall  - Remove gw from /usr/local/bin"
