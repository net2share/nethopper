.PHONY: build build-all clean test lint fmt deps build-singbox

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X github.com/nethopper/nethopper/internal/cli.Version=$(VERSION) \
	-X github.com/nethopper/nethopper/internal/cli.BuildTime=$(BUILD_TIME) \
	-X github.com/nethopper/nethopper/internal/cli.GitCommit=$(GIT_COMMIT) \
	-s -w"

# Sing-box source path (for local builds)
SINGBOX_SOURCE ?= ../sing-box-new

# Output binary name
BINARY := nethopper

# Build for current platform (requires embedded sing-box binaries)
build: deps
	go build $(LDFLAGS) -o $(BINARY) ./cmd/nethopper

# Build for all platforms
build-all: deps
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 ./cmd/nethopper
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64 ./cmd/nethopper
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 ./cmd/nethopper
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 ./cmd/nethopper
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe ./cmd/nethopper

# Build sing-box binaries for all platforms (for embedding)
# This cross-compiles sing-box from source and places binaries in internal/binary/embedded/
build-singbox:
	@echo "Building sing-box from source: $(SINGBOX_SOURCE)"
	@./scripts/build-singbox.sh "$(SINGBOX_SOURCE)"

# Build sing-box for current platform only (faster for testing)
build-singbox-local:
	@echo "Building sing-box for current platform from: $(SINGBOX_SOURCE)"
	@mkdir -p internal/binary/embedded
	@cd "$(SINGBOX_SOURCE)" && \
		export GOTOOLCHAIN=local && \
		CGO_ENABLED=0 go build -trimpath \
		-ldflags "-s -w -buildid= -checklinkname=0" \
		-tags "with_gvisor,with_quic,with_dhcp,with_wireguard,with_utls,with_acme,with_clash_api,with_tailscale,with_ccm,with_ocm,badlinkname,tfogo_checklinkname0" \
		-o "$(CURDIR)/internal/binary/embedded/sing-box-$$(go env GOOS)-$$(go env GOARCH)$$(if [ "$$(go env GOOS)" = "windows" ]; then echo ".exe"; fi)" \
		./cmd/sing-box
	@echo "Built: internal/binary/embedded/sing-box-$$(go env GOOS)-$$(go env GOARCH)"

# Full local build: build sing-box + build nethopper
local: build-singbox build
	@echo "Full local build complete!"

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -rf dist/
	rm -f internal/binary/embedded/sing-box-*

# Development helpers
run: build
	./$(BINARY)

run-server: build
	./$(BINARY) server

run-freenet: build
	./$(BINARY) freenet

# Show help
help:
	@echo "nethopper Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build              Build nethopper for current platform"
	@echo "  build-all          Build nethopper for all supported platforms"
	@echo "  build-singbox      Build sing-box for ALL platforms (cross-compile)"
	@echo "  build-singbox-local Build sing-box for current platform only"
	@echo "  local              Full local build (sing-box + nethopper)"
	@echo "  deps               Download dependencies"
	@echo "  test               Run tests"
	@echo "  lint               Run linter"
	@echo "  fmt                Format code"
	@echo "  clean              Clean build artifacts"
	@echo "  run                Build and run"
	@echo ""
	@echo "Environment variables:"
	@echo "  SINGBOX_SOURCE  Path to sing-box source (default: $(SINGBOX_SOURCE))"
	@echo "  VERSION         Version string (default: git describe)"
	@echo ""
	@echo "Examples:"
	@echo "  make build-singbox                    # Build sing-box for all platforms"
	@echo "  make build-singbox-local              # Build sing-box for current platform"
	@echo "  make local                            # Full build (sing-box + nethopper)"
	@echo "  SINGBOX_SOURCE=/path/to/sing-box make build-singbox  # Custom source path"
