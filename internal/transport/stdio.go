package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/shizhMSFT/echo-mcp/internal/protocol"
	"github.com/shizhMSFT/echo-mcp/internal/server"
)

// StdioTransport implements the stdio (standard input/output) transport for MCP
type StdioTransport struct {
	server *server.Server
	reader io.Reader
	writer io.Writer
}

// NewStdioTransport creates a new stdio transport instance
func NewStdioTransport(srv *server.Server, reader io.Reader, writer io.Writer) *StdioTransport {
	return &StdioTransport{
		server: srv,
		reader: reader,
		writer: writer,
	}
}

// Run starts the stdio transport loop
func (t *StdioTransport) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(t.reader)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Read next line
		if !scanner.Scan() {
			// EOF or error
			if err := scanner.Err(); err != nil {
				return err
			}
			return io.EOF
		}

		rawRequest := scanner.Bytes()

		// Process request
		response := t.handleRequest(rawRequest)

		// Write response
		responseBytes, err := json.Marshal(response)
		if err != nil {
			t.server.Logger.Error("Failed to marshal response", err, nil)
			continue
		}

		// Write response followed by newline
		if _, err := fmt.Fprintf(t.writer, "%s\n", responseBytes); err != nil {
			return err
		}
	}
}

// handleRequest processes a single MCP request
func (t *StdioTransport) handleRequest(rawRequest []byte) *protocol.JSONRPCResponse {
	// Parse request
	req, err := protocol.ParseRequest(rawRequest)
	if err != nil {
		rpcErr, ok := err.(*protocol.RPCError)
		if ok {
			return protocol.BuildErrorResponse(nil, rpcErr.Code, rpcErr.Message, rpcErr.Data)
		}
		return protocol.BuildErrorResponse(nil, protocol.InternalError, "Internal error", err.Error())
	}

	// Log request
	requestID := t.server.Logger.LogRequest(req.Method, rawRequest)

	// Route to appropriate handler
	var result interface{}
	var handleErr error

	switch req.Method {
	case "initialize":
		result, handleErr = protocol.HandleInitialize(req.Params)

	case "tools/list":
		result = protocol.HandleToolsList()

	case "tools/call":
		result, handleErr = protocol.HandleToolsCall(req.Params, server.HandleEcho)
		if handleErr == nil {
			// Log tool invocation
			var toolParams protocol.ToolCallParams
			if err := json.Unmarshal(req.Params, &toolParams); err == nil {
				t.server.Logger.LogToolInvocation(requestID, toolParams.Name, rawRequest)
			}
		}

	case "initialized":
		// Notification - no response needed
		return nil

	default:
		handleErr = &protocol.RPCError{
			Code:    protocol.MethodNotFound,
			Message: fmt.Sprintf("Method not found: %s", req.Method),
		}
	}

	// Handle error
	if handleErr != nil {
		rpcErr, ok := handleErr.(*protocol.RPCError)
		if ok {
			t.server.Logger.LogError(requestID, rpcErr.Message, handleErr, rawRequest)
			return protocol.BuildErrorResponse(req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data)
		}
		t.server.Logger.LogError(requestID, "Internal error", handleErr, rawRequest)
		return protocol.BuildErrorResponse(req.ID, protocol.InternalError, "Internal error", handleErr.Error())
	}

	// Log successful response
	t.server.Logger.LogResponse(requestID, req.Method)

	return protocol.BuildResponse(req.ID, result)
}
