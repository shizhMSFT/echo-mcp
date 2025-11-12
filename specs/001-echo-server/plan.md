# Implementation Plan: Echo MCP Server

**Branch**: `001-echo-server` | **Date**: 2025-11-12 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-echo-server/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build an MCP (Model Context Protocol) server in Go that provides an echo tool for testing MCP client implementations. The server accepts messages via the echo tool and returns them unchanged, while logging all requests (including MCP protocol messages like initialize) to stdout. The server supports both local transport (stdio) for development and remote transport (network) for distributed testing scenarios. Performance targets include <100ms p95 latency, 1000+ req/s throughput, and <100MB memory usage.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: 
- MCP Go SDK (official Model Context Protocol implementation for Go)
- Standard library (net/http for remote transport, encoding/json for protocol)
- Structured logging library (zerolog or similar for efficient JSON/text logging)

**Storage**: N/A (stateless server, no persistent storage required)  
**Testing**: Go testing framework (`go test`), table-driven tests, benchmark tests (`go test -bench`)  
**Target Platform**: Cross-platform (Linux, macOS, Windows), containerized deployment via Docker  
**Project Type**: Single project (server binary)  
**Performance Goals**: 
- 1,000 requests/second minimum throughput
- <100ms p95 latency for messages <1KB
- <100MB memory under normal load (100 req/s)

**Constraints**: 
- <100ms p95 latency for echo responses
- <100MB memory footprint
- 99.9% uptime (no crashes on malformed input)
- Zero race conditions (must pass `go test -race`)

**Scale/Scope**: 
- 100+ concurrent client connections (remote mode)
- 24+ hour continuous operation
- Messages up to 10MB+
- Suitable for use as testing infrastructure

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Code Quality Standards ✅

- **Static Analysis**: Will use `go vet` and `golangci-lint` in CI/CD pipeline
- **Formatting**: All Go code will be formatted with `gofmt` (enforced by pre-commit hooks)
- **Code Review**: GitHub PR review process with required approvals
- **Documentation**: All exported functions, types will have godoc comments
- **Complexity**: Will monitor with `gocyclo`; functions >15 complexity require justification
- **Dependencies**: Minimal dependencies (MCP SDK + logging); all will be security-scanned

**Status**: PASS - Standard Go tooling meets all requirements

### II. Test-First Development ✅

- **Red-Green-Refactor**: Tests written before implementation for all user stories
- **Coverage**: Target 80% minimum, 95%+ for critical paths (echo handler, transport layers)
- **Test Categories**:
  - **Unit tests**: Echo logic, logging formatters, transport handlers
  - **Contract tests**: MCP protocol compliance (initialize, tool calls, error responses)
  - **Integration tests**: End-to-end stdio and network transport flows
- **Test Independence**: Each test uses isolated server instances, no shared state
- **No Tests = No Merge**: Enforced via CI/CD checks

**Status**: PASS - Comprehensive test strategy defined

### III. User Experience Consistency ✅

- **MCP Protocol Compliance**: Will use official MCP SDK to ensure spec compliance
- **Error Messages**: Structured error responses with error codes, context, actionable messages
- **Response Format**: Consistent JSON-RPC 2.0 format per MCP specification
- **Echo Behavior**: Exact message preservation (no transformations) with explicit documentation
- **Logging**: Human-readable text format by default, structured JSON available via flag
- **Configuration**: Environment variables (ECHO_MCP_PORT, ECHO_MCP_LOG_FORMAT, etc.) with sensible defaults

**Status**: PASS - UX requirements mapped to design decisions

### IV. Performance Requirements ✅

- **Latency**: Architecture supports <50ms target (in-memory processing, zero I/O except logging)
- **Throughput**: Go's goroutine concurrency model supports 1000+ req/s easily
- **Memory**: Stateless design, bounded message buffers, no caching = minimal memory footprint
- **Concurrency**: Go's channel-based communication prevents race conditions
- **Graceful Degradation**: Connection limits, timeout handling, error recovery
- **Resource Cleanup**: Defer statements for all resources, context cancellation for goroutines

**Performance Testing Planned**:
- Benchmark tests for echo handler (`go test -bench`)
- Load testing with Apache Bench or similar (1000+ concurrent requests)
- Memory profiling (`go test -memprofile`) before release

**Status**: PASS - Go's runtime characteristics align with performance targets

### Constitution Compliance Summary

✅ **ALL GATES PASSED** - Ready to proceed to Phase 0 research

No complexity violations to justify. Project follows standard Go server architecture with minimal dependencies.

## Project Structure

