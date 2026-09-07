BINARY_NAME := gateway
VERSION ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all build test lint fmt run docker docker-buildx integration clean tidy

all: fmt lint test build

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) ./cmd/gateway

run: build
	./bin/$(BINARY_NAME) run --config config.example.yaml

test:
	go test ./... -race -count=1

integration:
	go test -tags=integration ./test/integration/... -count=1 -timeout=60s

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w . 2>/dev/null || true

tidy:
	go mod tidy

docker:
	docker buildx build --platform linux/arm64,linux/amd64 -t ghcr.io/tecxol/gatewaytecxol-driver:edge --load .

docker-buildx:
	docker buildx build --platform linux/arm64,linux/amd64 -t ghcr.io/tecxol/gatewaytecxol-driver:$(VERSION) --push .

clean:
	rm -rf bin/ dist/
