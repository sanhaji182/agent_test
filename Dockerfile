# --- Build stage ---
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /bin/server ./cmd/server
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /bin/mcp ./cmd/mcp

# Install Playwright driver in builder (cached layer)
# NOTE: Disabled for ARM64 compatibility. Using Steel Browser for headless browsing.
# RUN go run github.com/playwright-community/playwright-go/cmd/playwright@v0.5700.1 install --with-deps chromium

# --- Runtime stage (minimal, non-root) ---
FROM alpine:3.21

LABEL org.opencontainers.image.title="GoTest Agent"
LABEL org.opencontainers.image.description="AI-powered testing platform"

# Install only runtime system dependencies
RUN apk add --no-cache ca-certificates curl libc6-compat

# Create non-root user
RUN addgroup -S gotest && adduser -S -G gotest -h /app -s /sbin/nologin gotest

WORKDIR /app

# Copy binaries
COPY --from=builder /bin/server /usr/local/bin/server
COPY --from=builder /bin/mcp /usr/local/bin/mcp

# Copy Playwright driver + browsers from builder
# NOTE: Disabled for ARM64 compatibility.
# COPY --from=builder /root/.cache/ms-playwright-go /home/gotest/.cache/ms-playwright-go
# COPY --from=builder /root/.cache/ms-playwright /home/gotest/.cache/ms-playwright

# Create data directories
RUN mkdir -p /app/data/screenshots /app/data/reports /app/data/videos \
    && chown -R gotest:gotest /app

USER gotest

ENV SCREENSHOTS_PATH=/app/data/screenshots
ENV REPORTS_PATH=/app/data/reports

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -sf http://localhost:8080/health || exit 1

CMD ["server"]
