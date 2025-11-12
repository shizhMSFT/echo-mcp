# Tasks: Echo MCP Server

**Input**: Design documents from `/specs/001-echo-server/`  
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Following Test-Driven Development (TDD) per constitution - tests written before implementation

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: Repository root structure as defined in plan.md
- Paths follow Go standard layout: `cmd/`, `internal/`, `tests/`

---

## Phase 1: Setup (Project Initialization)

**Purpose**: Initialize Go project structure, dependencies, and tooling

- [X] T001 Initialize Go module with `go mod init github.com/shizhMSFT/echo-mcp` in repository root
- [X] T002 [P] Create directory structure: `cmd/echo-mcp/`, `internal/server/`, `internal/transport/`, `internal/protocol/`
- [X] T003 [P] Create test directories: `tests/contract/`, `tests/integration/`, `tests/unit/`
- [X] T004 [P] Create .gitignore with Go patterns (*.exe, *.test, vendor/, *.out, .env*, .DS_Store)
- [X] T005 [P] Create .dockerignore with build artifacts (.git, specs, .vscode, *.md except README.md, coverage.*)
- [X] T006 Add zerolog dependency: `go get github.com/rs/zerolog`
- [X] T007 [P] Create .golangci.yml with linter configuration (gofmt, go vet, gocyclo, staticcheck)
- [X] T008 [P] Create Makefile with targets: build, test, lint, run-local, run-remote, docker-build
- [X] T009 [P] Create README.md with project description, quick start, and usage examples
- [X] T010 [P] Create go.mod and verify all dependencies downloaded with `go mod download`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core MCP protocol infrastructure that MUST be complete before ANY user story implementation

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T011 Define MCP protocol types in internal/protocol/types.go (JSONRPCRequest, JSONRPCResponse, RPCError, InitializeParams, InitializeResult, ToolCallParams, ToolCallResult per data-model.md)
- [X] T012 [P] Define server info constants in internal/protocol/types.go (ServerName="echo-mcp", ProtocolVersion="2024-11-05")
- [X] T013 Implement MCP protocol message parser in internal/protocol/handlers.go (ParseRequest function with JSON unmarshaling and validation)
- [X] T014 Implement MCP protocol message builder in internal/protocol/handlers.go (BuildResponse, BuildErrorResponse functions)
- [X] T015 [P] Define logger interface and implementation in internal/server/logger.go (supports text and JSON formats, logs raw requests)
- [X] T016 Create server configuration struct in internal/server/server.go (Mode, Port, LogFormat, LogLevel fields)
- [X] T017 Implement configuration loading from environment variables in internal/server/server.go (ECHO_MCP_MODE, ECHO_MCP_PORT, ECHO_MCP_LOG_FORMAT, ECHO_MCP_LOG_LEVEL)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Basic Echo Functionality (Priority: P1) 🎯 MVP

**Goal**: Implement echo tool that accepts messages and returns them unchanged, with MCP protocol compliance

**Independent Test**: Send message to echo tool and verify response contains exact same message

### Tests for User Story 1 (TDD - Write First)

- [X] T018 [P] [US1] Contract test for tools/list in tests/contract/echo_tool_test.go (verify echo tool schema matches spec)
- [X] T019 [P] [US1] Contract test for tools/call with simple string in tests/contract/echo_tool_test.go (message="Hello, MCP!")
- [X] T020 [P] [US1] Contract test for tools/call with JSON string in tests/contract/echo_tool_test.go (verify JSON structure preserved)
- [X] T021 [P] [US1] Contract test for tools/call with empty string in tests/contract/echo_tool_test.go (message="")
- [X] T022 [P] [US1] Contract test for tools/call with unicode in tests/contract/echo_tool_test.go (message="Hello 世界 🌍")
- [X] T023 [P] [US1] Contract test for tools/call with special characters in tests/contract/echo_tool_test.go (tabs, newlines, quotes)
- [X] T024 [P] [US1] Unit test for echo handler in tests/unit/echo_handler_test.go (table-driven tests for message preservation)

### Implementation for User Story 1

- [X] T025 [US1] Implement echo tool definition in internal/server/echo_tool.go (ToolDefinition with name, description, inputSchema per data-model.md)
- [X] T026 [US1] Implement echo tool handler function in internal/server/echo_tool.go (HandleEcho that returns ToolCallResult with exact message copy)
- [X] T027 [US1] Implement tools/list handler in internal/protocol/handlers.go (returns array with echo tool definition)
- [X] T028 [US1] Implement tools/call router in internal/protocol/handlers.go (dispatches to echo handler based on tool name)
- [X] T029 [US1] Add error handling for missing message parameter in internal/server/echo_tool.go (return -32602 Invalid params error)
- [X] T030 [US1] Add error handling for unknown tool in internal/protocol/handlers.go (return -32601 Method not found error)
- [X] T031 [US1] Run all User Story 1 tests and verify they pass: `go test ./tests/contract/echo_tool_test.go ./tests/unit/echo_handler_test.go -v`

