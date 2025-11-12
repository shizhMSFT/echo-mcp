package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shizhMSFT/echo-mcp/internal/protocol"
	"github.com/shizhMSFT/echo-mcp/internal/server"
)

// HTTPTransport implements the HTTP transport for MCP
type HTTPTransport struct {
	server *server.Server
	addr   string
	srv    *http.Server
}

// NewHTTPTransport creates a new HTTP transport instance
func NewHTTPTransport(srv *server.Server, addr string) *HTTPTransport {
	return &HTTPTransport{
		server: srv,
		addr:   addr,
	}
}

// Run starts the HTTP server
func (t *HTTPTransport) Run(ctx context.Context) error {
	mux := http.NewServeMux()

	// MCP endpoint
	mux.HandleFunc("/mcp", t.handleMCP)

	// Health check endpoint
	mux.HandleFunc("/health", t.handleHealth)

	t.srv = &http.Server{
		Addr:    t.addr,
		Handler: mux,
	}

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		t.server.Logger.Info("Starting HTTP server", map[string]interface{}{
			"addr": t.addr,
		})
		if err := t.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or context cancellation
	select {
	case <-ctx.Done():
		return t.shutdown()
	case <-stop:
		return t.shutdown()
	case err := <-errChan:
		return err
	}
}

// shutdown gracefully shuts down the HTTP server
func (t *HTTPTransport) shutdown() error {
	t.server.Logger.Info("Shutting down HTTP server", nil)

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return t.srv.Shutdown(shutdownCtx)
}

// Handler returns the HTTP handler (for testing)
func (t *HTTPTransport) Handler(ctx context.Context) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", t.handleMCP)
	mux.HandleFunc("/health", t.handleHealth)
	return mux
}

// handleMCP handles MCP protocol requests
func (t *HTTPTransport) handleMCP(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only accept POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.sendError(w, nil, protocol.InternalError, "Failed to read request body", err.Error())
		return
	}
	defer r.Body.Close()

	// Process request
	response := t.handleRequest(body)

	// Send response
	w.Header().Set("Content-Type", "application/json")

	if response == nil {
		// No response needed (e.g., notification)
		w.WriteHeader(http.StatusOK)
		return
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		t.server.Logger.Error("Failed to marshal response", err, nil)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)
}

// handleHealth handles health check requests
func (t *HTTPTransport) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleRequest processes a single MCP request
func (t *HTTPTransport) handleRequest(rawRequest []byte) *protocol.JSONRPCResponse {
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

	case "ping":
		// Keep-alive ping (optional)
		result = map[string]interface{}{}

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

// sendError sends a JSON-RPC error response
func (t *HTTPTransport) sendError(w http.ResponseWriter, id interface{}, code int, message, data string) {
	response := protocol.BuildErrorResponse(id, code, message, data)
	responseBytes, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Still return 200 for JSON-RPC errors
	w.Write(responseBytes)
}
