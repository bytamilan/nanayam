# Nanayam Makefile
# Unified build system for the nanayam CLI and services

.PHONY: all build install test clean lint

BINARY_NAME=nanayam
BUILD_DIR=./build
CLI_DIR=./cli
GATEWAY_DIR=./services/gateway
CONSOLE_DIR=./apps/operator-console

VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-X github.com/bytamilan/nanayam/cli/cmd.version=$(VERSION) \
	-X github.com/bytamilan/nanayam/cli/cmd.commit=$(COMMIT) \
	-X github.com/bytamilan/nanayam/cli/cmd.date=$(DATE)"

all: build

## build: Build the nanayam CLI binary
build:
	@echo "Building nanayam $(VERSION)..."
	cd $(CLI_DIR) && go build $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Binary: $(BUILD_DIR)/$(BINARY_NAME)"

## install: Install nanayam to ~/.nanayam/bin
install: build
	@echo "Installing to ~/.nanayam/bin..."
	mkdir -p $(HOME)/.nanayam/bin
	cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/.nanayam/bin/
	@echo "Run 'export PATH=\"$$HOME/.nanayam/bin:$$PATH\"' or restart your terminal"

## build-all: Cross-compile for all platforms
build-all:
	@echo "Cross-compiling for all platforms..."
	mkdir -p $(BUILD_DIR)
	cd $(CLI_DIR) && GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	cd $(CLI_DIR) && GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	cd $(CLI_DIR) && GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	cd $(CLI_DIR) && GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	cd $(CLI_DIR) && GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Binaries built in $(BUILD_DIR)/"

## gateway: Build the Go gateway service
gateway:
	cd $(GATEWAY_DIR) && go build -o gateway .

## console: Build the Next.js operator console
console:
	cd $(CONSOLE_DIR) && npm run build

## test: Run all tests
test:
	cd $(CLI_DIR) && go test ./...
	cd $(GATEWAY_DIR) && go test ./...

## lint: Run linters
lint:
	cd $(CLI_DIR) && go vet ./...
	cd $(GATEWAY_DIR) && go vet ./...

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(GATEWAY_DIR)/gateway

## dev-install: Symlink dev binary for local testing
dev-install:
	mkdir -p $(HOME)/.nanayam/bin
	cd $(CLI_DIR) && go build -o $(HOME)/.nanayam/bin/$(BINARY_NAME) .
	@echo "Dev binary installed to $(HOME)/.nanayam/bin/$(BINARY_NAME)"

## fmt: Format Go code
fmt:
	cd $(CLI_DIR) && gofmt -w .
	cd $(GATEWAY_DIR) && gofmt -w .

## help: Show this help
help:
	@echo "Nanayam Build System"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
