# MCP Protocol Contracts: Echo Server

**Feature**: Echo MCP Server  
**Date**: 2025-11-12  
**Protocol**: Model Context Protocol (MCP) based on JSON-RPC 2.0

## Overview

This document defines the API contracts for the echo MCP server. All endpoints follow the MCP specification and JSON-RPC 2.0 message format.

## Transport Endpoints

### Stdio Transport (Local Mode)

**Endpoint**: Standard Input/Output Streams  
**Protocol**: JSON-RPC 2.0 messages over stdio  
**Message Format**: One JSON object per line (newline-delimited)

**Request Flow**:
1. Client writes JSON-RPC request to stdin
2. Server reads from stdin line-by-line
3. Server writes JSON-RPC response to stdout
4. Client reads from stdout line-by-line

**Example**:
```bash
# Input (stdin)
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test-client","version":"1.0.0"},"capabilities":{}}}

# Output (stdout)
{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","serverInfo":{"name":"echo-mcp","version":"0.1.0"},"capabilities":{"tools":{}}}}
```

---

### HTTP Transport (Remote Mode)

**Endpoint**: `POST /mcp`  
**Protocol**: HTTP with Server-Sent Events (SSE) for streaming  
**Headers**:
- `Content-Type: application/json`
- `Accept: text/event-stream` (for SSE responses)

**Request Flow**:
1. Client sends POST request with JSON-RPC message in body
2. Server processes request
3. Server sends response as SSE event or direct JSON response
4. Connection kept alive for multiple requests

**Example**:
```http
POST /mcp HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"test"}}}
```

**Response**:
```http
HTTP/1.1 200 OK
Content-Type: application/json

{"jsonrpc":"2.0","id":1,"result":{"content":[{"type":"text","text":"test"}]}}
```

---

## MCP Protocol Methods

### 1. initialize

**Description**: Establish MCP connection and exchange capabilities

**Method**: `initialize`  
**Required**: Yes (first message in session)

**Request**:
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
        "protocolVersion": "2024-11-05",
        "capabilities": {
            "roots": {
                "listChanged": true
            }
        },
        "clientInfo": {
            "name": "example-client",
            "version": "1.0.0"
        }
    }
}
```

**Response**:
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "protocolVersion": "2024-11-05",
        "capabilities": {
            "tools": {
                "listChanged": false
            },
            "logging": {}
        },
        "serverInfo": {
            "name": "echo-mcp",
            "version": "0.1.0"
        }
    }
}
```

**Validation**:
- `params.protocolVersion` MUST be "2024-11-05" or compatible
- `params.clientInfo.name` MUST be non-empty string
- Server responds with matching `protocolVersion`
- Server advertises `tools` capability

