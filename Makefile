BINARY  := docker-volume-exporter
VERSION ?= dev

.PHONY: build test lint fmt docker
build:
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/$(BINARY) ./cmd/$(BINARY)
test:
	go vet ./... && go test -race ./...
lint:
	golangci-lint run
fmt:
	gofmt -w .
docker:
	docker build --build-arg VERSION=$(VERSION) -t ghcr.io/yansetiaji/$(BINARY):$(VERSION) .
