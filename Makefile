# Nanayam Makefile
# Unified build system for the nanayam CLI and services

.PHONY: all build install test test-cli test-gateway test-console clean lint fmt-check validate \
        build-all release-assets local deploy-cloud server server-down

BINARY_NAME=nanayam
BUILD_DIR=./build
RELEASE_DIR=$(BUILD_DIR)/release
CLI_DIR=./cli
GATEWAY_DIR=./services/gateway
CONSOLE_DIR=./apps/org-console
PNPM?=pnpm

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

## local: Install the current checkout using the full installer flow
local:
	bash ./install.sh --dev-local --refresh --source "$(CURDIR)"

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

## release-assets: Build packaged release archives for install/upgrade flows
release-assets:
	@echo "Building packaged release assets for $(VERSION)..."
	rm -rf $(RELEASE_DIR)
	mkdir -p $(RELEASE_DIR)
	@set -e; \
	for platform in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64; do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		PACKAGE_DIR="$(RELEASE_DIR)/package-$${GOOS}-$${GOARCH}"; \
		ARCHIVE_BASENAME="$(BINARY_NAME)_$(VERSION)_$${GOOS}_$${GOARCH}"; \
		BINARY_FILE="$(BINARY_NAME)"; \
		if [ "$$GOOS" = "windows" ]; then BINARY_FILE="$(BINARY_NAME).exe"; fi; \
		rm -rf "$$PACKAGE_DIR"; \
		mkdir -p "$$PACKAGE_DIR"; \
		( cd $(CLI_DIR) && GOOS=$$GOOS GOARCH=$$GOARCH go build $(LDFLAGS) -o "../$$PACKAGE_DIR/$$BINARY_FILE" . ); \
		if [ "$$GOOS" = "windows" ]; then \
			( cd "$$PACKAGE_DIR" && zip -q "$(abspath $(RELEASE_DIR))/$$ARCHIVE_BASENAME.zip" "$$BINARY_FILE" ); \
		else \
			tar -czf "$(RELEASE_DIR)/$$ARCHIVE_BASENAME.tar.gz" -C "$$PACKAGE_DIR" "$$BINARY_FILE"; \
		fi; \
		rm -rf "$$PACKAGE_DIR"; \
		printf '%s\n' "Created $$ARCHIVE_BASENAME"; \
	done
	@echo "Release assets built in $(RELEASE_DIR)/"

## gateway: Build the Go gateway service
gateway:
	cd $(GATEWAY_DIR) && go build -o gateway .

## console: Build the Next.js org console
console:
	cd $(CONSOLE_DIR) && $(PNPM) install --frozen-lockfile && $(PNPM) build

## test: Run every test suite (CLI, gateway, console)
test: test-cli test-gateway test-console

## test-cli: Run the Go CLI unit tests
test-cli:
	cd $(CLI_DIR) && go test ./...

## test-gateway: Run the Go gateway unit tests
test-gateway:
	cd $(GATEWAY_DIR) && go test ./...

## test-console: Run the Next.js console unit tests
test-console:
	cd $(CONSOLE_DIR) && $(PNPM) install --frozen-lockfile && $(PNPM) test

## lint: Run linters
lint:
	cd $(CLI_DIR) && go vet ./...
	cd $(GATEWAY_DIR) && go vet ./...

## fmt-check: Fail if any Go file is not gofmt-formatted
fmt-check:
	@unformatted=$$(gofmt -l $(CLI_DIR) $(GATEWAY_DIR)); \
	if [ -n "$$unformatted" ]; then \
		echo "These files need gofmt:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi
	@echo "All Go files are formatted"

## validate: Run the full validation suite (format, lint, build, test)
validate: fmt-check lint build test
	@echo "Validation complete"

## deploy-cloud: Deploy Nanayam to a Kubernetes cluster
deploy-cloud:
	bash ./scripts/deploy-cloud.sh $(ARGS)

## server: Bring up the server stack (Fabric network + gateway, no clients) — Mac/Linux
server:
	bash ./scripts/start-server.sh

## server-down: Stop the server stack started by 'make server'
server-down:
	bash ./scripts/start-server.sh --down

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
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | awk -F ': ' '{ printf "  %-16s %s\n", $$1, $$2 }'
