BIN_SERVER := ./bin/server
BIN_CLI := ./bin/cli

RELEASE ?= develop
GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -s -w \
    -X main.release=$(RELEASE) \
    -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) \
    -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN_SERVER) -ldflags "$(LDFLAGS)" ./cmd/server
	go build -v -o $(BIN_CLI) -ldflags "$(LDFLAGS)" ./cmd/cli

version: build
	$(BIN_SERVER) version
	$(BIN_CLI) version

lint:
	golangci-lint run --timeout 5m ./...

test:
	go test -race -count 100 ./internal/...

clean:
	rm -rf ./bin

.PHONY: build version lint test clean
