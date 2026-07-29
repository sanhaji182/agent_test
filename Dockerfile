# --- Build stage ---
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/server ./cmd/server
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/mcp ./cmd/mcp

# Install Playwright driver in builder (cached layer)
RUN go run github.com/playwright-community/playwright-go/cmd/playwright@v0.5700.1 install --with-deps chromium

# --- Runtime stage (minimal, non-root) ---
FROM debian:bookworm-slim

LABEL org.opencontainers.image.title="GoTest Agent"
LABEL org.opencontainers.image.description="AI-powered testing platform"

# Install only runtime system dependencies for Chromium
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl \
    libnss3 libnspr4 libatk1.0-0 libatk-bridge2.0-0 libcups2 \
    libdrm2 libdbus-1-3 libxkbcommon0 libatspi2.0-0 libx11-6 \
    libxcomposite1 libxdamage1 libxext6 libxfixes3 libxrandr2 \
    libgbm1 libpango-1.0-0 libcairo2 libasound2 libxcb1 \
    fonts-liberation fonts-noto-color-emoji \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -r gotest && useradd -r -g gotest -d /app -s /sbin/nologin gotest

WORKDIR /app

# Copy binaries
COPY --from=builder /bin/server /usr/local/bin/server
COPY --from=builder /bin/mcp /usr/local/bin/mcp

# Copy Playwright driver + browsers from builder
COPY --from=builder /root/.cache/ms-playwright-go /home/gotest/.cache/ms-playwright-go
COPY --from=builder /root/.cache/ms-playwright /home/gotest/.cache/ms-playwright

# Create data directories
RUN mkdir -p /app/data/screenshots /app/data/reports /app/data/videos \
    && chown -R gotest:gotest /app /home/gotest/.cache

USER gotest

ENV SCREENSHOTS_PATH=/app/data/screenshots
ENV REPORTS_PATH=/app/data/reports

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -sf http://localhost:8080/health || exit 1

CMD ["server"]