**Checkpoint**: User Story 1 complete - Echo tool functional and independently testable

---

## Phase 4: User Story 3 - Local Transport Support (Priority: P1) 🎯 MVP

**Goal**: Enable stdio-based communication for local development and testing

**Independent Test**: Launch server as local process, send MCP messages via stdin, receive via stdout

**Note**: Implementing US3 before US2 because stdio transport is required for basic functionality testing

### Tests for User Story 3 (TDD - Write First)

- [X] T032 [P] [US3] Integration test for stdio initialize handshake in tests/integration/stdio_transport_test.go
- [X] T033 [P] [US3] Integration test for stdio echo request/response in tests/integration/stdio_transport_test.go
- [X] T034 [P] [US3] Integration test for stdio sequential requests in tests/integration/stdio_transport_test.go (verify no interleaving)
- [X] T035 [P] [US3] Integration test for stdio graceful shutdown in tests/integration/stdio_transport_test.go (client closes stdin)
- [ ] T036 [P] [US3] Unit test for stdio message reader in tests/unit/stdio_transport_test.go (line-delimited JSON parsing)

### Implementation for User Story 3

- [X] T037 [US3] Implement stdio transport in internal/transport/stdio.go (StdioTransport struct with Run method)
- [X] T038 [US3] Implement stdin reader using bufio.Scanner in internal/transport/stdio.go (reads line-delimited JSON)
- [X] T039 [US3] Implement stdout writer in internal/transport/stdio.go (writes JSON response followed by newline)
- [X] T040 [US3] Implement MCP request handling loop in internal/transport/stdio.go (read request → process → write response)
- [X] T041 [US3] Implement initialize handler integration in internal/transport/stdio.go (calls protocol handlers)
- [X] T042 [US3] Implement graceful shutdown on stdin EOF in internal/transport/stdio.go (context cancellation)
- [X] T043 [US3] Create main.go entry point in cmd/echo-mcp/main.go (CLI flags: --mode, --port, --log-format, --log-level)
- [X] T044 [US3] Implement server bootstrap in cmd/echo-mcp/main.go (initialize config, logger, transport based on mode flag)
- [X] T045 [US3] Run all User Story 3 tests and verify they pass: `go test ./tests/integration/stdio_transport_test.go -v`
- [X] T046 [US3] Manual test: echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test","version":"1.0"},"capabilities":{}}}' | go run cmd/echo-mcp/main.go --mode=local

**Checkpoint**: User Story 3 complete - Stdio transport functional, server can run in local mode

---

## Phase 5: User Story 2 - Request Logging and Debugging (Priority: P2)

**Goal**: Log all requests with raw request data to stdout for debugging

**Independent Test**: Send requests and verify logs contain complete raw request with timestamps

### Tests for User Story 2 (TDD - Write First)

- [ ] T047 [P] [US2] Unit test for logger raw request formatting in tests/unit/logger_test.go (verify raw JSON included)
- [ ] T048 [P] [US2] Unit test for logger timestamp format in tests/unit/logger_test.go (ISO 8601)
- [ ] T049 [P] [US2] Unit test for logger request ID generation in tests/unit/logger_test.go (UUID v4)
- [ ] T050 [P] [US2] Integration test for initialize request logging in tests/integration/logging_test.go (raw request logged)
- [ ] T051 [P] [US2] Integration test for tools/call request logging in tests/integration/logging_test.go (all requests logged chronologically)

### Implementation for User Story 2

- [ ] T052 [US2] Implement request ID generation in internal/server/logger.go (UUID v4 for each request)
- [ ] T053 [US2] Implement raw request logging in internal/server/logger.go (log complete JSON before parsing)
- [ ] T054 [US2] Implement structured log entry creation in internal/server/logger.go (timestamp, request_id, event_type, method, raw_request, parsed_data)
- [ ] T055 [US2] Add logging to initialize handler in internal/protocol/handlers.go (log raw request with event_type="request_received")
- [ ] T056 [US2] Add logging to tools/call handler in internal/protocol/handlers.go (log raw request with event_type="tool_invoked")
- [ ] T057 [US2] Add logging to error paths in internal/protocol/handlers.go (log malformed requests with event_type="error")
- [ ] T058 [US2] Integrate logger with stdio transport in internal/transport/stdio.go (log each request before processing)
- [ ] T059 [US2] Add configuration for log format in cmd/echo-mcp/main.go (--log-format flag: text or json)
- [ ] T060 [US2] Run all User Story 2 tests and verify they pass: `go test ./tests/unit/logger_test.go ./tests/integration/logging_test.go -v`

