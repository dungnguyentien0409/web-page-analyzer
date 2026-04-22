# Web Page Analyzer

A web application that analyzes HTML pages and extracts useful information such as HTML version, title, headings distribution, and link accessibility.

## Build and Run

### Prerequisites

- Go 1.25+
- Docker (optional)

### Local Development

```bash
# Build and run
make run

# Run tests
make test

# Run with coverage
make coverage
```

### Docker

```bash
# Build and run container
make docker-run

# Or with Prometheus + Grafana
make docker-compose-up
```

Services available at:
- **Web Analyzer**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

## Assumptions and Design Decisions

### Architecture

```
Request → [Inbound Rate Limit] → Handler → Fetcher → Analyzer → Response
                                                     ↓
                                              Link Checker
                                                     ↓
                                            [Outbound Rate Limit]
```

Clean separation of concerns with dedicated packages: `handler`, `fetcher`, `analyzer`, `ratelimit`, `metrics`.

### Worker Pool for Link Checking

Worker pool chosen over semaphore-based concurrency:
- Fixed goroutines with bounded memory usage
- Natural integration with rate limiting (workers block instead of spawning blocked goroutines)
- Clean shutdown via channel closure

### Two-Layer Rate Limiting

**Inbound** (per-IP): Protects our server from abuse
**Outbound** (global + per-host): Protects external servers, prevents being blocked

### Concurrent Analysis

HTML parsing and link extraction run in parallel via goroutines. Link checking uses configurable worker pool with retry logic.

### Observability

Prometheus metrics track HTTP requests, outbound requests, link checks, and rate limit rejections. Request tracing via `X-Request-ID` header.

### Reliability

- Graceful shutdown on SIGTERM/SIGINT
- Context timeouts at every layer
- Embedded templates for single-binary deployment

### Testing

Unit tests (96%+ coverage), integration tests against real HTTP server, and benchmarks for performance-critical paths.

## Future Improvements

### Performance

- **Link accessibility caching**: Cache HEAD request results with TTL to avoid re-checking same URLs across requests

### Features

- **REST API endpoint**: JSON response format for programmatic access
- **Batch analysis**: Analyze multiple URLs in one request

### Production Readiness

- **Circuit breaker**: Stop checking links if failure rate exceeds threshold
- **Input validation**: Stricter URL validation (protocol, format, reachable domain)
- **CI/CD pipeline**: GitHub Actions for automated testing and deployment