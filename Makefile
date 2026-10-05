.PHONY: build run clean build-all build-linux build-mac build-windows

APP_NAME=tui-cerebrum
BIN_DIR=bin

build:
	@echo "Building..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(APP_NAME) .

run: build
	@echo "Running..."
	@./$(BIN_DIR)/$(APP_NAME)

clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -f tui-cerebrum

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BIN_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 .

build-mac:
	@echo "Building for macOS (Apple Silicon)..."
	@mkdir -p $(BIN_DIR)
	@GOOS=darwin GOARCH=arm64 go build -o $(BIN_DIR)/$(APP_NAME)-darwin-arm64 .
	@echo "Building for macOS (Intel)..."
	@GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-darwin-amd64 .

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BIN_DIR)
	@GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-windows-amd64.exe .

build-all: build-linux build-mac build-windows
