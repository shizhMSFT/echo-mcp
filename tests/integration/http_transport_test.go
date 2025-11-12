package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sync"
	"testing"
	"time"
)

// T061: Integration test for HTTP initialize handshake
func TestHTTPInitializeHandshake(t *testing.T) {
	// Start server in remote mode
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	port := "18080"
	cmd := exec.CommandContext(ctx, "go", "run", "../../cmd/echo-mcp/main.go", "--mode=remote", "--port="+port)

	// Capture stderr for logs
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to get stderr pipe: %v", err)
	}

	logBuffer := &bytes.Buffer{}
	go func() {
		io.Copy(logBuffer, stderr)
	}()

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give server time to start
	time.Sleep(500 * time.Millisecond)

	// Send initialize request
	initRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]string{
				"name":    "test-client",
				"version": "1.0",
			},
			"capabilities": map[string]interface{}{},
		},
	}

	reqBody, _ := json.Marshal(initRequest)
	resp, err := http.Post(fmt.Sprintf("http://localhost:%s/mcp", port), "application/json", bytes.NewReader(reqBody))
	if err != nil {
		cmd.Process.Kill()
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Verify status code
	if resp.StatusCode != http.StatusOK {
		cmd.Process.Kill()
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify content type
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}

	// Parse response
	var jsonResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
		cmd.Process.Kill()
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if jsonResponse["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc=2.0, got %v", jsonResponse["jsonrpc"])
	}
	if jsonResponse["id"] != float64(1) {
		t.Errorf("Expected id=1, got %v", jsonResponse["id"])
	}

	// Verify result contains server info
	result, ok := jsonResponse["result"].(map[string]interface{})
	if !ok {
		cmd.Process.Kill()
		t.Fatalf("Expected result object, got %v", jsonResponse["result"])
	}

	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	if !ok {
		cmd.Process.Kill()
		t.Fatalf("Expected serverInfo object, got %v", result["serverInfo"])
	}

	if serverInfo["name"] != "echo-mcp" {
		t.Errorf("Expected server name echo-mcp, got %v", serverInfo["name"])
	}

	// Cleanup
	cmd.Process.Kill()
	cmd.Wait()
}

// T062: Integration test for HTTP echo request/response
func TestHTTPEchoRequestResponse(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	port := "18081"
	cmd := exec.CommandContext(ctx, "go", "run", "../../cmd/echo-mcp/main.go", "--mode=remote", "--port="+port)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Send echo request
	echoRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "echo",
			"arguments": map[string]interface{}{
				"message": "Hello from HTTP!",
			},
		},
	}

	reqBody, _ := json.Marshal(echoRequest)
	resp, err := http.Post(fmt.Sprintf("http://localhost:%s/mcp", port), "application/json", bytes.NewReader(reqBody))
	if err != nil {
		cmd.Process.Kill()
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		cmd.Process.Kill()
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var jsonResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
		cmd.Process.Kill()
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify echo response
	result, ok := jsonResponse["result"].(map[string]interface{})
	if !ok {
		cmd.Process.Kill()
		t.Fatalf("Expected result object, got %v", jsonResponse["result"])
	}

	content, ok := result["content"].([]interface{})
	if !ok || len(content) == 0 {
		cmd.Process.Kill()
		t.Fatalf("Expected content array, got %v", result["content"])
	}

	contentItem, ok := content[0].(map[string]interface{})
	if !ok {
		cmd.Process.Kill()
		t.Fatalf("Expected content item object, got %v", content[0])
	}

	if contentItem["text"] != "Hello from HTTP!" {
		t.Errorf("Expected message 'Hello from HTTP!', got %v", contentItem["text"])
	}

	cmd.Process.Kill()
	cmd.Wait()
}

