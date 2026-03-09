.PHONY: build build-all clean test lint fmt deps

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

SERVER_LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -s -w"
CLIENT_LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -s -w"

build: deps
	go build $(SERVER_LDFLAGS) -o nhserver ./cmd/nhserver
	go build $(CLIENT_LDFLAGS) -o nhclient ./cmd/nhclient

build-server: deps
	go build $(SERVER_LDFLAGS) -o nhserver ./cmd/nhserver

build-client: deps
	go build $(CLIENT_LDFLAGS) -o nhclient ./cmd/nhclient

build-all: deps
	@mkdir -p dist
	# Server (Linux only)
	GOOS=linux GOARCH=amd64 go build $(SERVER_LDFLAGS) -o dist/nhserver-linux-amd64 ./cmd/nhserver
	GOOS=linux GOARCH=arm64 go build $(SERVER_LDFLAGS) -o dist/nhserver-linux-arm64 ./cmd/nhserver
	# Client (cross-platform)
	GOOS=linux GOARCH=amd64 go build $(CLIENT_LDFLAGS) -o dist/nhclient-linux-amd64 ./cmd/nhclient
	GOOS=linux GOARCH=arm64 go build $(CLIENT_LDFLAGS) -o dist/nhclient-linux-arm64 ./cmd/nhclient
	GOOS=darwin GOARCH=amd64 go build $(CLIENT_LDFLAGS) -o dist/nhclient-darwin-amd64 ./cmd/nhclient
	GOOS=darwin GOARCH=arm64 go build $(CLIENT_LDFLAGS) -o dist/nhclient-darwin-arm64 ./cmd/nhclient
	GOOS=windows GOARCH=amd64 go build $(CLIENT_LDFLAGS) -o dist/nhclient-windows-amd64.exe ./cmd/nhclient

deps:
	go mod tidy

test:
	go test -v ./...

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

clean:
	rm -f nhserver nhclient
	rm -rf dist/

run-server: build-server
	./nhserver

run-client: build-client
	./nhclient
