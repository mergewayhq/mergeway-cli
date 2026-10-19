GO ?= go
GORUN = GOMODCACHE=$(PWD)/.cache/go GOCACHE=$(PWD)/.cache/gobuild $(GO)

MDBOOK ?= mdbook
DOCS_DIR ?= docs
HUGO ?= hugo
LINKCHECK ?= lychee
WEBSITE_DIR ?= website
HUGO_CACHEDIR ?= $(PWD)/.cache/hugo

PKG := github.com/mergewayhq/mergeway-cli
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X $(PKG)/internal/version.Commit=$(GIT_COMMIT) -X $(PKG)/internal/version.BuildDate=$(BUILD_DATE)

GOFILES := $(shell find . -type f -name '*.go' -not -path './.git/*' -not -path './.cache/*')

.PHONY: check-go build build-cli build-diff build-lsp build-mcp fmt fmt-check lint test race coverage ci clean release docs-build docs-serve website-build website-check website-check-routes website-check-links website-check-tutorial website-check-specification website-serve website-clean

check-go:
	@command -v $(GO) >/dev/null 2>&1 || { \
		echo "error: Go is not installed or not available on PATH (GO=$(GO))" >&2; \
		exit 1; \
	}

build: build-cli build-diff build-lsp build-mcp

build-cli: check-go
	mkdir -p bin
	$(GORUN) build -ldflags "$(LDFLAGS)" -o bin/mergeway-cli ./cmd/mergeway-cli

build-diff: check-go
	mkdir -p bin
	$(GORUN) build -ldflags "$(LDFLAGS)" -o bin/mergeway-diff ./cmd/mergeway-diff

build-lsp: check-go
	mkdir -p bin
	$(GORUN) build -o bin/mergeway-lsp ./cmd/mergeway-lsp

build-mcp: check-go
	mkdir -p bin
	$(GORUN) build -ldflags "$(LDFLAGS)" -o bin/mergeway-mcp ./cmd/mergeway-mcp

fmt: check-go
	gofmt -w $(GOFILES)

fmt-check: check-go
	@gofmt -l -d $(GOFILES)

lint: check-go
	golangci-lint run

test: check-go
	$(GORUN) test ./...

e2e: build-cli
	./e2e_test/mw_e2e_test.sh

race: check-go
	$(GORUN) test -race ./...

coverage: check-go
	$(GORUN) test -coverprofile=coverage.out ./...
	$(GORUN) tool cover -html=coverage.out -o coverage.html

ci: check-go fmt-check lint test race coverage

clean:
	rm -rf bin dist coverage.out .cache

docs-build:
	$(MDBOOK) build $(DOCS_DIR)

docs-serve:
	$(MDBOOK) serve $(DOCS_DIR)

website-build:
	cd $(WEBSITE_DIR) && HUGO_CACHEDIR=$(HUGO_CACHEDIR) $(HUGO) --environment production --baseURL https://mergewayhq.github.io/ --buildDrafts=false --buildFuture=false --cleanDestinationDir --gc --minify --panicOnWarning

website-check: website-build website-check-routes website-check-tutorial website-check-specification website-check-links

website-check-routes:
	sh $(WEBSITE_DIR)/scripts/check-routes.sh $(WEBSITE_DIR)/public

website-check-links:
	HUGO=$(HUGO) HUGO_CACHEDIR=$(HUGO_CACHEDIR) LINKCHECK=$(LINKCHECK) sh $(WEBSITE_DIR)/scripts/check-links.sh $(WEBSITE_DIR)/public

website-check-tutorial:
	GOMODCACHE=$(PWD)/.cache/go GOCACHE=$(PWD)/.cache/gobuild sh $(WEBSITE_DIR)/scripts/check-first-repository.sh

website-check-specification:
	GOMODCACHE=$(PWD)/.cache/go GOCACHE=$(PWD)/.cache/gobuild sh $(WEBSITE_DIR)/scripts/check-specification.sh

website-serve:
	cd $(WEBSITE_DIR) && HUGO_CACHEDIR=$(HUGO_CACHEDIR) $(HUGO) server --buildDrafts=false --buildFuture=false --disableFastRender

website-clean:
	cd $(WEBSITE_DIR) && rm -rf public resources/_gen .hugo_build.lock
