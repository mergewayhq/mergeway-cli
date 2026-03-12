# syntax=docker/dockerfile:1.7

FROM golang:1.24-alpine AS builder

WORKDIR /src

ARG VERSION=0.3.0-dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build \
      -trimpath \
      -ldflags="-s -w -X github.com/mergewayhq/mergeway-cli/internal/version.Number=${VERSION} -X github.com/mergewayhq/mergeway-cli/internal/version.Commit=${COMMIT} -X github.com/mergewayhq/mergeway-cli/internal/version.BuildDate=${BUILD_DATE}" \
      -o /out/mergeway-cli \
      .

FROM scratch

COPY --from=builder /out/mergeway-cli /mergeway-cli

ENTRYPOINT ["/mergeway-cli"]