**Error Cases**:
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "error": {
        "code": -32602,
        "message": "Invalid params: incompatible protocol version",
        "data": {
            "supported": "2024-11-05",
            "requested": "2023-01-01"
        }
    }
}
```

---

### 2. initialized (notification)

**Description**: Client confirms initialization complete

**Method**: `initialized`  
**Type**: Notification (no response expected)

**Request**:
```json
{
    "jsonrpc": "2.0",
    "method": "initialized",
    "params": {}
}
```

**Response**: None (notification)

---

### 3. tools/list

**Description**: List available tools on the server

**Method**: `tools/list`

**Request**:
```json
{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list",
    "params": {}
}
```

**Response**:
```json
{
    "jsonrpc": "2.0",
    "id": 2,
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

**Validation**:
- Response MUST include array of tools
- Echo tool MUST be present in array
- Each tool MUST have name, description, inputSchema

---

### 4. tools/call

**Description**: Execute the echo tool

**Method**: `tools/call`

**Request**:
```json
{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
        "name": "echo",
        "arguments": {
            "message": "Hello, MCP!"
        }
    }
}
```

**Response**:
```json
{
    "jsonrpc": "2.0",
    "id": 3,
    "result": {
        "content": [
            {
                "type": "text",
                "text": "Hello, MCP!"
            }
        ]
    }
}
```

**Validation**:
- `params.name` MUST be "echo"
- `params.arguments.message` MUST be present
- `result.content[0].text` MUST exactly match `params.arguments.message`
- Message preservation requirements:
  - ✅ Exact string match (case-sensitive)
  - ✅ Unicode characters preserved
  - ✅ Whitespace preserved
  - ✅ Special characters preserved
  - ✅ Empty strings supported
  - ✅ JSON structure preserved (if message is JSON string)

**Test Cases**:

**Simple String**:
```json
{"arguments": {"message": "test"}}
→ {"content": [{"type": "text", "text": "test"}]}
```

**Unicode**:
```json
{"arguments": {"message": "Hello 世界 🌍"}}
→ {"content": [{"type": "text", "text": "Hello 世界 🌍"}]}
```

**Empty String**:
```json
{"arguments": {"message": ""}}
→ {"content": [{"type": "text", "text": ""}]}
```

**JSON String**:
```json
{"arguments": {"message": "{\"key\":\"value\"}"}}
→ {"content": [{"type": "text", "text": "{\"key\":\"value\"}"}]}
```

**Special Characters**:
```json
{"arguments": {"message": "Tab:\t Newline:\n Quote:\" Backslash:\\"}}
→ {"content": [{"type": "text", "text": "Tab:\t Newline:\n Quote:\" Backslash:\\"}]}
```

**Error Cases**:

**Missing message parameter**:
```json
{
    "jsonrpc": "2.0",
    "id": 3,
    "error": {
        "code": -32602,
        "message": "Invalid params: missing required parameter 'message'",
        "data": {
            "required": ["message"],
            "received": []
        }
    }
}
```

**Unknown tool**:
```json
{
    "jsonrpc": "2.0",
    "id": 3,
    "error": {
        "code": -32601,
        "message": "Method not found: tool 'unknown' does not exist",
        "data": {
            "available_tools": ["echo"]
        }
    }
}
```

---

### 5. ping (optional)

**Description**: Keep-alive check

**Method**: `ping`

**Request**:
```json
{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "ping",
    "params": {}
}
```

**Response**:
```json
{
    "jsonrpc": "2.0",
    "id": 4,
    "result": {}
}
```

---

## Error Codes

Following JSON-RPC 2.0 specification:

| Code | Message | Description |
|------|---------|-------------|
| -32700 | Parse error | Invalid JSON received |
| -32600 | Invalid Request | JSON-RPC request structure invalid |
| -32601 | Method not found | Method does not exist or tool not available |
| -32602 | Invalid params | Invalid method parameters |
| -32603 | Internal error | Internal server error |

**Example Error Response**:
```json
{
    "jsonrpc": "2.0",
    "id": null,
    "error": {
        "code": -32700,
        "message": "Parse error: invalid JSON",
        "data": {
            "position": 15,
            "received": "{\"jsonrpc\":\"2.0\""
        }
    }
}
```

---

## Request/Response Examples

### Complete Session Flow

```jsonc
// 1. Client initiates connection
→ {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test-client","version":"1.0"},"capabilities":{}}}
← {"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","serverInfo":{"name":"echo-mcp","version":"0.1.0"},"capabilities":{"tools":{}}}}

// 2. Client confirms initialization
→ {"jsonrpc":"2.0","method":"initialized","params":{}}

// 3. Client lists available tools
→ {"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
← {"jsonrpc":"2.0","id":2,"result":{"tools":[{"name":"echo","description":"Echoes back the provided message exactly as received. Useful for testing MCP client implementations.","inputSchema":{"type":"object","properties":{"message":{"type":"string","description":"The message to echo back"}},"required":["message"]}}]}}

// 4. Client calls echo tool
→ {"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Hello, MCP!"}}}
← {"jsonrpc":"2.0","id":3,"result":{"content":[{"type":"text","text":"Hello, MCP!"}]}}

// 5. Client calls echo with different message
→ {"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Testing 123"}}}
← {"jsonrpc":"2.0","id":4,"result":{"content":[{"type":"text","text":"Testing 123"}]}}

// 6. Client closes connection (graceful shutdown)
→ EOF or close HTTP connection
```

---

## Logging Contract

All requests MUST be logged with the following information:

**Log Entry Format**:
```json
{
    "timestamp": "2025-11-12T10:30:45.123Z",
    "level": "info",
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "event": "request_received",
    "transport": "stdio|http",
    "raw_request": "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"echo\",\"arguments\":{\"message\":\"test\"}}}",
    "method": "tools/call",
    "parsed": {
        "tool": "echo",
        "message_length": 4
    }
}
```

**Required Log Events**:
- `server_started`: Server initialization
- `request_received`: Every incoming request (with raw request)
- `tool_invoked`: Tool execution
- `response_sent`: Response returned to client
- `error`: Any error condition
- `server_shutdown`: Graceful shutdown

---

## Performance Contracts

Based on success criteria from spec.md:

| Metric | Requirement | Measurement Method |
|--------|-------------|-------------------|
| Latency (p95) | <100ms for messages <1KB | Benchmark tests, percentile analysis |
| Throughput | ≥1000 req/s | Load testing with concurrent clients |
| Memory | <100MB under normal load | Memory profiling, monitoring |
| Uptime | 99.9% (no crashes) | Error handling tests, fuzzing |
| Concurrent Connections | ≥100 simultaneous clients | Integration tests with multiple clients |

**Performance Test Endpoints**:
```bash
# Throughput test
for i in {1..1000}; do
    echo '{"jsonrpc":"2.0","id":'$i',"method":"tools/call","params":{"name":"echo","arguments":{"message":"test"}}}' | ./echo-mcp &
done

# Latency test
time echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"test"}}}' | ./echo-mcp
```

---

## Contract Testing Checklist

- [ ] Initialize request returns correct capabilities
- [ ] Initialize with incompatible version returns error
- [ ] Tools/list includes echo tool with correct schema
- [ ] Tools/call with valid message returns exact echo
- [ ] Tools/call with empty string succeeds
- [ ] Tools/call with unicode preserves characters
- [ ] Tools/call with special characters preserves them
- [ ] Tools/call with missing message returns error
- [ ] Tools/call with unknown tool returns error
- [ ] Malformed JSON returns parse error (-32700)
- [ ] Invalid JSON-RPC structure returns invalid request (-32600)
- [ ] All requests logged with raw request data
- [ ] Initialize request logged before response
- [ ] Concurrent requests handled without data corruption
- [ ] Stdio transport handles sequential requests
- [ ] HTTP transport handles concurrent connections
- [ ] Server shuts down gracefully on SIGTERM
- [ ] Performance: <100ms p95 latency
- [ ] Performance: ≥1000 req/s throughput
- [ ] Performance: <100MB memory usage
