package protocol

import (
	"encoding/json"
	"fmt"
)

// ParseRequest parses a JSON-RPC request from raw bytes
func ParseRequest(data []byte) (*JSONRPCRequest, error) {
	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, &RPCError{
			Code:    ParseError,
			Message: "Parse error",
			Data:    err.Error(),
		}
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		return nil, &RPCError{
			Code:    InvalidRequest,
			Message: "Invalid Request: jsonrpc must be '2.0'",
		}
	}

	// Validate method is non-empty
	if req.Method == "" {
		return nil, &RPCError{
			Code:    InvalidRequest,
			Message: "Invalid Request: method is required",
		}
	}

	return &req, nil
}

// BuildResponse creates a successful JSON-RPC response
func BuildResponse(id interface{}, result interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// BuildErrorResponse creates an error JSON-RPC response
func BuildErrorResponse(id interface{}, code int, message string, data interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// HandleInitialize processes an initialize request
func HandleInitialize(params json.RawMessage) (*InitializeResult, error) {
	var initParams InitializeParams
	if err := json.Unmarshal(params, &initParams); err != nil {
		return nil, &RPCError{
			Code:    InvalidParams,
			Message: "Invalid params: failed to parse initialize parameters",
			Data:    err.Error(),
		}
	}

	// Validate protocol version
	if initParams.ProtocolVersion != ProtocolVersion {
		return nil, &RPCError{
			Code:    InvalidParams,
			Message: fmt.Sprintf("Invalid params: incompatible protocol version"),
			Data: map[string]string{
				"supported": ProtocolVersion,
				"requested": initParams.ProtocolVersion,
			},
		}
	}

	// Validate client info
	if initParams.ClientInfo.Name == "" {
		return nil, &RPCError{
			Code:    InvalidParams,
			Message: "Invalid params: clientInfo.name is required",
		}
	}

	// Build initialize result with server capabilities
	result := &InitializeResult{
		ProtocolVersion: ProtocolVersion,
		ServerInfo: ServerInfo{
			Name:    ServerName,
			Version: ServerVersion,
		},
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
			Logging: &LoggingCapability{},
		},
	}

	return result, nil
}

// HandleToolsList returns the list of available tools
func HandleToolsList() *ToolsListResult {
	return &ToolsListResult{
		Tools: []ToolDefinition{
			{
				Name:        "echo",
				Description: "Echoes back the provided message exactly as received. Useful for testing MCP client implementations.",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]PropertyDef{
						"message": {
							Type:        "string",
							Description: "The message to echo back",
						},
					},
					Required: []string{"message"},
				},
			},
		},
	}
}

// HandleToolsCall routes tool calls to the appropriate handler
func HandleToolsCall(params json.RawMessage, echoHandler func(message string) (*ToolCallResult, error)) (*ToolCallResult, error) {
	var toolParams ToolCallParams
	if err := json.Unmarshal(params, &toolParams); err != nil {
		return nil, &RPCError{
			Code:    InvalidParams,
			Message: "Invalid params: failed to parse tool call parameters",
			Data:    err.Error(),
		}
	}

	// Route to appropriate tool handler
	switch toolParams.Name {
	case "echo":
		// Extract message parameter
		message, ok := toolParams.Arguments["message"]
		if !ok {
			return nil, &RPCError{
				Code:    InvalidParams,
				Message: "Invalid params: 'message' parameter is required",
			}
		}

		messageStr, ok := message.(string)
		if !ok {
			return nil, &RPCError{
				Code:    InvalidParams,
				Message: "Invalid params: 'message' must be a string",
			}
		}

		return echoHandler(messageStr)

	default:
		return nil, &RPCError{
			Code:    MethodNotFound,
			Message: fmt.Sprintf("Method not found: unknown tool '%s'", toolParams.Name),
		}
	}
}
