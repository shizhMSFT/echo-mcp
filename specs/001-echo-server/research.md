# Research: Echo MCP Server

**Feature**: Echo MCP Server  
**Date**: 2025-11-12  
**Purpose**: Technical research and decisions for implementing an MCP server in Go

## Research Areas

### 1. MCP (Model Context Protocol) SDK for Go

**Decision**: Use official MCP Go SDK if available; otherwise implement MCP protocol directly using JSON-RPC 2.0 specification

**Rationale**: 
- MCP is a standardized protocol (JSON-RPC 2.0 based) for communication between AI assistants and context providers
- Official SDK ensures protocol compliance and reduces implementation risk
- If no official Go SDK exists, the protocol is well-documented and straightforward to implement
- Echo server is a simple reference implementation, making it ideal for minimal dependencies

**Alternatives Considered**:
- **Build from scratch without reference**: Rejected due to protocol compliance risk
- **Use generic JSON-RPC library**: Rejected because MCP has specific message structures beyond basic JSON-RPC

**Implementation Approach**:
- Check for official MCP Go SDK at https://github.com/modelcontextprotocol
- If available, use as primary dependency
- If not available, implement MCP protocol types and handlers directly
- Follow MCP specification for message formats (initialize, tool calls, errors)

**Resources**:
- MCP Specification: https://spec.modelcontextprotocol.io/
- MCP Protocol uses JSON-RPC 2.0 as transport layer
- Key message types: initialize, initialized, tools/list, tools/call

---

### 2. Go Logging Library for Structured and Human-Readable Logs

**Decision**: Use `zerolog` for efficient structured logging with both JSON and human-readable output

**Rationale**:
- Zero-allocation JSON logger (performance critical for high throughput)
- Supports both structured (JSON) and console (human-readable) output formats
- Simple API that matches constitution requirement for clear logging
- Popular, well-maintained, excellent performance benchmarks
- Built-in support for context, fields, and log levels

**Alternatives Considered**:
- **Standard library `log/slog`** (Go 1.21+): Good choice, but `zerolog` has better performance and more ergonomic API
- **zap**: Excellent performance but more complex API; zerolog offers better balance of performance and simplicity
- **logrus**: Popular but slower; zero-allocation is important for our throughput requirements

**Implementation Approach**:
```go
// Human-readable for development (default)
logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

// Structured JSON for production (flag-based)
logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
```

**Key Features Used**:
- `logger.Info().RawJSON("request", rawBytes).Msg("Received request")` - Log raw request
- `logger.With().Str("requestID", id).Logger()` - Contextual logging per request
- Console writer for human-readable output during development

---

### 3. Transport Implementations: Stdio and HTTP/SSE

**Decision**: 
- **Local transport**: Standard library `os.Stdin`/`os.Stdout` with `bufio.Scanner`
- **Remote transport**: HTTP with Server-Sent Events (SSE) for MCP protocol streaming

**Rationale**:
- **Stdio**: Standard library provides everything needed; no external dependencies
- **HTTP/SSE**: MCP remote transport typically uses SSE for server-to-client streaming
- SSE is simpler than WebSockets for this use case (server pushes, client requests)
- Standard library `net/http` is production-ready and performant

**Alternatives Considered**:
- **WebSockets for remote**: More complex than needed; SSE is sufficient for MCP
- **gRPC**: Overkill for simple echo server; adds significant complexity
- **Raw TCP**: Lower level than needed; HTTP provides better tooling and debugging

**Implementation Approach**:

**Stdio Transport**:
```go
scanner := bufio.NewScanner(os.Stdin)
for scanner.Scan() {
    request := scanner.Bytes()
    // Process MCP request
    response := handleRequest(request)
    fmt.Fprintln(os.Stdout, response)
}
```

**HTTP/SSE Transport**:
```go
http.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    
    // Handle MCP protocol over SSE
    flusher := w.(http.Flusher)
    // ... event streaming logic
})
```

---

### 4. Dockerfile Best Practices for Go Applications

**Decision**: Multi-stage Docker build with Alpine base for minimal image size

**Rationale**:
- Multi-stage builds separate build environment from runtime environment
- Alpine Linux provides minimal attack surface (<5MB base image)
- Static binary compilation in Go allows running without dependencies
- Smaller images = faster deployment, less bandwidth, improved security

**Implementation Approach**:

```dockerfile
# Stage 1: Build
FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o echo-mcp ./cmd/echo-mcp

# Stage 2: Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /build/echo-mcp .
EXPOSE 8080
CMD ["./echo-mcp", "--mode=remote"]
```

**Key Features**:
- `CGO_ENABLED=0`: Static binary (no C dependencies)
- `alpine:latest`: Minimal runtime image
- `ca-certificates`: For HTTPS if needed later
- Non-root user for security (can be added)

**Alternatives Considered**:
- **Scratch image**: Even smaller but lacks shell for debugging; Alpine provides good balance
- **Distroless**: Google's minimal images; Alpine is more familiar and well-documented
- **Full Debian/Ubuntu**: Unnecessarily large (100MB+ vs <10MB with Alpine)

**Testing Strategy**:
```bash
# Build image
docker build -t echo-mcp:latest .

# Test local mode (stdio)
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"echo","arguments":{"message":"test"}}}' | docker run -i echo-mcp:latest --mode=local

# Test remote mode (HTTP)
docker run -p 8080:8080 echo-mcp:latest --mode=remote
curl -X POST http://localhost:8080/mcp ...
```

---

### 5. Concurrency Patterns for Handling Multiple Requests

