BINARY_NAME := skiptui
BUILD_DIR   := bin
PKG         := skiptui/cmd/skiptui
VERSION     := 0.1.0
GIT_COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE  := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X 'skiptui/pkg/version.Version=$(VERSION)' \
           -X 'skiptui/pkg/version.GitCommit=$(GIT_COMMIT)' \
           -X 'skiptui/pkg/version.BuildDate=$(BUILD_DATE)'

.PHONY: all build clean test lint run setcap

all: build

build:
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME)..."
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(PKG)

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

test:
	@echo "Running tests..."
	go test -v ./...

lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null && golangci-lint run ./... || go vet ./...

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

setcap: build
	@chmod +x scripts/setup_caps.sh
	./scripts/setup_caps.sh $(BUILD_DIR)/$(BINARY_NAME)
