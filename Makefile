.PHONY: dev build-mac build-windows build-linux build-all clean install-wails

APP_NAME := claude-watch
BUILD_DIR := build/bin

# Install Wails CLI (run once)
install-wails:
	go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Development mode with hot reload
dev:
	wails dev

# Build for current platform
build:
	wails build

# Build for macOS (universal binary - Intel + Apple Silicon)
build-mac:
	wails build -platform darwin/universal
	@echo "Built: $(BUILD_DIR)/$(APP_NAME).app"

# Build for Windows
build-windows:
	wails build -platform windows/amd64
	@echo "Built: $(BUILD_DIR)/$(APP_NAME).exe"

# Build for Linux
build-linux:
	wails build -platform linux/amd64
	@echo "Built: $(BUILD_DIR)/$(APP_NAME)"

# Build for all platforms
build-all: build-mac build-windows build-linux
	@echo ""
	@echo "==================================="
	@echo "Built all platforms:"
	@echo "  macOS:   $(BUILD_DIR)/$(APP_NAME).app"
	@echo "  Windows: $(BUILD_DIR)/$(APP_NAME).exe"  
	@echo "  Linux:   $(BUILD_DIR)/$(APP_NAME)"
	@echo "==================================="

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -rf frontend/dist

# Run tests
test:
	go test ./...

# Generate bindings (useful after changing Go code)
generate:
	wails generate module

# Install app to /Applications (macOS)
install-mac: build-mac
	cp -r $(BUILD_DIR)/$(APP_NAME).app /Applications/
	@echo "Installed to /Applications/$(APP_NAME).app"

# Help
help:
	@echo "Claude Watch - Build Commands"
	@echo ""
	@echo "  make install-wails  - Install Wails CLI (run once)"
	@echo "  make dev            - Run in development mode"
	@echo "  make build          - Build for current platform"
	@echo "  make build-mac      - Build for macOS"
	@echo "  make build-windows  - Build for Windows"
	@echo "  make build-linux    - Build for Linux"
	@echo "  make build-all      - Build for all platforms"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make install-mac    - Install to /Applications (macOS)"
