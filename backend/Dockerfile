# Build stage
FROM golang:1.24-alpine AS builder

# Declare build args for multi-platform support
ARG TARGETARCH
ARG TARGETOS=linux

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.mod
COPY go.sum go.sum

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/

# Build with cache mounts for modules and build cache
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o server cmd/server/main.go

# Final stage - use minimal alpine
FROM alpine:latest

WORKDIR /

# Copy ca-certificates from builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /workspace/server .

# Copy schema.sql file for database initialization
COPY --from=builder /workspace/internal/store/schema.sql /internal/store/schema.sql

EXPOSE 8080

ENTRYPOINT ["/server"]