**Decision**: Goroutine per connection with channel-based coordination and context for cancellation

**Rationale**:
- Go's goroutines are lightweight (2KB initial stack) - can handle thousands concurrently
- Channels provide safe communication between goroutines (no race conditions)
- Context package enables graceful shutdown and request timeouts
- Standard pattern in Go HTTP servers (proven, well-understood)

**Implementation Approach**:

```go
// HTTP mode: net/http handles goroutine per request automatically
http.ListenAndServe(":8080", handler)

// Stdio mode: Single goroutine (only one stdin/stdout)
// But still use channels for internal coordination

// Shared request processing with concurrency safety
type Server struct {
    mu       sync.RWMutex
    requests chan Request
    shutdown chan struct{}
}

func (s *Server) processRequests(ctx context.Context) {
    for {
        select {
        case req := <-s.requests:
            // Process in goroutine to avoid blocking
            go s.handleRequest(ctx, req)
        case <-ctx.Done():
            return
        }
    }
}
```

**Concurrency Safety Measures**:
- No shared mutable state (each request processed independently)
- Logger is concurrency-safe (zerolog design)
- Use `sync.WaitGroup` for graceful shutdown
- Context cancellation propagates to all goroutines

**Alternatives Considered**:
- **Worker pool pattern**: Unnecessary complexity for stateless echo server
- **Single-threaded with async/await**: Go doesn't have async/await; goroutines are idiomatic
- **Thread pool**: OS threads are heavier than goroutines; not needed

---

### 6. Testing Strategy: Contract, Integration, Unit Tests

**Decision**: Three-tier testing strategy with emphasis on contract tests for MCP compliance

**Test Categories**:

**Contract Tests** (MCP Protocol Compliance):
```go
// tests/contract/mcp_protocol_test.go
func TestMCP_InitializeRequest(t *testing.T) {
    // Verify server responds to initialize with correct capabilities
}

func TestMCP_EchoToolSchema(t *testing.T) {
    // Verify echo tool advertised in tools/list matches spec
}

func TestMCP_EchoToolExecution(t *testing.T) {
    // Verify echo tool accepts message and returns it unchanged
}
```

**Integration Tests** (End-to-End Flows):
```go
// tests/integration/stdio_transport_test.go
func TestStdioTransport_FullSession(t *testing.T) {
    // Start server, send initialize, call echo, verify response, shutdown
}

// tests/integration/http_transport_test.go
func TestHTTPTransport_ConcurrentClients(t *testing.T) {
    // Multiple clients calling echo simultaneously
}
```

**Unit Tests** (Component Logic):
```go
// tests/unit/echo_handler_test.go
func TestEchoHandler_PreservesMessage(t *testing.T) {
    // Table-driven tests for various message types
}

// tests/unit/logger_test.go
func TestLogger_FormatsRawRequest(t *testing.T) {
    // Verify log output contains raw request
}
```

**Rationale**:
- Contract tests ensure MCP spec compliance (critical for interoperability)
- Integration tests verify transport layers work end-to-end
- Unit tests provide fast feedback during development
- Table-driven tests common in Go ecosystem

**Testing Tools**:
- Standard `testing` package
- `testify/assert` for cleaner assertions
- `httptest` for HTTP transport testing
- Custom test helpers for MCP message creation

---

### 7. Performance Testing and Benchmarking

**Decision**: Go built-in benchmarks + load testing with documented test procedure

**Benchmark Tests**:
```go
// tests/unit/echo_handler_test.go
func BenchmarkEchoHandler(b *testing.B) {
    handler := NewEchoHandler()
    message := []byte("test message")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        handler.Echo(message)
    }
}

func BenchmarkJSONMarshaling(b *testing.B) {
    // Benchmark MCP protocol message encoding/decoding
}
```

**Load Testing**:
```bash
# Using Apache Bench for HTTP transport
ab -n 10000 -c 100 http://localhost:8080/mcp

# Using custom script for stdio transport
for i in {1..1000}; do
    echo '{"jsonrpc":"2.0","id":'$i',"method":"tools/call",...}' | ./echo-mcp
done
```

**Performance Monitoring**:
```go
// Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

// CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

// Race detection
go test -race ./...
```

**Success Criteria** (from spec):
- <100ms p95 latency: Verified with benchmark percentile analysis
- 1000 req/s throughput: Verified with load testing
- <100MB memory: Verified with memory profiling under load
- No race conditions: Verified with `go test -race`

---

## Summary of Technical Decisions

| Decision Area | Choice | Key Rationale |
|---------------|--------|---------------|
| MCP SDK | Official SDK or direct implementation | Protocol compliance, minimal dependencies |
| Logging | zerolog | Zero-allocation, JSON + console output, performance |
| Stdio Transport | Standard library bufio | No dependencies needed, simple and reliable |
| HTTP Transport | net/http with SSE | Standard library, proven performance, MCP compatible |
| Containerization | Multi-stage Alpine Docker | Minimal image size, security, fast deployment |
| Concurrency | Goroutines + channels + context | Go idiomatic, safe, scalable |
| Testing | Contract + Integration + Unit | MCP compliance focus, comprehensive coverage |
| Performance Testing | Benchmarks + load tests + profiling | Quantitative validation of requirements |

All decisions align with constitution principles:
- ✅ **Code Quality**: Standard Go tooling (gofmt, go vet, golangci-lint)
- ✅ **Test-First**: Comprehensive test strategy defined before implementation
- ✅ **UX Consistency**: MCP spec compliance ensures consistent behavior
- ✅ **Performance**: Architectural choices support latency and throughput targets
