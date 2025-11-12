# Data Model: Echo MCP Server

**Feature**: Echo MCP Server  
**Date**: 2025-11-12  
**Purpose**: Define data structures and entities for MCP protocol implementation

## Overview

This document defines the data structures used in the echo MCP server. Since MCP is a protocol-driven server with no persistent storage, this focuses on in-memory data structures for protocol messages, requests, responses, and logging.

## Core Entities

### 1. MCP Protocol Messages

These structures represent the Model Context Protocol message formats (JSON-RPC 2.0 based).

#### JSONRPCRequest
```go
// JSONRPCRequest represents an incoming MCP request
type JSONRPCRequest struct {
    JSONRPC string          `json:"jsonrpc"` // Must be "2.0"
    ID      interface{}     `json:"id"`      // String, number, or null
    Method  string          `json:"method"`  // e.g., "initialize", "tools/call"
    Params  json.RawMessage `json:"params"`  // Method-specific parameters
}
```

**Validation Rules**:
- `JSONRPC` MUST equal "2.0"
- `ID` MUST be present for requests requiring responses
- `Method` MUST be non-empty string
- `Params` is optional, defaults to empty object

**State Transitions**: N/A (immutable once parsed)

---

#### JSONRPCResponse
```go
// JSONRPCResponse represents an outgoing MCP response
type JSONRPCResponse struct {
    JSONRPC string      `json:"jsonrpc"` // Must be "2.0"
    ID      interface{} `json:"id"`      // Matches request ID
    Result  interface{} `json:"result,omitempty"`
    Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

**Validation Rules**:
- Either `Result` or `Error` MUST be present, never both
- `ID` MUST match the corresponding request ID
- `Error.Code` follows JSON-RPC 2.0 error codes:
  - -32700: Parse error
  - -32600: Invalid request
  - -32601: Method not found
  - -32602: Invalid params
  - -32603: Internal error

---

### 2. MCP Initialize Protocol

#### InitializeParams
```go
// InitializeParams contains client capabilities sent during initialization
type InitializeParams struct {
    ProtocolVersion string                 `json:"protocolVersion"`
    Capabilities    ClientCapabilities     `json:"capabilities"`
    ClientInfo      ClientInfo             `json:"clientInfo"`
}

type ClientCapabilities struct {
    Roots      *RootsCapability      `json:"roots,omitempty"`
    Sampling   *SamplingCapability   `json:"sampling,omitempty"`
}

type ClientInfo struct {
    Name    string `json:"name"`
    Version string `json:"version"`
}
```

**Validation Rules**:
- `ProtocolVersion` MUST be compatible version (e.g., "2024-11-05")
- `ClientInfo.Name` MUST be non-empty

---

#### InitializeResult
```go
// InitializeResult contains server capabilities returned to client
type InitializeResult struct {
    ProtocolVersion string             `json:"protocolVersion"`
    Capabilities    ServerCapabilities `json:"capabilities"`
    ServerInfo      ServerInfo         `json:"serverInfo"`
}

type ServerCapabilities struct {
    Tools      *ToolsCapability      `json:"tools,omitempty"`
    Logging    *LoggingCapability    `json:"logging,omitempty"`
}

type ToolsCapability struct {
    ListChanged bool `json:"listChanged,omitempty"`
}

type ServerInfo struct {
    Name    string `json:"name"`
    Version string `json:"version"`
}
```

**Validation Rules**:
- `ProtocolVersion` MUST match supported version
- `ServerInfo.Name` MUST be "echo-mcp"
- `Capabilities.Tools` MUST be present (server provides echo tool)

---

### 3. Echo Tool Structures

#### ToolDefinition
```go
// ToolDefinition describes the echo tool schema
type ToolDefinition struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
    Type       string                 `json:"type"` // "object"
    Properties map[string]PropertyDef `json:"properties"`
    Required   []string               `json:"required,omitempty"`
}

