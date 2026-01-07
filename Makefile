GO ?= go
GORUN = GOMODCACHE=$(PWD)/.cache/go GOCACHE=$(PWD)/.cache/gobuild $(GO)

MDBOOK ?= mdbook
DOCS_DIR ?= docs

PKG := github.com/mergewayhq/mergeway-cli
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X $(PKG)/internal/version.Commit=$(GIT_COMMIT) -X $(PKG)/internal/version.BuildDate=$(BUILD_DATE)

GOFILES := $(shell find . -type f -name '*.go' -not -path './.git/*' -not -path './.cache/*')

.PHONY: build fmt fmt-check lint test race coverage ci clean release docs-build docs-serve

build:
	$(GORUN) build -ldflags "$(LDFLAGS)" -o bin/mw .

fmt:
	gofmt -w $(GOFILES)

fmt-check:
	@gofmt -l -d $(GOFILES)

lint:
	golangci-lint run

test:
	$(GORUN) test ./...

race:
	$(GORUN) test -race ./...

coverage:
	$(GORUN) test -coverprofile=coverage.out ./...
	$(GORUN) tool cover -html=coverage.out -o coverage.html

ci: fmt-check lint test race coverage

clean:
	rm -rf bin dist coverage.out .cache

release:
	./scripts/build_binaries.sh

docs-build:
	$(MDBOOK) build $(DOCS_DIR)

docs-serve:
	$(MDBOOK) serve $(DOCS_DIR)
