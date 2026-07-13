BINARY_NAME=groovy-check
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: build test install clean lint fmt release

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

test:
	go test -v ./...

install:
	go install $(LDFLAGS) .

clean:
	rm -rf bin/ dist/

lint:
	golangci-lint run

fmt:
	go fmt ./...

release:
	goreleaser release --snapshot --clean
