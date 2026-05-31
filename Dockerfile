FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server
RUN CGO_ENABLED=0 go build -o /bin/mcp ./cmd/mcp

FROM alpine:3.20
RUN apk add --no-cache ca-certificates curl
COPY --from=builder /bin/server /bin/server
COPY --from=builder /bin/mcp /bin/mcp
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1
CMD ["/bin/server"]
