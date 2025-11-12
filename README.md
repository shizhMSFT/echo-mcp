# echo-mcp

A minimal Model Context Protocol (MCP) server that echoes back client messages. Designed for testing, debugging, and demonstrating MCP integrations with simple request-response behavior.

## Features

- **Echo Tool**: Accepts messages and returns them unchanged for testing MCP client implementations
- **Dual Transport**: Supports both local (stdio) and remote (HTTP) communication modes
- **Request Logging**: Logs all requests with raw data to stdout for debugging
- **MCP Protocol Compliance**: Full implementation of MCP specification based on JSON-RPC 2.0
- **High Performance**: <100ms p95 latency, 1000+ req/s throughput, <100MB memory usage

## Prerequisites

- **Go**: Version 1.21 or higher
- **Docker**: For containerized deployment (optional)

## Installation

### From Source

```bash
# Clone repository
git clone https://github.com/shizhMSFT/echo-mcp.git
cd echo-mcp

# Download dependencies
go mod download

# Build binary
make build

# Or use go directly
go build -o echo-mcp ./cmd/echo-mcp
```

### Using Docker

```bash
# Build Docker image
make docker-build

# Or use docker directly
docker build -t echo-mcp:latest .
```

## Quick Start

### Local Mode (Stdio Transport)

```bash
# Start server in local mode
./echo-mcp --mode=local

# Send test message
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test","version":"1.0"},"capabilities":{}}}' | ./echo-mcp --mode=local
```

### Remote Mode (HTTP Transport)

```bash
# Start server in remote mode (default port 8080)
./echo-mcp --mode=remote --port=8080

# Send test message via HTTP
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Hello, MCP!"}}}'
```

## Configuration

Configure the server using command-line flags or environment variables:

| Flag | Environment Variable | Default | Description |
|------|---------------------|---------|-------------|
| `--mode` | `ECHO_MCP_MODE` | `local` | Transport mode: `local` (stdio) or `remote` (HTTP) |
| `--port` | `ECHO_MCP_PORT` | `8080` | HTTP server port (remote mode only) |
| `--log-format` | `ECHO_MCP_LOG_FORMAT` | `text` | Log format: `text` or `json` |
| `--log-level` | `ECHO_MCP_LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |

## Usage Examples

### List Available Tools

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./echo-mcp --mode=local
```

### Call Echo Tool

```bash
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"echo","arguments":{"message":"test"}}}' | ./echo-mcp --mode=local
```

### Docker Examples

```bash
# Local mode
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"docker-test","version":"1.0"},"capabilities":{}}}' | docker run -i echo-mcp:latest --mode=local

# Remote mode
docker run -p 8080:8080 echo-mcp:latest --mode=remote
```

## Development

### Build and Test

```bash
# Run all tests
make test

# Run linters
make lint

# Build binary
make build

# Run all checks (lint + test + build)
make all
```

### Test Coverage

```bash
# Generate coverage report
make coverage

# View coverage in browser
open coverage.html
```

## Architecture

The server follows Go standard project layout:

- `cmd/echo-mcp/` - Main application entry point
- `internal/protocol/` - MCP protocol types and handlers
- `internal/server/` - Core server logic and echo tool implementation
- `internal/transport/` - Transport layer (stdio and HTTP)
- `tests/` - Contract, integration, and unit tests

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please ensure:
- All tests pass (`make test`)
- Code is formatted (`gofmt -s -w .`)
- Linters pass (`make lint`)
- Test coverage ≥80%

