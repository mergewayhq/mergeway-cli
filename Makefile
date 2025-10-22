GO ?= go
GORUN = GOMODCACHE=$(PWD)/.cache/go GOCACHE=$(PWD)/.cache/gobuild $(GO)

MDBOOK ?= mdbook
DOCS_DIR ?= docs

PKG := github.com/mergewayhq/mergeway-cli
VERSION_FILE := internal/version/version.txt
VERSION := $(shell cat $(VERSION_FILE))
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X $(PKG)/internal/version.Number=$(VERSION) -X $(PKG)/internal/version.Commit=$(GIT_COMMIT) -X $(PKG)/internal/version.BuildDate=$(BUILD_DATE)

GOFILES := $(shell find . -type f -name '*.go' -not -path './.git/*' -not -path './.cache/*')

.PHONY: build fmt fmt-check lint test race coverage ci clean release docs-build docs-serve

build:
	$(GORUN) build -ldflags "$(LDFLAGS)" -o bin/mw .

fmt:
	gofmt -w $(GOFILES)

fmt-check:
	@./scripts/check_gofmt.sh

lint:
	golangci-lint run

test:
	$(GORUN) test ./...

race:
	$(GORUN) test -race ./...

coverage:
	$(GORUN) test -coverprofile=coverage.out ./...

ci: fmt-check lint test race coverage

clean:
	rm -rf bin dist coverage.out .cache

release:
	./scripts/build_binaries.sh

docs-build:
	$(MDBOOK) build $(DOCS_DIR)

docs-serve:
	$(MDBOOK) serve $(DOCS_DIR)
