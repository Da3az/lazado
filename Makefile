BINARY_NAME := lazado
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR := build
GO ?= go
GOFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

# Legacy bash install path
LEGACY_PREFIX ?= $(HOME)/.lazado

.PHONY: build install clean test lint run help legacy-install legacy-uninstall

## Go TUI targets

build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/lazado

install: build
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/.local/bin/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to ~/.local/bin/"
	@echo "Run 'lazado init' to set up authentication."

run:
	$(GO) run $(GOFLAGS) ./cmd/lazado

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...

clean:
	@rm -rf $(BUILD_DIR)

## Legacy bash targets (for backward compatibility)

legacy-install:
	@mkdir -p $(LEGACY_PREFIX)/lib $(LEGACY_PREFIX)/config
	@cp lazado.sh $(LEGACY_PREFIX)/
	@cp lib/*.sh $(LEGACY_PREFIX)/lib/
	@cp config/*.template $(LEGACY_PREFIX)/config/
	@echo "Legacy bash version installed to $(LEGACY_PREFIX)"

legacy-uninstall:
	@rm -rf $(LEGACY_PREFIX)
	@echo "Removed $(LEGACY_PREFIX)"

help:
	@echo "lazado Makefile targets:"
	@echo "  make build           Build the Go TUI binary"
	@echo "  make install         Build and install to ~/.local/bin/"
	@echo "  make run             Run directly with go run"
	@echo "  make test            Run tests"
	@echo "  make lint            Run go vet"
	@echo "  make clean           Remove build artifacts"
	@echo "  make legacy-install  Install the original bash version"
