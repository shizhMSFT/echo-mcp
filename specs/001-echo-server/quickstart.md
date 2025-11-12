# Quickstart Guide: Echo MCP Server

**Feature**: Echo MCP Server  
**Date**: 2025-11-12  
**Purpose**: Quick setup and usage guide for developers testing MCP clients

## Overview

The echo-mcp server is a minimal MCP (Model Context Protocol) server that provides an echo tool for testing MCP client implementations. It accepts messages and returns them unchanged while logging all activity for debugging.

## Prerequisites

- **Go**: Version 1.21 or higher
- **Docker**: For containerized deployment (optional)
- **MCP Client**: Any MCP-compatible client for testing

## Installation

### Option 1: Build from Source

```bash
# Clone repository
git clone https://github.com/shizhMSFT/echo-mcp.git
cd echo-mcp

# Download dependencies
go mod download

# Build binary
go build -o echo-mcp ./cmd/echo-mcp

# Verify installation
./echo-mcp --version
```

### Option 2: Docker

```bash
# Build Docker image
docker build -t echo-mcp:latest .

# Verify image
docker images echo-mcp
```

## Quick Start

### Local Mode (Stdio Transport)

Most common for development and testing:

```bash
# Start server in local mode
./echo-mcp --mode=local

# Or using Docker
docker run -i echo-mcp:latest --mode=local
```

The server now reads from stdin and writes to stdout. Send MCP protocol messages:

```bash
# Send initialize request
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test","version":"1.0"},"capabilities":{}}}' | ./echo-mcp

# Expected response:
# {"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","serverInfo":{"name":"echo-mcp","version":"0.1.0"},"capabilities":{"tools":{}}}}
```

### Remote Mode (HTTP Transport)

For distributed testing:

```bash
# Start server in remote mode (default port 8080)
./echo-mcp --mode=remote --port=8080

# Or using Docker
docker run -p 8080:8080 echo-mcp:latest --mode=remote
```

Send requests via HTTP:

```bash
# Initialize connection
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"curl-client","version":"1.0"},"capabilities":{}}}'

# Call echo tool
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Hello, MCP!"}}}'
```

## Basic Usage Examples

### Example 1: Simple Echo Test

```bash
# Interactive session (local mode)
./echo-mcp --mode=local

# Type this and press Enter:
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"test"}}}

# Server responds:
{"jsonrpc":"2.0","id":1,"result":{"content":[{"type":"text","text":"test"}]}}
```

### Example 2: List Available Tools

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./echo-mcp --mode=local

# Response shows echo tool schema:
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "tools": [
      {
        "name": "echo",
        "description": "Echoes back the provided message exactly as received. Useful for testing MCP client implementations.",
        "inputSchema": {
          "type": "object",
          "properties": {
            "message": {
              "type": "string",
              "description": "The message to echo back"
            }
          },
          "required": ["message"]
        }
      }
    ]
  }
}
```

### Example 3: Testing Unicode and Special Characters

```bash
# Unicode emoji test
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Hello 🌍 世界"}}}' | ./echo-mcp --mode=local

# Response preserves all characters:
{"jsonrpc":"2.0","id":1,"result":{"content":[{"type":"text","text":"Hello 🌍 世界"}]}}

# Special characters test
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Tab:\tNewline:\nQuote:\""}}}' | ./echo-mcp --mode=local
```

### Example 4: Error Handling

```bash
# Missing required parameter
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{}}}' | ./echo-mcp --mode=local

# Error response:
{"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"Invalid params: missing required parameter 'message'"}}

# Unknown tool
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"unknown","arguments":{"message":"test"}}}' | ./echo-mcp --mode=local

# Error response:
{"jsonrpc":"2.0","id":2,"error":{"code":-32601,"message":"Method not found: tool 'unknown' does not exist"}}
```

## Configuration

### Command-Line Flags

```bash
./echo-mcp [flags]

Flags:
  --mode string        Transport mode: "local" (stdio) or "remote" (HTTP) (default "local")
  --port int           Port for remote mode (default 8080)
  --log-format string  Log format: "text" or "json" (default "text")
  --log-level string   Log level: "debug", "info", "warn", "error" (default "info")
  --version            Show version information
  --help               Show help message
```

### Environment Variables

```bash
# Set via environment
export ECHO_MCP_MODE=remote
export ECHO_MCP_PORT=9000
export ECHO_MCP_LOG_FORMAT=json
export ECHO_MCP_LOG_LEVEL=debug

./echo-mcp
```

### Docker Configuration

```bash
# Environment variables
docker run -p 9000:9000 \
  -e ECHO_MCP_MODE=remote \
  -e ECHO_MCP_PORT=9000 \
  -e ECHO_MCP_LOG_FORMAT=json \
  echo-mcp:latest