### Documentation (this feature)

```text
specs/001-echo-server/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── mcp-protocol.md  # MCP protocol contract definitions
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
echo-mcp/
├── cmd/
│   └── echo-mcp/
│       └── main.go           # Server entry point, CLI flags, bootstrap
├── internal/
│   ├── server/
│   │   ├── server.go         # Core MCP server implementation
│   │   ├── echo_tool.go      # Echo tool handler
│   │   └── logger.go         # Request logging logic
│   ├── transport/
│   │   ├── stdio.go          # Local stdio transport implementation
│   │   └── http.go           # Remote HTTP/SSE transport implementation
│   └── protocol/
│       ├── types.go          # MCP protocol types and models
│       └── handlers.go       # Protocol message handlers (initialize, etc.)
├── pkg/
│   └── # Empty for now - public APIs if needed later
├── tests/
│   ├── contract/
│   │   ├── mcp_protocol_test.go      # MCP spec compliance tests
│   │   └── echo_tool_test.go         # Echo tool contract tests
│   ├── integration/
│   │   ├── stdio_transport_test.go   # End-to-end stdio tests
│   │   └── http_transport_test.go    # End-to-end HTTP tests
│   └── unit/
│       ├── echo_handler_test.go      # Echo logic unit tests
│       └── logger_test.go            # Logging unit tests
├── Dockerfile                  # Multi-stage Docker build
├── .dockerignore
├── go.mod                      # Go module dependencies
├── go.sum                      # Dependency checksums
├── .golangci.yml              # Linter configuration
├── Makefile                    # Build, test, lint targets
└── README.md                   # Usage, setup, examples
```

**Structure Decision**: Single project structure selected. This is a standalone MCP server binary with no frontend/mobile components. The `internal/` directory enforces encapsulation (cannot be imported by external projects), while `cmd/echo-mcp/` provides the executable entry point. The `tests/` directory is organized by test type (contract/integration/unit) for clear test categorization per constitution requirements.

## Complexity Tracking

> No constitution violations to justify - standard Go server architecture with minimal complexity.

## Phase 0: Research Summary

**Status**: ✅ Complete

**Artifacts**: [research.md](./research.md)

**Key Decisions**:

1. **MCP SDK**: Use official MCP Go SDK if available; otherwise implement protocol directly using JSON-RPC 2.0 spec
2. **Logging**: `zerolog` for zero-allocation, dual-format (JSON/text) logging
3. **Stdio Transport**: Standard library `bufio.Scanner` for line-delimited JSON
4. **HTTP Transport**: Standard library `net/http` with Server-Sent Events (SSE)
5. **Docker**: Multi-stage Alpine build for minimal image size (<10MB)
6. **Concurrency**: Goroutines per connection with channel-based coordination
7. **Testing**: Three-tier strategy (contract/integration/unit) with emphasis on MCP compliance

All NEEDS CLARIFICATION items from Technical Context resolved through research.

## Phase 1: Design Summary

**Status**: ✅ Complete

**Artifacts**: 
- [data-model.md](./data-model.md) - MCP protocol data structures
- [contracts/mcp-protocol.md](./contracts/mcp-protocol.md) - API contracts and test cases
- [quickstart.md](./quickstart.md) - Developer quick-start guide

**Key Deliverables**:

1. **Data Model**: Defined all MCP protocol types (JSONRPCRequest, InitializeParams, ToolCallParams, etc.)
2. **API Contracts**: 
   - Initialize handshake flow
   - tools/list endpoint
   - tools/call (echo) endpoint with comprehensive test cases
   - Error handling contracts
   - Logging contracts
3. **Quickstart Guide**: Complete setup, usage examples, troubleshooting

**Design Validation**:
- ✅ All entities from spec.md mapped to Go structs
- ✅ All functional requirements have corresponding contracts
- ✅ Performance contracts defined (latency, throughput, memory)
- ✅ Constitution re-check: All principles satisfied by design

## Phase 2: Implementation Readiness

**Next Command**: `/speckit.tasks` to generate task breakdown

**Implementation Prerequisites**:
- ✅ Technical decisions documented
- ✅ Data model defined
- ✅ API contracts specified
- ✅ Test strategy established
- ✅ Performance targets quantified
- ✅ Docker strategy defined

**Expected Implementation Flow** (per constitution TDD requirements):

1. **Setup Phase**: Initialize Go module, configure tooling
2. **Foundational Phase**: MCP protocol types, logging infrastructure
3. **User Story 1 (P1)**: Basic echo functionality
4. **User Story 2 (P2)**: Request logging with raw request capture
5. **User Story 3 (P1)**: Local stdio transport
6. **User Story 4 (P2)**: Remote HTTP transport
7. **Dockerfile**: Multi-stage build with testing
8. **Documentation**: README, examples, CI/CD setup