type PropertyDef struct {
    Type        string `json:"type"`
    Description string `json:"description,omitempty"`
}
```

**Example - Echo Tool Schema**:
```json
{
    "name": "echo",
    "description": "Echoes back the provided message exactly as received",
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
```

**Validation Rules**:
- `Name` MUST be "echo"
- `InputSchema.Type` MUST be "object"
- `Properties` MUST include "message" field

---

#### ToolCallParams
```go
// ToolCallParams represents parameters for tools/call request
type ToolCallParams struct {
    Name      string                 `json:"name"`
    Arguments map[string]interface{} `json:"arguments"`
}
```

**Validation Rules**:
- `Name` MUST be "echo" (only tool supported)
- `Arguments` MUST contain "message" key

---

#### ToolCallResult
```go
// ToolCallResult represents the result of a tool execution
type ToolCallResult struct {
    Content []ContentItem `json:"content"`
    IsError bool          `json:"isError,omitempty"`
}

type ContentItem struct {
    Type string `json:"type"` // "text"
    Text string `json:"text"`
}
```

**Example - Echo Response**:
```json
{
    "content": [
        {
            "type": "text",
            "text": "Hello, MCP!"
        }
    ]
}
```

**Validation Rules**:
- `Content` MUST contain at least one item
- For echo tool, `Content[0].Text` MUST exactly match input message

---

### 4. Logging Structures

#### LogEntry
```go
// LogEntry represents a structured log entry for server operations
type LogEntry struct {
    Timestamp   time.Time              `json:"timestamp"`
    RequestID   string                 `json:"request_id"`
    EventType   EventType              `json:"event_type"`
    Method      string                 `json:"method,omitempty"`
    RawRequest  json.RawMessage        `json:"raw_request,omitempty"`
    ParsedData  map[string]interface{} `json:"parsed_data,omitempty"`
    Error       string                 `json:"error,omitempty"`
    DurationMs  int64                  `json:"duration_ms,omitempty"`
}

type EventType string

const (
    EventRequestReceived  EventType = "request_received"
    EventToolInvoked      EventType = "tool_invoked"
    EventResponseSent     EventType = "response_sent"
    EventError            EventType = "error"
    EventServerStarted    EventType = "server_started"
    EventServerShutdown   EventType = "server_shutdown"
)
```

**Validation Rules**:
- `Timestamp` MUST be ISO 8601 format
- `RequestID` MUST be unique per request (UUID v4)
- `RawRequest` MUST contain complete unprocessed JSON for `EventRequestReceived`
- `EventType` MUST be one of the defined constants

**Example Log Entry**:
```json
{
    "timestamp": "2025-11-12T10:30:45.123Z",
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "event_type": "request_received",
    "method": "tools/call",
    "raw_request": "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"echo\",\"arguments\":{\"message\":\"test\"}}}",
    "parsed_data": {
        "tool": "echo",
        "message": "test"
    }
}
```

---

### 5. Transport Session Structures

#### TransportSession
```go
// TransportSession manages a communication session between client and server
type TransportSession struct {
    ID           string
    Type         TransportType
    State        SessionState
    CreatedAt    time.Time
    LastActivity time.Time
    ClientID     string // For remote transport; empty for stdio
}

type TransportType string

const (
    TransportStdio TransportType = "stdio"
    TransportHTTP  TransportType = "http"
)

type SessionState string

const (
    SessionActive SessionState = "active"
    SessionClosed SessionState = "closed"
    SessionError  SessionState = "error"
)
```

**Validation Rules**:
- `ID` MUST be unique (UUID v4)
- `Type` MUST be either "stdio" or "http"
- `State` transitions: active → closed OR active → error
- `LastActivity` updated on every request/response

**State Transitions**:
```
[Created] → active
active → active (on request/response)
active → closed (graceful shutdown)
active → error (connection failure)
```

---

## Relationships

### Entity Relationship Diagram

```
┌─────────────────┐
│ TransportSession│
│  - ID           │
│  - Type         │
│  - State        │
└────────┬────────┘
         │ handles
         ↓
┌─────────────────┐
│ JSONRPCRequest  │
│  - ID           │──────┐
│  - Method       │      │ generates
│  - Params       │      ↓
└────────┬────────┘  ┌──────────┐
         │ triggers  │ LogEntry │
         ↓           │  - Raw   │
┌─────────────────┐  │  - Event │
│ ToolCallParams  │  └──────────┘
│  - Name: "echo" │      ↑
│  - Arguments    │      │ logs
└────────┬────────┘      │
         │ produces      │
         ↓               │
┌─────────────────┐      │
│ ToolCallResult  │      │
│  - Content      │──────┘
└────────┬────────┘
         │ wrapped in
         ↓
┌─────────────────┐
│ JSONRPCResponse │
│  - ID (matches) │
│  - Result       │
└─────────────────┘
```

---

## Validation Summary

| Entity | Key Validations |
|--------|-----------------|
| JSONRPCRequest | JSONRPC="2.0", Method non-empty, Valid JSON |
| JSONRPCResponse | ID matches request, Result XOR Error |
| InitializeParams | ProtocolVersion compatible, ClientInfo.Name present |
| ToolCallParams | Name="echo", Arguments contains "message" |
| ToolCallResult | Content non-empty, Text matches input exactly |
| LogEntry | Timestamp ISO8601, RequestID unique, RawRequest complete |
| TransportSession | ID unique, Type valid, State transitions correct |

---

## Performance Considerations

### Memory Management
- All message structures use `json.RawMessage` for deferred parsing (reduces allocations)
- No caching or persistent state (stateless server)
- Bounded message size to prevent OOM (10MB max per spec edge cases)

### Concurrency Safety
- All structures are immutable after creation
- Logging uses thread-safe zerolog library
- No shared mutable state between requests

### Serialization
- Standard library `encoding/json` for MCP protocol
- Zero-copy where possible (RawMessage references original buffer)
- Pre-allocated buffers for common message sizes

---

## References

- MCP Protocol Specification: https://spec.modelcontextprotocol.io/
- JSON-RPC 2.0 Specification: https://www.jsonrpc.org/specification
- JSON Schema for tool definitions
