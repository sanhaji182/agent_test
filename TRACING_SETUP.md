# OpenTelemetry Tracing Setup

## Overview

GoTest Agent includes distributed tracing using OpenTelemetry, which provides:
- Automatic HTTP request tracing
- LLM call instrumentation with retry tracking
- End-to-end visibility across all operations
- Performance monitoring and debugging

## Setup

### 1. Install Dependencies

Run this command to download OpenTelemetry packages and update go.sum:

```bash
go mod tidy
```

This will fetch:
- `go.opentelemetry.io/otel` - Core OpenTelemetry API
- `go.opentelemetry.io/otel/sdk` - SDK implementation
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` - OTLP gRPC exporter
- `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` - HTTP middleware
- And all transitive dependencies

### 2. Verify Build

```bash
go build -o agent-test ./cmd/server
```

Expected: Build completes without errors.

### 3. Run Tests

```bash
go test ./internal/tracing -v
go test ./internal/ai -v
go test ./... -short
```

Expected: All tests pass.

## Configuration

### Environment Variables

Add these to your `.env` file:

```bash
# Enable/disable tracing (default: false)
TRACING_ENABLED=true

# OTLP endpoint for trace exporter (default: localhost:4317)
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317

# Service version (default: 1.0.0)
SERVICE_VERSION=1.0.0
```

### Viewing Traces

#### Option 1: Jaeger (Recommended)

Run Jaeger locally:

```bash
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  jaegertracing/all-in-one:latest
```

Access Jaeger UI: http://localhost:16686

#### Option 2: Zipkin

Run Zipkin locally:

```bash
docker run -d --name zipkin \
  -p 9411:9411 \
  openzipkin/zipkin
```

Access Zipkin UI: http://localhost:9411

#### Option 3: Cloud Backends

- **Grafana Tempo**: Configure via OTLP endpoint
- **Datadog**: Use OTLP exporter with Datadog agent
- **Honeycomb**: Direct OTLP support
- **Lightstep**: Native OpenTelemetry support

## How It Works

### HTTP Request Tracing

Every HTTP request automatically creates a span with:
- HTTP method and path
- Request/response status
- Duration
- Error information

Example trace:
```
HTTP GET /api/v1/runs
└─ 45ms total
   ├─ Database query: 12ms
   └─ Response serialization: 8ms
```

### LLM Call Tracing

AI client calls are instrumented with:
- Retry attempts
- Backoff delays
- Circuit breaker state
- Error details

Example trace:
```
LLM Call (Claude)
└─ 2.3s total
   ├─ Attempt 1: 500ms (429 - Rate Limited)
   ├─ Backoff: 1s
   ├─ Attempt 2: 800ms (Success)
   └─ Total retries: 1
```

### End-to-End Example

Trace a complete test run:
```
POST /api/v1/runs
└─ 15.2s total
   ├─ Request parsing: 2ms
   ├─ Database insert: 15ms
   ├─ Launch goroutine: 1ms
   └─ Agent execution: 15.1s
      ├─ LLM: Analyze requirements: 2.1s
      ├─ LLM: Generate test plan: 3.5s (1 retry)
      ├─ Playwright: Execute tests: 8.2s
      │  ├─ Navigate: 1.2s
      │  ├─ Fill form: 0.8s
      │  └─ Assert: 0.5s
      └─ LLM: Generate report: 1.3s
```

## Troubleshooting

### Build Fails with "missing go.sum entry"

Run `go mod tidy` to update dependencies.

### No Traces Appearing

1. Verify `TRACING_ENABLED=true` in `.env`
2. Check `OTEL_EXPORTER_OTLP_ENDPOINT` is correct
3. Verify the trace backend (Jaeger/Zipkin) is running
4. Check application logs for errors: `grep -i "tracing" logs/app.log`

### Traces Show but Missing Details

Ensure spans are properly closed in your code:
```go
ctx, span := tracing.Tracer("my-service").Start(ctx, "operation-name")
defer span.End()  // Important!
```

## Architecture

```
┌─────────────────┐
│   HTTP Request  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Tracing         │ ← Automatic span creation
│ Middleware      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   API Handler   │
└────────┬────────┘
         │
         ├──────────────┬──────────────┐
         │              │              │
         ▼              ▼              ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│  Database   │  │ LLM Client  │  │  Playwright │
│   Query     │  │  (with      │  │   Browser   │
│             │  │  retry      │  │             │
│             │  │  tracking)  │  │             │
└─────────────┘  └─────────────┘  └─────────────┘
         │              │              │
         └──────────────┴──────────────┘
                        │
                        ▼
              ┌─────────────────┐
              │ OTLP Exporter   │
              └────────┬────────┘
                       │
                       ▼
              ┌─────────────────┐
              │  Trace Backend  │
              │ (Jaeger/Zipkin) │
              └─────────────────┘
```

## Performance Impact

- **CPU overhead**: <1% (sampling reduces this further)
- **Memory overhead**: ~5-10MB for span buffer
- **Network overhead**: Minimal (batched exports)
- **Latency impact**: <1ms per request

## Next Steps

1. Run `go mod tidy`
2. Verify build with `go build`
3. Start Jaeger: `docker run ...`
4. Run the application with `TRACING_ENABLED=true`
5. Make some API requests
6. View traces in Jaeger UI
7. Analyze performance bottlenecks
8. Optimize slow operations

## Support

For issues:
- Check application logs
- Verify OTLP endpoint connectivity
- Review OpenTelemetry documentation: https://opentelemetry.io/docs/
