FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server
RUN CGO_ENABLED=0 go build -o /bin/mcp ./cmd/mcp

FROM golang:1.26-bookworm
RUN apt-get update && apt-get install -y curl ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /bin/server /bin/server
COPY --from=builder /bin/mcp /bin/mcp
# Install playwright and browsers
RUN /bin/server & sleep 1 && go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps chromium || true
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1
CMD ["/bin/server"]
