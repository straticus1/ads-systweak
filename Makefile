.PHONY: build clean run test

APP_NAME = ads-systweak
VERSION = 1.0.0
BUILD_DIR = bin

build: clean
	@echo "Building $(APP_NAME)..."
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64
	# Create universal binary
	lipo -create -output $(BUILD_DIR)/$(APP_NAME) $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(BUILD_DIR)/$(APP_NAME)-darwin-amd64
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME) (Universal)"

clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)

test:
	@echo "Running tests..."
	go test ./... -v