// T063: Integration test for HTTP concurrent clients
func TestHTTPConcurrentClients(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	port := "18082"
	cmd := exec.CommandContext(ctx, "go", "run", "../../cmd/echo-mcp/main.go", "--mode=remote", "--port="+port)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Launch multiple concurrent clients
	numClients := 10
	var wg sync.WaitGroup
	errors := make(chan error, numClients)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			// Send echo request
			echoRequest := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      clientID,
				"method":  "tools/call",
				"params": map[string]interface{}{
					"name": "echo",
					"arguments": map[string]interface{}{
						"message": fmt.Sprintf("Client %d", clientID),
					},
				},
			}

			reqBody, _ := json.Marshal(echoRequest)
			resp, err := http.Post(fmt.Sprintf("http://localhost:%s/mcp", port), "application/json", bytes.NewReader(reqBody))
			if err != nil {
				errors <- fmt.Errorf("Client %d: %v", clientID, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("Client %d: expected status 200, got %d", clientID, resp.StatusCode)
				return
			}

			var jsonResponse map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
				errors <- fmt.Errorf("Client %d: %v", clientID, err)
				return
			}

			// Verify response ID matches request ID
			if jsonResponse["id"] != float64(clientID) {
				errors <- fmt.Errorf("Client %d: expected id=%d, got %v", clientID, clientID, jsonResponse["id"])
				return
			}

			// Verify echo content
			result, ok := jsonResponse["result"].(map[string]interface{})
			if !ok {
				errors <- fmt.Errorf("Client %d: invalid result", clientID)
				return
			}

			content, ok := result["content"].([]interface{})
			if !ok || len(content) == 0 {
				errors <- fmt.Errorf("Client %d: invalid content", clientID)
				return
			}

			contentItem, ok := content[0].(map[string]interface{})
			if !ok {
				errors <- fmt.Errorf("Client %d: invalid content item", clientID)
				return
			}

			expectedMsg := fmt.Sprintf("Client %d", clientID)
			if contentItem["text"] != expectedMsg {
				errors <- fmt.Errorf("Client %d: expected '%s', got '%v'", clientID, expectedMsg, contentItem["text"])
				return
			}
		}(i)
	}

	// Wait for all clients
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}

	cmd.Process.Kill()
	cmd.Wait()
}

// T064: Integration test for HTTP client disconnect (verify resource cleanup)
func TestHTTPClientDisconnect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	port := "18083"
	cmd := exec.CommandContext(ctx, "go", "run", "../../cmd/echo-mcp/main.go", "--mode=remote", "--port="+port)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Send request and immediately close connection
	client := &http.Client{
		Timeout: 100 * time.Millisecond,
	}

	echoRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "echo",
			"arguments": map[string]interface{}{
				"message": "test",
			},
		},
	}

	reqBody, _ := json.Marshal(echoRequest)

	// Send request
	resp, err := client.Post(fmt.Sprintf("http://localhost:%s/mcp", port), "application/json", bytes.NewReader(reqBody))
	if err == nil {
		resp.Body.Close()
	}

	// Server should still be responsive after client disconnect
	time.Sleep(100 * time.Millisecond)

	// Send another request to verify server is still running
	resp2, err := client.Post(fmt.Sprintf("http://localhost:%s/mcp", port), "application/json", bytes.NewReader(reqBody))
	if err != nil {
		cmd.Process.Kill()
		t.Fatalf("Server not responsive after client disconnect: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		cmd.Process.Kill()
		t.Errorf("Expected status 200, got %d", resp2.StatusCode)
	}

	cmd.Process.Kill()
	cmd.Wait()
}

// Test health check endpoint
func TestHTTPHealthCheck(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	port := "18084"
	cmd := exec.CommandContext(ctx, "go", "run", "../../cmd/echo-mcp/main.go", "--mode=remote", "--port="+port)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Check health endpoint
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/health", port))
	if err != nil {
		cmd.Process.Kill()
		t.Fatalf("Failed to check health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		cmd.Process.Kill()
		t.Errorf("Expected health check status 200, got %d", resp.StatusCode)
	}

	cmd.Process.Kill()
	cmd.Wait()
}

// Test CORS headers
func TestHTTPCORSHeaders(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	port := "18085"
	cmd := exec.CommandContext(ctx, "go", "run", "../../cmd/echo-mcp/main.go", "--mode=remote", "--port="+port)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Send request and check CORS headers
	echoRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	}

	reqBody, _ := json.Marshal(echoRequest)
	resp, err := http.Post(fmt.Sprintf("http://localhost:%s/mcp", port), "application/json", bytes.NewReader(reqBody))
	if err != nil {
		cmd.Process.Kill()
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Verify CORS header
	if cors := resp.Header.Get("Access-Control-Allow-Origin"); cors != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin: *, got %s", cors)
	}

	cmd.Process.Kill()
	cmd.Wait()
}