# Command flags
docker run -p 8080:8080 \
  echo-mcp:latest \
  --mode=remote --port=8080 --log-format=json
```

## Logging and Debugging

### Log Output

All requests are logged to stdout (or stderr for errors) by default:

**Text Format** (default, human-readable):
```
2025-11-12T10:30:45.123Z INFO  [550e8400-e29b] request_received method=tools/call transport=stdio
2025-11-12T10:30:45.124Z INFO  [550e8400-e29b] tool_invoked tool=echo message_length=11
2025-11-12T10:30:45.125Z INFO  [550e8400-e29b] response_sent duration_ms=2
```

**JSON Format** (structured, machine-parsable):
```json
{"timestamp":"2025-11-12T10:30:45.123Z","level":"info","request_id":"550e8400-e29b","event":"request_received","method":"tools/call","transport":"stdio","raw_request":"{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"echo\",\"arguments\":{\"message\":\"Hello, MCP!\"}}}"}
```

### Debug Mode

Enable debug logging for detailed information:

```bash
./echo-mcp --log-level=debug

# Logs include:
# - Full request/response bodies
# - Internal state transitions
# - Performance metrics
```

### Redirecting Logs

Separate logs from MCP protocol responses:

```bash
# Logs to stderr, responses to stdout
./echo-mcp --mode=local 2>server.log

# View logs in real-time
tail -f server.log
```

## Testing Your MCP Client

### Complete Session Example

```bash
#!/bin/bash
# test-session.sh - Complete MCP session test

# Start echo server
./echo-mcp --mode=local > responses.txt 2> logs.txt &
SERVER_PID=$!

# Wait for startup
sleep 1

# Send initialize
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test","version":"1.0"},"capabilities":{}}}' >&1

# Send initialized notification
echo '{"jsonrpc":"2.0","method":"initialized","params":{}}' >&1

# List tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' >&1

# Call echo tool
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Hello, MCP!"}}}' >&1

# Cleanup
kill $SERVER_PID
```

### Using with MCP Client Libraries

**Python Example**:
```python
import subprocess
import json

# Start server
server = subprocess.Popen(
    ['./echo-mcp', '--mode=local'],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE
)

# Send request
request = {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
        "name": "echo",
        "arguments": {"message": "Hello from Python!"}
    }
}
server.stdin.write((json.dumps(request) + '\n').encode())
server.stdin.flush()

# Read response
response = server.stdout.readline()
print(json.loads(response))

server.terminate()
```

## Performance Testing

### Latency Test

```bash
# Measure single request latency
time echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"test"}}}' | ./echo-mcp --mode=local

# Expected: <100ms total time
```

### Throughput Test

```bash
# Send 1000 requests
for i in {1..1000}; do
    echo '{"jsonrpc":"2.0","id":'$i',"method":"tools/call","params":{"name":"echo","arguments":{"message":"test"}}}' | ./echo-mcp --mode=local &
done
wait

# Expected: All complete successfully within 1 second
```

### Load Test (Remote Mode)

```bash
# Using Apache Bench
ab -n 10000 -c 100 -p request.json -T application/json http://localhost:8080/mcp

# request.json content:
# {"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"test"}}}

# Expected:
# - Requests per second: >1000
# - 95th percentile latency: <100ms
```

## Troubleshooting

### Server Won't Start

```bash
# Check Go version
go version  # Should be 1.21+

# Check for port conflicts (remote mode)
lsof -i :8080
netstat -an | grep 8080

# Try different port
./echo-mcp --mode=remote --port=9000
```

### Responses Not Appearing

```bash
# Ensure newline-delimited JSON
echo '{"jsonrpc":"2.0","id":1,"method":"ping"}' | ./echo-mcp

# Check stderr for errors
./echo-mcp 2>&1 | tee server.log
```

### Invalid JSON Errors

```bash
# Validate JSON before sending
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"test"}}}' | jq .

# If jq succeeds, JSON is valid
```

### Docker Container Issues

```bash
# Check container logs
docker logs <container-id>

# Run interactively for debugging
docker run -it echo-mcp:latest /bin/sh

# Test inside container
./echo-mcp --version
```

## Next Steps

- **Integrate with your MCP client**: Use echo-mcp to validate your client implementation
- **Contract testing**: Verify your client adheres to MCP specification
- **Performance testing**: Benchmark your client against echo-mcp
- **Development**: Use as reference for building your own MCP servers

## Resources

- **MCP Specification**: https://spec.modelcontextprotocol.io/
- **Source Code**: https://github.com/shizhMSFT/echo-mcp
- **Issue Tracker**: https://github.com/shizhMSFT/echo-mcp/issues
- **Examples**: See `/examples` directory for client implementations

## Support

For issues, questions, or contributions:
- Open an issue on GitHub
- Check existing issues for solutions
- Contribute improvements via pull requests