**Checkpoint**: User Story 2 complete - All requests logged with raw data

---

## Phase 6: User Story 4 - Remote Transport Support (Priority: P2)

**Goal**: Enable HTTP-based communication for distributed testing scenarios

**Independent Test**: Run server in remote mode, connect via HTTP, exchange echo messages

### Tests for User Story 4 (TDD - Write First)

- [ ] T061 [P] [US4] Integration test for HTTP initialize handshake in tests/integration/http_transport_test.go
- [ ] T062 [P] [US4] Integration test for HTTP echo request/response in tests/integration/http_transport_test.go
- [ ] T063 [P] [US4] Integration test for HTTP concurrent clients in tests/integration/http_transport_test.go (multiple simultaneous connections)
- [ ] T064 [P] [US4] Integration test for HTTP client disconnect in tests/integration/http_transport_test.go (verify resource cleanup)
- [ ] T065 [P] [US4] Unit test for HTTP handler in tests/unit/http_transport_test.go (POST /mcp endpoint)

### Implementation for User Story 4

- [ ] T066 [US4] Implement HTTP transport in internal/transport/http.go (HTTPTransport struct with Run method)
- [ ] T067 [US4] Implement POST /mcp handler in internal/transport/http.go (reads JSON from request body)
- [ ] T068 [US4] Implement request processing in internal/transport/http.go (parse → handle → respond)
- [ ] T069 [US4] Implement JSON response writer in internal/transport/http.go (writes JSON response with proper content-type)
- [ ] T070 [US4] Add CORS headers for cross-origin testing in internal/transport/http.go (Access-Control-Allow-Origin: *)
- [ ] T071 [US4] Implement concurrent request handling in internal/transport/http.go (goroutine per request via net/http)
- [ ] T072 [US4] Implement graceful shutdown on SIGTERM in internal/transport/http.go (http.Server.Shutdown with context)
- [ ] T073 [US4] Integrate logger with HTTP transport in internal/transport/http.go (log each request before processing)
- [ ] T074 [US4] Add health check endpoint GET /health in internal/transport/http.go (returns 200 OK)
- [ ] T075 [US4] Update main.go to support remote mode in cmd/echo-mcp/main.go (start HTTP server when --mode=remote)
- [ ] T076 [US4] Run all User Story 4 tests and verify they pass: `go test ./tests/integration/http_transport_test.go -v`
- [ ] T077 [US4] Manual test: Start server with `go run cmd/echo-mcp/main.go --mode=remote --port=8080`, then curl -X POST http://localhost:8080/mcp with test request

**Checkpoint**: User Story 4 complete - HTTP transport functional, server supports remote mode

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final quality improvements, documentation, containerization, and validation

### MCP Protocol Compliance Tests

- [ ] T078 [P] Contract test for initialize with incompatible version in tests/contract/mcp_protocol_test.go (verify error response)
- [ ] T079 [P] Contract test for malformed JSON in tests/contract/mcp_protocol_test.go (verify parse error -32700)
- [ ] T080 [P] Contract test for invalid JSON-RPC structure in tests/contract/mcp_protocol_test.go (verify invalid request -32600)
- [ ] T081 [P] Contract test for ping method in tests/contract/mcp_protocol_test.go (optional keep-alive)

### Performance and Reliability

- [ ] T082 [P] Benchmark test for echo handler in tests/unit/echo_handler_test.go (BenchmarkEchoHandler, verify <100ms p95)
- [ ] T083 [P] Benchmark test for JSON marshaling in tests/unit/protocol_test.go (BenchmarkJSONMarshaling)
- [ ] T084 [P] Race condition test: run `go test -race ./...` and verify no data races
- [ ] T085 [P] Load test script for stdio transport in tests/integration/load_test.sh (1000 sequential requests)
- [ ] T086 [P] Load test script for HTTP transport in tests/integration/load_test.sh (1000 concurrent requests with ab)
- [ ] T087 [P] Memory profiling test in tests/integration/memory_test.go (verify <100MB under load)

### Docker and Deployment

- [ ] T088 Create multi-stage Dockerfile (builder stage with Go 1.21-alpine, tester stage with tests, runtime stage with Alpine)
- [ ] T089 Add CGO_ENABLED=0 static build to Dockerfile (ensures no C dependencies)
- [ ] T090 Add healthcheck to Dockerfile (CMD echo '{"jsonrpc":"2.0","id":1,"method":"ping"}' | ./echo-mcp --mode=local)
- [ ] T091 [P] Test Docker build: `docker build -t echo-mcp:latest .`
- [ ] T092 [P] Test Docker local mode: echo test request | docker run -i echo-mcp:latest --mode=local
- [ ] T093 [P] Test Docker remote mode: docker run -p 8080:8080 echo-mcp:latest --mode=remote, then curl test
- [ ] T094 [P] Test Docker memory usage: docker stats (verify <100MB)

