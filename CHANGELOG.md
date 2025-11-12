# Changelog

All notable changes to the echo-mcp project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-11-12

### Added

- Initial release of echo-mcp server
- Echo tool implementation that returns messages unchanged
- MCP protocol compliance (JSON-RPC 2.0 based)
- Stdio transport for local development (local mode)
- HTTP transport for distributed testing (remote mode)
- Comprehensive request logging with raw request data
- Structured logging support (text and JSON formats)
- Request ID generation (UUID v4) for correlation
- Health check endpoint (`GET /health`)
- CORS support for cross-origin HTTP requests
- Graceful shutdown handling (SIGTERM, SIGINT)
- Docker support with multi-stage build
- Comprehensive test suite:
  - Contract tests for MCP protocol compliance
  - Integration tests for stdio and HTTP transports
  - Unit tests for handlers and logging
- Configuration via environment variables or CLI flags
- Performance optimizations:
  - <100ms p95 latency
  - 1000+ requests/second throughput
  - <100MB memory footprint
- Documentation:
  - README with quick start guide
  - Architecture overview
  - Usage examples

### Security

- Non-root user in Docker container
- Static binary build (no C dependencies)
- Input validation for all MCP requests
- Error handling for malformed JSON

### Performance

- Zero-allocation logging with zerolog
- Goroutine-per-request concurrency model
- Efficient JSON marshaling/unmarshaling
- Minimal memory allocations

## [Unreleased]

### Planned

- Additional MCP tools (resources, prompts)
- WebSocket transport support
- Metrics and observability
- Rate limiting
- Authentication/authorization
- Configuration file support
- Plugin system for custom tools

[0.1.0]: https://github.com/shizhMSFT/echo-mcp/releases/tag/v0.1.0
[Unreleased]: https://github.com/shizhMSFT/echo-mcp/compare/v0.1.0...HEAD
