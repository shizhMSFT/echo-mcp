package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shizhMSFT/echo-mcp/internal/server"
	"github.com/shizhMSFT/echo-mcp/internal/transport"
)

// T065: Unit test for HTTP handler (POST /mcp endpoint)
func TestHTTPHandler(t *testing.T) {
	// Create server
	config := &server.Config{
		Mode:      "remote",
		Port:      8080,
		LogFormat: "json",
		LogLevel:  "info",
	}
	srv := server.NewServer(config)

	// Create HTTP transport
	httpTransport := transport.NewHTTPTransport(srv, ":8080")

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		validateResp   func(*testing.T, *http.Response)
	}{
		{
			name:           "POST /mcp with initialize request",
			method:         "POST",
			path:           "/mcp",
			expectedStatus: http.StatusOK,
			body: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "initialize",
				"params": map[string]interface{}{
					"protocolVersion": "2024-11-05",
					"clientInfo": map[string]string{
						"name":    "test",
						"version": "1.0",
					},
					"capabilities": map[string]interface{}{},
				},
			},
			validateResp: func(t *testing.T, resp *http.Response) {
				// Verify content type
				if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", ct)
				}

				// Parse response
				var jsonResp map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				// Verify response structure
				if jsonResp["jsonrpc"] != "2.0" {
					t.Errorf("Expected jsonrpc=2.0, got %v", jsonResp["jsonrpc"])
				}
				if jsonResp["id"] != float64(1) {
					t.Errorf("Expected id=1, got %v", jsonResp["id"])
				}
				if jsonResp["result"] == nil {
					t.Errorf("Expected result, got nil")
				}
			},
		},
		{
			name:           "POST /mcp with tools/call request",
			method:         "POST",
			path:           "/mcp",
			expectedStatus: http.StatusOK,
			body: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      2,
				"method":  "tools/call",
				"params": map[string]interface{}{
					"name": "echo",
					"arguments": map[string]interface{}{
						"message": "test message",
					},
				},
			},
			validateResp: func(t *testing.T, resp *http.Response) {
				var jsonResp map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				result, ok := jsonResp["result"].(map[string]interface{})
				if !ok {
					t.Fatalf("Expected result object, got %v", jsonResp["result"])
				}

				content, ok := result["content"].([]interface{})
				if !ok || len(content) == 0 {
					t.Fatalf("Expected content array, got %v", result["content"])
				}

				contentItem, ok := content[0].(map[string]interface{})
				if !ok {
					t.Fatalf("Expected content item object, got %v", content[0])
				}

				if contentItem["text"] != "test message" {
					t.Errorf("Expected message 'test message', got %v", contentItem["text"])
				}
			},
		},
		{
			name:           "POST /mcp with malformed JSON",
			method:         "POST",
			path:           "/mcp",
			expectedStatus: http.StatusOK,
			body:           `{"invalid json`,
			validateResp: func(t *testing.T, resp *http.Response) {
				var jsonResp map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				// Should return JSON-RPC error
				if jsonResp["error"] == nil {
					t.Error("Expected error in response")
				}
			},
		},
		{
			name:           "GET /health",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
			body:           nil,
			validateResp: func(t *testing.T, resp *http.Response) {
				// Health check should return 200 OK
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
			},
		},
		{
			name:           "POST /mcp with tools/list",
			method:         "POST",
			path:           "/mcp",
			expectedStatus: http.StatusOK,
			body: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      3,
				"method":  "tools/list",
				"params":  map[string]interface{}{},
			},
			validateResp: func(t *testing.T, resp *http.Response) {
				var jsonResp map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				result, ok := jsonResp["result"].(map[string]interface{})
				if !ok {
					t.Fatalf("Expected result object, got %v", jsonResp["result"])
				}

				tools, ok := result["tools"].([]interface{})
				if !ok {
					t.Fatalf("Expected tools array, got %v", result["tools"])
				}

				if len(tools) == 0 {
					t.Error("Expected at least one tool")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			var reqBody []byte
			if tt.body != nil {
				switch v := tt.body.(type) {
				case string:
					reqBody = []byte(v)
				default:
					reqBody, _ = json.Marshal(v)
				}
			}

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(reqBody))
			if tt.method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Handle request using the HTTP transport's handler
			ctx := context.Background()
			handler := httpTransport.Handler(ctx)
			handler.ServeHTTP(w, req)

			// Verify status code
			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Run custom validation if provided
			if tt.validateResp != nil {
				tt.validateResp(t, resp)
			}
		})
	}
}

// Test CORS headers are set correctly
func TestHTTPCORSHeaders(t *testing.T) {
	config := &server.Config{
		Mode:      "remote",
		Port:      8080,
		LogFormat: "json",
		LogLevel:  "info",
	}
	srv := server.NewServer(config)

	httpTransport := transport.NewHTTPTransport(srv, ":8080")

	reqBody, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	})

	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx := context.Background()
	handler := httpTransport.Handler(ctx)
	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Verify CORS header
	if cors := resp.Header.Get("Access-Control-Allow-Origin"); cors != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin: *, got %s", cors)
	}
}

// Test graceful shutdown
func TestHTTPGracefulShutdown(t *testing.T) {
	config := &server.Config{
		Mode:      "remote",
		Port:      8080,
		LogFormat: "json",
		LogLevel:  "info",
	}
	srv := server.NewServer(config)

	httpTransport := transport.NewHTTPTransport(srv, ":18090")

	// Start server in background
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		httpTransport.Run(ctx)
	}()

	// Give server time to start
	// Note: We can't easily test actual shutdown in unit test
	// This would require integration test

	// Cancel context to trigger shutdown
	cancel()

	// Note: In a real implementation, we'd verify:
	// 1. Server stops accepting new connections
	// 2. Existing connections are allowed to finish
	// 3. Server returns from Run() without error
}