### Documentation and Code Quality

- [ ] T095 [P] Add godoc comments to all exported types and functions in internal/ packages
- [ ] T096 [P] Update README.md with installation, usage, configuration, examples from quickstart.md
- [ ] T097 [P] Create CHANGELOG.md with v0.1.0 release notes
- [ ] T098 [P] Run golangci-lint and fix all issues: `golangci-lint run`
- [ ] T099 [P] Run gofmt on all Go files: `gofmt -s -w .`
- [ ] T100 [P] Verify test coverage ≥80%: `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out`

### Final Validation

- [ ] T101 Run complete test suite: `go test -v -race -coverprofile=coverage.out ./...`
- [ ] T102 Build production binary: `go build -o echo-mcp ./cmd/echo-mcp`
- [ ] T103 Manual end-to-end test: Complete session flow from contracts/mcp-protocol.md (initialize → tools/list → tools/call)
- [ ] T104 Verify all acceptance scenarios from spec.md (4 user stories × 4-5 scenarios each)
- [ ] T105 Generate coverage report: `go tool cover -html=coverage.out -o coverage.html`

---

## Dependencies

### User Story Completion Order

```text
Phase 1: Setup
    ↓
Phase 2: Foundational (blocking)
    ↓
    ├─→ Phase 3: US1 (Echo Functionality) ✓ MVP
    │       ↓
    ├─→ Phase 4: US3 (Stdio Transport) ✓ MVP - depends on US1
    │       ↓
    ├─→ Phase 5: US2 (Logging) - enhances US1+US3
    │       ↓
    └─→ Phase 6: US4 (HTTP Transport) - depends on US1, parallel to US2
            ↓
        Phase 7: Polish
```

**MVP Scope** (Minimum Viable Product):
- Phase 1: Setup ✓
- Phase 2: Foundational ✓
- Phase 3: US1 - Basic Echo Functionality ✓
- Phase 4: US3 - Stdio Transport ✓

After MVP, phases 5-7 can be implemented incrementally.

### Parallel Execution Examples

**Phase 1 Setup** - Can run in parallel:
- T002, T003, T004, T005, T007, T008, T009 (different files, no dependencies)

**Phase 3 Tests** - Can run in parallel:
- T018, T019, T020, T021, T022, T023, T024 (independent test files)

**Phase 5 Tests** - Can run in parallel:
- T047, T048, T049 (unit tests in same file, different test functions)
- T050, T051 (integration tests)

**Phase 7 Polish** - Can run in parallel:
- T082, T083, T084, T085, T086, T087 (independent tests)
- T091, T092, T093, T094 (Docker tests, sequential within Docker but parallel to other tasks)
- T095, T096, T097, T098, T099 (documentation and linting)

---

## Implementation Strategy

### MVP-First Approach

1. **Week 1**: Phases 1-2 (Setup + Foundation)
2. **Week 2**: Phases 3-4 (US1 Echo + US3 Stdio) → **MVP READY**
3. **Week 3**: Phases 5-6 (US2 Logging + US4 HTTP)
4. **Week 4**: Phase 7 (Polish, Docker, Documentation)

### Independent Story Delivery

Each user story phase is independently testable:
- **US1**: Can test echo functionality standalone
- **US3**: Can test stdio transport with US1
- **US2**: Can test logging with US1+US3
- **US4**: Can test HTTP transport with US1 (independent of US3)

### TDD Workflow per Task

1. **Red**: Run failing test (T018-T024 for US1, etc.)
2. **Green**: Implement minimum code (T025-T030 for US1)
3. **Refactor**: Improve quality while keeping tests green
4. **Validate**: Run `go test`, `go vet`, `golangci-lint`

---

## Task Summary

**Total Tasks**: 105

**By Phase**:
- Phase 1 (Setup): 10 tasks
- Phase 2 (Foundational): 7 tasks
- Phase 3 (US1 - Echo): 14 tasks (7 tests + 7 implementation)
- Phase 4 (US3 - Stdio): 15 tasks (5 tests + 10 implementation)
- Phase 5 (US2 - Logging): 13 tasks (5 tests + 8 implementation)
- Phase 6 (US4 - HTTP): 17 tasks (5 tests + 12 implementation)
- Phase 7 (Polish): 29 tasks (tests, Docker, docs, validation)

**Parallel Opportunities**: 45 tasks marked [P] can run in parallel with others in same phase

**Test Tasks**: 35 (following TDD - all written before implementation)
**Implementation Tasks**: 52
**Setup/Polish Tasks**: 18

**MVP Tasks**: 46 (Phases 1-4 only)
**Post-MVP Tasks**: 59 (Phases 5-7)
