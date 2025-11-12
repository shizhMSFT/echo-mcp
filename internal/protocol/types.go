package protocol

import "encoding/json"

// JSONRPCRequest represents an incoming MCP request following JSON-RPC 2.0 specification
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"` // Must be "2.0"
	ID      interface{}     `json:"id"`      // String, number, or null
	Method  string          `json:"method"`  // e.g., "initialize", "tools/call"
	Params  json.RawMessage `json:"params"`  // Method-specific parameters
}

// JSONRPCResponse represents an outgoing MCP response following JSON-RPC 2.0 specification
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"` // Must be "2.0"
	ID      interface{} `json:"id"`      // Matches request ID
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error with standard error codes
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface
func (e *RPCError) Error() string {
	return e.Message
}

// Standard JSON-RPC 2.0 error codes
const (
	ParseError     = -32700 // Invalid JSON
	InvalidRequest = -32600 // Invalid JSON-RPC structure
	MethodNotFound = -32601 // Unknown method
	InvalidParams  = -32602 // Invalid method parameters
	InternalError  = -32603 // Internal server error
)

// InitializeParams contains client capabilities sent during initialization
type InitializeParams struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

// ClientCapabilities describes what the client supports
type ClientCapabilities struct {
	Roots    *RootsCapability    `json:"roots,omitempty"`
	Sampling *SamplingCapability `json:"sampling,omitempty"`
}

// RootsCapability describes client's root management capability
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability describes client's sampling capability
type SamplingCapability struct{}

// ClientInfo contains information about the client
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult contains server capabilities returned to client
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

// ServerCapabilities describes what the server supports
type ServerCapabilities struct {
	Tools   *ToolsCapability   `json:"tools,omitempty"`
	Logging *LoggingCapability `json:"logging,omitempty"`
}

// ToolsCapability describes server's tool capability
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// LoggingCapability describes server's logging capability
type LoggingCapability struct{}

// ServerInfo contains information about the server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Server constants
const (
	ServerName      = "echo-mcp"
	ServerVersion   = "0.1.0"
	ProtocolVersion = "2024-11-05"
)

// ToolCallParams contains parameters for a tool call
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ToolCallResult contains the result of a tool call
type ToolCallResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ContentItem represents a piece of content in the result
type ContentItem struct {
	Type string `json:"type"` // "text", "image", "resource"
	Text string `json:"text,omitempty"`
}

// ToolDefinition describes a tool's schema
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema defines the JSON schema for tool input
type InputSchema struct {
	Type       string                 `json:"type"` // "object"
	Properties map[string]PropertyDef `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

// PropertyDef defines a property in the input schema
type PropertyDef struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// ToolsListResult contains the list of available tools
type ToolsListResult struct {
	Tools []ToolDefinition `json:"tools"`
}