Each user story follows Red-Green-Refactor:
- Write failing tests (contract, integration, unit)
- Implement minimum code to pass tests
- Refactor for quality while maintaining green tests
- Run quality gates (go vet, golangci-lint, go test -race)

## Dockerfile Implementation Plan

**File**: `Dockerfile` (repository root)

**Multi-Stage Build Strategy**:

```dockerfile
# Stage 1: Build
FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o echo-mcp ./cmd/echo-mcp

# Stage 2: Test
FROM builder AS tester
RUN go test -v -race -coverprofile=coverage.out ./...
RUN go tool cover -html=coverage.out -o coverage.html

# Stage 3: Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /build/echo-mcp .
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD echo '{"jsonrpc":"2.0","id":1,"method":"ping"}' | ./echo-mcp --mode=local || exit 1
CMD ["./echo-mcp", "--mode=remote"]
```

**Testing Strategy**:

```bash
# Build all stages
docker build -t echo-mcp:latest .

# Test local mode (stdio)
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Docker test"}}}' | \
  docker run -i echo-mcp:latest --mode=local

# Test remote mode (HTTP)
docker run -d -p 8080:8080 --name echo-test echo-mcp:latest --mode=remote
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"HTTP test"}}}'
docker stop echo-test && docker rm echo-test

# Performance test in container
docker run -d -p 8080:8080 --name echo-perf echo-mcp:latest --mode=remote
ab -n 10000 -c 100 -p request.json -T application/json http://localhost:8080/mcp
docker stats echo-perf  # Verify <100MB memory
docker stop echo-perf && docker rm echo-perf

# Multi-architecture build (optional)
docker buildx build --platform linux/amd64,linux/arm64 -t echo-mcp:latest .
```

**.dockerignore**:
```
.git
.github
specs
.vscode
*.md
!README.md
Dockerfile
.dockerignore
coverage.out
coverage.html
*.test
```

**Validation Checklist**:
- [ ] Image builds successfully
- [ ] Image size <20MB (target <10MB)
- [ ] Local mode works via docker run -i
- [ ] Remote mode works with port mapping
- [ ] Healthcheck passes
- [ ] Performance tests meet requirements in container
- [ ] Memory usage <100MB under load
- [ ] Multi-architecture support (amd64, arm64)

## Post-Phase 1 Constitution Re-Check

**Re-validation**: All constitution principles verified against final design

### I. Code Quality Standards ✅
- Go standard project layout (`cmd/`, `internal/`, `pkg/`)
- Godoc comments planned for all exported symbols
- golangci-lint config defined
- Makefile targets for quality gates

### II. Test-First Development ✅
- Contract tests defined in contracts/mcp-protocol.md
- Integration test scenarios specified (stdio, HTTP)
- Unit test targets identified (echo handler, logger)
- Coverage targets: 80% minimum, 95% for critical paths

### III. User Experience Consistency ✅
- MCP protocol compliance ensured via official SDK
- Error messages follow JSON-RPC 2.0 standard codes
- Configuration via env vars with sensible defaults
- Logging dual-format (text for humans, JSON for machines)

### IV. Performance Requirements ✅
- Architecture supports <50ms latency (in-memory, no I/O)
- Goroutine concurrency enables 1000+ req/s
- Stateless design ensures <100MB memory
- Performance benchmarks and load tests defined

**Final Status**: ✅ ALL PRINCIPLES SATISFIED - Ready for implementation

---

## Summary

**Planning Complete** - All phases executed successfully:

- ✅ **Phase 0**: Research completed, technical decisions documented
- ✅ **Phase 1**: Design completed, data model and contracts defined
- ✅ **Dockerfile**: Implementation plan and testing strategy defined
- ✅ **Constitution**: All principles validated pre and post-design

**Artifacts Generated**:
1. `plan.md` - This comprehensive implementation plan
2. `research.md` - Technical research and decisions (7 areas)
3. `data-model.md` - MCP protocol data structures and validation rules
4. `contracts/mcp-protocol.md` - API contracts with test cases
5. `quickstart.md` - User-facing quick-start guide
6. `.github/copilot-instructions.md` - Updated with Go 1.21+ context

**Next Step**: Run `/speckit.tasks` to generate detailed task breakdown for implementation

**Branch**: `001-echo-server`  
**Ready for**: Implementation with TDD workflow
