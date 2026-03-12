GO = $(shell which go 2>/dev/null)

APP             := learn-neo4j
ASK_CYPHER_APP  := ask-cypher
VERSION         ?= v0.1.0
LDFLAGS         := -ldflags "-X main.AppVersion=$(VERSION)"

.PHONY: all build build-ask clean run run-ask test

all: clean build

clean:
	$(GO) clean -testcache
	$(RM) -rf bin/*
build:
	$(GO) build -o bin/$(APP) $(LDFLAGS) cmd/$(APP)/*.go
build-ask:
	$(GO) build -o bin/$(ASK_CYPHER_APP) cmd/$(ASK_CYPHER_APP)/*.go
run:
	$(GO) run $(LDFLAGS) cmd/$(APP)/*.go
run-ask:
	$(GO) run cmd/$(ASK_CYPHER_APP)/*.go $(ARGS)
seed:
	$(GO) run cmd/seed-data/main.go
test:
	$(GO) test -v ./...
