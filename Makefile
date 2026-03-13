BIN_SERVER := ./bin/server
BIN_CLI := ./bin/cli

RELEASE ?= develop
GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -s -w \
    -X main.release=$(RELEASE) \
    -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) \
    -X main.gitHash=$(GIT_HASH)

export LDFLAGS

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

run: build
	$(BIN_SERVER)

cli: build
	$(BIN_CLI)

# ==============================
# Release
# ==============================
up:
	@echo "****** Release $(RELEASE) is building *****"
	@COMPOSE_BAKE=1 docker-compose up -d --build
	@echo "****** Release $(RELEASE) has started *****"

down:
	@docker-compose down
	@docker image prune -f

# ==============================
# Integration Tests
# ==============================
integration-tests:
	@bash -c '\
		set -e; \
	  	echo "************ Integration tests ************"; \
		COMPOSE_BAKE=1 docker-compose -f docker-compose.yaml -f docker-compose.integration.override.yaml up -d --build; \
		docker wait integration_tests > /dev/null; \
		docker-compose -f docker-compose.yaml -f docker-compose.integration.override.yaml logs integration-tests; \
		CODE=$$(docker inspect -f "{{.State.ExitCode}}" integration_tests); \
		echo "************ Tests finish with exit code = $$CODE ************"; \
		docker-compose -f docker-compose.yaml -f docker-compose.integration.override.yaml down -v; \
		docker image prune -f; \
		exit $$CODE \
	'

.PHONY: build version lint test clean run up down integration-tests
