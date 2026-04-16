GO ?= go
GORUN = GOMODCACHE=$(PWD)/.cache/go GOCACHE=$(PWD)/.cache/gobuild $(GO)

MDBOOK ?= mdbook
DOCS_DIR ?= docs

PKG := github.com/mergewayhq/mergeway-cli
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X $(PKG)/internal/version.Commit=$(GIT_COMMIT) -X $(PKG)/internal/version.BuildDate=$(BUILD_DATE)

GOFILES := $(shell find . -type f -name '*.go' -not -path './.git/*' -not -path './.cache/*')

.PHONY: check-go build fmt fmt-check lint test race coverage ci clean release docs-build docs-serve

check-go:
	@command -v $(GO) >/dev/null 2>&1 || { \
		echo "error: Go is not installed or not available on PATH (GO=$(GO))" >&2; \
		exit 1; \
	}

build: check-go
	$(GORUN) build -ldflags "$(LDFLAGS)" -o bin/mergeway-cli .

fmt: check-go
	gofmt -w $(GOFILES)

fmt-check: check-go
	@gofmt -l -d $(GOFILES)

lint: check-go
	golangci-lint run

test: check-go
	$(GORUN) test ./...

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
